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
	"net/http"

	"go.gearno.de/kit/httpserver"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type DeleteUserAPIKeyRequest struct {
	ID string `json:"id"`
}

type DeleteUserAPIKeyResponse struct {
	ID string `json:"id"`
}

func DeleteUserAPIKeyHandler(authSvc *authsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := UserFromContext(ctx)

		var req DeleteUserAPIKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
			return
		}

		if req.ID == "" {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "user api key id is required",
			})
			return
		}

		userAPIKeyID, err := gid.ParseGID(req.ID)
		if err != nil {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid user api key id",
			})
			return
		}

		if err := authSvc.DeleteUserAPIKey(ctx, userAPIKeyID, user.ID); err != nil {
			var errNotFound *coredata.ErrUserAPIKeyNotFound
			if errors.As(err, &errNotFound) {
				httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{
					"error": "user api key not found",
				})
				return
			}
			httpserver.RenderJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to delete user api key",
			})
			return
		}

		response := DeleteUserAPIKeyResponse{
			ID: req.ID,
		}

		httpserver.RenderJSON(w, http.StatusOK, response)
	}
}
