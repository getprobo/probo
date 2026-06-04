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

package connector

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

const OktaProvider = "OKTA"

// oktaDomainRe matches a dotted DNS hostname (at least two labels; each
// label 1-63 chars of [a-z0-9-], not starting or ending with a hyphen).
// Okta supports both *.okta.com / *.oktapreview.com orgs and fully custom
// domains, so the suffix is intentionally unrestricted — the host shape and
// the IP-literal rejection below, plus the transport's SSRF protection, are
// the guards.
var oktaDomainRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)+$`)

// NormalizeOktaDomain extracts and validates the bare Okta org host from
// operator input. It accepts either a bare host ("acme.okta.com") or a full
// URL ("https://acme.okta.com/"), strips any scheme/path, lowercases, and
// rejects explicit ports, IP literals, and malformed hostnames. The returned
// host is what the driver and name resolver interpolate into the per-org API
// host (https://<host>/api/v1/...), so it is the single SSRF-sensitive input
// and must be validated here on the write path.
func NormalizeOktaDomain(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", fmt.Errorf("cannot normalize okta domain: empty")
	}

	// url.Parse needs a scheme to populate Host; add a placeholder for bare
	// hosts. The scheme itself is discarded — only the hostname is kept.
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", fmt.Errorf("cannot normalize okta domain: invalid")
	}

	if u.Port() != "" {
		return "", fmt.Errorf("cannot normalize okta domain: ports are not allowed")
	}

	host := strings.ToLower(u.Hostname())
	if !IsValidOktaDomain(host) {
		return "", fmt.Errorf("cannot normalize okta domain: invalid host")
	}

	return host, nil
}

// IsValidOktaDomain reports whether host is a syntactically valid Okta org
// domain (a dotted DNS hostname, not an IP literal). It re-validates the
// stored domain at driver/name-resolver construction time as defense in
// depth, regardless of how the connector row was populated.
func IsValidOktaDomain(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}

	// Reject IP literals: an Okta org is always a DNS name, and an IP host
	// would sidestep the hostname shape check below.
	if net.ParseIP(host) != nil {
		return false
	}

	return oktaDomainRe.MatchString(host)
}
