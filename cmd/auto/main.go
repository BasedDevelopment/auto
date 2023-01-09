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
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/BasedDevelopment/auto/internal/config"
	"github.com/BasedDevelopment/auto/internal/server"
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
	if _, err := os.Stat(config.Config.TLSPath + "/auto.key"); err != nil {
		log.Fatal().Err(err).Msg("Failed to load TLS key")
	}
	if _, err := os.Stat(config.Config.TLSPath + "/auto.pem"); err != nil {
		log.Fatal().Err(err).Msg("Failed to load TLS certificate")
	}

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
	err := srv.ListenAndServeTLS(config.Config.TLSPath+"/auto.pem", config.Config.TLSPath+"/auto.key")

	if err != nil && err != http.ErrServerClosed {
		log.Fatal().
			Err(err).
			Msg("Failed to start HTTP listener")
	}

	// Wait for server context to be stopped
	<-srvCtx.Done()
	log.Info().Msg("Graceful shutdown complete. Thank you for using auto!")
}
