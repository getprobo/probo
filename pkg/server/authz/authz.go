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

package authz

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/server/session"
)

type Config struct {
	Auth         *authsvc.Service
	Authz        *authz.Service
	Logger       *log.Logger
	CookieName   string
	CookieSecret string
	CookieSecure bool
}

type Server struct {
	router *chi.Mux
}

type ctxKey struct{ name string }

var (
	userContextKey = &ctxKey{name: "user"}
)

func NewServer(cfg Config) (*Server, error) {
	router := chi.NewRouter()

	// Apply authentication middleware to all routes
	router.Use(requireAuth(cfg.Auth, cfg.Authz, cfg.CookieName, cfg.CookieSecret, cfg.CookieSecure))

	router.Get("/{organizationID}/permissions", PermissionsHandler(cfg.Authz, UserFromContext, cfg.Logger))

	return &Server{
		router: router,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// requireAuth is a middleware that requires authentication
func requireAuth(
	authService *authsvc.Service,
	authzService *authz.Service,
	cookieName string,
	cookieSecret string,
	cookieSecure bool,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			authResult := session.TryAuth(ctx, w, r, authService, authzService, sessionAuthCfg, errorHandler)
			if authResult == nil {
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
				return
			}

			ctx = context.WithValue(ctx, userContextKey, authResult.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) *coredata.User {
	user, _ := ctx.Value(userContextKey).(*coredata.User)
	return user
}
