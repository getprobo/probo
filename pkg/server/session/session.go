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

package session

import (
	"context"
	"errors"
	"net/http"

	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/securecookie"
	"github.com/getprobo/probo/pkg/auth"
)

type AuthConfig struct {
	CookieName   string
	CookieSecret string
}

type AuthResult struct {
	Session    *coredata.Session
	User       *coredata.User
	TenantIDs  []gid.TenantID
	AuthErrors map[gid.TenantID]error // Maps tenant ID to authentication error
}

type ErrorHandler struct {
	OnCookieError  func(err error)
	OnParseError   func(w http.ResponseWriter, authCfg AuthConfig)
	OnSessionError func(w http.ResponseWriter, authCfg AuthConfig)
	OnUserError    func(w http.ResponseWriter, authCfg AuthConfig)
	OnTenantError  func(err error)
}

func TryAuth(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	authSvc *auth.Service,
	authzSvc *authz.Service,
	authCfg AuthConfig,
	errorHandler ErrorHandler,
) *AuthResult {
	cookieValue, err := securecookie.Get(r, securecookie.DefaultConfig(
		authCfg.CookieName,
		authCfg.CookieSecret,
	))
	if err != nil {
		if !errors.Is(err, securecookie.ErrCookieNotFound) && errorHandler.OnCookieError != nil {
			errorHandler.OnCookieError(err)
		}
		return nil
	}

	sessionID, err := gid.ParseGID(cookieValue)
	if err != nil {
		if errorHandler.OnParseError != nil {
			errorHandler.OnParseError(w, authCfg)
		}
		return nil
	}

	session, err := authSvc.GetSession(ctx, sessionID)
	if err != nil {
		if errorHandler.OnSessionError != nil {
			errorHandler.OnSessionError(w, authCfg)
		}
		return nil
	}

	user, err := authSvc.GetUserBySession(ctx, sessionID)
	if err != nil {
		if errorHandler.OnUserError != nil {
			errorHandler.OnUserError(w, authCfg)
		}
		return nil
	}

	organizations, err := authzSvc.GetAllUserOrganizations(ctx, user.ID)
	if err != nil {
		if errorHandler.OnTenantError != nil {
			errorHandler.OnTenantError(err)
		}
		return nil
	}

	// Validate organization access based on authentication requirements
	// Only include organizations the user has proper authentication for
	allowedTenantIDs := make([]gid.TenantID, 0, len(organizations))
	authErrors := make(map[gid.TenantID]error)

	// Extract organization IDs for batch check
	orgIDs := make([]gid.GID, len(organizations))
	for i, org := range organizations {
		orgIDs[i] = org.ID
	}

	// Batch check access to all organizations in a single query
	accessResults, err := authSvc.CheckOrganizationAccess(ctx, user, orgIDs, session)
	if err != nil {
		if errorHandler.OnTenantError != nil {
			errorHandler.OnTenantError(err)
		}
		return nil
	}

	// Process results
	for _, org := range organizations {
		result := accessResults[org.ID]
		if result.Allowed {
			// User has proper authentication for this org
			allowedTenantIDs = append(allowedTenantIDs, org.ID.TenantID())
		} else {
			// Store the authentication error for later use
			authErrors[org.ID.TenantID()] = result.ToError(authSvc.BaseURL())
		}
	}

	return &AuthResult{
		Session:   session,
		User:      user,
		TenantIDs: allowedTenantIDs,
		AuthErrors: authErrors,
	}
}

func ClearCookie(w http.ResponseWriter, authCfg AuthConfig) {
	securecookie.Clear(w, securecookie.DefaultConfig(
		authCfg.CookieName,
		authCfg.CookieSecret,
	))
}
