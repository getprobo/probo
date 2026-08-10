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

package baseurl

import (
	"fmt"
	"net/url"
	"strings"
)

// CSPOrigin returns a canonical http(s)://host[:port] origin safe to embed as
// a single Content-Security-Policy source. Go's url.Parse (and Parse) accept
// hosts containing ';' which would terminate a CSP directive and inject
// another (e.g. a weaker frame-ancestors). Empty raw is allowed and returned
// unchanged for callers that omit the origin.
func CSPOrigin(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid CSP origin: %w", err)
	}

	return cspOriginFromURL(parsed)
}

// CSPOrigin returns the BaseURL's scheme://host origin in a form safe for CSP.
func (b *BaseURL) CSPOrigin() (string, error) {
	if b == nil || b.parsed == nil {
		return "", fmt.Errorf("base URL is nil")
	}

	return cspOriginFromURL(b.parsed)
}

func cspOriginFromURL(u *url.URL) (string, error) {
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("CSP origin scheme must be http or https")
	}

	if u.User != nil {
		return "", fmt.Errorf("CSP origin must not include userinfo")
	}

	if u.Opaque != "" {
		return "", fmt.Errorf("CSP origin must not be opaque")
	}

	if u.Path != "" && u.Path != "/" {
		return "", fmt.Errorf("CSP origin must not include a path")
	}

	if u.RawQuery != "" || u.Fragment != "" {
		return "", fmt.Errorf("CSP origin must not include query or fragment")
	}

	host := u.Host
	if host == "" {
		return "", fmt.Errorf("CSP origin must include a host")
	}

	// CSP sources are space-separated; ';' starts a new directive. Quotes and
	// wildcards must not appear in an application origin either.
	if strings.ContainsAny(host, " \t\r\n;,\"'\\*") {
		return "", fmt.Errorf("CSP origin host contains invalid characters")
	}

	if strings.HasSuffix(host, ":") {
		return "", fmt.Errorf("CSP origin host has an empty port")
	}

	hostname := u.Hostname()
	if hostname == "" || hostname == "." {
		return "", fmt.Errorf("CSP origin hostname is invalid")
	}

	if port := u.Port(); port != "" {
		for _, c := range port {
			if c < '0' || c > '9' {
				return "", fmt.Errorf("CSP origin port is invalid")
			}
		}
	}

	return u.Scheme + "://" + host, nil
}
