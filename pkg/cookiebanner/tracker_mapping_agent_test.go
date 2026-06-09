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

func TestNormalizeAlnum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "letters lowercased", input: "Letaido", expected: "letaido"},
		{name: "strips punctuation and spaces", input: "letaido.com - Inc", expected: "letaidocominc"},
		{name: "keeps digits", input: "auth0", expected: "auth0"},
		{name: "empty", input: "", expected: ""},
		{name: "only punctuation", input: "-_.:", expected: ""},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, normalizeAlnum(tt.input))
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
