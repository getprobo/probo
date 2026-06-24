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
