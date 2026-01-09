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

package authn

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securetoken"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func NewAPIKeyMiddleware(svc *iam.Service, tokenSecret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				tokenValue, err := securetoken.Get(r, tokenSecret)
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
					httpserver.RenderJSON(
						w,
						http.StatusUnauthorized,
						&graphql.Response{
							Errors: gqlerror.List{
								gqlutils.Conflictf(ctx, "API key authentication cannot be used with session authentication"),
							},
						},
					)
					return
				}

				apiKey, err := svc.APIKeyService.GetAPIKey(ctx, keyID)
				if err != nil {
					var errPersonalAPIKeyNotFound *iam.ErrPersonalAPIKeyNotFound
					var errPersonalAPIKeyExpired *iam.ErrPersonalAPIKeyExpired

					if errors.As(err, &errPersonalAPIKeyNotFound) || errors.As(err, &errPersonalAPIKeyExpired) {
						next.ServeHTTP(w, r)
						return
					}

					panic(fmt.Errorf("cannot get personal API key: %w", err))
				}

				identity, err := svc.AccountService.GetIdentity(ctx, apiKey.IdentityID)
				if err != nil {
					var errIdentityNotFound *iam.ErrIdentityNotFound
					if errors.As(err, &errIdentityNotFound) {
						next.ServeHTTP(w, r)
						return
					}

					panic(fmt.Errorf("cannot get identity: %w", err))
				}

				ctx = ContextWithAPIKey(ctx, apiKey)
				ctx = ContextWithIdentity(ctx, identity)

				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
