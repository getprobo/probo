// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/agentrun"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/geoloc"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/riskmanagement"
	"go.probo.inc/probo/pkg/securecookie"
	connect_v1 "go.probo.inc/probo/pkg/server/api/connect/v1"
	console_v1 "go.probo.inc/probo/pkg/server/api/console/v1"
	cookiebanner_v1 "go.probo.inc/probo/pkg/server/api/cookiebanner/v1"
	files_v1 "go.probo.inc/probo/pkg/server/api/files/v1"
	mcp_v1 "go.probo.inc/probo/pkg/server/api/mcp/v1"
	slack_v1 "go.probo.inc/probo/pkg/server/api/slack/v1"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/slack"
	"go.probo.inc/probo/pkg/thirdparty"
)

type (
	Config struct {
		BaseURL           *baseurl.BaseURL
		AllowedOrigins    []string
		Probo             *probo.Service
		ResourceAlias     *resourcealias.Service
		File              *filemanager.Service
		IAM               *iam.Service
		Visitor           *visitor.Service
		ESign             *esign.Service
		Management        *management.Service
		CertManager       *certmanager.Service
		AccessReview      *accessreview.Service
		AgentRun          *agentrun.Service
		Slack             *slack.Service
		Mailman           *mailman.Service
		CookieBanner      *cookiebanner.Service
		Geoloc            *geoloc.Service
		ThirdParty        *thirdparty.Service
		RiskManagement    *riskmanagement.Service
		Cookie            securecookie.Config
		TokenSecret       string
		ConnectorRegistry *connector.ConnectorRegistry
		ProviderRegistry  *provider.Registry
		CustomDomainCname string
		GraphQLLimits     gqlutils.Limits
		Logger            *log.Logger
	}

	MCPConfig struct {
		Version        string
		RequestTimeout time.Duration
		MaxRequestSize int64
	}

	Server struct {
		cfg                 Config
		csrf                *http.CrossOriginProtection
		consoleHandler      http.Handler
		cookieBannerHandler http.Handler
		filesHandler        http.Handler
		mcpHandler          http.Handler
		slackHandler        http.Handler
		connectHandler      http.Handler
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

	csrf := http.NewCrossOriginProtection()
	for _, origin := range cfg.AllowedOrigins {
		if err := csrf.AddTrustedOrigin(origin); err != nil {
			return nil, fmt.Errorf("cannot add trusted origin %q: %w", origin, err)
		}
	}

	// The SAML Assertion Consumer Service endpoint receives cross-origin
	// POSTs from external identity providers by design.
	csrf.AddInsecureBypassPattern("POST /connect/v1/saml/2.0/consume")

	// The cookie banner API is called cross-origin from customer websites
	// by the JS SDK. CORS is handled by the cookie banner middleware.
	// GET and OPTIONS are safe methods (always allowed), but we bypass
	// POST explicitly since it comes from customer origins.
	csrf.AddInsecureBypassPattern("POST /cookie-banner/v1/{rest...}")

	// OAuth2 token, introspection, revocation, and device authorization
	// endpoints receive cross-origin POSTs from external clients.
	csrf.AddInsecureBypassPattern("POST /connect/v1/oauth2/token")
	csrf.AddInsecureBypassPattern("POST /connect/v1/oauth2/introspect")
	csrf.AddInsecureBypassPattern("POST /connect/v1/oauth2/revoke")
	csrf.AddInsecureBypassPattern("POST /connect/v1/oauth2/device")

	csrf.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpserver.RenderJSON(
			w,
			http.StatusForbidden,
			map[string]string{
				"error": "cross-origin request denied",
			},
		)
	}))

	return &Server{
		cfg:  cfg,
		csrf: csrf,
		consoleHandler: console_v1.NewMux(
			cfg.Logger.Named("console.v1"),
			cfg.Probo,
			cfg.ResourceAlias,
			cfg.IAM,
			cfg.ESign,
			cfg.Management,
			cfg.CertManager,
			cfg.AccessReview,
			cfg.AgentRun,
			cfg.Mailman,
			cfg.CookieBanner,
			cfg.Cookie,
			cfg.TokenSecret,
			cfg.ConnectorRegistry,
			cfg.ProviderRegistry,
			cfg.File,
			cfg.BaseURL,
			cfg.CustomDomainCname,
			cfg.ThirdParty,
			cfg.RiskManagement,
			cfg.GraphQLLimits,
		),
		cookieBannerHandler: cookiebanner_v1.NewMux(
			cfg.Logger.Named("cookiebanner.v1"),
			cfg.CookieBanner,
			cfg.Geoloc,
		),
		filesHandler: files_v1.NewMux(
			cfg.Logger.Named("files.v1"),
			cfg.File,
			cfg.Probo,
			cfg.IAM,
			cfg.Cookie,
			cfg.TokenSecret,
			cfg.BaseURL,
		),
		mcpHandler: mcp_v1.NewMux(
			cfg.Logger.Named("mcp.v1"),
			cfg.Probo,
			cfg.Management,
			cfg.CertManager,
			cfg.ResourceAlias,
			cfg.ThirdParty,
			cfg.IAM,
			cfg.AccessReview,
			cfg.CookieBanner,
			cfg.RiskManagement,
			cfg.TokenSecret,
			cfg.File,
			cfg.BaseURL,
		),
		slackHandler: slack_v1.NewMux(
			cfg.Logger.Named("slack.v1"),
			cfg.Slack,
			cfg.Visitor,
		),
		connectHandler: connect_v1.NewMux(
			cfg.Logger.Named("connect.v1"),
			cfg.IAM,
			cfg.Visitor,
			cfg.Cookie,
			cfg.TokenSecret,
			cfg.File,
			cfg.BaseURL,
			func(ctx context.Context, host string) bool {
				if host == cfg.BaseURL.Host() {
					return true
				}

				return cfg.Visitor.IsVerifiedRedirectHost(ctx, host)
			},
			cfg.GraphQLLimits,
		),
	}, nil
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

	// Cookie banner has its own per-banner CORS middleware; mount it
	// outside the global CORS handler so OPTIONS preflights from
	// customer websites are not swallowed by the stricter AllowedOrigins
	// list that applies to console/connect routes.
	router.Mount("/cookie-banner/v1", http.StripPrefix("/cookie-banner/v1", s.cookieBannerHandler))

	router.Group(func(r chi.Router) {
		r.Use(cors.Handler(corsOpts))
		r.Mount("/console/v1", http.StripPrefix("/console/v1", s.consoleHandler))
		r.Mount("/connect/v1", http.StripPrefix("/connect/v1", s.connectHandler))
		r.Mount("/files/v1", http.StripPrefix("/files/v1", s.filesHandler))
		r.Mount("/mcp/v1", http.StripPrefix("/mcp/v1", s.mcpHandler))
		r.Mount("/slack/v1", http.StripPrefix("/slack/v1", s.slackHandler))
	})

	s.csrf.Handler(router).ServeHTTP(w, r)
}
