//go:generate go tool github.com/99designs/gqlgen generate

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

package trust_v1

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
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
		TokenSecret       string
		Scope             string
		TokenType         string
		CookieSecure      bool
	}

	Resolver struct {
		trust         *trust.Service
		logger        *log.Logger
		iam           *iam.Service
		sessionCookie *authn.Cookie
		baseURL       *baseurl.BaseURL
	}
)

type ctxKey struct{ name string }

var (
	trustCenterIDKey = &ctxKey{name: "trust_center_id"}
)

func TrustCenterIDFromContext(ctx context.Context) gid.GID {
	if trustCenterID, ok := ctx.Value(trustCenterIDKey).(gid.GID); ok {
		return trustCenterID
	}

	return gid.Nil
}

func ContextWithTrustCenterID(ctx context.Context, trustCenterID gid.GID) context.Context {
	return context.WithValue(ctx, trustCenterIDKey, trustCenterID)
}

func NewMux(
	logger *log.Logger,
	iamSvc *iam.Service,
	trustSvc *trust.Service,
	cookieConfig securecookie.Config,
	baseURL *baseurl.BaseURL,
) *chi.Mux {
	r := chi.NewMux()

	r.Use(compliancepage.NewCompliancePagePresenceMiddleware())
	r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
	r.Use(compliancepage.NewMembershipMiddleware(trustSvc, logger))

	graphqlHandler := NewGraphQLHandler(iamSvc, trustSvc, logger, baseURL, cookieConfig)

	r.Handle("/graphql", graphqlHandler)

	return r
}

func (r *Resolver) RootTrustService(ctx context.Context) *trust.TenantService {
	return r.trust.WithTenant(gid.NewTenantID())
}

func (r *Resolver) TrustService(ctx context.Context, tenantID gid.TenantID) *trust.TenantService {
	return r.trust.WithTenant(tenantID)
}
