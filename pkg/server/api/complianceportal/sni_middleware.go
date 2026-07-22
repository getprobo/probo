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

package complianceportal

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func NewSNIMiddleware(visitorSvc *visitor.Service) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			if r.TLS == nil {
				next.ServeHTTP(w, r)
				return
			}

			compliancePage, err := visitorSvc.GetPortalByDomainName(ctx, r.TLS.ServerName)
			if err != nil {
				if errors.Is(err, visitor.ErrPageNotFound) {
					next.ServeHTTP(w, r)
					return
				}

				httpserver.LoggerFromContext(ctx).ErrorCtx(
					ctx,
					"cannot get compliance portal by domain name",
					log.Error(err),
					log.String("server_name", r.TLS.ServerName),
				)

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

			// Redirect secondary domains to the canonical host so a compliance
			// page is only ever served under a single origin. ACME HTTP-01
			// challenges are handled upstream and never reach this middleware.
			if !strings.HasPrefix(r.URL.Path, "/.well-known/") {
				canonicalHost, err := visitorSvc.GetPortalEffectiveCanonicalHost(ctx, compliancePage.ID)
				if err != nil {
					httpserver.LoggerFromContext(ctx).ErrorCtx(
						ctx,
						"cannot get compliance portal canonical host",
						log.Error(err),
						log.String("compliance_portal_id", compliancePage.ID.String()),
					)

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

				if canonicalHost != "" && canonicalHost != r.Host {
					target := &url.URL{
						Scheme:   "https",
						Host:     canonicalHost,
						Path:     r.URL.Path,
						RawQuery: r.URL.RawQuery,
					}

					http.Redirect(w, r, target.String(), http.StatusPermanentRedirect)

					return
				}
			}

			// Origin only — consumers append their own paths (SEO, sitemap,
			// robots, brand assets, OAuth). Including r.URL.Path here would
			// duplicate the route (e.g. /fr/documents/fr/documents).
			baseURL := &url.URL{
				Host:   r.Host,
				Scheme: "https",
			}
			baseURLString := baseURL.String()

			ctx = context.WithValue(
				ctx,
				compliancePortalBaseURLKey,
				&baseURLString,
			)
			r = r.WithContext(ctx)

			if compliancePage.Active {
				ctx = context.WithValue(ctx, compliancePortalKey, compliancePage)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
