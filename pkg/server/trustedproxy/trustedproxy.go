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
