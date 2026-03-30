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

//go:generate go tool github.com/99designs/gqlgen generate

package trust_v1

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/compliancepage"
	"go.probo.inc/probo/pkg/trust"
)

type (
	TrustAuthConfig struct {
		CookieName        string
		CookieDomain      string
		CookieDuration    time.Duration
		TokenDuration     time.Duration
		ReportURLDuration time.Duration
		Scope             string
		TokenType         string
		CookieSecure      bool
	}

	Resolver struct {
		trust         *trust.Service
		esign         *esign.Service
		mailman       *mailman.Service
		logger        *log.Logger
		iam           *iam.Service
		sessionCookie *authn.Cookie
		baseURL       *baseurl.BaseURL
	}
)

func NewMux(
	logger *log.Logger,
	iamSvc *iam.Service,
	trustSvc *trust.Service,
	esignSvc *esign.Service,
	mailmanSvc *mailman.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	baseURL *baseurl.BaseURL,
) *chi.Mux {
	r := chi.NewMux()

	r.Use(compliancepage.NewCompliancePagePresenceMiddleware())
	r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
	r.Use(compliancepage.NewMemberProvisioningMiddleware(trustSvc, logger))

	graphqlHandler := NewGraphQLHandler(iamSvc, trustSvc, esignSvc, mailmanSvc, logger, baseURL, cookieConfig, tokenSecret)
	r.Handle("/graphql", graphqlHandler)

	return r
}

func (r *Resolver) TrustService(ctx context.Context, tenantID gid.TenantID) *trust.TenantService {
	return r.trust.WithTenant(tenantID)
}
