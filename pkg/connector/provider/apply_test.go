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

package provider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
)

// TestApplyOAuth2Defaults_AuthURLTemplating verifies that operator-supplied
// AuthURLParams (for example Vercel's "{integration_slug}") are substituted
// into the static provider AuthURL when the connector is initialized.
// Providers without placeholders are unaffected.
func TestApplyOAuth2Defaults_AuthURLTemplating(t *testing.T) {
	t.Parallel()

	t.Run("placeholder is substituted when AuthURLParams is supplied", func(t *testing.T) {
		t.Parallel()

		r := provider.NewBuiltinRegistry()
		c := &connector.OAuth2Connector{
			ClientID:     "id",
			ClientSecret: "secret",
			AuthURLParams: map[string]string{
				"integration_slug": "acme",
			},
		}

		// VERCEL uses a templated AuthURL with the
		// "{integration_slug}" placeholder.
		r.ApplyOAuth2Defaults("VERCEL", "https://example.com/cb", c)

		assert.Equal(t, "https://vercel.com/integrations/acme/new", c.AuthURL)
		assert.Equal(t, "https://api.vercel.com/v2/oauth/access_token", c.TokenURL)
	})

	t.Run("placeholder remains literal when AuthURLParams is empty", func(t *testing.T) {
		t.Parallel()

		r := provider.NewBuiltinRegistry()
		c := &connector.OAuth2Connector{
			ClientID:     "id",
			ClientSecret: "secret",
		}

		r.ApplyOAuth2Defaults("VERCEL", "https://example.com/cb", c)

		// No substitution requested; the placeholder is preserved
		// verbatim so a misconfiguration is visible at the
		// authorization step rather than silently masked.
		assert.Equal(t, "https://vercel.com/integrations/{integration_slug}/new", c.AuthURL)
	})
}

// TestApplyOAuth2Defaults_PKCEDefaults asserts that the registered
// PAGERDUTY provider defaults flip RequiresPKCE on so the downstream
// Initiate/Complete flow generates a verifier and replays it.
func TestApplyOAuth2Defaults_PKCEDefaults(t *testing.T) {
	t.Parallel()

	for _, p := range []string{"PAGERDUTY"} {
		t.Run(p, func(t *testing.T) {
			t.Parallel()

			r := provider.NewBuiltinRegistry()
			c := &connector.OAuth2Connector{ClientID: "id", ClientSecret: "secret"}
			r.ApplyOAuth2Defaults(p, "https://example.com/cb", c)
			assert.True(t, c.RequiresPKCE,
				"provider %s must enable PKCE so Initiate generates a verifier", p)
		})
	}
}
