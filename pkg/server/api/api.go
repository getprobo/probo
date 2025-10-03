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

	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/saferedirect"
	console_v1 "github.com/getprobo/probo/pkg/server/api/console/v1"
	trust_v1 "github.com/getprobo/probo/pkg/server/api/trust/v1"
	"github.com/getprobo/probo/pkg/trust"
	"github.com/getprobo/probo/pkg/usrmgr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
)

type (
	ConsoleAuthConfig struct {
		CookieName      string
		CookieDomain    string
		SessionDuration time.Duration
		CookieSecret    string
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
	}

	Config struct {
		AllowedOrigins      []string
		Probo               *probo.Service
		Usrmgr              *usrmgr.Service
		Trust               *trust.Service
		Auth                ConsoleAuthConfig
		TrustAuth           TrustAuthConfig
		ConnectorRegistry   *connector.ConnectorRegistry
		SafeRedirect        *saferedirect.SafeRedirect
		CustomDomainCname   string
		Logger              *log.Logger
	}

	Server struct {
		cfg Config
	}
)

var (
	ErrMissingProboService  = errors.New("server configuration requires a valid probo.Service instance")
	ErrMissingUsrmgrService = errors.New("server configuration requires a valid usrmgr.Service instance")
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

	if cfg.Usrmgr == nil {
		return nil, ErrMissingUsrmgrService
	}

	return &Server{
		cfg: cfg,
	}, nil
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
			s.cfg.Usrmgr,
			console_v1.AuthConfig{
				CookieName:      s.cfg.Auth.CookieName,
				CookieDomain:    s.cfg.Auth.CookieDomain,
				SessionDuration: s.cfg.Auth.SessionDuration,
				CookieSecret:    s.cfg.Auth.CookieSecret,
			},
			s.cfg.ConnectorRegistry,
			s.cfg.SafeRedirect,
			s.cfg.CustomDomainCname,
		),
	)

	// Mount the trust API with authentication
	router.Mount(
		"/trust/v1",
		trust_v1.NewMux(
			s.cfg.Logger.Named("trust.v1"),
			s.cfg.Usrmgr,
			s.cfg.Trust,
			console_v1.AuthConfig{
				CookieName:      s.cfg.Auth.CookieName,
				CookieDomain:    s.cfg.Auth.CookieDomain,
				SessionDuration: s.cfg.Auth.SessionDuration,
				CookieSecret:    s.cfg.Auth.CookieSecret,
			},
			trust_v1.TrustAuthConfig{
				CookieName:        s.cfg.TrustAuth.CookieName,
				CookieDomain:      s.cfg.TrustAuth.CookieDomain,
				CookieDuration:    s.cfg.TrustAuth.CookieDuration,
				TokenDuration:     s.cfg.TrustAuth.TokenDuration,
				ReportURLDuration: s.cfg.TrustAuth.ReportURLDuration,
				TokenSecret:       s.cfg.TrustAuth.TokenSecret,
				Scope:             s.cfg.TrustAuth.Scope,
				TokenType:         s.cfg.TrustAuth.TokenType,
			},
		),
	)

	router.ServeHTTP(w, r)
}
