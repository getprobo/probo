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

package compliancepage

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/trust"
)

func NewSNIMiddleware(trustSvc *trust.Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if r.TLS == nil {
				next.ServeHTTP(w, r)
				return
			}

			compliancePage, err := trustSvc.GetByDomainName(ctx, r.TLS.ServerName)
			if err != nil {
				if errors.Is(err, trust.ErrPageNotFound) {
					next.ServeHTTP(w, r)
					return
				}

				httpserver.RenderJSON(
					w,
					http.StatusInternalServerError,
					&graphql.Response{
						Errors: gqlerror.List{
							gqlutils.Internal(ctx),
						},
					},
				)

				return
			}

			baseURL := &url.URL{
				Host:   r.Host,
				Path:   r.URL.Path,
				Scheme: "https",
			}

			ctx = context.WithValue(
				ctx,
				compliancePageBaseURLKey,
				new(baseURL.String()),
			)
			r = r.WithContext(ctx)

			if compliancePage.Active {
				ctx = context.WithValue(ctx, compliancePageKey, compliancePage)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
