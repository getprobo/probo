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

	"errors"

	authsvc "go.probo.inc/probo/pkg/auth"
	"go.gearno.de/kit/httpserver"
)

type (
	ResetPasswordRequest struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	ResetPasswordResponse struct {
		Success bool `json:"success"`
	}
)

func ResetPasswordHandler(authSvc *authsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot decode body: %w", err))
			return
		}

		err := authSvc.ResetPassword(r.Context(), req.Token, req.Password)
		if err != nil {
			var invalidPasswordErr *authsvc.ErrInvalidPassword
			var invalidTokenErr *authsvc.ErrInvalidTokenType

			if errors.As(err, &invalidPasswordErr) {
				httpserver.RenderError(w, http.StatusBadRequest, err)
				return
			}

			if errors.As(err, &invalidTokenErr) {
				httpserver.RenderError(w, http.StatusBadRequest, err)
				return
			}

			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("cannot reset password: %w", err))
			return
		}

		httpserver.RenderJSON(w, http.StatusOK, ResetPasswordResponse{
			Success: true,
		})
	}
}
