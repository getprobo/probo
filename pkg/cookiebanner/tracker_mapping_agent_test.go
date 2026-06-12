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

package cookiebanner

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestNameMatchesSiteDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		vendorName string
		siteOrigin string
		expected   bool
	}{
		{
			name:       "site brand matches own domain",
			vendorName: "Letaido",
			siteOrigin: "https://letaido.com",
			expected:   true,
		},
		{
			name:       "case and spacing insensitive",
			vendorName: "  LET AIDO ",
			siteOrigin: "https://letaido.com",
			expected:   true,
		},
		{
			name:       "matches against subdomain origin",
			vendorName: "Letaido",
			siteOrigin: "https://app.letaido.com",
			expected:   true,
		},
		{
			name:       "matches full domain form",
			vendorName: "letaido.com",
			siteOrigin: "https://letaido.com",
			expected:   true,
		},
		{
			name:       "unrelated vendor is not the site",
			vendorName: "Google Analytics",
			siteOrigin: "https://letaido.com",
			expected:   false,
		},
		{
			name:       "empty vendor name",
			vendorName: "",
			siteOrigin: "https://letaido.com",
			expected:   false,
		},
		{
			name:       "empty origin",
			vendorName: "Letaido",
			siteOrigin: "",
			expected:   false,
		},
		{
			name:       "unparseable origin",
			vendorName: "Letaido",
			siteOrigin: "not a url",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, nameMatchesSiteDomain(tt.vendorName, tt.siteOrigin))
			},
		)
	}
}

func TestNameIsCookieDatabaseAggregator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		vendor   string
		expected bool
	}{
		{name: "cookifi is denied", vendor: "Cookifi", expected: true},
		{name: "cookiepedia is denied", vendor: "cookiepedia", expected: true},
		{name: "cookie database is denied", vendor: "Cookie Database", expected: true},
		{name: "cookieserve is denied", vendor: "CookieServe", expected: true},
		{name: "spacing and casing insensitive", vendor: "  COOK IFI ", expected: true},
		{name: "punctuation insensitive", vendor: "Cookie_Database", expected: true},
		{name: "cookiedatabase domain form is denied", vendor: "cookiedatabase.org", expected: true},
		{name: "cookifi domain form is denied", vendor: "cookifi.com", expected: true},
		{name: "cookiepedia url form is denied", vendor: "https://www.cookiepedia.co.uk/list", expected: true},
		{name: "cookieserve subdomain form is denied", vendor: "scan.cookieserve.com", expected: true},
		{name: "onetrust is allowed", vendor: "OneTrust", expected: false},
		{name: "cookiebot is allowed", vendor: "Cookiebot", expected: false},
		{name: "cookiebot domain form is allowed", vendor: "cookiebot.com", expected: false},
		{name: "cookieyes is allowed", vendor: "CookieYes", expected: false},
		{name: "cookie-script is allowed", vendor: "Cookie-Script", expected: false},
		{name: "unrelated vendor is allowed", vendor: "Google Analytics", expected: false},
		{name: "empty name", vendor: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, nameIsCookieDatabaseAggregator(tt.vendor))
			},
		)
	}
}

func TestBuildAgentPrompt(t *testing.T) {
	t.Parallel()

	tp := coredata.TrackerPattern{
		Pattern:     "ethereum-https://letaido.com",
		TrackerType: coredata.TrackerTypeLocalStorage,
		MatchType:   coredata.TrackerPatternMatchTypeExact,
	}

	t.Run(
		"emits scanned_site when domain supplied",
		func(t *testing.T) {
			t.Parallel()

			prompt := buildAgentPrompt(tp, nil, "letaido.com")
			assert.Contains(t, prompt, "<scanned_site> letaido.com </scanned_site>")
		},
	)

	t.Run(
		"omits scanned_site when domain empty",
		func(t *testing.T) {
			t.Parallel()

			prompt := buildAgentPrompt(tp, nil, "")
			assert.NotContains(t, prompt, "<scanned_site>")
		},
	)

	t.Run(
		"emits both scanned_site and observed_domains",
		func(t *testing.T) {
			t.Parallel()

			prompt := buildAgentPrompt(tp, []string{"doubleclick.net"}, "letaido.com")
			assert.Contains(t, prompt, "<scanned_site> letaido.com </scanned_site>")
			assert.Contains(t, prompt, "<observed_domains> doubleclick.net </observed_domains>")
			assert.Less(t, strings.Index(prompt, "<scanned_site>"), strings.Index(prompt, "<observed_domains>"))
		},
	)
}
