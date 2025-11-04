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

package api

import (
	"errors"
	"net/http"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/saferedirect"
	console_v1 "go.probo.inc/probo/pkg/server/api/console/v1"
	trust_v1 "go.probo.inc/probo/pkg/server/api/trust/v1"
	"go.probo.inc/probo/pkg/trust"
)

type (
	ConsoleAuthConfig struct {
		CookieName      string
		CookieDomain    string
		SessionDuration time.Duration
		CookieSecret    string
		CookieSecure    bool
	}

	TrustAuthConfig struct {
		CookieName        string
		CookieDomain      string
		CookieDuration    time.Duration
		TokenDuration     time.Duration
		ReportURLDuration time.Duration
		TokenSecret       string
		Scope             string
		TokenType         string
		CookieSecure      bool
	}

	Config struct {
		AllowedOrigins    []string
		Probo             *probo.Service
		Auth              *auth.Service
		Authz             *authz.Service
		Trust             *trust.Service
		SAML              *auth.SAMLService
		ConsoleAuth       ConsoleAuthConfig
		TrustAuth         TrustAuthConfig
		ConnectorRegistry *connector.ConnectorRegistry
		SafeRedirect      *saferedirect.SafeRedirect
		CustomDomainCname string
		Logger            *log.Logger
	}

	Server struct {
		cfg             Config
		trustAPIHandler http.Handler
	}
)

var (
	ErrMissingProboService = errors.New("server configuration requires a valid probo.Service instance")
	ErrMissingAuthService  = errors.New("server configuration requires a valid auth.Service instance")
	ErrMissingAuthzService = errors.New("server configuration requires a valid authz.Service instance")
)

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	httpserver.RenderJSON(
		w,
		http.StatusMethodNotAllowed,
		map[string]string{
			"error": "method not allowed",
		},
	)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	httpserver.RenderJSON(
		w,
		http.StatusNotFound,
		map[string]string{
			"error": "not found",
		},
	)
}

func NewServer(cfg Config) (*Server, error) {
	if cfg.Probo == nil {
		return nil, ErrMissingProboService
	}

	if cfg.Auth == nil {
		return nil, ErrMissingAuthService
	}

	if cfg.Authz == nil {
		return nil, ErrMissingAuthzService
	}

	// Create trust API handler once
	trustAPIHandler := trust_v1.NewMux(
		cfg.Logger.Named("trust.v1"),
		cfg.Auth,
		cfg.Authz,
		cfg.Trust,
		console_v1.AuthConfig{
			CookieName:      cfg.ConsoleAuth.CookieName,
			CookieDomain:    cfg.ConsoleAuth.CookieDomain,
			SessionDuration: cfg.ConsoleAuth.SessionDuration,
			CookieSecret:    cfg.ConsoleAuth.CookieSecret,
			CookieSecure:    cfg.ConsoleAuth.CookieSecure,
		},
		trust_v1.TrustAuthConfig{
			CookieName:        cfg.TrustAuth.CookieName,
			CookieDomain:      cfg.TrustAuth.CookieDomain,
			CookieDuration:    cfg.TrustAuth.CookieDuration,
			TokenDuration:     cfg.TrustAuth.TokenDuration,
			ReportURLDuration: cfg.TrustAuth.ReportURLDuration,
			TokenSecret:       cfg.TrustAuth.TokenSecret,
			Scope:             cfg.TrustAuth.Scope,
			TokenType:         cfg.TrustAuth.TokenType,
			CookieSecure:      cfg.TrustAuth.CookieSecure,
		},
	)

	return &Server{
		cfg:             cfg,
		trustAPIHandler: trustAPIHandler,
	}, nil
}

func (s *Server) TrustAPIHandler() http.Handler {
	return s.trustAPIHandler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	corsOpts := cors.Options{
		AllowedOrigins:     s.cfg.AllowedOrigins,
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
		AllowedHeaders:     []string{"content-type", "traceparent", "authorization"},
		ExposedHeaders:     []string{"x-Request-id"},
		AllowCredentials:   true,
		MaxAge:             600, // 10 minutes (chrome >= 76 maximum value c.f. https://source.chromium.org/chromium/chromium/src/+/main:services/network/public/cpp/cors/preflight_result.cc;drc=52002151773d8cd9ffc5f557cd7cc880fddcae3e;l=36)
		OptionsPassthrough: false,
		Debug:              false,
	}

	// Default API security headers
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Permissions-Policy", "microphone=(), camera=(), geolocation=()")

	// Default API security headers
	router := chi.NewRouter()
	router.MethodNotAllowed(methodNotAllowed)
	router.NotFound(notFound)

	router.Use(cors.Handler(corsOpts))

	// Mount the console API with authentication
	router.Mount(
		"/console/v1",
		console_v1.NewMux(
			s.cfg.Logger.Named("console.v1"),
			s.cfg.Probo,
			s.cfg.Auth,
			s.cfg.Authz,
			console_v1.AuthConfig{
				CookieName:      s.cfg.ConsoleAuth.CookieName,
				CookieDomain:    s.cfg.ConsoleAuth.CookieDomain,
				SessionDuration: s.cfg.ConsoleAuth.SessionDuration,
				CookieSecret:    s.cfg.ConsoleAuth.CookieSecret,
				CookieSecure:    s.cfg.ConsoleAuth.CookieSecure,
			},
			s.cfg.ConnectorRegistry,
			s.cfg.SafeRedirect,
			s.cfg.CustomDomainCname,
			s.cfg.SAML,
		),
	)

	// Mount the trust API with authentication
	router.Mount("/trust/v1", s.trustAPIHandler)

	router.ServeHTTP(w, r)
}
