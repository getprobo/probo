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
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/securecookie"
	connect_v1 "go.probo.inc/probo/pkg/server/api/connect/v1"
	console_v1 "go.probo.inc/probo/pkg/server/api/console/v1"
	mcp_v1 "go.probo.inc/probo/pkg/server/api/mcp/v1"
	slack_v1 "go.probo.inc/probo/pkg/server/api/slack/v1"
	trust_v1 "go.probo.inc/probo/pkg/server/api/trust/v1"
	"go.probo.inc/probo/pkg/slack"
	"go.probo.inc/probo/pkg/trust"
)

type (
	Config struct {
		BaseURL           *baseurl.BaseURL
		AllowedOrigins    []string
		Probo             *probo.Service
		IAM               *iam.Service
		Trust             *trust.Service
		Slack             *slack.Service
		Cookie            securecookie.Config
		TokenSecret       string
		ConnectorRegistry *connector.ConnectorRegistry
		CustomDomainCname string
		Logger            *log.Logger
	}

	MCPConfig struct {
		Version        string
		RequestTimeout time.Duration
		MaxRequestSize int64
	}

	Server struct {
		cfg                   Config
		compliancePageHandler http.Handler
		consoleHandler        http.Handler
		mcpHandler            http.Handler
		slackHandler          http.Handler
		connectHandler        http.Handler
	}
)

var (
	ErrMissingProboService = errors.New("server configuration requires a valid probo.Service instance")
	ErrMissingIAMService   = errors.New("server configuration requires a valid iam.Service instance")
	ErrMissingSlackService = errors.New("server configuration requires a valid slack.Service instance")
)

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	httpserver.RenderJSON(
		w,
		http.StatusMethodNotAllowed,
		map[string]string{
			"error": "method not allowed",
		},
	)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

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

	if cfg.IAM == nil {
		return nil, ErrMissingIAMService
	}

	if cfg.Slack == nil {
		return nil, ErrMissingSlackService
	}

	return &Server{
		cfg: cfg,
		compliancePageHandler: trust_v1.NewMux(
			cfg.Logger.Named("trust.v1"),
			cfg.IAM,
			cfg.Trust,
			cfg.Cookie,
			cfg.BaseURL,
		),
		consoleHandler: console_v1.NewMux(
			cfg.Logger.Named("console.v1"),
			cfg.Probo,
			cfg.IAM,
			cfg.Cookie,
			cfg.TokenSecret,
			cfg.ConnectorRegistry,
			cfg.BaseURL,
			cfg.CustomDomainCname,
		),
		mcpHandler: mcp_v1.NewMux(
			cfg.Logger.Named("mcp.v1"),
			cfg.Probo,
			cfg.IAM,
			cfg.TokenSecret,
		),
		slackHandler: slack_v1.NewMux(
			cfg.Logger.Named("slack.v1"),
			cfg.Slack,
			cfg.Trust,
		),
		connectHandler: connect_v1.NewMux(
			cfg.Logger.Named("connect.v1"),
			cfg.IAM,
			cfg.Cookie,
			cfg.TokenSecret,
			cfg.BaseURL,
		),
	}, nil
}

func (s *Server) CompliancePageHandler() http.Handler {
	return s.compliancePageHandler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	corsOpts := cors.Options{
		AllowedOrigins:     s.cfg.AllowedOrigins,
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
		AllowedHeaders:     []string{"content-type", "traceparent", "authorization"},
		ExposedHeaders:     []string{"x-request-id"},
		AllowCredentials:   true,
		MaxAge:             600, // 10 minutes (chrome >= 76 maximum value c.f. https://source.chromium.org/chromium/chromium/src/+/main:services/network/public/cpp/cors/preflight_result.cc;drc=52002151773d8cd9ffc5f557cd7cc880fddcae3e;l=36)
		OptionsPassthrough: false,
		Debug:              false,
	}

	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Protection", "0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'self'")
	w.Header().Set("Permissions-Policy", "microphone=(), camera=(), geolocation=()")

	router := chi.NewRouter()
	router.MethodNotAllowed(methodNotAllowed)
	router.NotFound(notFound)

	router.Use(cors.Handler(corsOpts))

	router.Mount("/console/v1", http.StripPrefix("/console/v1", s.consoleHandler))
	router.Mount("/connect/v1", http.StripPrefix("/connect/v1", s.connectHandler))
	router.Mount("/trust/v1", http.StripPrefix("/trust/v1", s.compliancePageHandler))
	router.Mount("/mcp/v1", http.StripPrefix("/mcp/v1", s.mcpHandler))
	router.Mount("/slack/v1", http.StripPrefix("/slack/v1", s.slackHandler))

	router.ServeHTTP(w, r)
}
