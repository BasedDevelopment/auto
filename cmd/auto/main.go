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
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/BasedDevelopment/auto/internal/config"
	"github.com/BasedDevelopment/auto/internal/controllers"
	"github.com/BasedDevelopment/auto/internal/server"
	"github.com/BasedDevelopment/eve/pkg/fwdlog"
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

	tlsPath string
	crtPath string
	keyPath string
	caPath  string
)

func init() {
	configureLogger()
	checkPaths()

	// Load configuration
	log.Info().Msg("Loading configuration")

	if err := config.Load(*configPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	tlsPath := config.Config.TLSPath

	if tlsPath[len(tlsPath)-1:] != "/" {
		tlsPath += "/"
	}

	crtPath = tlsPath + config.Config.Hostname + ".crt"
	keyPath = tlsPath + config.Config.Hostname + ".key"
	caPath = tlsPath + "ca.crt"
}

func main() {
	log.Info().
		Str("host", config.Config.API.Host).
		Int("port", config.Config.API.Port).
		Msg("HTTPS server listening")

	// TLS config
	// Add the CA to the pool
	caPool := x509.NewCertPool()
	caCertBytes, err := ioutil.ReadFile(caPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load CA certificate")
	}
	caPool.AppendCertsFromPEM(caCertBytes)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		ClientCAs:  caPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			// Verify the serial number of the certificate
			if verifiedChains[0][0].SerialNumber.String() != config.Config.Eve.Serial {
				return fmt.Errorf("serial number mismatch")
			}
			return nil
		},
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:      config.Config.API.Host + ":" + strconv.Itoa(config.Config.API.Port),
		Handler:   server.Service(),
		TLSConfig: tlsConfig,
		ErrorLog:  fwdlog.Logger(),
	}

	srvCtx, srvStopCtx := context.WithCancel(context.Background())

	// Initialize the hypervisor
	hv := controllers.Hypervisor
	hv.IP = net.ParseIP(config.Config.Libvirt.Host)
	hv.Port = config.Config.Libvirt.Port
	if err := hv.Init(); err != nil {
		log.Error().Err(err).Msg("Failed to initialize hypervisor")
	}

	if err := controllers.CheckStorage(); err != nil {
		log.Error().Err(err).Msg("Failed to initialize storage")
	}

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
		controllers.Hypervisor.Libvirt.Close()

		srvStopCtx()
	}()

	// Start the server
	err = srv.ListenAndServeTLS(crtPath, keyPath)

	if err != nil && err != http.ErrServerClosed {
		log.Fatal().
			Err(err).
			Msg("Failed to start HTTP listener")
	}

	// Wait for server context to be stopped
	<-srvCtx.Done()
	log.Info().Msg("Graceful shutdown complete. Thank you for using auto!")
}
