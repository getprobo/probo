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

package scanner

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/netcheck"
)

// origin is a scheme, host and port. The host has no leading "www." and
// default ports are dropped, matching CanonicalizeOrigin plus port folding.
// Userinfo is rejected.
type origin struct {
	scheme string
	host   string
	port   string
}

func parseOrigin(raw string) (origin, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil {
		return origin{}, false
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return origin{}, false
	}

	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	if host == "" {
		return origin{}, false
	}

	return origin{
		scheme: u.Scheme,
		host:   host,
		port:   normalizedPort(u.Scheme, u.Port()),
	}, true
}

func normalizedPort(scheme, port string) string {
	if (scheme == "https" && port == "443") || (scheme == "http" && port == "80") {
		return ""
	}

	return port
}

func (o origin) hostPort() string {
	if o.port == "" {
		return o.host
	}

	return net.JoinHostPort(o.host, o.port)
}

// String is the origin without a path, such as "https://example.com".
func (o origin) String() string {
	return o.scheme + "://" + o.hostPort()
}

func (o origin) root() string {
	return (&url.URL{Scheme: o.scheme, Host: o.hostPort(), Path: "/"}).String()
}

func (o origin) wwwRoot() string {
	host := "www." + o.host
	if o.port != "" {
		host = net.JoinHostPort(host, o.port)
	}

	return (&url.URL{Scheme: o.scheme, Host: host, Path: "/"}).String()
}

func (o origin) pathURL(path string) string {
	return (&url.URL{Scheme: o.scheme, Host: o.hostPort(), Path: path}).String()
}

// contains reports whether rawURL has this origin. "www." and default ports
// do not make a URL a different origin.
func (o origin) contains(rawURL string) bool {
	other, ok := parseOrigin(rawURL)
	if !ok {
		return false
	}

	return o == other
}

// chooseStart picks the URL of the first page. When the apex host does not
// resolve, the www host is used instead. A private apex is not replaced.
func chooseStart(site origin, apex, www hostVerdict) (string, error) {
	switch apex {
	case hostPublic:
		return site.root(), nil
	case hostUnknown:
		if www == hostPublic {
			return site.wwwRoot(), nil
		}

		return "", fmt.Errorf("cannot start scan: cannot resolve host %q", site.host)
	default:
		return "", fmt.Errorf("cannot start scan: %w: host %q", netcheck.ErrBlockedAddress, site.host)
	}
}
