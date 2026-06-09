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
