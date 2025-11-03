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
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/httpserver"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/gid"
)

type (
	ListUserAPIKeysResponse struct {
		UserAPIKeys []UserAPIKeyResponse `json:"apiKeys"`
	}

	UserAPIKeyResponse struct {
		ID            gid.GID                            `json:"id"`
		Name          string                             `json:"name"`
		ExpiresAt     time.Time                          `json:"expiresAt"`
		CreatedAt     time.Time                          `json:"createdAt"`
		Organizations []UserAPIKeyOrganizationMembership `json:"organizations"`
	}

	UserAPIKeyOrganizationMembership struct {
		OrganizationID   gid.GID `json:"organizationId"`
		OrganizationName string  `json:"organizationName"`
		Role             string  `json:"role"`
	}
)

func ListUserAPIKeysHandler(authSvc *authsvc.Service, authzSvc *authz.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := UserFromContext(ctx)

		organizations, err := authzSvc.GetAllUserOrganizations(ctx, user.ID)
		if err != nil {
			panic(fmt.Errorf("cannot list organizations for user: %w", err))
		}

		tenantIDs := make([]gid.TenantID, 0, len(organizations))
		for _, org := range organizations {
			tenantIDs = append(tenantIDs, org.ID.TenantID())
		}

		userAPIKeysWithMemberships, err := authSvc.ListUserAPIKeysWithMemberships(ctx, user.ID, tenantIDs)
		if err != nil {
			panic(fmt.Errorf("cannot list user api keys: %w", err))
		}

		response := ListUserAPIKeysResponse{
			UserAPIKeys: make([]UserAPIKeyResponse, 0, len(userAPIKeysWithMemberships)),
		}

		for _, keyWithMemberships := range userAPIKeysWithMemberships {
			organizations := make([]UserAPIKeyOrganizationMembership, 0, len(keyWithMemberships.Memberships))
			for _, membership := range keyWithMemberships.Memberships {
				organizations = append(organizations, UserAPIKeyOrganizationMembership{
					OrganizationID:   membership.OrganizationID,
					OrganizationName: membership.OrganizationName,
					Role:             membership.Role.String(),
				})
			}

			response.UserAPIKeys = append(response.UserAPIKeys, UserAPIKeyResponse{
				ID:            keyWithMemberships.UserAPIKey.ID,
				Name:          keyWithMemberships.UserAPIKey.Name,
				ExpiresAt:     keyWithMemberships.UserAPIKey.ExpiresAt,
				CreatedAt:     keyWithMemberships.UserAPIKey.CreatedAt,
				Organizations: organizations,
			})
		}

		httpserver.RenderJSON(w, http.StatusOK, response)
	}
}
