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

package connector_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
)

func TestZendeskAuthorizeURL(t *testing.T) {
	t.Parallel()

	got, err := connector.ZendeskAuthorizeURL("acme")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.zendesk.com/oauth/authorizations/new", got)

	// An invalid subdomain must error rather than build a host.
	_, err = connector.ZendeskAuthorizeURL("evil.example")
	require.Error(t, err)
}

func TestZendeskTokenURL(t *testing.T) {
	t.Parallel()

	got, err := connector.ZendeskTokenURL("acme")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.zendesk.com/oauth/tokens", got)

	_, err = connector.ZendeskTokenURL("acme/../evil")
	require.Error(t, err)
}

// TestIsValidZendeskSubdomain exercises the SSRF guard that gates every Zendesk
// URL host. Anything that could escape the host position of
// <subdomain>.zendesk.com must be rejected.
func TestIsValidZendeskSubdomain(t *testing.T) {
	t.Parallel()

	for _, s := range []string{
		"acme",
		"my-company",
		"a",                     // single label
		"ABC123",                // DNS is case-insensitive
		"a--b",                  // interior double hyphen is allowed
		strings.Repeat("a", 63), // max DNS label length
	} {
		assert.True(t, connector.IsValidZendeskSubdomain(s), s)
	}

	for _, s := range []string{
		"",                      // empty
		strings.Repeat("a", 64), // one over the 63-char label cap
		"-acme",                 // leading hyphen
		"acme-",                 // trailing hyphen
		"acme.evil",             // dot — would add a host segment
		"acme/evil",             // slash — path escape
		"acme:1234",             // colon — port/authority escape
		"acme@evil",             // userinfo escape
		"acme evil",             // whitespace
		"acme_evil",             // underscore is not a DNS label char
		"acmé",                  // non-ASCII
	} {
		assert.False(t, connector.IsValidZendeskSubdomain(s), s)
	}
}
