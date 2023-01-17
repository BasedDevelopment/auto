/*
 * auto - hypervisor agent for eve
 * Copyright (C) 2022-2023  BNS Services LLC

 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.

 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.

 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package server

import (
	"net/http"
	"time"

	"github.com/BasedDevelopment/auto/internal/server/routes"
	"github.com/BasedDevelopment/eve/pkg/middleware"
	"github.com/go-chi/chi/v5"
	cm "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func Service() *chi.Mux {
	r := chi.NewMux()

	// Middlewares
	r.Use(cm.RequestID)
	r.Use(middleware.Logger)
	r.Use(httprate.LimitByIP(100, 1*time.Minute))
	r.Use(cm.AllowContentType("application/json"))
	r.Use(cm.CleanPath)
	r.Use(cm.NoCache)
	r.Use(cm.Heartbeat("/"))
	r.Use(middleware.Recoverer)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	r.Route("/libvirt", func(r chi.Router) {
		r.Get("/", routes.GetHV)
		r.Route("/domain", func(r chi.Router) {
			r.Get("/", routes.GetDomains)
			//r.Post("/", routes.CreateDomain)
			r.Route("/{domain}", func(r chi.Router) {
				r.Get("/", routes.GetDomain)
				//r.Put("/", routes.UpdateDomain)
				//r.Delete("/", routes.DeleteDomain)
			})
		})
	})

	return r
}
