// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package authn

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
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
					if _, ok := errors.AsType[*iam.ErrPersonalAPIKeyNotFound](err); ok {
						next.ServeHTTP(w, r)
						return
					}

					if _, ok := errors.AsType[*iam.ErrPersonalAPIKeyExpired](err); ok {
						next.ServeHTTP(w, r)
						return
					}

					panic(fmt.Errorf("cannot get personal API key: %w", err))
				}

				identity, err := svc.AccountService.GetIdentity(ctx, apiKey.IdentityID)
				if err != nil {
					if _, ok := errors.AsType[*iam.ErrIdentityNotFound](err); ok {
						next.ServeHTTP(w, r)
						return
					}

					panic(fmt.Errorf("cannot get identity: %w", err))
				}

				ctx = ContextWithAPIKey(ctx, apiKey)
				ctx = ContextWithIdentity(ctx, identity)

				httpserver.LoggerFromContext(ctx).InfoCtx(
					ctx,
					"api key authenticated",
					log.String("identity_id", identity.ID.String()),
					log.String("api_key_id", apiKey.ID.String()),
				)

				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
