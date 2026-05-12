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
	"fmt"
	"net"
	"net/http"
	"strings"
)

var forwardedHeaders = []string{
	"Forwarded",
	"X-Forwarded-For",
}

// NewMiddleware returns an HTTP middleware that strips forwarded
// headers from requests that did not originate from one of the given
// trusted proxies.  Each entry in trusted may be either a single IP
// address (e.g. "10.0.0.1") or a CIDR range (e.g. "10.0.0.0/24").
// When the list is empty every request is treated as untrusted and
// the headers are always removed.  An error is returned if any entry
// is neither a valid IP nor a valid CIDR.
func NewMiddleware(trusted []string) (func(http.Handler) http.Handler, error) {
	ips, nets, err := parseTrusted(trusted)
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isTrusted(r.RemoteAddr, ips, nets) {
				for _, h := range forwardedHeaders {
					r.Header.Del(h)
				}
			}
			next.ServeHTTP(w, r)
		})
	}, nil
}

func parseTrusted(trusted []string) ([]net.IP, []*net.IPNet, error) {
	ips := make([]net.IP, 0, len(trusted))
	nets := make([]*net.IPNet, 0, len(trusted))
	for _, entry := range trusted {
		if strings.Contains(entry, "/") {
			_, ipNet, err := net.ParseCIDR(entry)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot parse CIDR %q: %w", entry, err)
			}
			nets = append(nets, ipNet)
			continue
		}

		ip := net.ParseIP(entry)
		if ip == nil {
			return nil, nil, fmt.Errorf("cannot parse IP address %q", entry)
		}
		ips = append(ips, ip)
	}
	return ips, nets, nil
}

func isTrusted(remoteAddr string, ips []net.IP, nets []*net.IPNet) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	for _, t := range ips {
		if t.Equal(ip) {
			return true
		}
	}

	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}

	return false
}
