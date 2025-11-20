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
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/gid"
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

		// Extract API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.WarnCtx(ctx, "MCP auth: missing Authorization header",
				log.String("correlation_id", correlationID),
			)
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		// Expect "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			logger.WarnCtx(ctx, "MCP auth: invalid Authorization header format",
				log.String("correlation_id", correlationID),
			)
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		apiKeyToken := parts[1]

		// Validate the API key
		user, userAPIKey, err := authSvc.ValidateUserAPIKey(ctx, apiKeyToken)
		if err != nil {
			logger.WarnCtx(ctx, "MCP auth: invalid API key",
				log.Error(err),
				log.String("correlation_id", correlationID),
			)
			http.Error(w, "invalid api key", http.StatusUnauthorized)
			return
		}

		// Get organizations for this API key through the service layer
		organizations, err := authzSvc.GetAllOrganizationsForUserAPIKeyId(ctx, userAPIKey.ID)
		if err != nil {
			logger.ErrorCtx(ctx, "MCP auth: failed to load API key organizations",
				log.Error(err),
				log.String("correlation_id", correlationID),
			)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		// Extract tenant IDs from organizations
		tenantIDs := make([]gid.TenantID, 0, len(organizations))
		for _, org := range organizations {
			tenantIDs = append(tenantIDs, org.ID.TenantID())
		}

		// Create MCP context with user and accessible tenants
		mcpCtx := &MCPContext{
			UserID:    user.ID,
			TenantIDs: tenantIDs,
		}

		ctx = ContextWithMCPContext(ctx, mcpCtx)

		logger.InfoCtx(ctx, "MCP authentication successful",
			log.String("correlation_id", correlationID),
			log.String("user_id", user.ID.String()),
			log.String("api_key_id", userAPIKey.ID.String()),
			log.Int("accessible_tenants", len(tenantIDs)),
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
