//go:generate go run github.com/99designs/gqlgen generate

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
	"fmt"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	console_v1 "github.com/getprobo/probo/pkg/server/api/console/v1"
	"github.com/getprobo/probo/pkg/server/api/trust/v1/auth"
	"github.com/getprobo/probo/pkg/server/api/trust/v1/schema"
	gqlutils "github.com/getprobo/probo/pkg/server/graphql"
	"github.com/getprobo/probo/pkg/server/session"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/getprobo/probo/pkg/trust"
	"github.com/getprobo/probo/pkg/usrmgr"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
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
	}

	Resolver struct {
		trustCenterSvc *trust.Service
		authCfg        console_v1.AuthConfig
		trustAuthCfg   TrustAuthConfig
	}

	ctxKey struct{ name string }
)

var (
	sessionContextKey     = &ctxKey{name: "session"}
	userContextKey        = &ctxKey{name: "user"}
	userTenantContextKey  = &ctxKey{name: "user_tenants"}
	tokenAccessContextKey = &ctxKey{name: "token_access"}
)

func SessionFromContext(ctx context.Context) *coredata.Session {
	session, _ := ctx.Value(sessionContextKey).(*coredata.Session)
	return session
}

func UserFromContext(ctx context.Context) *coredata.User {
	user, _ := ctx.Value(userContextKey).(*coredata.User)
	return user
}

func TokenAccessFromContext(ctx context.Context) *auth.TokenAccessData {
	tokenAccess, _ := ctx.Value(tokenAccessContextKey).(*auth.TokenAccessData)
	return tokenAccess
}

// UserFromContext implements auth.ContextAccessor interface
func (r *Resolver) UserFromContext(ctx context.Context) *coredata.User {
	return UserFromContext(ctx)
}

// TokenAccessFromContext implements auth.ContextAccessor interface
func (r *Resolver) TokenAccessFromContext(ctx context.Context) *auth.TokenAccessData {
	return TokenAccessFromContext(ctx)
}

func NewMux(
	logger *log.Logger,
	usrmgrSvc *usrmgr.Service,
	trustSvc *trust.Service,
	authCfg console_v1.AuthConfig,
	trustAuthCfg TrustAuthConfig,
) *chi.Mux {
	r := chi.NewMux()

	r.Handle("/graphql", graphqlHandler(logger, usrmgrSvc, trustSvc, authCfg, trustAuthCfg))

	r.Post("/auth/authenticate", authTokenHandler(trustSvc, trustAuthCfg))
	r.Delete("/auth/logout", trustCenterLogoutHandler(authCfg, trustAuthCfg))

	return r
}

func graphqlHandler(logger *log.Logger, usrmgrSvc *usrmgr.Service, trustSvc *trust.Service, authCfg console_v1.AuthConfig, trustAuthCfg TrustAuthConfig) http.HandlerFunc {
	resolver := &Resolver{
		trustCenterSvc: trustSvc,
		authCfg:        authCfg,
		trustAuthCfg:   trustAuthCfg,
	}

	c := schema.Config{
		Resolvers: resolver,
	}

	c.Directives.MustBeAuthenticated = auth.MustBeAuthenticatedDirective(resolver)

	es := schema.NewExecutableSchema(c)

	srv := handler.New(es)

	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.Options{})

	srv.Use(extension.Introspection{})

	srv.SetRecoverFunc(gqlutils.RecoverFunc)

	return WithSession(usrmgrSvc, trustSvc, authCfg, trustAuthCfg, srv.ServeHTTP)
}

// TrustService returns a trust service scoped to the given tenant
func (r *Resolver) TrustService(ctx context.Context, tenantID gid.TenantID) *trust.TenantService {
	return r.trustCenterSvc.WithTenant(tenantID)
}

// GetTenantService returns a tenant service for the given tenant ID
func (r *Resolver) GetTenantService(ctx context.Context, tenantID gid.TenantID) *trust.TenantService {
	return r.trustCenterSvc.WithTenant(tenantID)
}

func WithSession(usrmgrSvc *usrmgr.Service, trustSvc *trust.Service, authCfg console_v1.AuthConfig, trustAuthCfg TrustAuthConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if authCtx := tryTokenAuth(ctx, w, r, trustSvc, trustAuthCfg); authCtx != nil {
			next(w, r.WithContext(authCtx))
			return
		}

		if authCtx := trySessionAuth(ctx, w, r, usrmgrSvc, authCfg); authCtx != nil {
			next(w, r.WithContext(authCtx))
			updateSessionIfNeeded(authCtx, usrmgrSvc)
			return
		}

		next(w, r.WithContext(ctx))
	}
}

func trySessionAuth(ctx context.Context, w http.ResponseWriter, r *http.Request, usrmgrSvc *usrmgr.Service, authCfg console_v1.AuthConfig) context.Context {
	sessionAuthCfg := session.AuthConfig{
		CookieName:   authCfg.CookieName,
		CookieSecret: authCfg.CookieSecret,
	}

	errorHandler := session.ErrorHandler{
		OnParseError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
			session.ClearCookie(w, authCfg)
		},
		OnSessionError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
			session.ClearCookie(w, authCfg)
		},
		OnUserError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
			session.ClearCookie(w, authCfg)
		},
		OnTenantError: func(err error) {
			session.ClearCookie(w, sessionAuthCfg)
		},
	}

	authResult := session.TryAuth(ctx, w, r, usrmgrSvc, sessionAuthCfg, errorHandler)
	if authResult == nil {
		return nil
	}

	ctx = context.WithValue(ctx, sessionContextKey, authResult.Session)
	ctx = context.WithValue(ctx, userContextKey, authResult.User)
	ctx = context.WithValue(ctx, userTenantContextKey, &authResult.TenantIDs)

	return ctx
}

func tryTokenAuth(ctx context.Context, w http.ResponseWriter, r *http.Request, trustSvc *trust.Service, trustAuthCfg TrustAuthConfig) context.Context {
	cookie, err := r.Cookie(trustAuthCfg.CookieName)
	if err != nil {
		return nil
	}

	basicPayload, err := statelesstoken.ValidateToken[probo.TrustCenterAccessData](
		trustAuthCfg.TokenSecret,
		trustAuthCfg.TokenType,
		cookie.Value,
	)
	if err != nil {
		clearTokenCookie(w, trustAuthCfg)
		return nil
	}

	tenantID := basicPayload.Data.TrustCenterID.TenantID()

	tenantSvc := trustSvc.WithTenant(tenantID)
	payload, err := tenantSvc.TrustCenterAccesses.ValidateToken(ctx, cookie.Value)
	if err != nil {
		clearTokenCookie(w, trustAuthCfg)
		return nil
	}

	tokenAccess := &auth.TokenAccessData{
		TrustCenterID: payload.TrustCenterID,
		Email:         payload.Email,
		TenantID:      tenantID,
		Scope:         trustAuthCfg.Scope,
	}

	return context.WithValue(ctx, tokenAccessContextKey, tokenAccess)
}

func clearTokenCookie(w http.ResponseWriter, trustAuthCfg TrustAuthConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     trustAuthCfg.CookieName,
		Value:    "",
		Domain:   trustAuthCfg.CookieDomain,
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func updateSessionIfNeeded(ctx context.Context, usrmgrSvc *usrmgr.Service) {
	session := SessionFromContext(ctx)
	if session != nil {
		if err := usrmgrSvc.UpdateSession(ctx, session); err != nil {
			panic(fmt.Errorf("failed to update session: %w", err))
		}
	}
}
