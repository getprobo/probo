// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// Package server provides functionality for serving the SPA frontend.
package server

import (
	"net/http"
	"strings"

	"github.com/getprobo/probo/pkg/agents"
	"github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/saferedirect"
	"github.com/getprobo/probo/pkg/server/api"
	"github.com/getprobo/probo/pkg/server/trust"
	"github.com/getprobo/probo/pkg/server/web"
	trust_pkg "github.com/getprobo/probo/pkg/trust"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
)

// Config holds the configuration for the server
type Config struct {
	AllowedOrigins    []string
	ExtraHeaderFields map[string]string
	Probo             *probo.Service
	Auth              *auth.Service
	Authz             *authz.Service
	Trust             *trust_pkg.Service
	ConsoleAuth       api.ConsoleAuthConfig
	TrustAuth         api.TrustAuthConfig
	ConnectorRegistry *connector.ConnectorRegistry
	Agent             *agents.Agent
	SafeRedirect      *saferedirect.SafeRedirect
	Logger            *log.Logger
}

// Server represents the main server that handles both API and frontend requests
type Server struct {
	apiServer         *api.Server
	webServer         *web.Server
	trustServer       *trust.Server
	router            *chi.Mux
	extraHeaderFields map[string]string
}

// NewServer creates a new server instance
func NewServer(cfg Config) (*Server, error) {
	// Create API server
	apiCfg := api.Config{
		AllowedOrigins:    cfg.AllowedOrigins,
		Probo:             cfg.Probo,
		Auth:              cfg.Auth,
		Authz:             cfg.Authz,
		Trust:             cfg.Trust,
		ConsoleAuth:       cfg.ConsoleAuth,
		TrustAuth:         cfg.TrustAuth,
		ConnectorRegistry: cfg.ConnectorRegistry,
		SafeRedirect:      cfg.SafeRedirect,
		Logger:            cfg.Logger.Named("api"),
	}
	apiServer, err := api.NewServer(apiCfg)
	if err != nil {
		return nil, err
	}

	// Create web server for console SPA
	webServer, err := web.NewServer()
	if err != nil {
		return nil, err
	}

	// Create trust server for trust SPA
	trustServer, err := trust.NewServer()
	if err != nil {
		return nil, err
	}

	// Create main router
	router := chi.NewRouter()

	server := &Server{
		apiServer:         apiServer,
		webServer:         webServer,
		trustServer:       trustServer,
		router:            router,
		extraHeaderFields: cfg.ExtraHeaderFields,
	}

	// Set up routes
	server.setupRoutes()

	return server, nil
}

// setupRoutes configures the routing for the server
func (s *Server) setupRoutes() {
	// API routes under /api
	s.router.Mount("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip the /api prefix from the path
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		s.apiServer.ServeHTTP(w, r)
	}))

	// Trust routes go to the trust SPA
	s.router.Route("/trust", func(r chi.Router) {
		r.Mount("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/trust")
			if req.URL.Path == "" {
				req.URL.Path = "/"
			}
			s.trustServer.ServeHTTP(w, req)
		}))
	})

	// All other routes go to the console SPA frontend
	s.router.Mount("/", s.webServer)
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for key, value := range s.extraHeaderFields {
		w.Header().Set(key, value)
	}

	s.router.ServeHTTP(w, r)
}
