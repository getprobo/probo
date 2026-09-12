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
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
)

func NewServiceAccountMiddleware(svc *iam.Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				if PrincipalIDFromContext(ctx) != gid.Nil {
					next.ServeHTTP(w, r)
					return
				}

				token, err := bearertoken.Parse(r.Header.Get("Authorization"))
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}

				account, credential, err := svc.ServiceAccounts.Authenticate(ctx, token)
				if err != nil {
					next.ServeHTTP(w, r)
					return
				}

				ctx = ContextWithServiceAccount(ctx, account)
				ctx = iam.ContextWithServiceAccountCredential(ctx, credential)

				httpserver.LoggerFromContext(ctx).InfoCtx(
					ctx,
					"service account authenticated",
					log.String("service_account_id", account.ID.String()),
					log.String("credential_id", credential.ID.String()),
				)

				next.ServeHTTP(w, r.WithContext(ctx))
			},
		)
	}
}
