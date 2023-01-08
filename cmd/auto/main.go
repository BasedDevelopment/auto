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
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/BasedDevelopment/auto/internal/config"
	"github.com/BasedDevelopment/auto/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	shutdownTimeout = 5 * time.Second
	version         = "0.0.1"
)

var (
	configPath = flag.String("config", "/etc/auto/config.toml", "path to config file")
	logLevel   = flag.String("log-level", "debug", "Log level (trace, debug, info, warn, error, fatal, panic)")
	logFormat  = flag.String("log-format", "json", "Log format (json, pretty)")
	noSplash   = flag.Bool("nosplash", false, "Disable splash screen")
)

func init() {
	configureLogger()

	// Load configuration
	log.Info().Msg("Loading configuration")

	if err := config.Load(*configPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}
}

func main() {
	log.Info().Msg("Fetching TLS key")
	_, err := os.Stat(config.Config.TLSPath + "/auto.key")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal().Err(err).Msg("Failed to fetch TLS key")
		} else {
			log.Info().Msg("TLS key not found, generating")
			initCSR()
			log.Info().Msg("TLS key generated and sent")
			return
		}
	}
	log.Info().Msg("TLS key found, fetching TLS certificate")
	_, err = os.Stat(config.Config.TLSPath + "/auto.pem")
	if err != nil {
		if !os.IsNotExist(err) {
			log.Fatal().Err(err).Msg("Failed to fetch TLS certificate")
		} else {
			log.Info().Msg("TLS certificate not found, checking eve")
			//getCert()
			log.Info().Msg("TLS certificate fetched, starting")
		}
	}
	log.Info().Msg("TLS certificate found, starting")
	log.Info().
		Str("host", config.Config.API.Host).
		Int("port", config.Config.API.Port).
		Msg("HTTPS server listening")

	// Create HTTP server
	srv := &http.Server{
		Addr:    config.Config.API.Host + ":" + strconv.Itoa(config.Config.API.Port),
		Handler: server.Service(),
	}

	srvCtx, srvStopCtx := context.WithCancel(context.Background())

	// Watch for OS signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(srvCtx, 30*time.Second)
		defer shutdownCtxCancel() // release srvCtx if we take too long to shut down

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Warn().Msg("Graceful shutdown timed out... forcing regular exit.")
			}
		}()

		// Gracefully shut down services
		log.Info().Msg("Gracefully shutting down services")
		// Webserver
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to shutdown HTTP listener")
		} else {
			log.Info().Msg("Webserver shutdown success")
		}

		// Libvirt connections
		log.Info().Msg("Libvirt connections shutdown success")

		srvStopCtx()
	}()

	// Start the server
	err = srv.ListenAndServeTLS(config.Config.TLSPath+"/auto.pem", config.Config.TLSPath+"/auto.key")

	if err != nil && err != http.ErrServerClosed {
		log.Fatal().
			Err(err).
			Msg("Failed to start HTTP listener")
	}

	// Wait for server context to be stopped
	<-srvCtx.Done()
	log.Info().Msg("Graceful shutdown complete. Thank you for using auto!")
}

func initCSR() {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate ed25519 key pair")
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	})
	// Save key
	if err := os.WriteFile(config.Config.TLSPath+"/auto.key", keyPem, 0600); err != nil {
		log.Fatal().Err(err).Msg("Failed to save TLS key")
	}
	// generate CSR
	csrTemplate := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "auto",
		},
		DNSNames: []string{config.Config.Name},
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, priv)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate CSR")
	}
	csrPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csr,
	})
	// Save CSR
	if err := os.WriteFile(config.Config.TLSPath+"/auto.csr", csrPem, 0600); err != nil {
		log.Fatal().Err(err).Msg("Failed to save TLS CSR")
	}
	// Marshal CSR request
	reqJson, err := json.Marshal(map[string]string{
		"csr": string(csrPem),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to marshal CSR signing request")
	}
	// Send CSR request
	req, err := http.NewRequest("POST", config.Config.Eve.URL+"/auto", bytes.NewBuffer(reqJson))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create CSR signing request")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to send CSR signing request")
	}
	defer resp.Body.Close()
	// Check if response is OK
	if resp.StatusCode == http.StatusOK {
		return
	} else {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read CSR signing response")
		}
		body := string(bodyBytes)
		log.Fatal().
			Int("status", resp.StatusCode).
			Str("body", body).
			Msg("Eve returned an error")
	}
}

func configureLogger() {
	flag.Parse()

	if *logFormat == "pretty" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	switch *logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Init logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if !*noSplash {
		log.Info().Msg("+-----------------------------------+")
		log.Info().Msg("|  Auto - Hypervisor agent for eve  |")
		log.Info().Msg("|               v" + version + "              |")
		log.Info().Msg("+-----------------------------------+")
	} else {
		log.Info().Msg("Auto - hypervisor agent for eve v" + version)
	}
}
