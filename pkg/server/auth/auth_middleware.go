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

package auth

import (
	"context"
	"fmt"
	"net/http"

	"go.gearno.de/kit/httpserver"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/session"
)

type ctxKey struct{ name string }

var (
	sessionContextKey = &ctxKey{name: "session"}
	userContextKey    = &ctxKey{name: "user"}
)

func RequireAuth(
	authSvc *authsvc.Service,
	authzSvc *authz.Service,
	cookieName string,
	cookieSecret string,
	cookieSecure bool,
	next http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionAuthCfg := session.AuthConfig{
			CookieName:   cookieName,
			CookieSecret: cookieSecret,
			CookieSecure: cookieSecure,
		}

		errorHandler := session.ErrorHandler{
			OnCookieError: func(err error) {
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("invalid session"))
			},
			OnParseError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				session.ClearCookie(w, authCfg)
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("invalid session"))
			},
			OnSessionError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				session.ClearCookie(w, authCfg)
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("session expired"))
			},
			OnUserError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				session.ClearCookie(w, authCfg)
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
			},
			OnTenantError: func(err error) {
				panic(fmt.Errorf("cannot list tenants for user: %w", err))
			},
		}

		authResult := session.TryAuth(ctx, w, r, authSvc, authzSvc, sessionAuthCfg, errorHandler)
		if authResult == nil {
			httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
			return
		}

		ctx = context.WithValue(ctx, sessionContextKey, authResult.Session)
		ctx = context.WithValue(ctx, userContextKey, authResult.User)

		next(w, r.WithContext(ctx))
	}
}

func SessionFromContext(ctx context.Context) *coredata.Session {
	session, _ := ctx.Value(sessionContextKey).(*coredata.Session)
	return session
}

func UserFromContext(ctx context.Context) *coredata.User {
	user, _ := ctx.Value(userContextKey).(*coredata.User)
	return user
}
