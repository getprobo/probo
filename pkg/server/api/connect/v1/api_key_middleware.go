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

package connect_v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securetoken"
)

var (
	apiKeyContextKey = &ctxKey{name: "api_key"}
)

func APIKeyFromContext(ctx context.Context) *coredata.UserAPIKey {
	apiKey, _ := ctx.Value(apiKeyContextKey).(*coredata.UserAPIKey)
	return apiKey
}

func NewAPIKeyMiddleware(svc *iam.Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				tokenValue, err := securetoken.Get(r, "")
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}

				keyID, err := gid.ParseGID(tokenValue)
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}

				session := SessionFromContext(ctx)
				if keyID != gid.Nil && session != nil {
					httpserver.RenderError(w, http.StatusBadRequest, errors.New("api key authentication cannot be used with session authentication"))
					return
				}

				apiKey, err := svc.APIKeyService.GetAPIKey(ctx, keyID)
				if err != nil {
					var errUserAPIKeyNotFound *iam.ErrUserAPIKeyNotFound
					var errUserAPIKeyExpired *iam.ErrUserAPIKeyExpired

					if errors.As(err, &errUserAPIKeyNotFound) || errors.As(err, &errUserAPIKeyExpired) {
						next.ServeHTTP(w, r)
						return
					}

					panic(fmt.Errorf("cannot get user API key: %w", err))
				}

				user, err := svc.AccountService.GetIdentity(ctx, apiKey.UserID)
				if err != nil {
					var errUserNotFound *iam.ErrUserNotFound
					if errors.As(err, &errUserNotFound) {
						next.ServeHTTP(w, r)
						return
					}

					panic(fmt.Errorf("cannot get user: %w", err))
				}

				ctx = context.WithValue(ctx, apiKeyContextKey, apiKey)
				ctx = context.WithValue(ctx, identityContextKey, user)

				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
