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

package trustedproxy

import (
	"net"
	"net/http"
)

var forwardedHeaders = []string{
	"Forwarded",
	"X-Forwarded-For",
}

// NewMiddleware returns an HTTP middleware that strips forwarded
// headers from requests that did not originate from one of the given
// trusted proxy IPs.  When the list is empty every request is treated
// as untrusted and the headers are always removed.
func NewMiddleware(trusted []net.IP) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isTrusted(r.RemoteAddr, trusted) {
				for _, h := range forwardedHeaders {
					r.Header.Del(h)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isTrusted(remoteAddr string, trusted []net.IP) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	for _, t := range trusted {
		if t.Equal(ip) {
			return true
		}
	}

	return false
}
