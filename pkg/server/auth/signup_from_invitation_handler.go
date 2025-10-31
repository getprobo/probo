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

	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/securecookie"
	"go.gearno.de/kit/httpserver"
)

type (
	SignupFromInvitationRequest struct {
		Token    string `json:"token"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
	}

	SignupFromInvitationResponse struct {
	}
)

func SignupFromInvitationHandler(authSvc *authsvc.Service, cookieName string, cookieSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignupFromInvitationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot decode body: %w", err))
			return
		}

		user, session, err := authSvc.SignupFromInvitation(r.Context(), req.Token, req.Password, req.FullName)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, err)
			return
		}

		securecookie.Set(
			w,
			securecookie.DefaultConfig(
				cookieName,
				cookieSecret,
			),
			session.ID.String(),
		)

		httpserver.RenderJSON(
			w,
			http.StatusOK,
			SignUpResponse{
				User: UserResponse{
					ID:        user.ID,
					Email:     user.EmailAddress,
					FullName:  user.FullName,
					CreatedAt: user.CreatedAt,
					UpdatedAt: user.UpdatedAt,
				},
			},
		)
	}
}
