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

package mcp_v1

import (
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	serverauth "go.probo.inc/probo/pkg/server/auth"
)

// WithMCPAuth wraps an HTTP handler with MCP authentication middleware
// It authenticates using API keys from the Authorization header
func WithMCPAuth(
	logger *log.Logger,
	authSvc *auth.Service,
	authzSvc *authz.Service,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		correlationID := r.Header.Get("X-Request-ID")
		if correlationID == "" {
			correlationID = r.Header.Get("X-Correlation-ID")
		}

		logger.InfoCtx(ctx, "MCP authentication attempt",
			log.String("correlation_id", correlationID),
			log.String("path", r.URL.Path),
		)

		// Authenticate using API key from shared function
		authCtx := serverauth.AuthenticateWithAPIKey(ctx, r, authSvc, authzSvc)
		if authCtx == nil {
			logger.WarnCtx(ctx, "MCP auth: authentication required",
				log.String("correlation_id", correlationID),
			)
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		user := serverauth.UserFromContext(authCtx)
		userAPIKey := serverauth.UserAPIKeyFromContext(authCtx)
		tenantAccess := serverauth.UserTenantAccessFromContext(authCtx)

		logger.InfoCtx(authCtx, "MCP authentication successful",
			log.String("correlation_id", correlationID),
			log.String("user_id", user.ID.String()),
			log.String("api_key_id", userAPIKey.ID.String()),
			log.Int("accessible_tenants", len(tenantAccess.TenantIDs)),
		)

		next.ServeHTTP(w, r.WithContext(authCtx))
	})
}
