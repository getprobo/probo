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
	"context"
	"net/http"
	"strings"

	"github.com/getprobo/probo/pkg/agents"
	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/saferedirect"
	"github.com/getprobo/probo/pkg/server/api"
	trust_v1 "github.com/getprobo/probo/pkg/server/api/trust/v1"
	"github.com/getprobo/probo/pkg/server/trust"
	"github.com/getprobo/probo/pkg/server/web"
	trust_pkg "github.com/getprobo/probo/pkg/trust"
	"github.com/getprobo/probo/pkg/usrmgr"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
)

// Config holds the configuration for the server
type Config struct {
	AllowedOrigins    []string
	ExtraHeaderFields map[string]string
	Probo             *probo.Service
	Usrmgr            *usrmgr.Service
	Trust             *trust_pkg.Service
	Auth              api.ConsoleAuthConfig
	TrustAuth         api.TrustAuthConfig
	ConnectorRegistry *connector.ConnectorRegistry
	Agent             *agents.Agent
	SafeRedirect      *saferedirect.SafeRedirect
	CustomDomainCname string
	Logger            *log.Logger
}

// Server represents the main server that handles both API and frontend requests
type Server struct {
	apiServer         *api.Server
	webServer         *web.Server
	trustServer       *trust.Server
	router            *chi.Mux
	extraHeaderFields map[string]string
	proboService      *probo.Service
	logger            *log.Logger
}

// NewServer creates a new server instance
func NewServer(cfg Config) (*Server, error) {
	// Create API server
	apiCfg := api.Config{
		AllowedOrigins:    cfg.AllowedOrigins,
		Probo:             cfg.Probo,
		Usrmgr:            cfg.Usrmgr,
		Trust:             cfg.Trust,
		Auth:              cfg.Auth,
		TrustAuth:         cfg.TrustAuth,
		ConnectorRegistry: cfg.ConnectorRegistry,
		SafeRedirect:      cfg.SafeRedirect,
		CustomDomainCname: cfg.CustomDomainCname,
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
		proboService:      cfg.Probo,
		logger:            cfg.Logger,
	}

	// Set up routes
	server.setupRoutes()

	return server, nil
}

// setupRoutes configures the routing for the server
func (s *Server) setupRoutes() {
	// API routes
	s.router.Mount("/api", s.apiServer)

	// Trust center routes by slug or ID
	s.router.Route("/trust/{slugOrId}", func(r chi.Router) {
		r.Use(s.loadTrustCenterBySlugOrID)
		r.Use(s.stripTrustPrefix)
		r.Mount("/", s.trustCenterRouter())
	})

	// Console SPA (catch-all)
	s.router.Mount("/", s.webServer)
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.setExtraHeaders(w)
	s.router.ServeHTTP(w, r)
}

// setExtraHeaders adds configured extra headers to the response
func (s *Server) setExtraHeaders(w http.ResponseWriter) {
	for key, value := range s.extraHeaderFields {
		w.Header().Set(key, value)
	}
}

