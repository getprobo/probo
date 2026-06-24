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
