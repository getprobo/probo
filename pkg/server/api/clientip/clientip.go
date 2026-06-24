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

package clientip

import (
	"net"
	"net/http"
	"strings"
)

// Extract resolves the client IP address from standard proxy headers
// in priority order: RFC 7239 Forwarded, then X-Forwarded-For, then
// the connection's remote address. It takes the rightmost (last)
// entry from multi-value headers — the one appended by the trusted
// load balancer closest to us.
func Extract(r *http.Request) string {
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		if ip := parseForwardedFor(fwd); ip != "" {
			return ip
		}
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.LastIndexByte(xff, ','); i != -1 {
			xff = xff[i+1:]
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

// parseForwardedFor extracts the client IP from the last "for=" directive
// of an RFC 7239 Forwarded header value.
func parseForwardedFor(header string) string {
	if i := strings.LastIndexByte(header, ','); i != -1 {
		header = header[i+1:]
	}

	for part := range strings.SplitSeq(header, ";") {
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
