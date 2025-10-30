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

	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/gid"
	"go.gearno.de/kit/httpserver"
)

type (
	ListInvitationsResponse struct {
		Invitations []InvitationResponse `json:"invitations"`
	}

	InvitationResponse struct {
		ID           gid.GID                     `json:"id"`
		Email        string                      `json:"email"`
		FullName     string                      `json:"fullName"`
		Role         string                      `json:"role"`
		ExpiresAt    string                      `json:"expiresAt"`
		AcceptedAt   *string                     `json:"acceptedAt,omitempty"`
		CreatedAt    string                      `json:"createdAt"`
		Organization OrganizationResponseSummary `json:"organization"`
	}

	OrganizationResponseSummary struct {
		ID   gid.GID `json:"id"`
		Name string  `json:"name"`
	}
)

func ListInvitationsHandler(authzSvc *authz.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := UserFromContext(ctx)

		invitations, err := authzSvc.GetUserPendingInvitations(ctx, user.EmailAddress)
		if err != nil {
			panic(fmt.Errorf("cannot list invitations for user: %w", err))
		}

		response := ListInvitationsResponse{
			Invitations: make([]InvitationResponse, 0, len(invitations)),
		}

		for _, invitation := range invitations {
			invitationResp := InvitationResponse{
				ID:        invitation.ID,
				Email:     invitation.Email,
				FullName:  invitation.FullName,
				Role:      invitation.Role.String(),
				ExpiresAt: invitation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
				CreatedAt: invitation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				Organization: OrganizationResponseSummary{
					ID:   invitation.Organization.ID,
					Name: invitation.Organization.Name,
				},
			}

			if invitation.AcceptedAt != nil {
				acceptedAtStr := invitation.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
				invitationResp.AcceptedAt = &acceptedAtStr
			}

			response.Invitations = append(response.Invitations, invitationResp)
		}

		httpserver.RenderJSON(w, http.StatusOK, response)
	}
}
