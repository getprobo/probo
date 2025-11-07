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
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// PermissionsHandler returns permissions for the current user's role in an organization
// It uses the centralized permissions map
func PermissionsHandler(
	authzService *authz.Service,
	userFromContext func(ctx context.Context) *coredata.User,
	logger *log.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		orgIDStr := chi.URLParam(r, "organizationID")
		if orgIDStr == "" {
			http.Error(w, "organizationID parameter required", http.StatusBadRequest)
			return
		}

		orgID, err := gid.ParseGID(orgIDStr)
		if err != nil {
			http.Error(w, "invalid organizationID", http.StatusBadRequest)
			return
		}

		user := userFromContext(ctx)
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tenantAuthzSvc := authzService.WithTenant(orgID.TenantID())

		memberRole, err := tenantAuthzSvc.GetUserRoleInOrganization(ctx, user.ID, orgID)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot get user role: %v", err), http.StatusInternalServerError)
			return
		}
		userRole := authz.Role(memberRole.String())

		permissions := authz.GetPermissionsByRole(userRole)

		response := map[string]any{
			"permissions": permissions,
			"role":        memberRole.String(),
		}

		httpserver.RenderJSON(w, http.StatusOK, response)
	}
}
