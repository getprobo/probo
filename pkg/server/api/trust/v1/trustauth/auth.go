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

package trustauth

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/server/api/trust/v1/types"
)

type TokenAccessData struct {
	TrustCenterID gid.GID
	Email         mail.Addr
	TenantID      gid.TenantID
	Scope         string
}

type ContextAccessor interface {
	IdentityFromContext(ctx context.Context) *coredata.Identity
	TokenAccessFromContext(ctx context.Context) *TokenAccessData
}

func ValidateTenantAccess(ctx context.Context, accessor ContextAccessor, userTenantContextKey interface{}, resourceTenantID gid.TenantID) error {
	tokenAccess := accessor.TokenAccessFromContext(ctx)
	if tokenAccess != nil {
		if tokenAccess.TenantID != resourceTenantID {
			return fmt.Errorf("access denied: token not authorized for this organization")
		}
		return nil
	}

	identity := accessor.IdentityFromContext(ctx)
	if identity != nil {
		userTenants, ok := ctx.Value(userTenantContextKey).(*[]gid.TenantID)
		if !ok || userTenants == nil {
			return fmt.Errorf("access denied: no tenant information available")
		}

		for _, tenantID := range *userTenants {
			if tenantID == resourceTenantID {
				return nil
			}
		}
		return fmt.Errorf("access denied: not authorized for this organization")
	}

	return fmt.Errorf("access denied: authentication required")
}

func GetCurrentUserRole(ctx context.Context, accessor ContextAccessor) types.Role {
	identity := accessor.IdentityFromContext(ctx)
	tokenAccess := accessor.TokenAccessFromContext(ctx)

	if identity != nil || tokenAccess != nil {
		return types.RoleUser
	}
	return types.RoleNone
}

func MustBeAuthenticatedDirective(accessor ContextAccessor) func(ctx context.Context, obj any, next graphql.Resolver, role *types.Role) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver, role *types.Role) (any, error) {
		currentRole := GetCurrentUserRole(ctx, accessor)

		if role != nil && *role == types.RoleUser && currentRole == types.RoleNone {
			return nil, fmt.Errorf("access denied: authentication required")
		}

		return next(ctx)
	}
}
