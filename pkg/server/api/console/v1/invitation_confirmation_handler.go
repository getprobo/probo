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

	"github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"go.gearno.de/kit/httpserver"
)

type (
	InvitationConfirmationRequest struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	InvitationConfirmationResponse struct {
	}
)

func InvitationConfirmationHandler(authSvc *auth.Service, authzSvc *authz.Service, authCfg AuthConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req InvitationConfirmationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot decode body: %w", err))
			return
		}

		payload, err := statelesstoken.ValidateToken[coredata.InvitationData](
			authCfg.CookieSecret,
			authz.TokenTypeOrganizationInvitation,
			req.Token,
		)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid invitation token: %w", err))
			return
		}

		user, _, err := authSvc.SignUp(r.Context(), payload.Data.Email, req.Password, payload.Data.FullName)
		if err != nil {
			var errUserAlreadyExists *auth.ErrUserAlreadyExists
			if errors.As(err, &errUserAlreadyExists) {
				user, err = authSvc.GetUserByEmail(r.Context(), payload.Data.Email)
				if err != nil {
					httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("failed to load existing user: %w", err))
					return
				}
			} else {
				httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("failed to create user: %w", err))
				return
			}
		}

		err = authzSvc.AcceptInvitation(r.Context(), req.Token, user.ID)
		if err != nil {
			httpserver.RenderError(w, http.StatusInternalServerError, err)
			return
		}

		httpserver.RenderJSON(w, http.StatusOK, InvitationConfirmationResponse{})
	}
}
