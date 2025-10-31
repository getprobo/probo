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

	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.gearno.de/kit/httpserver"
)

type (
	AuthenticationStatus string

	ListOrganizationsResponse struct {
		Organizations []OrganizationResponse `json:"organizations"`
	}

	OrganizationResponse struct {
		ID                   gid.GID              `json:"id"`
		Name                 string               `json:"name"`
		LogoURL              *string              `json:"logoUrl,omitempty"`
		AuthenticationMethod string               `json:"authenticationMethod"` // "password", "saml", or "any"
		AuthStatus           AuthenticationStatus `json:"authStatus"`           // "authenticated", "unauthenticated", "expired"
		LoginURL             string               `json:"loginUrl"`             // URL to login (SAML or password login page)
	}
)

const (
	AuthStatusAuthenticated   AuthenticationStatus = "authenticated"
	AuthStatusUnauthenticated AuthenticationStatus = "unauthenticated"
	AuthStatusExpired         AuthenticationStatus = "expired"
)

func buildOrganizationResponse(
	org *coredata.Organization,
	accessResult authsvc.AccessResult,
	sessionData coredata.SessionData,
) OrganizationResponse {
	// Generate logo URL path if organization has a logo
	var logoURL *string
	if org.LogoFileID != nil {
		url := fmt.Sprintf("/connect/organizations/%s/logo", org.ID)
		logoURL = &url
	}

	orgResponse := OrganizationResponse{
		ID:      org.ID,
		Name:    org.Name,
		LogoURL: logoURL,
	}

	// User does not have required authentication
	if !accessResult.Allowed {
		orgResponse.AuthStatus = AuthStatusUnauthenticated

		switch accessResult.MissingAuth {
		case authsvc.AuthMethodSAML, authsvc.AuthMethodAny:
			orgResponse.AuthenticationMethod = "saml"
			if accessResult.SAMLConfig != nil {
				orgResponse.LoginURL = fmt.Sprintf("/connect/saml/login/%s", accessResult.SAMLConfig.ID)
			}
		case authsvc.AuthMethodPassword:
			orgResponse.AuthenticationMethod = "password"
			orgResponse.LoginURL = "/auth/login?method=password"
		}
		return orgResponse
	}

	// User has required authentication
	orgResponse.AuthStatus = AuthStatusAuthenticated

	if sessionData.PasswordAuthenticated {
		orgResponse.AuthenticationMethod = "password"
		orgResponse.LoginURL = "/auth/login?method=password"
	} else if samlInfo, ok := sessionData.SAMLAuthenticatedOrgs[org.ID.String()]; ok {
		orgResponse.AuthenticationMethod = "saml"
		orgResponse.LoginURL = fmt.Sprintf("/connect/saml/login/%s", samlInfo.SAMLConfigID)
	} else {
		orgResponse.AuthenticationMethod = "any"
		orgResponse.LoginURL = "/auth/login?method=password"
	}

	return orgResponse
}

func ListOrganizationsHandler(authSvc *authsvc.Service, authzSvc *authz.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := UserFromContext(ctx)
		sess := SessionFromContext(ctx)

		organizations, err := authzSvc.GetAllUserOrganizations(ctx, user.ID)
		if err != nil {
			panic(fmt.Errorf("cannot list organizations for user: %w", err))
		}

		orgIDs := make([]gid.GID, len(organizations))
		for i, org := range organizations {
			orgIDs[i] = org.ID
		}
		accessResults, err := authSvc.CheckOrganizationAccess(ctx, user, orgIDs, sess)
		if err != nil {
			panic(fmt.Errorf("cannot check organization access: %w", err))
		}

		response := ListOrganizationsResponse{
			Organizations: make([]OrganizationResponse, 0, len(organizations)),
		}
		for _, org := range organizations {
			accessResult := accessResults[org.ID]
			orgResponse := buildOrganizationResponse(org, accessResult, sess.Data)
			response.Organizations = append(response.Organizations, orgResponse)
		}

		httpserver.RenderJSON(w, http.StatusOK, response)
	}
}
