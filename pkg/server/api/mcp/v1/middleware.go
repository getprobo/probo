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

package v1

import (
	"context"
	"fmt"
	"net/http"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/server/session"
	"github.com/getprobo/probo/pkg/usrmgr"
	"go.gearno.de/kit/log"
)

// WithMCPAuth wraps an HTTP handler with MCP authentication middleware
// It authenticates the user session and stores accessible tenants in context
func WithMCPAuth(
	logger *log.Logger,
	usrmgrSvc *usrmgr.Service,
	authCfg AuthConfig,
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

		sessionAuthCfg := session.AuthConfig{
			CookieName:   authCfg.CookieName,
			CookieSecret: authCfg.CookieSecret,
		}

		errorHandler := session.ErrorHandler{
			OnCookieError: func(err error) {
				logger.ErrorCtx(ctx, "MCP auth: failed to get session cookie",
					log.Error(err),
					log.String("correlation_id", correlationID),
				)
				http.Error(w, "authentication required", http.StatusUnauthorized)
			},
			OnParseError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				logger.WarnCtx(ctx, "MCP auth: failed to parse session",
					log.String("correlation_id", correlationID),
				)
				session.ClearCookie(w, authCfg)
				http.Error(w, "invalid session", http.StatusUnauthorized)
			},
			OnSessionError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				logger.WarnCtx(ctx, "MCP auth: session not found or expired",
					log.String("correlation_id", correlationID),
				)
				session.ClearCookie(w, authCfg)
				http.Error(w, "session expired", http.StatusUnauthorized)
			},
			OnUserError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				logger.WarnCtx(ctx, "MCP auth: user not found",
					log.String("correlation_id", correlationID),
				)
				session.ClearCookie(w, authCfg)
				http.Error(w, "user not found", http.StatusUnauthorized)
			},
			OnTenantError: func(err error) {
				logger.ErrorCtx(ctx, "MCP auth: failed to list tenants for user",
					log.Error(err),
					log.String("correlation_id", correlationID),
				)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			},
		}

		authResult := session.TryAuth(ctx, w, r, usrmgrSvc, sessionAuthCfg, errorHandler)
		if authResult == nil {
			logger.WarnCtx(ctx, "MCP auth: authentication failed",
				log.String("correlation_id", correlationID),
			)
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		// Create MCP context with user and accessible tenants
		// Each tool will validate organization access from its arguments
		mcpCtx := &MCPContext{
			UserID:    authResult.User.ID,
			TenantIDs: authResult.TenantIDs,
		}

		ctx = ContextWithMCPContext(ctx, mcpCtx)

		logger.InfoCtx(ctx, "MCP authentication successful",
			log.String("correlation_id", correlationID),
			log.String("user_id", authResult.User.ID.String()),
			log.Int("accessible_tenants", len(authResult.TenantIDs)),
		)

		// Update session after the handler completes
		defer func() {
			if err := usrmgrSvc.UpdateSession(ctx, authResult.Session); err != nil {
				logger.ErrorCtx(ctx, "MCP auth: failed to update session",
					log.Error(err),
					log.String("correlation_id", correlationID),
					log.String("session_id", authResult.Session.ID.String()),
				)
			}
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateOrganizationAccess validates that the user has access to the given organization
func ValidateOrganizationAccess(ctx context.Context, organizationID gid.GID) error {
	mcpCtx := MCPContextFromContext(ctx)
	if mcpCtx == nil {
		return fmt.Errorf("authentication context not found")
	}

	tenantID := organizationID.TenantID()
	for _, tid := range mcpCtx.TenantIDs {
		if tid == tenantID {
			return nil
		}
	}

	return fmt.Errorf("access denied: user does not have access to organization %s", organizationID.String())
}

// ValidateJSONSchema validates tool arguments against a JSON schema
// This is a placeholder for proper JSON schema validation
func ValidateJSONSchema(ctx context.Context, args interface{}, schemaName string) error {
	// TODO: Implement proper JSON schema validation
	// The MCP SDK's jsonschema validation doesn't seem to work as expected
	// For now, we'll rely on Go's type system and add explicit validation in handlers

	if args == nil {
		return fmt.Errorf("arguments are required for %s", schemaName)
	}

	return nil
}
