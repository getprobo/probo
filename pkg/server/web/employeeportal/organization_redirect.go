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

package employeeportal

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type OldestPortalLookup func(ctx context.Context, organizationID gid.GID) (gid.GID, error)

// OrganizationGIDRedirectMiddleware 302s /{organizationId}/… to the
// organization's oldest employee portal. The handler is mounted after
// StripPrefix(PathPrefix), so r.URL.Path is the SPA-relative path.
func OrganizationGIDRedirectMiddleware(lookup OldestPortalLookup, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if lookup == nil {
			next.ServeHTTP(w, r)
			return
		}

		trimmed := strings.TrimPrefix(r.URL.Path, "/")
		if trimmed == "" {
			next.ServeHTTP(w, r)
			return
		}

		first, rest, _ := strings.Cut(trimmed, "/")
		parsed, err := gid.ParseGID(first)
		if err != nil || parsed.EntityType() != coredata.OrganizationEntityType {
			next.ServeHTTP(w, r)
			return
		}

		portalID, err := lookup(r.Context(), parsed)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				http.NotFound(w, r)
				return
			}

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		location, err := OrganizationPath(portalID.String(), splitPath(rest)...)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		redirectURL := url.URL{Path: location, RawQuery: r.URL.RawQuery}
		http.Redirect(w, r, redirectURL.String(), http.StatusFound)
	})
}

func splitPath(rest string) []string {
	if rest == "" {
		return nil
	}

	return strings.Split(rest, "/")
}
