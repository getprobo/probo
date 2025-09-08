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
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	console_v1 "github.com/getprobo/probo/pkg/server/api/console/v1"
	"github.com/getprobo/probo/pkg/server/api/trust/v1/schema"
	"github.com/getprobo/probo/pkg/server/api/trust/v1/trustauth"
	gqlutils "github.com/getprobo/probo/pkg/server/graphql"
	"github.com/getprobo/probo/pkg/server/session"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/getprobo/probo/pkg/trust"
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

func TokenAccessFromContext(ctx context.Context) *trustauth.TokenAccessData {
	tokenAccess, _ := ctx.Value(tokenAccessContextKey).(*trustauth.TokenAccessData)
	return tokenAccess
}

// UserFromContext implements trustauth.ContextAccessor interface
func (r *Resolver) UserFromContext(ctx context.Context) *coredata.User {
	return UserFromContext(ctx)
}

// TokenAccessFromContext implements trustauth.ContextAccessor interface
func (r *Resolver) TokenAccessFromContext(ctx context.Context) *trustauth.TokenAccessData {
	return TokenAccessFromContext(ctx)
}

func NewMux(
	logger *log.Logger,
	authSvc *auth.Service,
	authzSvc *authz.Service,
	trustSvc *trust.Service,
	authCfg console_v1.AuthConfig,
	trustAuthCfg TrustAuthConfig,
) *chi.Mux {
	r := chi.NewMux()

	r.Handle("/graphql", graphqlHandler(logger, authSvc, authzSvc, trustSvc, authCfg, trustAuthCfg))

	r.Post("/auth/authenticate", authTokenHandler(trustSvc, trustAuthCfg))
	r.Delete("/auth/logout", trustCenterLogoutHandler(authCfg, trustAuthCfg))

	return r
}

func graphqlHandler(logger *log.Logger, authSvc *auth.Service, authzSvc *authz.Service, trustSvc *trust.Service, authCfg console_v1.AuthConfig, trustAuthCfg TrustAuthConfig) http.HandlerFunc {
	resolver := &Resolver{
		trustCenterSvc: trustSvc,
		authCfg:        authCfg,
		trustAuthCfg:   trustAuthCfg,
	}

	c := schema.Config{
		Resolvers: resolver,
	}

	c.Directives.MustBeAuthenticated = trustauth.MustBeAuthenticatedDirective(resolver)

	es := schema.NewExecutableSchema(c)

	srv := handler.New(es)

	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.Options{})

	srv.Use(extension.Introspection{})
	srv.Use(gqlutils.NewTracingExtension(logger))

	srv.SetRecoverFunc(gqlutils.RecoverFunc)

	return WithSession(authSvc, authzSvc, trustSvc, authCfg, trustAuthCfg, srv.ServeHTTP)
}

func (r *Resolver) RootTrustService(ctx context.Context) *trust.TenantService {
	return r.trustCenterSvc.WithTenant(gid.NewTenantID())
}

func (r *Resolver) PublicTrustService(ctx context.Context, tenantID gid.TenantID) *trust.TenantService {
	return r.trustCenterSvc.WithTenant(tenantID)
}

func (r *Resolver) PrivateTrustService(ctx context.Context, tenantID gid.TenantID) (*trust.TenantService, error) {
	if err := trustauth.ValidateTenantAccess(ctx, r, userTenantContextKey, tenantID); err != nil {
		return nil, fmt.Errorf("cannot access trust center: %w", err)
	}

	return r.trustCenterSvc.WithTenant(tenantID), nil
}

func WithSession(authSvc *auth.Service, authzSvc *authz.Service, trustSvc *trust.Service, authCfg console_v1.AuthConfig, trustAuthCfg TrustAuthConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ip := extractIPAddress(r)
		ctx = context.WithValue(ctx, coredata.ContextKeyIPAddress, ip)

		if authCtx := tryTokenAuth(ctx, w, r, trustSvc, trustAuthCfg); authCtx != nil {
			next(w, r.WithContext(authCtx))
			return
		}

		if authCtx := trySessionAuth(ctx, w, r, authSvc, authzSvc, authCfg); authCtx != nil {
			next(w, r.WithContext(authCtx))
			updateSessionIfNeeded(authCtx, authSvc)
			return
		}

		next(w, r.WithContext(ctx))
	}
}

func trySessionAuth(ctx context.Context, w http.ResponseWriter, r *http.Request, authSvc *auth.Service, authzSvc *authz.Service, authCfg console_v1.AuthConfig) context.Context {
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

	authResult := session.TryAuth(ctx, w, r, authSvc, authzSvc, sessionAuthCfg, errorHandler)
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
	if err := tenantSvc.TrustCenterAccesses.ValidateToken(ctx, basicPayload.Data.TrustCenterID, basicPayload.Data.Email); err != nil {
		clearTokenCookie(w, trustAuthCfg)
		return nil
	}

	tokenAccess := &trustauth.TokenAccessData{
		TrustCenterID: basicPayload.Data.TrustCenterID,
		Email:         basicPayload.Data.Email,
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

func updateSessionIfNeeded(ctx context.Context, authSvc *auth.Service) {
	session := SessionFromContext(ctx)
	if session != nil {
		if _, err := authSvc.UpdateSession(ctx, session.ID); err != nil {
			panic(fmt.Errorf("failed to update session: %w", err))
		}
	}
}

func extractIPAddress(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := strings.Split(xff, ",")[0]; ip != "" {
			return strings.TrimSpace(ip)
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	if ip := strings.Split(r.RemoteAddr, ":")[0]; ip != "" {
		return ip
	}

	return "unknown"
}
