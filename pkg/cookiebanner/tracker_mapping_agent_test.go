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
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
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

func TestEvidenceSupportsAttribution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		evidence string
		expected bool
	}{
		{name: "database match supports", evidence: evidenceSourceDatabaseMatch, expected: true},
		{name: "naming convention supports", evidence: evidenceSourceNamingConvention, expected: true},
		{name: "web search supports", evidence: evidenceSourceWebSearch, expected: true},
		{name: "browser page supports", evidence: evidenceSourceBrowserPage, expected: true},
		{name: "none does not support", evidence: evidenceSourceNone, expected: false},
		{name: "empty does not support", evidence: "", expected: false},
		{name: "unknown value does not support", evidence: "vibes", expected: false},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()
				assert.Equal(t, tt.expected, evidenceSupportsAttribution(tt.evidence))
			},
		)
	}
}

func TestInterpretCatalogRow(t *testing.T) {
	t.Parallel()

	vendorID := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)

	t.Run(
		"first party verdict is terminal",
		func(t *testing.T) {
			t.Parallel()

			adopt, untrusted, firstParty := interpretCatalogRow(coredata.CommonTrackerPattern{
				CommonThirdPartyID: &vendorID,
				Confidence:         1,
				Attribution:        coredata.CommonTrackerPatternAttributionFirstParty,
			})

			assert.Nil(t, adopt)
			assert.Nil(t, untrusted)
			assert.True(t, firstParty)
		},
	)

	t.Run(
		"trusted vendor is adopted",
		func(t *testing.T) {
			t.Parallel()

			adopt, untrusted, firstParty := interpretCatalogRow(coredata.CommonTrackerPattern{
				CommonThirdPartyID: &vendorID,
				Confidence:         trustedAttributionConfidence,
				Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
			})

			require.NotNil(t, adopt)
			assert.Equal(t, vendorID, *adopt)
			assert.Nil(t, untrusted)
			assert.False(t, firstParty)
		},
	)

	t.Run(
		"low-confidence vendor is untrusted, not adopted",
		func(t *testing.T) {
			t.Parallel()

			adopt, untrusted, firstParty := interpretCatalogRow(coredata.CommonTrackerPattern{
				CommonThirdPartyID: &vendorID,
				Confidence:         agentSourceConfidence,
				Attribution:        coredata.CommonTrackerPatternAttributionThirdParty,
			})

			assert.Nil(t, adopt)
			require.NotNil(t, untrusted)
			assert.Equal(t, vendorID, *untrusted)
			assert.False(t, firstParty)
		},
	)

	t.Run(
		"unlinked row yields nothing",
		func(t *testing.T) {
			t.Parallel()

			adopt, untrusted, firstParty := interpretCatalogRow(coredata.CommonTrackerPattern{
				Confidence:  0.5,
				Attribution: coredata.CommonTrackerPatternAttributionUndetermined,
			})

			assert.Nil(t, adopt)
			assert.Nil(t, untrusted)
			assert.False(t, firstParty)
		},
	)
}

func TestIsPreExistingSource(t *testing.T) {
	t.Parallel()

	preExisting := coredata.CookieSourcePreExisting
	script := coredata.CookieSourceScript

	assert.True(t, isPreExistingSource(coredata.TrackerPattern{Source: &preExisting}))
	assert.False(t, isPreExistingSource(coredata.TrackerPattern{Source: &script}))
	assert.False(t, isPreExistingSource(coredata.TrackerPattern{Source: nil}))
}

func TestVendorAttributionRejected(t *testing.T) {
	t.Parallel()

	h := newMappingHandler(nil)
	ctx := context.Background()
	tp := coredata.TrackerPattern{Pattern: "_x", TrackerType: coredata.TrackerTypeCookie}

	confident := func(mut func(*TrackerMappingAgentResult)) TrackerMappingAgentResult {
		r := TrackerMappingAgentResult{
			ThirdPartyName:       "Acme Analytics",
			Category:             coredata.ThirdPartyCategoryAnalytics,
			ThirdPartyConfidence: 0.9,
			EvidenceSource:       evidenceSourceNamingConvention,
		}
		if mut != nil {
			mut(&r)
		}

		return r
	}

	t.Run(
		"accepts a confident, evidence-backed attribution",
		func(t *testing.T) {
			t.Parallel()
			assert.False(t, h.vendorAttributionRejected(ctx, tp, confident(nil), "https://example.com"))
		},
	)

	t.Run(
		"rejects below confidence threshold",
		func(t *testing.T) {
			t.Parallel()
			r := confident(func(r *TrackerMappingAgentResult) { r.ThirdPartyConfidence = 0.3 })
			assert.True(t, h.vendorAttributionRejected(ctx, tp, r, "https://example.com"))
		},
	)

	t.Run(
		"rejects empty name",
		func(t *testing.T) {
			t.Parallel()
			r := confident(func(r *TrackerMappingAgentResult) { r.ThirdPartyName = "" })
			assert.True(t, h.vendorAttributionRejected(ctx, tp, r, "https://example.com"))
		},
	)

	t.Run(
		"rejects when evidence source is none",
		func(t *testing.T) {
			t.Parallel()
			r := confident(func(r *TrackerMappingAgentResult) { r.EvidenceSource = evidenceSourceNone })
			assert.True(t, h.vendorAttributionRejected(ctx, tp, r, "https://example.com"))
		},
	)

	t.Run(
		"rejects when evidence source is empty",
		func(t *testing.T) {
			t.Parallel()
			r := confident(func(r *TrackerMappingAgentResult) { r.EvidenceSource = "" })
			assert.True(t, h.vendorAttributionRejected(ctx, tp, r, "https://example.com"))
		},
	)

	t.Run(
		"rejects when name matches scanned site",
		func(t *testing.T) {
			t.Parallel()
			r := confident(func(r *TrackerMappingAgentResult) { r.ThirdPartyName = "Example" })
			assert.True(t, h.vendorAttributionRejected(ctx, tp, r, "https://example.com"))
		},
	)

	t.Run(
		"rejects cookie-database aggregator",
		func(t *testing.T) {
			t.Parallel()
			r := confident(func(r *TrackerMappingAgentResult) { r.ThirdPartyName = "Cookiepedia" })
			assert.True(t, h.vendorAttributionRejected(ctx, tp, r, "https://example.com"))
		},
	)
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
