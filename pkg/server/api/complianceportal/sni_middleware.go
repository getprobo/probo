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

// portalHost is the host a compliance portal request is served under.
type portalHost struct {
	// serverName is the hostname the compliance page is resolved by: the TLS
	// SNI, or the Host header when TLS is terminated upstream.
	serverName string

	// current is the host the request came in on, compared against the
	// canonical host to decide whether the request must be redirected.
	current string
}

// NewSNIMiddleware resolves the compliance page served under the request
// hostname. The hostname normally comes from the TLS SNI of the portal's own
// HTTPS listener; when tlsTerminatedByProxy is set, public TLS is terminated by
// an upstream proxy and the plain-HTTP request carries it in the Host header
// instead.
func NewSNIMiddleware(visitorSvc *visitor.Service, tlsTerminatedByProxy bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, ok := portalHostFromRequest(r, tlsTerminatedByProxy)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			servePortal(w, r, next, visitorSvc, host)
		})
	}
}

// portalHostFromRequest returns the host a compliance page is served under, and
// whether the request carries one at all. A request that reaches the portal in
// clear text only does when an upstream proxy terminates public TLS.
func portalHostFromRequest(r *http.Request, tlsTerminatedByProxy bool) (portalHost, bool) {
	if r.TLS != nil {
		return portalHost{serverName: r.TLS.ServerName, current: r.Host}, true
	}

	if !tlsTerminatedByProxy {
		return portalHost{}, false
	}

	host := hostWithoutPort(r.Host)

	return portalHost{serverName: host, current: host}, true
}

// servePortal resolves the request host to a compliance page and hands the
// request over to next with the page in its context.
func servePortal(
	w http.ResponseWriter,
	r *http.Request,
	next http.Handler,
	visitorSvc *visitor.Service,
	host portalHost,
) {
	ctx := r.Context()

	compliancePage, err := visitorSvc.GetPortalByDomainName(ctx, host.serverName)
	if err != nil {
		if errors.Is(err, visitor.ErrPageNotFound) {
			next.ServeHTTP(w, r)
			return
		}

		httpserver.LoggerFromContext(ctx).ErrorCtx(
			ctx,
			"cannot get compliance portal by domain name",
			log.Error(err),
			log.String("server_name", host.serverName),
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

		if canonicalHost != "" && canonicalHost != host.current {
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
	//
	// The scheme is always https: either this listener terminates TLS itself,
	// or an upstream proxy terminates it and forwards in clear text.
	baseURL := (&url.URL{
		Host:   host.current,
		Scheme: "https",
	}).String()
	ctx = ContextWithCompliancePortalBaseURL(ctx, baseURL)
	r = r.WithContext(ctx)

	if compliancePage.Active {
		ctx = ContextWithCompliancePortal(ctx, compliancePage)
		next.ServeHTTP(w, r.WithContext(ctx))

		return
	}

	next.ServeHTTP(w, r)
}

// hostWithoutPort strips the port from a Host header value. Unlike net.SplitHostPort
// it leaves a value that carries no port untouched.
func hostWithoutPort(host string) string {
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}

	return host
}
