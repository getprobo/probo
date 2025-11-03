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
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/gid"
)

type GetUserAPIKeyResponse struct {
	Key string `json:"key"`
}

func GetUserAPIKeyHandler(authSvc *authsvc.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := UserFromContext(ctx)

		userAPIKeyIDStr := chi.URLParam(r, "id")
		if userAPIKeyIDStr == "" {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "user api key id is required",
			})
			return
		}

		userAPIKeyID, err := gid.ParseGID(userAPIKeyIDStr)
		if err != nil {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid user api key id",
			})
			return
		}

		key, err := authSvc.GetUserAPIKey(ctx, userAPIKeyID, user.ID)
		if err != nil {
			httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{
				"error": "user api key not found",
			})
			return
		}

		response := GetUserAPIKeyResponse{
			Key: key,
		}

		httpserver.RenderJSON(w, http.StatusOK, response)
	}
}
