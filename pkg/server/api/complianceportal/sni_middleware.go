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
	"errors"
	"net"
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

func NewDomainMiddleware(
	visitorSvc *visitor.Service,
	externallyTerminatedTLS bool,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			host, requestProto, ok := compliancePortalRequest(r, externallyTerminatedTLS)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			compliancePage, err := visitorSvc.GetPortalByDomainName(ctx, host)
			if err != nil {
				if errors.Is(err, visitor.ErrPageNotFound) {
					next.ServeHTTP(w, r)
					return
				}

				httpserver.LoggerFromContext(ctx).ErrorCtx(
					ctx,
					"cannot get compliance portal by domain name",
					log.Error(err),
					log.String("host", host),
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

			if requestProto == "http" {
				target := &url.URL{
					Scheme:   "https",
					Host:     host,
					Path:     r.URL.Path,
					RawQuery: r.URL.RawQuery,
				}

				http.Redirect(w, r, target.String(), http.StatusMovedPermanently)

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

				canonicalHost, ok := normalizeCompliancePortalHost(canonicalHost)
				if canonicalHost != "" && ok && canonicalHost != host {
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
			baseURL := (&url.URL{
				Host:   host,
				Scheme: "https",
			}).String()
			ctx = ContextWithCompliancePortalBaseURL(ctx, baseURL)
			if externallyTerminatedTLS && r.TLS == nil {
				r.Host = host
			}
			r = r.WithContext(ctx)

			if compliancePage.Active {
				ctx = ContextWithCompliancePortal(ctx, compliancePage)
				next.ServeHTTP(w, r.WithContext(ctx))

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func compliancePortalRequest(r *http.Request, externallyTerminatedTLS bool) (string, string, bool) {
	if r.TLS != nil {
		host, ok := normalizeCompliancePortalHost(r.TLS.ServerName)
		if !ok {
			return "", "", false
		}

		return host, "https", true
	}

	if !externallyTerminatedTLS {
		return "", "", false
	}

	requestProto, ok := singleHeaderValue(r.Header, "X-Forwarded-Proto")
	if !ok {
		return "", "", false
	}

	requestProto = strings.ToLower(requestProto)
	if requestProto != "http" && requestProto != "https" {
		return "", "", false
	}

	// External mode establishes the reverse proxy as the TLS and host trust
	// boundary. Deployments must prevent direct access to this listener and use
	// a proxy that overwrites these forwarded headers rather than appending to
	// client-supplied values.
	host := r.Host
	if len(r.Header.Values("X-Forwarded-Host")) > 0 {
		forwardedHost, ok := singleHeaderValue(r.Header, "X-Forwarded-Host")
		if !ok || forwardedHost == "" {
			return "", "", false
		}

		host = forwardedHost
	}

	host, ok = normalizeCompliancePortalHost(host)
	if !ok {
		return "", "", false
	}

	return host, requestProto, true
}

func singleHeaderValue(header http.Header, name string) (string, bool) {
	values := header.Values(name)
	if len(values) != 1 || strings.Contains(values[0], ",") {
		return "", false
	}

	return strings.TrimSpace(values[0]), true
}

func normalizeCompliancePortalHost(value string) (string, bool) {
	host := strings.TrimSpace(value)
	if host == "" || strings.Contains(host, ",") {
		return "", false
	}

	if strings.Contains(host, ":") {
		var err error

		host, _, err = net.SplitHostPort(host)
		if err != nil {
			return "", false
		}
	}

	host = strings.ToLower(strings.TrimSuffix(host, "."))

	return host, host != ""
}