// loadTrustCenterBySlugOrID middleware loads trust center info from slug or ID and adds to context
func (s *Server) loadTrustCenterBySlugOrID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		slugOrId := chi.URLParam(r, "slugOrId")

		// Try to parse as GID first
		var trustCenter *probo.TrustCenterInfo
		var err error

		if id, parseErr := gid.ParseGID(slugOrId); parseErr == nil {
			// It's a valid ID, load by ID
			s.logger.InfoCtx(ctx, "loading trust center by ID",
				log.String("id", id.String()),
				log.String("path", r.URL.Path),
			)

			trustCenter, err = s.proboService.LoadTrustCenterByID(ctx, id)
			if err != nil {
				s.logger.WarnCtx(ctx, "trust center not found",
					log.String("id", id.String()),
					log.Error(err),
				)
				http.Error(w, "Trust center not found", http.StatusNotFound)
				return
			}

			s.logger.InfoCtx(ctx, "trust center loaded by ID",
				log.String("id", id.String()),
				log.String("trust_center_id", trustCenter.ID.String()),
				log.String("organization_id", trustCenter.OrganizationID.String()),
			)
		} else {
			// Not a valid ID, treat as slug
			s.logger.InfoCtx(ctx, "loading trust center by slug",
				log.String("slug", slugOrId),
				log.String("path", r.URL.Path),
			)

			trustCenter, err = s.proboService.LoadTrustCenterBySlug(ctx, slugOrId)
			if err != nil {
				s.logger.WarnCtx(ctx, "trust center not found",
					log.String("slug", slugOrId),
					log.Error(err),
				)
				http.Error(w, "Trust center not found", http.StatusNotFound)
				return
			}

			s.logger.InfoCtx(ctx, "trust center loaded by slug",
				log.String("slug", slugOrId),
				log.String("trust_center_id", trustCenter.ID.String()),
				log.String("organization_id", trustCenter.OrganizationID.String()),
			)
		}

		ctx = s.addTrustCenterToContext(ctx, trustCenter.ID.TenantID(), trustCenter.OrganizationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loadTrustCenterByDomain middleware loads trust center info from custom domain and adds to context
func (s *Server) loadTrustCenterByDomain(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.TLS == nil || r.TLS.ServerName == "" {
			next.ServeHTTP(w, r)
			return
		}

		domain := r.TLS.ServerName

		s.logger.InfoCtx(ctx, "loading organization by custom domain",
			log.String("domain", domain),
			log.String("path", r.URL.Path),
		)

		organizationID, err := s.proboService.LoadOrganizationByDomain(ctx, domain)
		if err != nil {
			s.logger.WarnCtx(ctx, "organization not found for domain",
				log.String("domain", domain),
				log.Error(err),
			)
			next.ServeHTTP(w, r)
			return
		}

		s.logger.InfoCtx(ctx, "organization loaded",
			log.String("domain", domain),
			log.String("organization_id", organizationID.String()),
		)

		ctx = s.addTrustCenterToContext(ctx, organizationID.TenantID(), organizationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// addTrustCenterToContext adds trust center identification to context
func (s *Server) addTrustCenterToContext(ctx context.Context, tenantID, organizationID interface{}) context.Context {
	ctx = context.WithValue(ctx, trust_v1.CustomDomainTenantIDKey, tenantID)
	ctx = context.WithValue(ctx, trust_v1.CustomDomainOrganizationIDKey, organizationID)
	return ctx
}

// stripTrustPrefix middleware strips /trust/{slugOrId} from the path
func (s *Server) stripTrustPrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slugOrId := chi.URLParam(r, "slugOrId")
		prefix := "/trust/" + slugOrId

		// Strip the prefix from the path
		if r.URL.Path == prefix {
			// Redirect to trailing slash for proper asset resolution
			http.Redirect(w, r, prefix+"/", http.StatusMovedPermanently)
			return
		}

		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}

		next.ServeHTTP(w, r)
	})
}

// trustCenterRouter returns a router for trust center content (API + frontend)
func (s *Server) trustCenterRouter() chi.Router {
	r := chi.NewRouter()

	// Trust API routes
	r.Mount("/api/trust/v1", s.apiServer.TrustAPIHandler())

	// Trust center frontend (catch-all)
	r.Handle("/*", s.trustServer)

	return r
}

// TrustCenterHandler returns an HTTP handler for serving trust centers on custom domains
func (s *Server) TrustCenterHandler() http.Handler {
	r := chi.NewRouter()

	// Set security headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; preload")
			s.setExtraHeaders(w)
			next.ServeHTTP(w, r)
		})
	})

	// Load organization by custom domain
	r.Use(s.loadTrustCenterByDomain)

	// Mount trust center content
	r.Mount("/", s.trustCenterRouter())

	return r
}
