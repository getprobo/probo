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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/connector"
)

func TestNormalizeOktaDomain(t *testing.T) {
	t.Parallel()

	t.Run("valid inputs normalize to the bare host", func(t *testing.T) {
		t.Parallel()

		cases := map[string]string{
			"acme.okta.com":                 "acme.okta.com",
			"  acme.okta.com  ":             "acme.okta.com",
			"https://acme.okta.com":         "acme.okta.com",
			"https://acme.okta.com/":        "acme.okta.com",
			"http://acme.okta.com/sso/saml": "acme.okta.com",
			"ACME.OKTA.COM":                 "acme.okta.com",
			"dev-12345.okta.com":            "dev-12345.okta.com",
			"login.acme.com":                "login.acme.com",
			"acme.oktapreview.com":          "acme.oktapreview.com",
		}

		for input, want := range cases {
			got, err := connector.NormalizeOktaDomain(input)
			require.NoErrorf(t, err, "input %q", input)
			assert.Equalf(t, want, got, "input %q", input)
		}
	})

	t.Run("invalid inputs are rejected", func(t *testing.T) {
		t.Parallel()

		inputs := []string{
			"",
			"   ",
			"localhost",
			"okta",
			"acme.okta.com:8080",
			"https://acme.okta.com:443",
			"127.0.0.1",
			"169.254.169.254",
			"::1",
			"acme .okta.com",
			"-acme.okta.com",
		}

		for _, input := range inputs {
			_, err := connector.NormalizeOktaDomain(input)
			assert.Errorf(t, err, "input %q should be rejected", input)
		}
	})
}

func TestIsValidOktaDomain(t *testing.T) {
	t.Parallel()

	assert.True(t, connector.IsValidOktaDomain("acme.okta.com"))
	assert.True(t, connector.IsValidOktaDomain("dev-12345.okta.com"))
	assert.True(t, connector.IsValidOktaDomain("login.acme.co.uk"))

	assert.False(t, connector.IsValidOktaDomain(""))
	assert.False(t, connector.IsValidOktaDomain("localhost"))
	assert.False(t, connector.IsValidOktaDomain("192.168.0.1"))
	assert.False(t, connector.IsValidOktaDomain("acme.okta.com:443"))
	assert.False(t, connector.IsValidOktaDomain("ACME.OKTA.COM"))
}
