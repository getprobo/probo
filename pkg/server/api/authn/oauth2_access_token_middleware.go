// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"fmt"
	"net/http"

	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/iam"
)

func NewOAuth2AccessTokenMiddleware(svc *iam.Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				if IdentityFromContext(ctx) != nil {
					next.ServeHTTP(w, r)
					return
				}

				tokenValue, err := bearertoken.Parse(r.Header.Get("Authorization"))
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}

				accessToken, err := svc.OAuth2ServerService.LoadAccessToken(ctx, tokenValue)
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}

				identity, err := svc.AccountService.GetIdentity(ctx, accessToken.IdentityID)
				if err != nil {
					panic(fmt.Errorf("cannot get identity for oauth2 access token: %w", err))
				}

				ctx = ContextWithIdentity(ctx, identity)

				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
