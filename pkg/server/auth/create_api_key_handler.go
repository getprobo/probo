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
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/httpserver"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	CreateUserAPIKeyRequest struct {
		Name          string                                    `json:"name"`
		ExpiresAt     time.Time                                 `json:"expiresAt"`
		Organizations []UserAPIKeyOrganizationMembershipRequest `json:"organizations"`
	}

	UserAPIKeyOrganizationMembershipRequest struct {
		OrganizationID string `json:"organizationId"`
		Role           string `json:"role"`
	}

	CreateUserAPIKeyResponse struct {
		UserAPIKey UserAPIKeyResponse `json:"apiKey"`
		Key        string             `json:"key"`
	}
)

func CreateUserAPIKeyHandler(authSvc *authsvc.Service, authzSvc *authz.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := UserFromContext(ctx)

		var req CreateUserAPIKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
			return
		}

		if req.ExpiresAt.IsZero() {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "expiresAt is required",
			})
			return
		}

		if req.ExpiresAt.Before(time.Now()) {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "expiration date must be in the future",
			})
			return
		}

		if len(req.Organizations) == 0 {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "at least one organization is required",
			})
			return
		}

		name := req.Name
		if name == "" {
			name = time.Now().Format("2006-01-02")
		}

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
					"error": "only owners can create API keys for this organization",
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

		userAPIKey, key, err := authSvc.CreateUserAPIKey(ctx, user.ID, name, req.ExpiresAt, memberships)
		if err != nil {
			panic(fmt.Errorf("cannot create user api key: %w", err))
		}

		response := CreateUserAPIKeyResponse{
			UserAPIKey: UserAPIKeyResponse{
				ID:        userAPIKey.ID,
				Name:      userAPIKey.Name,
				ExpiresAt: userAPIKey.ExpiresAt,
				CreatedAt: userAPIKey.CreatedAt,
			},
			Key: key,
		}

		httpserver.RenderJSON(w, http.StatusCreated, response)
	}
}
