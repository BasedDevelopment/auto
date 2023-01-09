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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	version = "0.0.1"
)

var (
	configPath = flag.String("config-path", "/etc/auto/config.toml", "Path to TLS key and certificate")
	genKey     = flag.Bool("gen-key", false, "Generate TLS key")
	makeCSR    = flag.Bool("make-csr", false, "Make CSR")

	keyFile string
	csrFile string
	crtFile string
	caFile  string
)

func init() {
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("Auto - hypervisor agent for eve v" + version)

	if err := config.Load(*configPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	keyFile = config.Config.TLSPath + "/" + config.Config.Hostname + ".key"
	csrFile = config.Config.TLSPath + "/" + config.Config.Hostname + ".csr"
	crtFile = config.Config.TLSPath + "/" + config.Config.Hostname + ".crt"
	caFile = config.Config.TLSPath + "/ca.crt"
}

func main() {
	if *genKey {
		createKey()
		return
	}

	if *makeCSR {
		createCSR()
		return
	}

	log.Info().Msg("No action specified, checking and verifying TLS")

	// Keyfile
	_, err := os.Stat(keyFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal().Err(err).Msg("Failed to fetch TLS key")
		} else {
			log.Info().Msg("TLS key not found, generating")
			createKey()
		}
	} else {
		log.Info().Msg("TLS key found")
	}

	// Certificate Signing Request
	_, err = os.Stat(csrFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal().Err(err).Msg("Failed to fetch TLS CSR")
		} else {
			log.Info().Msg("TLS CSR not found, generating")
			createCSR()
		}
	} else {
		log.Info().Msg("TLS CSR found")
	}
	csrSum()

	// Certificate
	_, err = os.Stat(crtFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal().Err(err).Msg("Failed to fetch TLS certificate")
		} else {
			log.Fatal().Msg("TLS certificate not found, please sign the CSR via eve-tools")
		}
	} else {
		log.Info().Msg("TLS certificate found")
	}
	certSum()

	// CA Certificate
	_, err = os.Stat(caFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal().Err(err).Msg("Failed to fetch CA certificate")
		} else {
			log.Fatal().Msg("CA certificate not found, please fetch the CA certificate from eve")
		}
	} else {
		log.Info().Msg("CA certificate found")
		validateCert()
	}
	caSum()

	log.Info().Msg("Nothing more to do, exiting")
}
