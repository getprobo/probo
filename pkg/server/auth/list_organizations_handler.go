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
	"errors"
	"fmt"
	"net/http"
	"time"

	authsvc "github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/filemanager"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/server/session"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/pg"
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

// generateLogoURL generates a presigned URL for an organization's logo
func generateLogoURL(
	ctx context.Context,
	fileManager *filemanager.Service,
	conn pg.Conn,
	logoFileID *gid.GID,
) (*string, error) {
	if logoFileID == nil {
		return nil, nil
	}

	var file coredata.File
	// Load file without scope since we're in auth context (cross-tenant)
	q := `SELECT bucket_name, file_key, file_name, mime_type, file_size FROM files WHERE id = $1`
	err := conn.QueryRow(ctx, q, logoFileID).Scan(
		&file.BucketName,
		&file.FileKey,
		&file.FileName,
		&file.MimeType,
		&file.FileSize,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load file: %w", err)
	}

	presignedURL, err := fileManager.GenerateFileUrl(ctx, &file, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func ListOrganizationsHandler(authSvc *authsvc.Service, authzSvc *authz.Service, authCfg RoutesConfig) http.HandlerFunc {
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

		// Get all organizations for the user (without filtering by authentication state)
		organizations, err := authzSvc.GetAllUserOrganizations(ctx, authResult.User.ID)
		if err != nil {
			panic(fmt.Errorf("failed to list organizations for user: %w", err))
		}

		// Build response with authentication requirements for each organization
		response := ListOrganizationsResponse{
			Organizations: make([]OrganizationResponse, 0, len(organizations)),
		}

		for _, org := range organizations {
			orgResponse := OrganizationResponse{
				ID:   org.ID,
				Name: org.Name,
			}

			// Generate logo URL if available
			if authCfg.FileManager != nil && authCfg.PGClient != nil {
				err := authCfg.PGClient.WithConn(ctx, func(conn pg.Conn) error {
					logoURL, err := generateLogoURL(ctx, authCfg.FileManager, conn, org.LogoFileID)
					if err != nil {
						// Log error but don't fail the request
						return nil
					}
					orgResponse.LogoURL = logoURL
					return nil
				})
				if err != nil {
					// Log error but continue
				}
			}

			// Check authentication requirements for this organization
			err := authSvc.CheckOrganizationAccess(ctx, authResult.User, org.ID, authResult.Session)
			if err != nil {
				// User needs additional authentication
				var errSAMLRequired authsvc.ErrSAMLAuthRequired
				if errors.As(err, &errSAMLRequired) {
					orgResponse.AuthenticationMethod = "saml"
					orgResponse.AuthStatus = AuthStatusUnauthenticated
					orgResponse.LoginURL = fmt.Sprintf("/auth/saml/login/%s", errSAMLRequired.ConfigID)
				} else {
					orgResponse.AuthenticationMethod = "password"
					orgResponse.AuthStatus = AuthStatusUnauthenticated
					orgResponse.LoginURL = "/authentication/login?method=password"
				}
			} else {
				// User has proper authentication
				orgResponse.AuthStatus = AuthStatusAuthenticated

				// Determine which auth method they used
				if authResult.Session.Data.PasswordAuthenticated {
					orgResponse.AuthenticationMethod = "password"
					orgResponse.LoginURL = "/authentication/login?method=password"
				} else if len(authResult.Session.Data.SAMLAuthenticatedOrgs) > 0 {
					// Find SAML config for this org
					orgResponse.AuthenticationMethod = "saml"
					// Try to find the SAML config ID for login URL
					if samlInfo, ok := authResult.Session.Data.SAMLAuthenticatedOrgs[org.ID.String()]; ok {
						orgResponse.LoginURL = fmt.Sprintf("/auth/saml/login/%s", samlInfo.SAMLConfigID)
					} else {
						orgResponse.LoginURL = "/authentication/login?method=password"
					}
				} else {
					orgResponse.AuthenticationMethod = "any"
					orgResponse.LoginURL = "/authentication/login?method=password"
				}
			}

			response.Organizations = append(response.Organizations, orgResponse)
		}

		httpserver.RenderJSON(w, http.StatusOK, response)
	}
}
