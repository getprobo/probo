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

	"github.com/99designs/gqlgen/graphql"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/trust/v1/schema"
	"go.probo.inc/probo/pkg/server/api/trust/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
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
	}
)

type ctxKey struct{ name string }

var (
	TrustCenterKey = &ctxKey{name: "trust_center"}
)

func TrustCenterFromContext(ctx context.Context) probo.TrustCenterInfo {
	trustCenter, _ := ctx.Value(TrustCenterKey).(probo.TrustCenterInfo)
	return trustCenter
}

func ContextWithTrustCenter(ctx context.Context, trustCenter probo.TrustCenterInfo) context.Context {
	return context.WithValue(ctx, TrustCenterKey, trustCenter)
}

func NewMux(
	logger *log.Logger,
	iamSvc *iam.Service,
	trustSvc *trust.Service,
	cookieConfig securecookie.Config,
) *chi.Mux {
	r := chi.NewMux()

	sessionMiddleware := authn.NewSessionMiddleware(iamSvc, cookieConfig)
	r.Use(sessionMiddleware)

	config := schema.Config{
		Resolvers: &Resolver{
			iam:           iamSvc,
			trust:         trustSvc,
			logger:        logger,
			sessionCookie: authn.NewCookie(&cookieConfig),
		},
		Directives: schema.DirectiveRoot{
			MustBeAuthenticated: func(ctx context.Context, obj any, next graphql.Resolver, role *types.Role) (any, error) {
				identity := authn.IdentityFromContext(ctx)

				if identity == nil {
					return nil, gqlutils.Unauthenticatedf(ctx, "authentication required")
				}

				return next(ctx)
			},
		},
	}
	es := schema.NewExecutableSchema(config)
	graphqlHandler := gqlutils.NewHandler(es, logger)

	r.Handle("/graphql", graphqlHandler)

	return r
}

func (r *Resolver) RootTrustService(ctx context.Context) *trust.TenantService {
	return r.trust.WithTenant(gid.NewTenantID())
}

func (r *Resolver) TrustService(ctx context.Context, tenantID gid.TenantID) *trust.TenantService {
	return r.trust.WithTenant(tenantID)
}
