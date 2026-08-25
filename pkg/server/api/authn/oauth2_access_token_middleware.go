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
	"fmt"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oauth2"
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

				authorization := r.Header.Get("Authorization")

				tokenValue, err := bearertoken.Parse(authorization)
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
				ctx = oauth2.ContextWithAccessToken(ctx, accessToken)

				httpserver.LoggerFromContext(ctx).InfoCtx(
					ctx,
					"access token authenticated",
					log.String("identity_id", identity.ID.String()),
					log.String("access_token_id", accessToken.ID.String()),
				)

				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
