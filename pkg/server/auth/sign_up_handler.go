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
	"errors"
	"fmt"
	"net/http"

	"go.gearno.de/kit/httpserver"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/securecookie"
)

type (
	SignUpRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
	}

	SignUpResponse struct {
		User UserResponse `json:"user"`
	}
)

func SignUpHandler(authSvc *authsvc.Service, cookieName string, cookieSecret string, cookieSecure bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot decode body: %w", err))
			return
		}

		user, session, err := authSvc.SignUp(
			r.Context(),
			req.Email,
			req.Password,
			req.FullName,
		)
		if err != nil {
			var errUserAlreadyExists *authsvc.ErrUserAlreadyExists
			if errors.As(err, &errUserAlreadyExists) {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot register user: %w", err))
				return
			}

			var errSignupDisabled *authsvc.ErrSignupDisabled
			if errors.As(err, &errSignupDisabled) {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot register user: %w", err))
				return
			}

			panic(fmt.Errorf("cannot register user: %w", err))
		}

		securecookie.Set(
			w,
			securecookie.DefaultConfig(
				cookieName,
				cookieSecret,
				cookieSecure,
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
