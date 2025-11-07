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

package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agents"
	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/server/api"
	trust_v1 "go.probo.inc/probo/pkg/server/api/trust/v1"
	auth_server "go.probo.inc/probo/pkg/server/auth"
	authz_server "go.probo.inc/probo/pkg/server/authz"
	"go.probo.inc/probo/pkg/server/trust"
	"go.probo.inc/probo/pkg/server/web"
	trust_pkg "go.probo.inc/probo/pkg/trust"
)

type Config struct {
	AllowedOrigins    []string
	ExtraHeaderFields map[string]string
	Probo             *probo.Service
	Auth              *auth.Service
	Authz             *authz.Service
	Trust             *trust_pkg.Service
	SAML              *auth.SAMLService
	ConsoleAuth       api.ConsoleAuthConfig
	TrustAuth         api.TrustAuthConfig
	ConnectorRegistry *connector.ConnectorRegistry
	Agent             *agents.Agent
	SafeRedirect      *saferedirect.SafeRedirect
	CustomDomainCname string
	FileManager       *filemanager.Service
	PGClient          *pg.Client
	Logger            *log.Logger
}

type Server struct {
	apiServer         *api.Server
	webServer         *web.Server
	trustServer       *trust.Server
	authServer        *auth_server.Server
	authzServer       *authz_server.Server
	router            *chi.Mux
	extraHeaderFields map[string]string
	proboService      *probo.Service
	logger            *log.Logger
}

func NewServer(cfg Config) (*Server, error) {
	apiCfg := api.Config{
		AllowedOrigins:    cfg.AllowedOrigins,
		Probo:             cfg.Probo,
		Auth:              cfg.Auth,
		Authz:             cfg.Authz,
		Trust:             cfg.Trust,
		SAML:              cfg.SAML,
		ConsoleAuth:       cfg.ConsoleAuth,
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

	webServer, err := web.NewServer()
	if err != nil {
		return nil, err
	}

	trustServer, err := trust.NewServer()
	if err != nil {
		return nil, err
	}

	authServer, err := auth_server.NewServer(auth_server.Config{
		Auth:            cfg.Auth,
		Authz:           cfg.Authz,
		SAML:            cfg.SAML,
		CookieName:      cfg.ConsoleAuth.CookieName,
		CookieDomain:    cfg.ConsoleAuth.CookieDomain,
		SessionDuration: cfg.ConsoleAuth.SessionDuration,
		CookieSecret:    cfg.ConsoleAuth.CookieSecret,
		CookieSecure:    cfg.ConsoleAuth.CookieSecure,
		FileManager:     cfg.FileManager,
		Logger:          cfg.Logger.Named("auth"),
	})
	if err != nil {
		return nil, err
	}

	authzServer, err := authz_server.NewServer(authz_server.Config{
		Auth:         cfg.Auth,
		Authz:        cfg.Authz,
		Logger:       cfg.Logger.Named("authz"),
		CookieName:   cfg.ConsoleAuth.CookieName,
		CookieSecret: cfg.ConsoleAuth.CookieSecret,
		CookieSecure: cfg.ConsoleAuth.CookieSecure,
	})
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()

	server := &Server{
		apiServer:         apiServer,
		webServer:         webServer,
		trustServer:       trustServer,
		authServer:        authServer,
		authzServer:       authzServer,
		router:            router,
		extraHeaderFields: cfg.ExtraHeaderFields,
		proboService:      cfg.Probo,
		logger:            cfg.Logger,
	}

	server.setupRoutes()

	return server, nil
}

func (s *Server) setupRoutes() {
	s.router.Mount("/api", s.apiServer)
	s.router.Mount("/connect", s.authServer)
	s.router.Mount("/authz", s.authzServer)

	s.router.Route("/trust/{slugOrId}", func(r chi.Router) {
		r.Use(s.loadTrustCenterBySlugOrID)
		r.Use(s.stripTrustPrefix)
		r.Mount("/", s.trustCenterRouter())
	})

	s.router.Mount("/", s.webServer)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.setExtraHeaders(w)
	s.router.ServeHTTP(w, r)
}

func (s *Server) setExtraHeaders(w http.ResponseWriter) {
	for key, value := range s.extraHeaderFields {
		w.Header().Set(key, value)
	}
}

func (s *Server) handleCustomDomain404(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.TLS == nil {
		domain := r.Host

		_, err := s.proboService.LoadOrganizationByDomain(ctx, domain)
		if err == nil {
			httpsURL := "https://" + r.Host + r.URL.RequestURI()
			s.logger.InfoCtx(ctx, "404 on HTTP custom domain, redirecting to HTTPS",
				log.String("domain", domain),
				log.String("from", r.URL.RequestURI()),
				log.String("to", httpsURL),
			)
			http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
			return
		}
	}

	httpserver.RenderError(w, http.StatusNotFound, errors.New("not found"))
}

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

func (s *Server) addTrustCenterToContext(ctx context.Context, tenantID, organizationID interface{}) context.Context {
	ctx = context.WithValue(ctx, trust_v1.CustomDomainTenantIDKey, tenantID)
	ctx = context.WithValue(ctx, trust_v1.CustomDomainOrganizationIDKey, organizationID)
	return ctx
}

func (s *Server) stripTrustPrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slugOrId := chi.URLParam(r, "slugOrId")
		prefix := "/trust/" + slugOrId

		if r.URL.Path == prefix {
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

func (s *Server) trustCenterRouter() chi.Router {
	r := chi.NewRouter()

	r.Mount("/api/trust/v1", s.apiServer.TrustAPIHandler())
	r.Handle("/*", s.trustServer)

	return r
}

func (s *Server) TrustCenterHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; preload")
			s.setExtraHeaders(w)
			next.ServeHTTP(w, r)
		})
	})

	r.Use(s.loadTrustCenterByDomain)
	r.NotFound(s.handleCustomDomain404)

	r.Mount("/", s.trustCenterRouter())

	return r
}
