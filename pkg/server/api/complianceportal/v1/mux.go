// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package complianceportal_v1

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	complianceportalstatics "go.probo.inc/probo/apps/compliance-portal"
	"go.probo.inc/probo/pkg/baseurl"
	visitor "go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/complianceportal"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

type MuxConfig struct {
	BaseURL           *baseurl.BaseURL
	ExtraHeaderFields map[string]string
	AllowedOrigins    []string
	Logger            *log.Logger
	IAM               *iam.Service
	Visitor           *visitor.Service
	ResourceAlias     *resourcealias.Service
	File              *filemanager.Service
	ESign             *esign.Service
	Mailman           *mailman.Service
	Cookie            securecookie.Config
	TokenSecret       string
	GraphQLLimits     gqlutils.Limits
}

func NewMux(cfg MuxConfig) (http.Handler, error) {
	webServer, err := NewServer(compliancePageHeadData())
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()

	appOrigin := ""
	if cfg.BaseURL != nil {
		appOrigin = cfg.BaseURL.Scheme() + "://" + cfg.BaseURL.Host()
	}

	csp, err := complianceportalstatics.ContentSecurityPolicy(appOrigin)
	if err != nil {
		return nil, fmt.Errorf("cannot build content security policy: %w", err)
	}

	r.Use(complianceportal.NewSNIMiddleware(cfg.Visitor))
	r.Use(
		server.NewSecurityHeadersMiddleware(
			server.SecurityHeadersOptions{
				ExtraHeaderFields:     cfg.ExtraHeaderFields,
				ContentSecurityPolicy: csp,
			},
		),
	)

	markdownHandler := complianceportal.NewHandler(cfg.Visitor)

	r.Get("/llms.txt", markdownHandler.HandleLLMsTxt)
	r.Get("/robots.txt", markdownHandler.HandleRobotsTxt)
	r.Get("/sitemap.xml", markdownHandler.HandleSitemap)

	allowedHost := saferedirect.Any(
		saferedirect.Origins(cfg.AllowedOrigins...),
		cfg.Visitor.IsVerifiedRedirectHost,
	)

	oauthInitiateHandler := NewOAuthInitiateHandler(
		cfg.BaseURL,
		cfg.Visitor,
		allowedHost,
		cfg.Logger,
	)
	oauthCallbackHandler := NewOAuthCallbackHandler(
		cfg.IAM,
		cfg.Visitor,
		cfg.Cookie,
		allowedHost,
		cfg.Logger,
	)

	graphqlHandler := NewGraphQLHandler(
		cfg.IAM,
		cfg.Visitor,
		cfg.ResourceAlias,
		cfg.File,
		cfg.ESign,
		cfg.Mailman,
		cfg.Logger,
		cfg.BaseURL,
		cfg.Cookie,
		cfg.TokenSecret,
		cfg.GraphQLLimits,
	)

	r.Group(
		func(r chi.Router) {
			r.Use(complianceportal.NewCompliancePortalPresenceMiddleware())

			r.Method(http.MethodGet, complianceportal.CIMDMetadataPath, NewOAuthClientMetadataHandler(cfg.Visitor))
			r.Method(http.MethodGet, complianceportal.BrandLogoPath, NewBrandLogoHandler(cfg.Logger, cfg.File))
			r.Method(http.MethodGet, complianceportal.BrandDarkLogoPath, NewBrandDarkLogoHandler(cfg.Logger, cfg.File))
			r.Method(http.MethodGet, complianceportal.OAuthInitiatePath, oauthInitiateHandler)
			r.Method(http.MethodGet, complianceportal.OAuthCallbackPath, oauthCallbackHandler)

			r.Group(
				func(r chi.Router) {
					r.Use(authn.NewSessionMiddleware(cfg.IAM, cfg.Cookie))
					r.Use(complianceportal.NewSessionHostMiddleware(cfg.Cookie))
					r.Use(complianceportal.NewMemberProvisioningMiddleware(cfg.Visitor, cfg.Logger))
					r.Handle(complianceportal.GraphQLPath, graphqlHandler)
				},
			)

			r.Handle("/*", webServer)
		},
	)

	return r, nil
}

func compliancePageHeadData() HeadDataFunc {
	return func(r *http.Request) HeadData {
		tc := complianceportal.CompliancePortalFromContext(r.Context())
		if tc == nil {
			return HeadData{Title: "Compliance Page"}
		}

		compliancePageBaseURL := complianceportal.CompliancePortalBaseURLFromContext(r.Context())

		pageTitle := "Compliance at " + tc.EntityName + "."

		description := pageTitle
		if tc.Description != nil && *tc.Description != "" {
			description = *tc.Description
		}

		htmlLang, canonical, hreflang := SEOFromRequest(r)

		headData := HeadData{
			Title:        pageTitle,
			Description:  description,
			HTMLLang:     htmlLang,
			OGURL:        canonical,
			CanonicalURL: canonical,
			Hreflang:     hreflang,
		}

		if tc.LogoFileID != nil && compliancePageBaseURL != nil {
			faviconURL, err := visitor.BrandLogoURL(*compliancePageBaseURL)
			if err == nil {
				headData.FaviconURL = faviconURL
			}
		}

		return headData
	}
}
