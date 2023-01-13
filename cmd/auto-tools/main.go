/*
 * auto - hypervisor agent for eve
 * Copyright (C) 2022-2023  BNS Services LLC
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"os"

	"github.com/BasedDevelopment/auto/internal/config"
	"github.com/BasedDevelopment/eve/pkg/pki"
	"github.com/BasedDevelopment/eve/pkg/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	version = "0.0.1"
)

var (
	configPath = flag.String("config-path", "/etc/auto/config.toml", "Path to TLS key and certificate")
	makeKey    = flag.Bool("make-key", false, "Make private key")
	makeCSR    = flag.Bool("make-csr", false, "Make CSR")
	checkSum   = flag.String("checksum", "", "Check the checksum of a pem encoded file")

	keyPath string
	csrPath string
	crtPath string
	caPath  string
)

func init() {
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("auto-tools, tools to manage certificate on an auto instance " + version)

	if err := config.Load(*configPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	tlsPath := config.Config.TLSPath

	// Ensure TLS path has a slash at the end
	if tlsPath[len(tlsPath)-1:] != "/" {
		tlsPath += "/"
	}

	keyPath = tlsPath + config.Config.Hostname + ".key"
	csrPath = tlsPath + config.Config.Hostname + ".csr"
	crtPath = tlsPath + config.Config.Hostname + ".crt"
	caPath = tlsPath + "ca.crt"

	// Ensure TLS path exists
	if _, err := os.Stat(tlsPath); os.IsNotExist(err) {
		log.Info().
			Str("path", tlsPath).
			Msg("TLS path does not exist, creating")
		if err := os.MkdirAll(tlsPath, 0700); err != nil {
			log.Fatal().
				Err(err).
				Str("path", tlsPath).
				Msg("Failed to create TLS path")
		}
	}
}

func main() {
	if *makeKey {
		log.Info().Msg("Creating key")
		b := pki.GenKey()
		util.WriteFile(keyPath, b)
		log.Info().
			Str("path", keyPath).
			Msg("Key written")
		return
	}

	if *makeCSR {
		log.Info().Msg("Creating CSR")
		privBytes := util.ReadFile(keyPath)
		priv := pki.ReadKey(privBytes)
		csrBytes := pki.GenCSR(priv, config.Config.Hostname)
		util.WriteFile(csrPath, csrBytes)
		sum := pki.PemSum(csrBytes)
		log.Info().
			Str("path", csrPath).
			Str("SHA1", sum).
			Msg("CSR written")
		return
	}

	if *checkSum != "" {
		b := util.ReadFile(*checkSum)
		result := pki.PemSum(b)
		log.Info().
			Str("path", *checkSum).
			Str("SHA1", result).
			Msg("Checksum")
		return
	}

	log.Info().Msg("No action specified, verifying PKI")
	checkPKI()
}
