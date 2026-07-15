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

package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/agentrun"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/complianceportal/management"
	trust "go.probo.inc/probo/pkg/complianceportal/visitor"
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
	"go.probo.inc/probo/pkg/server/api"
	connect_v1 "go.probo.inc/probo/pkg/server/api/connect/v1"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/server/mailactions"
	console_web "go.probo.inc/probo/pkg/server/web"
	"go.probo.inc/probo/pkg/slack"
	"go.probo.inc/probo/pkg/thirdparty"
	"go.probo.inc/probo/pkg/uri"
)

type Config struct {
	BaseURL           *baseurl.BaseURL
	AllowedOrigins    []string
	ExtraHeaderFields map[string]string
	Probo             *probo.Service
	ResourceAlias     *resourcealias.Service
	File              *filemanager.Service
	IAM               *iam.Service
	Trust             *trust.Service
	ESign             *esign.Service
	CustomDomain      *management.Service
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

type Server struct {
	cfg                Config
	apiServer          *api.Server
	mailActionsHandler http.Handler
	consoleWebServer   *console_web.Server
	router             *chi.Mux
	extraHeaderFields  map[string]string
	baseURL            string
	proboService       *probo.Service
	iamService         *iam.Service
	logger             *log.Logger
}

func NewServer(cfg Config) (*Server, error) {
	apiCfg := api.Config{
		BaseURL:           cfg.BaseURL,
		AllowedOrigins:    cfg.AllowedOrigins,
		Probo:             cfg.Probo,
		ResourceAlias:     cfg.ResourceAlias,
		File:              cfg.File,
		IAM:               cfg.IAM,
		Trust:             cfg.Trust,
		ESign:             cfg.ESign,
		CustomDomain:      cfg.CustomDomain,
		AccessReview:      cfg.AccessReview,
		AgentRun:          cfg.AgentRun,
		Slack:             cfg.Slack,
		Mailman:           cfg.Mailman,
		CookieBanner:      cfg.CookieBanner,
		Geoloc:            cfg.Geoloc,
		ThirdParty:        cfg.ThirdParty,
		RiskManagement:    cfg.RiskManagement,
		Cookie:            cfg.Cookie,
		TokenSecret:       cfg.TokenSecret,
		ConnectorRegistry: cfg.ConnectorRegistry,
		ProviderRegistry:  cfg.ProviderRegistry,
		CustomDomainCname: cfg.CustomDomainCname,
		GraphQLLimits:     cfg.GraphQLLimits,
		Logger:            cfg.Logger.Named("api"),
	}

	apiServer, err := api.NewServer(apiCfg)
	if err != nil {
		return nil, err
	}

	consoleWebServer, err := console_web.NewServer()
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()

	server := &Server{
		cfg:                cfg,
		apiServer:          apiServer,
		mailActionsHandler: mailactions.NewMux(cfg.Mailman, cfg.TokenSecret),
		consoleWebServer:   consoleWebServer,
		router:             router,
		extraHeaderFields:  cfg.ExtraHeaderFields,
		baseURL:            cfg.BaseURL.String(),
		proboService:       cfg.Probo,
		iamService:         cfg.IAM,
		logger:             cfg.Logger,
	}

	server.setupRoutes()

	return server, nil
}

func (s *Server) setupRoutes() {
	// OIDC Discovery 1.0 §4 and RFC 8414 §3 both require the metadata
	// document at the issuer root under well-known paths.
	s.router.Get("/.well-known/openid-configuration", s.oidcDiscoveryHandler)
	s.router.Get("/.well-known/oauth-authorization-server", s.oidcDiscoveryHandler)
	s.router.Get("/.well-known/oauth-protected-resource", s.protectedResourceMetadataHandler)

	s.router.Mount("/api", http.StripPrefix("/api", s.apiServer))
	s.router.Mount("/mail-actions", http.StripPrefix("/mail-actions", s.mailActionsHandler))

	s.router.Mount("/", s.consoleWebServer)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.setExtraHeaders(w)
	s.router.ServeHTTP(w, r)
}

func (s *Server) setExtraHeaders(w http.ResponseWriter) {
	ApplyExtraHeaders(w, s.extraHeaderFields)
}

func (s *Server) oidcDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	metadata := connect_v1.OAuth2ServerMetadata(
		s.cfg.BaseURL,
		s.iamService.OAuth2ScopeRegistry.RegisteredScopes(),
	)

	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpserver.RenderJSON(w, http.StatusOK, metadata)
}

func (s *Server) protectedResourceMetadataHandler(w http.ResponseWriter, r *http.Request) {
	resource := uri.URI(s.baseURL)
	metadata := s.iamService.OAuth2ProtectedResourceMetadata(resource)

	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpserver.RenderJSON(w, http.StatusOK, metadata)
}
