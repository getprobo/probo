// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"net/url"
)

const ZendeskProvider = "ZENDESK"

// IsValidZendeskSubdomain reports whether s is a single DNS label safe to use
// as the host component of <subdomain>.zendesk.com. The subdomain is
// customer-supplied and feeds a URL host on every authorize, token, and API
// request, so it MUST be validated before use to close an SSRF vector: only
// ASCII letters, digits, and interior hyphens are allowed (no dots, slashes,
// colons, '@', or whitespace that could escape the host position), bounded to
// a 63-character DNS label. DNS is case-insensitive, so mixed case is accepted
// and used verbatim.
func IsValidZendeskSubdomain(s string) bool {
	if len(s) == 0 || len(s) > 63 {
		return false
	}

	// i is a byte offset; because every accepted character is single-byte
	// ASCII, i equals the character position, so the i < len(s)-1 bound below
	// correctly identifies the final character. A non-ASCII rune falls through
	// to the default case and is rejected before that assumption matters.
	for i, c := range s {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '-' && i > 0 && i < len(s)-1:
		default:
			return false
		}
	}

	return true
}

// ZendeskAuthorizeURL returns the OAuth2 authorize endpoint for a Zendesk
// customer subdomain (e.g. "acme"). It errors on any subdomain that is not a
// valid single DNS label.
func ZendeskAuthorizeURL(subdomain string) (string, error) {
	if !IsValidZendeskSubdomain(subdomain) {
		return "", fmt.Errorf("cannot build authorize URL: invalid zendesk subdomain")
	}

	u := url.URL{Scheme: "https", Host: subdomain + ".zendesk.com", Path: "/oauth/authorizations/new"}

	return u.String(), nil
}

// ZendeskTokenURL returns the OAuth2 token endpoint for a Zendesk customer
// subdomain (e.g. "acme"). It errors on any subdomain that is not a valid
// single DNS label.
func ZendeskTokenURL(subdomain string) (string, error) {
	if !IsValidZendeskSubdomain(subdomain) {
		return "", fmt.Errorf("cannot build token URL: invalid zendesk subdomain")
	}

	u := url.URL{Scheme: "https", Host: subdomain + ".zendesk.com", Path: "/oauth/tokens"}

	return u.String(), nil
}
