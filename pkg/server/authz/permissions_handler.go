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
func PermissionsHandler(
	authzService *authz.Service,
	userFromContext func(ctx context.Context) *coredata.User,
	logger *log.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get organization ID from URL parameter
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

		// Get user from context
		user := userFromContext(ctx)
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get authz service for the tenant
		tenantAuthzSvc := authzService.WithTenant(orgID.TenantID())

		// Get user's role in the organization
		memberRole, err := tenantAuthzSvc.GetUserRoleInOrganization(ctx, user.ID, orgID)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot get user role: %v", err), http.StatusInternalServerError)
			return
		}
		role := authz.Role(memberRole.String())

		// Build permissions map for the role with entity type names
		permissions := make(map[string]map[string]bool)

		entityPerms, ok := authz.RolePermissions[role]
		if !ok {
			http.Error(w, "Invalid role", http.StatusInternalServerError)
			return
		}

		for entityType, actionPerms := range entityPerms {
			entityName, ok := coredata.EntityTypes[entityType]
			if !ok || entityName == "" {
				logger.ErrorCtx(ctx, "Missing entity name in coredata.EntityTypes",
					log.Int("entity_type", int(entityType)),
				)
				continue
			}
			permissions[entityName] = make(map[string]bool)
			for action, allowed := range actionPerms {
				permissions[entityName][string(action)] = allowed
			}
		}

		httpserver.RenderJSON(w, http.StatusOK, permissions)
	}
}
