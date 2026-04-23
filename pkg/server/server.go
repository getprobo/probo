// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/domainconnect"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/file"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oauth2server"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api"
	"go.probo.inc/probo/pkg/server/api/compliancepage"
	"go.probo.inc/probo/pkg/server/mailactions"
	trust_web "go.probo.inc/probo/pkg/server/trust"
	console_web "go.probo.inc/probo/pkg/server/web"
	"go.probo.inc/probo/pkg/slack"
	"go.probo.inc/probo/pkg/trust"
	"go.probo.inc/probo/pkg/uri"
)

type Config struct {
	BaseURL           *baseurl.BaseURL
	AllowedOrigins    []string
	ExtraHeaderFields map[string]string
	Probo             *probo.Service
	File              *file.Service
	IAM               *iam.Service
	Trust             *trust.Service
	ESign             *esign.Service
	AccessReview      *accessreview.Service
	Slack             *slack.Service
	Mailman           *mailman.Service
	CookieBanner      *cookiebanner.Service
	Cookie            securecookie.Config
	TokenSecret       string
	ConnectorRegistry *connector.ConnectorRegistry
	CustomDomainCname string
	DomainConnect     domainconnect.Config
	ResolverAddr      string
	Logger            *log.Logger
}

type Server struct {
	apiServer          *api.Server
	mailActionsHandler http.Handler
	consoleWebServer   *console_web.Server
	trustWebServer     *trust_web.Server
	router             *chi.Mux
	extraHeaderFields  map[string]string
	baseURL            string
	proboService       *probo.Service
	iamService         *iam.Service
	trustService       *trust.Service
	logger             *log.Logger
}

func NewServer(cfg Config) (*Server, error) {
	apiCfg := api.Config{
		BaseURL:           cfg.BaseURL,
		AllowedOrigins:    cfg.AllowedOrigins,
		Probo:             cfg.Probo,
		File:              cfg.File,
		IAM:               cfg.IAM,
		Trust:             cfg.Trust,
		ESign:             cfg.ESign,
		AccessReview:      cfg.AccessReview,
		Slack:             cfg.Slack,
		Mailman:           cfg.Mailman,
		CookieBanner:      cfg.CookieBanner,
		Cookie:            cfg.Cookie,
		TokenSecret:       cfg.TokenSecret,
		ConnectorRegistry: cfg.ConnectorRegistry,
		CustomDomainCname: cfg.CustomDomainCname,
		DomainConnect:     cfg.DomainConnect,
		ResolverAddr:      cfg.ResolverAddr,
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

	trustWebServer, err := trust_web.NewServer(compliancePageHeadData(cfg.BaseURL, cfg.Trust))
	if err != nil {
		return nil, err
	}

	router := chi.NewRouter()

	server := &Server{
		apiServer:          apiServer,
		mailActionsHandler: mailactions.NewMux(cfg.Mailman, cfg.TokenSecret),
		consoleWebServer:   consoleWebServer,
		trustWebServer:     trustWebServer,
		router:             router,
		extraHeaderFields:  cfg.ExtraHeaderFields,
		baseURL:            cfg.BaseURL.String(),
		proboService:       cfg.Probo,
		iamService:         cfg.IAM,
		trustService:       cfg.Trust,
		logger:             cfg.Logger,
	}

	server.setupRoutes(cfg.BaseURL.String())

	return server, nil
}

func (s *Server) setupRoutes(baseURL string) {
	// OIDC Discovery 1.0 §4 and RFC 8414 §3 both require the metadata
	// document at the issuer root under well-known paths.
	s.router.Get("/.well-known/openid-configuration", s.oidcDiscoveryHandler)
	s.router.Get("/.well-known/oauth-authorization-server", s.oidcDiscoveryHandler)

	s.router.Mount("/api", http.StripPrefix("/api", s.apiServer))
	s.router.Mount("/mail-actions", http.StripPrefix("/mail-actions", s.mailActionsHandler))

	s.router.Route("/trust/{slugOrId}", func(r chi.Router) {
		r.Use(compliancepage.NewIDMiddleware(s.trustService, baseURL))
		r.Use(s.stripTrustPrefix)
		r.Mount("/", s.trustCenterRouter())
	})

	s.router.Mount("/", s.consoleWebServer)
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

func (s *Server) oidcDiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	api := s.baseURL + "/api/connect/v1"

	endpoints := oauth2server.Endpoints{
		Authorization:       uri.URI(api + "/oauth2/authorize"),
		Token:               uri.URI(api + "/oauth2/token"),
		Userinfo:            uri.URI(api + "/oauth2/userinfo"),
		JWKS:                uri.URI(api + "/oauth2/jwks"),
		Registration:        uri.URI(api + "/oauth2/register"),
		Introspection:       uri.URI(api + "/oauth2/introspect"),
		Revocation:          uri.URI(api + "/oauth2/revoke"),
		DeviceAuthorization: uri.URI(api + "/oauth2/device"),
	}

	metadata := s.iamService.OAuth2ServerService.Metadata(endpoints)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpserver.RenderJSON(w, http.StatusOK, metadata)
}

func (s *Server) handleCustomDomain404(w http.ResponseWriter, r *http.Request) {
	httpserver.RenderError(w, http.StatusNotFound, errors.New("not found"))
}

func (s *Server) stripTrustPrefix(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slugOrId := chi.URLParam(r, "slugOrId")
		prefix := "/trust/" + slugOrId

		if r.URL.Path == prefix {
			cleanPath := path.Clean(prefix) + "/"
			http.Redirect(w, r, cleanPath, http.StatusMovedPermanently)
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

	h := compliancepage.NewHandler(s.trustService)

	r.Mount("/api/trust/v1", s.apiServer.CompliancePageHandler())
	r.Get("/llms.txt", h.HandleLLMsTxt)
	r.Get("/robots.txt", h.HandleRobotsTxt)
	r.Get("/sitemap.xml", h.HandleSitemap)
	r.Handle("/*", s.trustWebServer)

	return r
}

func (s *Server) TrustCenterHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(compliancepage.NewSNIMiddleware(s.trustService))
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; preload")
			s.setExtraHeaders(w)
			next.ServeHTTP(w, r)
		})
	})

	r.NotFound(s.handleCustomDomain404)

	r.Mount("/", s.trustCenterRouter())

	return r
}

func compliancePageHeadData(baseURL *baseurl.BaseURL, trustService *trust.Service) trust_web.HeadDataFunc {
	return func(r *http.Request) trust_web.HeadData {
		tc := compliancepage.CompliancePageFromContext(r.Context())
		if tc == nil {
			return trust_web.HeadData{Title: "Compliance Page"}
		}

		org, err := trustService.GetOrganizationByTrustCenterID(r.Context(), tc.ID)
		if err != nil || org == nil {
			return trust_web.HeadData{Title: "Compliance Page"}
		}

		compliancePageBaseURL := compliancepage.CompliancePageBaseURLFromContext(r.Context())

		headData := trust_web.HeadData{
			Title:       org.Name + " — Compliance",
			Description: org.Name + " Compliance Page",
			OGURL:       ref.UnrefOrZero(compliancePageBaseURL),
		}

		if tc.LogoFileID != nil {
			faviconURL, err := baseURL.WithPath("/api/files/v1/" + tc.LogoFileID.String()).String()
			if err == nil {
				headData.FaviconURL = faviconURL
			}
		}

		return headData
	}
}
