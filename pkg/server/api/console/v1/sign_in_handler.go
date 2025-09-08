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

package console_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/securecookie"
	"github.com/getprobo/probo/pkg/auth"
	"go.gearno.de/kit/httpserver"
)

type (
	SignInRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	SignInResponse struct {
		User UserResponse `json:"user"`
	}

	UserResponse struct {
		ID        gid.GID   `json:"id"`
		Email     string    `json:"email"`
		FullName  string    `json:"fullName"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}
)

func SignInHandler(authSvc *auth.Service, authCfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req SignInRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot decode body: %w", err))
			return
		}

		session, user, err := authSvc.SignIn(r.Context(), req.Email, req.Password)
		if err != nil {
			var ErrInvalidCredentials *auth.ErrInvalidCredentials
			if errors.As(err, &ErrInvalidCredentials) {
				httpserver.RenderError(w, http.StatusUnauthorized, err)
				return
			}

			panic(fmt.Errorf("cannot sign in: %w", err))
		}

		securecookie.Set(
			w,
			securecookie.DefaultConfig(
				authCfg.CookieName,
				authCfg.CookieSecret,
			),
			session.ID.String(),
		)

		httpserver.RenderJSON(
			w,
			http.StatusOK,
			SignInResponse{
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
