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

func TestDatadogAuthorizeURL(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		site string
		want string
	}{
		{"US1", "https://app.datadoghq.com/oauth2/v1/authorize"},
		{"US3", "https://us3.datadoghq.com/oauth2/v1/authorize"},
		{"US5", "https://us5.datadoghq.com/oauth2/v1/authorize"},
		{"EU1", "https://app.datadoghq.eu/oauth2/v1/authorize"},
		{"AP1", "https://ap1.datadoghq.com/oauth2/v1/authorize"},
		{"AP2", "https://ap2.datadoghq.com/oauth2/v1/authorize"},
		{"US1-FED", "https://app.ddog-gov.com/oauth2/v1/authorize"},
	} {
		got, err := connector.DatadogAuthorizeURL(tc.site)
		require.NoError(t, err, tc.site)
		assert.Equal(t, tc.want, got, tc.site)
	}

	_, err := connector.DatadogAuthorizeURL("BOGUS")
	require.Error(t, err)
}

func TestDatadogTokenURL(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		domain string
		want   string
	}{
		{"datadoghq.com", "https://api.datadoghq.com/oauth2/v1/token"},
		{"us3.datadoghq.com", "https://api.us3.datadoghq.com/oauth2/v1/token"},
		{"us5.datadoghq.com", "https://api.us5.datadoghq.com/oauth2/v1/token"},
		{"datadoghq.eu", "https://api.datadoghq.eu/oauth2/v1/token"},
		{"ap1.datadoghq.com", "https://api.ap1.datadoghq.com/oauth2/v1/token"},
		{"ap2.datadoghq.com", "https://api.ap2.datadoghq.com/oauth2/v1/token"},
		{"ddog-gov.com", "https://api.ddog-gov.com/oauth2/v1/token"},
	} {
		got, err := connector.DatadogTokenURL(tc.domain)
		require.NoError(t, err, tc.domain)
		assert.Equal(t, tc.want, got, tc.domain)
	}

	_, err := connector.DatadogTokenURL("evil.example.com")
	require.Error(t, err)
}

func TestIsValidDatadogDomain(t *testing.T) {
	t.Parallel()

	for _, domain := range []string{
		"datadoghq.com", "us3.datadoghq.com", "us5.datadoghq.com",
		"datadoghq.eu", "ap1.datadoghq.com", "ap2.datadoghq.com", "ddog-gov.com",
	} {
		assert.True(t, connector.IsValidDatadogDomain(domain), domain)
	}

	for _, domain := range []string{
		"",
		"api.datadoghq.com", // the API host, not an apiDomain — must be rejected
		"datadoghq.com.evil.com",
		"attacker.datadoghq.com.evil.com",
		"evil.example.com",
	} {
		assert.False(t, connector.IsValidDatadogDomain(domain), domain)
	}
}

func TestDatadogSiteForDomain(t *testing.T) {
	t.Parallel()

	site, ok := connector.DatadogSiteForDomain("us5.datadoghq.com")
	require.True(t, ok)
	assert.Equal(t, "US5", site)

	_, ok = connector.DatadogSiteForDomain("nope.com")
	assert.False(t, ok)
}
