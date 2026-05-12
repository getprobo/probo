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

package browser

import (
	"go.probo.inc/probo/pkg/agent/tools/internal/netcheck"
)

// validatePublicURL checks that a URL uses an http(s) scheme and that its
// host does not resolve to a private, loopback, or link-local IP address.
// This prevents SSRF attacks where the LLM could be tricked into requesting
// internal network endpoints.
func validatePublicURL(rawURL string) error {
	return netcheck.ValidatePublicURL(rawURL)
}

// validatePublicDomain checks that a domain does not resolve to a private,
// loopback, or link-local IP address.
func validatePublicDomain(domain string) error {
	return netcheck.ValidatePublicDomain(domain)
}
