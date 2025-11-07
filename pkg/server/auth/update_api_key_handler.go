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
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"go.gearno.de/kit/httpserver"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type UpdateUserAPIKeyRequest struct {
	ID            string                                    `json:"id"`
	Name          *string                                   `json:"name,omitempty"`
	Organizations []UserAPIKeyOrganizationMembershipRequest `json:"organizations"`
}

func UpdateUserAPIKeyHandler(authSvc *authsvc.Service, authzSvc *authz.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := UserFromContext(ctx)

		var req UpdateUserAPIKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
			return
		}

		if req.ID == "" {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "user api key id is required",
			})
			return
		}

		userAPIKeyID, err := gid.ParseGID(req.ID)
		if err != nil {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid user api key id",
			})
			return
		}

		if req.Name == nil && len(req.Organizations) == 0 {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "at least one field must be provided for update",
			})
			return
		}

		if req.Name != nil && *req.Name != "" {
			if err := authSvc.UpdateUserAPIKeyName(ctx, userAPIKeyID, user.ID, *req.Name); err != nil {
				var errNotFound *coredata.ErrUserAPIKeyNotFound
				if errors.As(err, &errNotFound) {
					httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{
						"error": "user api key not found",
					})
					return
				}
				if strings.Contains(err.Error(), "does not belong to user") {
					httpserver.RenderJSON(w, http.StatusForbidden, map[string]string{
						"error": "access denied",
					})
					return
				}
				httpserver.RenderJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "failed to update user api key name",
				})
				return
			}
		}

		if len(req.Organizations) > 0 {
			orgInputs := make([]authsvc.UserAPIKeyOrganizationRequest, len(req.Organizations))
			for i, org := range req.Organizations {
				orgID, err := gid.ParseGID(org.OrganizationID)
				if err != nil {
					httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
						"error": "invalid organization id",
					})
					return
				}

				// Check if user is an OWNER for this organization
				tenantAuthzSvc := authzSvc.WithTenant(orgID.TenantID())
				role, err := tenantAuthzSvc.GetUserRoleInOrganization(ctx, user.ID, orgID)
				if err != nil {
					httpserver.RenderJSON(w, http.StatusForbidden, map[string]string{
						"error": "user does not have access to this organization",
					})
					return
				}
				if role != coredata.MembershipRoleOwner {
					httpserver.RenderJSON(w, http.StatusForbidden, map[string]string{
						"error": "only owners can update API keys for this organization",
					})
					return
				}

				orgInputs[i] = authsvc.UserAPIKeyOrganizationRequest{
					OrganizationID: orgID,
					Role:           coredata.APIRole(org.Role),
				}
			}

			memberships, err := authSvc.ValidateAndBuildUserAPIKeyMemberships(ctx, user.ID, orgInputs)
			if err != nil {
				httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
					"error": "invalid organizations",
				})
				return
			}

			if err := authSvc.UpdateUserAPIKeyMemberships(ctx, userAPIKeyID, user.ID, memberships); err != nil {
				var errNotFound *coredata.ErrUserAPIKeyNotFound
				if errors.As(err, &errNotFound) {
					httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{
						"error": "user api key not found",
					})
					return
				}
				if strings.Contains(err.Error(), "does not belong to user") {
					httpserver.RenderJSON(w, http.StatusForbidden, map[string]string{
						"error": "access denied",
					})
					return
				}
				httpserver.RenderJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "failed to update user api key",
				})
				return
			}
		}

		httpserver.RenderJSON(w, http.StatusOK, map[string]string{
			"message": "User API key updated successfully",
		})
	}
}
