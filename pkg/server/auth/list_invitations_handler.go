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
	"fmt"
	"net/http"

	authsvc "github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/getprobo/probo/pkg/server/session"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/pg"
)

type (
	ListInvitationsResponse struct {
		Invitations []InvitationResponse `json:"invitations"`
	}

	InvitationResponse struct {
		ID           gid.GID              `json:"id"`
		Email        string               `json:"email"`
		FullName     string               `json:"fullName"`
		Role         string               `json:"role"`
		ExpiresAt    string               `json:"expiresAt"`
		AcceptedAt   *string              `json:"acceptedAt,omitempty"`
		CreatedAt    string               `json:"createdAt"`
		Organization OrganizationSummary  `json:"organization"`
	}

	OrganizationSummary struct {
		ID   gid.GID `json:"id"`
		Name string  `json:"name"`
	}
)

// loadOrganizationByID loads an organization by ID without tenant scope
func loadOrganizationByID(
	ctx context.Context,
	conn pg.Conn,
	orgID gid.GID,
) (*coredata.Organization, error) {
	query := `
SELECT
	id,
	tenant_id,
	name,
	logo_file_id,
	horizontal_logo_file_id,
	description,
	website_url,
	email,
	headquarter_address,
	custom_domain_id,
	created_at,
	updated_at
FROM
	authz_organizations
WHERE
	id = $1
`

	row := conn.QueryRow(ctx, query, orgID)

	var org coredata.Organization
	err := row.Scan(
		&org.ID,
		&org.TenantID,
		&org.Name,
		&org.LogoFileID,
		&org.HorizontalLogoFileID,
		&org.Description,
		&org.WebsiteURL,
		&org.Email,
		&org.HeadquarterAddress,
		&org.CustomDomainID,
		&org.CreatedAt,
		&org.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load organization: %w", err)
	}

	return &org, nil
}

func ListInvitationsHandler(authSvc *authsvc.Service, authzSvc *authz.Service, authCfg RoutesConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionAuthCfg := session.AuthConfig{
			CookieName:   authCfg.CookieName,
			CookieSecret: authCfg.CookieSecret,
		}

		errorHandler := session.ErrorHandler{
			OnCookieError: func(err error) {
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("invalid session"))
			},
			OnParseError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				session.ClearCookie(w, authCfg)
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("invalid session"))
			},
			OnSessionError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				session.ClearCookie(w, authCfg)
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("session expired"))
			},
			OnUserError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				session.ClearCookie(w, authCfg)
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("user not found"))
			},
			OnTenantError: func(err error) {
				panic(fmt.Errorf("failed to list tenants for user: %w", err))
			},
		}

		authResult := session.TryAuth(ctx, w, r, authSvc, authzSvc, sessionAuthCfg, errorHandler)
		if authResult == nil {
			httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
			return
		}

		// Get pending invitations for the user
		cursor := page.NewCursor(
			1000,
			nil,
			page.Head,
			page.OrderBy[coredata.InvitationOrderField]{
				Field:     coredata.InvitationOrderFieldCreatedAt,
				Direction: page.OrderDirectionDesc,
			},
		)

		invitationFilter := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})

		invitationsPage, err := authzSvc.GetUserInvitations(ctx, authResult.User.EmailAddress, cursor, invitationFilter)
		if err != nil {
			panic(fmt.Errorf("failed to list invitations for user: %w", err))
		}

		// Build response
		response := ListInvitationsResponse{
			Invitations: make([]InvitationResponse, 0, len(invitationsPage.Data)),
		}

		// Load organization data for each invitation
		err = authCfg.PGClient.WithConn(ctx, func(conn pg.Conn) error {
			for _, invitation := range invitationsPage.Data {
				invitationResp := InvitationResponse{
					ID:        invitation.ID,
					Email:     invitation.Email,
					FullName:  invitation.FullName,
					Role:      invitation.Role,
					ExpiresAt: invitation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
					CreatedAt: invitation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
				}

				if invitation.AcceptedAt != nil {
					acceptedAtStr := invitation.AcceptedAt.Format("2006-01-02T15:04:05Z07:00")
					invitationResp.AcceptedAt = &acceptedAtStr
				}

				// Load organization details
				org, err := loadOrganizationByID(ctx, conn, invitation.OrganizationID)
				if err != nil {
					// Log error but continue - organization might have been deleted
					return nil
				}

				invitationResp.Organization = OrganizationSummary{
					ID:   org.ID,
					Name: org.Name,
				}

				response.Invitations = append(response.Invitations, invitationResp)
			}

			return nil
		})
		if err != nil {
			panic(fmt.Errorf("failed to load organization details: %w", err))
		}

		httpserver.RenderJSON(w, http.StatusOK, response)
	}
}
