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

	"github.com/getprobo/probo/pkg/managederror"
	"github.com/getprobo/probo/pkg/securecookie"
	"github.com/getprobo/probo/pkg/usrmgr"
	"go.gearno.de/kit/httpserver"
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

func SignUpHandler(usrmgrSvc *usrmgr.Service, authCfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot decode body: %w", err))
			return
		}

		user, session, err := usrmgrSvc.SignUp(
			r.Context(),
			req.Email,
			req.Password,
			req.FullName,
		)
		if err != nil {
			var errManaged *managederror.ErrorManaged
			if errors.As(err, &errManaged) && errManaged.Code == managederror.CodeAlreadyExists {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot register user: %w", err))
				return
			}

			if errors.As(err, &errManaged) && errManaged.Code == managederror.CodeOperationNotAllowed {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot register user: %w", err))
				return
			}

			panic(fmt.Errorf("cannot register user: %w", err))
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
