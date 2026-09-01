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
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// OrganizationPath builds an employee-portal path under PathPrefix for the
// given organization and optional extra segments (for example "signatures").
func OrganizationPath(organizationID string, segments ...string) (string, error) {
	parts := make([]string, 0, 2+len(segments))

	parts = append(parts, PathPrefix, url.PathEscape(organizationID))
	for _, segment := range segments {
		parts = append(parts, url.PathEscape(segment))
	}

	path, err := url.JoinPath(parts[0], parts[1:]...)
	if err != nil {
		return "", fmt.Errorf("cannot build employee portal path: %w", err)
	}

	return path, nil
}

// LegacyRedirectPath maps a console employee request path to the
// employee-portal destination. ok is false when the request should stay on
// console.
func LegacyRedirectPath(requestPath string) (string, bool) {
	path := strings.TrimSuffix(requestPath, "/")

	if path == "/enroll" {
		location, err := url.JoinPath(PathPrefix, "enroll")
		if err != nil {
			return "", false
		}

		return location, true
	}

	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 3 || parts[0] != "organizations" || parts[2] != "employee" {
		return "", false
	}

	organizationID := parts[1]
	rest := parts[3:]

	switch {
	case len(rest) == 0:
		return organizationPathOK(organizationID)
	case len(rest) == 1:
		switch rest[0] {
		case "signatures", "approvals", "devices", "bind", "bindings":
			return organizationPathOK(organizationID, rest[0])
		default:
			return organizationPathOK(organizationID, "signatures", rest[0])
		}
	case len(rest) == 2 && (rest[0] == "signatures" || rest[0] == "approvals"):
		return organizationPathOK(organizationID, rest[0], rest[1])
	default:
		return "", false
	}
}

func organizationPathOK(organizationID string, segments ...string) (string, bool) {
	path, err := OrganizationPath(organizationID, segments...)
	if err != nil {
		return "", false
	}

	return path, true
}

// LegacyRedirectMiddleware issues a 302 to the employee portal for legacy
// console employee paths, then falls through to next.
func LegacyRedirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		location, ok := LegacyRedirectPath(r.URL.Path)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		redirectURL := url.URL{Path: location, RawQuery: r.URL.RawQuery}
		http.Redirect(w, r, redirectURL.String(), http.StatusFound)
	})
}
