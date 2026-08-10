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

package server

import "net/http"

const (
	strictTransportSecurityValue = "max-age=31536000; includeSubDomains; preload"
	permissionsPolicyValue       = "microphone=(), camera=(), geolocation=()"
)

type SecurityHeadersOptions struct {
	ExtraHeaderFields     map[string]string
	ContentSecurityPolicy string
}

func ApplyExtraHeaders(w http.ResponseWriter, extraHeaderFields map[string]string) {
	for key, value := range extraHeaderFields {
		w.Header().Set(key, value)
	}
}

func NewSecurityHeadersMiddleware(opts SecurityHeadersOptions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				// Extras first so generated security headers cannot be
				// weakened or removed by a colliding extra-header-fields key.
				ApplyExtraHeaders(w, opts.ExtraHeaderFields)

				w.Header().Set("Strict-Transport-Security", strictTransportSecurityValue)

				if opts.ContentSecurityPolicy != "" {
					w.Header().Set("Content-Security-Policy", opts.ContentSecurityPolicy)
					w.Header().Set("X-Frame-Options", "DENY")
					w.Header().Set("X-Content-Type-Options", "nosniff")
					w.Header().Set("Referrer-Policy", "no-referrer")
					w.Header().Set("Permissions-Policy", permissionsPolicyValue)
				}

				next.ServeHTTP(w, r)
			},
		)
	}
}
