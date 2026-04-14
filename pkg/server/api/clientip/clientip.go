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

package clientip

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type ctxKey struct{}

// NewMiddleware returns an HTTP middleware that extracts the client IP
// from standard proxy headers and stores it in the request context.
func NewMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := Extract(r)
			ctx := context.WithValue(r.Context(), ctxKey{}, ip)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// FromContext returns the client IP stored by the middleware, or an
// empty string if the middleware has not run.
func FromContext(ctx context.Context) string {
	if ip, ok := ctx.Value(ctxKey{}).(string); ok {
		return ip
	}
	return ""
}

// Extract resolves the client IP address from standard proxy headers
// in priority order: RFC 7239 Forwarded, then X-Forwarded-For, then
// the connection's remote address.
func Extract(r *http.Request) string {
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		if ip := parseForwardedFor(fwd); ip != "" {
			return ip
		}
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i != -1 {
			xff = xff[:i]
		}
		xff = strings.TrimSpace(xff)

		if ip, _, err := net.SplitHostPort(xff); err == nil {
			return ip
		}
		return xff
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

// parseForwardedFor extracts the client IP from the first "for=" directive
// of an RFC 7239 Forwarded header value.
func parseForwardedFor(header string) string {
	if i := strings.IndexByte(header, ','); i != -1 {
		header = header[:i]
	}

	for _, part := range strings.Split(header, ";") {
		part = strings.TrimSpace(part)
		if !strings.HasPrefix(strings.ToLower(part), "for=") {
			continue
		}

		val := part[4:]
		val = strings.Trim(val, "\"")

		if strings.HasPrefix(val, "[") {
			if end := strings.IndexByte(val, ']'); end != -1 {
				return val[1:end]
			}
		}

		if ip, _, err := net.SplitHostPort(val); err == nil {
			return ip
		}
		return val
	}

	return ""
}
