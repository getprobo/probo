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
	"net/http"
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

var (
	// UserContextKey is the context key for the authenticated user
	UserContextKey = &ctxKey{name: "user"}
	// UserTenantContextKey is the context key for tenant access information
	UserTenantContextKey = &ctxKey{name: "user_tenants"}
	// UserAPIKeyContextKey is the context key for the API key used for authentication
	UserAPIKeyContextKey = &ctxKey{name: "user_api_key"}
)

type UserTenantAccess struct {
	TenantIDs  []gid.TenantID
	AuthErrors map[gid.TenantID]error
}

// UserFromContext extracts the authenticated user from the context.
func UserFromContext(ctx context.Context) *coredata.User {
	user, _ := ctx.Value(UserContextKey).(*coredata.User)
	return user
}

// UserAPIKeyFromContext extracts the API key from the context.
func UserAPIKeyFromContext(ctx context.Context) *coredata.UserAPIKey {
	userAPIKey, _ := ctx.Value(UserAPIKeyContextKey).(*coredata.UserAPIKey)
	return userAPIKey
}

// UserTenantAccessFromContext extracts the tenant access information from the context.
func UserTenantAccessFromContext(ctx context.Context) *UserTenantAccess {
	access, _ := ctx.Value(UserTenantContextKey).(*UserTenantAccess)
	return access
}

// AuthenticateWithAPIKey attempts to authenticate using an API key from the Authorization header.
// It returns a context with authentication information if successful, or nil if no API key
// was provided or authentication failed. This function does not return errors - it silently
// fails to allow fallback to other authentication methods.
func AuthenticateWithAPIKey(ctx context.Context, r *http.Request, authSvc *auth.Service, authzSvc *authz.Service) context.Context {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil
	}

	apiKeyString := strings.TrimPrefix(authHeader, "Bearer ")

	user, userAPIKey, err := authSvc.ValidateUserAPIKey(ctx, apiKeyString)
	if err != nil {
		return nil
	}

	organizations, err := authzSvc.GetAllOrganizationsForUserAPIKeyId(ctx, userAPIKey.ID)
	if err != nil {
		return nil
	}

	tenantIDs := make([]gid.TenantID, 0, len(organizations))
	for _, org := range organizations {
		tenantIDs = append(tenantIDs, org.ID.TenantID())
	}

	ctx = context.WithValue(ctx, UserContextKey, user)
	ctx = context.WithValue(ctx, UserAPIKeyContextKey, userAPIKey)
	ctx = context.WithValue(ctx, UserTenantContextKey, &UserTenantAccess{
		TenantIDs:  tenantIDs,
		AuthErrors: make(map[gid.TenantID]error),
	})

	return ctx
}

// RequireTenantAccess ensures that the authenticated user has access to the specified tenant.
// It panics with an authz.TenantAccessError if access is denied.
func RequireTenantAccess(ctx context.Context, tenantID gid.TenantID) {
	access := UserTenantAccessFromContext(ctx)

	if access == nil {
		panic(&authz.TenantAccessError{Message: "tenant not found"})
	}

	if !slices.Contains(access.TenantIDs, tenantID) {
		if access.AuthErrors != nil {
			if authErr := access.AuthErrors[tenantID]; authErr != nil {
				panic(authErr)
			}
		}

		panic(&authz.TenantAccessError{Message: "tenant not found"})
	}
}
