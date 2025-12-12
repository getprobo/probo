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

	"go.gearno.de/kit/httpserver"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/session"
)

type (
	AcceptInvitationRequest struct {
		InvitationID gid.GID `json:"invitationId"`
	}

	AcceptInvitationResponse struct {
		InvitationID gid.GID `json:"invitationId"`
	}
)

func AcceptInvitationHandler(authSvc *authsvc.Service, authzSvc *authz.Service, cookieName string, cookieSecret string, cookieSecure bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionAuthCfg := session.AuthConfig{
			CookieName:   cookieName,
			CookieSecret: cookieSecret,
			CookieSecure: cookieSecure,
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
				panic(fmt.Errorf("cannot list tenants for user: %w", err))
			},
		}

		authResult := session.TryAuth(ctx, w, r, authSvc, authzSvc, sessionAuthCfg, errorHandler)
		if authResult == nil {
			httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
			return
		}

		// Parse request body
		var req AcceptInvitationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
			return
		}

		// Accept the invitation
		_, err := authzSvc.AcceptInvitationByID(ctx, req.InvitationID, authResult.User.ID)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, err)
			return
		}

		response := AcceptInvitationResponse(req)

		httpserver.RenderJSON(w, http.StatusOK, response)
	}
}
