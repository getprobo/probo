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

package thirdparty

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/slug"
)

// catalogEntry builds a CatalogEntry with a fresh id, for readability in
// the table tests below.
func catalogEntry(name string, domains ...string) *CatalogEntry {
	return &CatalogEntry{
		Party: &coredata.CommonThirdParty{
			ID:        gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
			Name:      name,
			Slug:      slug.Make(name),
			Category:  coredata.ThirdPartyCategoryAnalytics,
			CreatedAt: time.Now().UTC(),
		},
		Domains: domains,
	}
}

func TestNormalizeCatalogName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "plain name is unchanged", input: "Hotjar", expected: "hotjar"},
		{name: "lowercases", input: "HotJar", expected: "hotjar"},
		{name: "trims", input: "  Hotjar  ", expected: "hotjar"},
		{name: "strips a legal form", input: "Acme Inc", expected: "acme"},
		{name: "strips a comma legal form", input: "Acme, Inc.", expected: "acme"},
		{name: "strips gmbh", input: "LiveZilla GmbH", expected: "livezilla"},
		{name: "strips limited", input: "PowerLinks Media Limited", expected: "powerlinks media"},
		{name: "strips a parenthesised qualifier", input: "Hotjar (UK)", expected: "hotjar"},
		{name: "strips parentheses then legal form", input: "Hotjar Ltd (UK)", expected: "hotjar"},
		// Geographic qualifiers are preserved: a country often marks a
		// separately incorporated entity with its own data residency, and
		// short forms collide with ordinary words (.ai is a country code).
		{name: "keeps a country name", input: "Acme France", expected: "acme france"},
		{name: "keeps a country code", input: "OVHcloud US", expected: "ovhcloud us"},
		{name: "keeps a country after a legal form is stripped", input: "Acme France Inc", expected: "acme france"},
		{name: "keeps an AI suffix", input: "Fireworks AI", expected: "fireworks ai"},
		{name: "keeps a region word", input: "Acme EMEA", expected: "acme emea"},
		{name: "collapses internal whitespace", input: "Acme   Analytics", expected: "acme analytics"},
		// Conservatism: these words are not qualifiers, so the names must
		// survive intact rather than over-merging onto a shorter brand.
		{name: "keeps a non-qualifier trailing word", input: "Cisco Systems", expected: "cisco systems"},
		{name: "keeps a product name", input: "Amazon Web Services", expected: "amazon web services"},
		{name: "keeps group", input: "Criteo Group", expected: "criteo group"},
		{name: "empty stays empty", input: "", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, normalizeCatalogName(tt.input))
		})
	}
}

func TestFindDuplicates_Signals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		a             *CatalogEntry
		b             *CatalogEntry
		minScore      float64
		expectPaired  bool
		expectedScore float64
	}{
		{
			name:          "identical names",
			a:             catalogEntry("Hotjar"),
			b:             catalogEntry("hotjar"),
			expectPaired:  true,
			expectedScore: DuplicateScoreExactName,
		},
		{
			name:          "legal form variant",
			a:             catalogEntry("Hotjar"),
			b:             catalogEntry("Hotjar Ltd"),
			expectPaired:  true,
			expectedScore: DuplicateScoreNormalizedName,
		},
		{
			// A country qualifier is not noise: "OVHcloud US" is a separately
			// incorporated entity with its own processing agreement, so this
			// must not be offered as a merge.
			name:         "country qualified name is not a duplicate",
			a:            catalogEntry("Acme Analytics"),
			b:            catalogEntry("Acme Analytics France"),
			expectPaired: false,
		},
		{
			// .ai is a country code, so any country-code stripping would
			// reduce this to "Fireworks" and pair it with an unrelated vendor.
			name:         "an AI-suffixed brand is not a duplicate of its stem",
			a:            catalogEntry("Fireworks"),
			b:            catalogEntry("Fireworks AI"),
			expectPaired: false,
		},
		{
			name:          "parenthesised variant",
			a:             catalogEntry("Hotjar"),
			b:             catalogEntry("Hotjar Ltd (UK)"),
			expectPaired:  true,
			expectedScore: DuplicateScoreNormalizedName,
		},
		{
			// A shared corporate domain alone is weak: the catalog gives a
			// vendor's whole product family the parent's domain, so this
			// must stay below the default threshold.
			name:         "shared domain alone is below the default threshold",
			a:            catalogEntry("Adobe Analytics", "adobe.com"),
			b:            catalogEntry("Adobe ColdFusion", "adobe.com"),
			expectPaired: false,
		},
		{
			name:          "shared domain alone surfaces when the threshold is lowered",
			a:             catalogEntry("Adobe Analytics", "adobe.com"),
			b:             catalogEntry("Adobe ColdFusion", "adobe.com"),
			minScore:      DuplicateScoreSharedDomain,
			expectPaired:  true,
			expectedScore: DuplicateScoreSharedDomain,
		},
		{
			name:          "shared domain is case insensitive",
			a:             catalogEntry("Segment", "Segment.COM"),
			b:             catalogEntry("Twilio CDP", "segment.com"),
			minScore:      DuplicateScoreSharedDomain,
			expectPaired:  true,
			expectedScore: DuplicateScoreSharedDomain,
		},
		{
			// A prefix plus a shared domain finds true duplicates like
			// "Tawk.to" / "Tawk.to Chat", but is the same shape as a product
			// family, so it stays opt-in.
			name:          "brand prefix with a shared domain surfaces when the threshold is lowered",
			a:             catalogEntry("Tawk.to", "tawk.to"),
			b:             catalogEntry("Tawk.to Chat", "tawk.to"),
			minScore:      DuplicateScoreBrandPrefixWithDomain,
			expectPaired:  true,
			expectedScore: DuplicateScoreBrandPrefixWithDomain,
		},
		{
			// The decisive negative, from real catalog rows: a parent and its
			// products all carry the parent's domain, and merging them would
			// destroy distinct entries with their own DPAs.
			name:         "a parent and its product are not paired by default",
			a:            catalogEntry("Google", "google.com"),
			b:            catalogEntry("Google Analytics", "google.com"),
			expectPaired: false,
		},
		{
			// The load-bearing negative: distinct products of one parent are
			// legitimately separate entries, so this must not be suggested
			// for merging at the default threshold.
			name:         "brand prefix alone is below the default threshold",
			a:            catalogEntry("Google"),
			b:            catalogEntry("Google Analytics"),
			expectPaired: false,
		},
		{
			name:          "brand prefix alone surfaces when the threshold is lowered",
			a:             catalogEntry("Google"),
			b:             catalogEntry("Google Analytics"),
			minScore:      DuplicateScoreBrandPrefix,
			expectPaired:  true,
			expectedScore: DuplicateScoreBrandPrefix,
		},
		{
			name:         "unrelated vendors are not paired",
			a:            catalogEntry("Hotjar", "hotjar.com"),
			b:            catalogEntry("Mixpanel", "mixpanel.com"),
			minScore:     DuplicateScoreBrandPrefix,
			expectPaired: false,
		},
		{
			// Two unrelated projects hosted on one corporate domain: real
			// catalog rows that must not be suggested for merging by default.
			name:         "distinct products on one corporate domain are not paired by default",
			a:            catalogEntry("Github", "github.com"),
			b:            catalogEntry("TimescaleDB", "github.com"),
			expectPaired: false,
		},
		{
			name:         "a non-qualifier trailing word does not merge",
			a:            catalogEntry("Cisco"),
			b:            catalogEntry("Cisco Systems"),
			expectPaired: false,
		},
		{
			name:         "a short prefix does not merge",
			a:            catalogEntry("Go"),
			b:            catalogEntry("Go Cardless"),
			minScore:     DuplicateScoreBrandPrefix,
			expectPaired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			clusters := FindDuplicates([]*CatalogEntry{tt.a, tt.b}, tt.minScore)

			if !tt.expectPaired {
				assert.Empty(t, clusters)
				return
			}

			require.Len(t, clusters, 1)
			require.Len(t, clusters[0].Losers, 1)
			require.Len(t, clusters[0].Pairs, 1)
			assert.InDelta(t, tt.expectedScore, clusters[0].Pairs[0].Score, 0.0001)
			assert.NotEmpty(t, clusters[0].Pairs[0].Reason)
		})
	}
}

// TestFindDuplicates_ClustersTransitively pins that variants of one vendor
// come back as a single cluster. A and C need not match each other directly
// — they are connected through B — so the operator gets one merge to
// perform rather than a set of pairs to reassemble.
func TestFindDuplicates_ClustersTransitively(t *testing.T) {
	t.Parallel()

	a := catalogEntry("Acme")
	b := catalogEntry("Acme Inc")
	c := catalogEntry("Acme Ltd (EU)")
	unrelated := catalogEntry("Mixpanel")

	clusters := FindDuplicates([]*CatalogEntry{a, b, c, unrelated}, 0)

	require.Len(t, clusters, 1)
	assert.Len(t, clusters[0].Losers, 2)

	ids := map[gid.GID]bool{clusters[0].Winner.Party.ID: true}
	for _, l := range clusters[0].Losers {
		ids[l.Party.ID] = true
	}

	assert.True(t, ids[a.Party.ID])
	assert.True(t, ids[b.Party.ID])
	assert.True(t, ids[c.Party.ID])
	assert.False(t, ids[unrelated.Party.ID], "an unrelated vendor must not join the cluster")
}

// TestFindDuplicates_WinnerSelection pins the winner heuristic tier by
// tier. Seeded must win outright: the next seed run recreates that slug, so
// merging it away would resurrect the duplicate.
func TestFindDuplicates_WinnerSelection(t *testing.T) {
	t.Parallel()

	t.Run("seeded row wins over a better linked row", func(t *testing.T) {
		t.Parallel()

		seeded := catalogEntry("Acme")
		seeded.Seeded = true

		linked := catalogEntry("Acme Inc")
		linked.OrgThirdParties = 99
		linked.TrackerPatterns = 99

		clusters := FindDuplicates([]*CatalogEntry{linked, seeded}, 0)

		require.Len(t, clusters, 1)
		assert.Equal(t, seeded.Party.ID, clusters[0].Winner.Party.ID)
	})

	t.Run("more organization links wins", func(t *testing.T) {
		t.Parallel()

		few := catalogEntry("Acme")
		few.OrgThirdParties = 1
		few.TrackerPatterns = 50

		many := catalogEntry("Acme Inc")
		many.OrgThirdParties = 5

		clusters := FindDuplicates([]*CatalogEntry{few, many}, 0)

		require.Len(t, clusters, 1)
		assert.Equal(t, many.Party.ID, clusters[0].Winner.Party.ID)
	})

	t.Run("more pattern links breaks an organization tie", func(t *testing.T) {
		t.Parallel()

		few := catalogEntry("Acme")
		few.TrackerPatterns = 2

		many := catalogEntry("Acme Inc")
		many.TrackerPatterns = 7

		clusters := FindDuplicates([]*CatalogEntry{few, many}, 0)

		require.Len(t, clusters, 1)
		assert.Equal(t, many.Party.ID, clusters[0].Winner.Party.ID)
	})

	t.Run("more resolved enrichment fields breaks a link tie", func(t *testing.T) {
		t.Parallel()

		// Both names are equally clean and both runs report "partial", so only
		// the number of fields actually resolved separates them.
		thin := catalogEntry("Acme Ltd")
		thin.Party.Enrichment = json.RawMessage(
			`{"status":"partial","fields":{"website_url":{"status":"found"},"privacy_policy_url":{"status":"not_found"}}}`,
		)

		full := catalogEntry("Acme Inc")
		full.Party.Enrichment = json.RawMessage(
			`{"status":"partial","fields":{"website_url":{"status":"found"},"privacy_policy_url":{"status":"found"},"legal_name":{"status":"exists_external"}}}`,
		)

		clusters := FindDuplicates([]*CatalogEntry{thin, full}, 0)

		require.Len(t, clusters, 1)
		assert.Equal(t, full.Party.ID, clusters[0].Winner.Party.ID)
	})

	t.Run("a cleaner name breaks an otherwise total tie", func(t *testing.T) {
		t.Parallel()

		// Equal on every reference count and on enrichment, so the winner must
		// not be decided by whichever row the agent happened to create first.
		noisy := catalogEntry("h265ify (browser extension)")
		noisy.Party.CreatedAt = time.Now().UTC().Add(-48 * time.Hour)

		clean := catalogEntry("h265ify")

		clusters := FindDuplicates([]*CatalogEntry{noisy, clean}, 0)

		require.Len(t, clusters, 1)
		assert.Equal(t, clean.Party.ID, clusters[0].Winner.Party.ID)
	})

	t.Run("reference counts outrank a cleaner name", func(t *testing.T) {
		t.Parallel()

		clean := catalogEntry("Microsoft Advertising")

		used := catalogEntry("Microsoft Advertising (Bing)")
		used.OrgThirdParties = 1
		used.TrackerPatterns = 10

		clusters := FindDuplicates([]*CatalogEntry{clean, used}, 0)

		require.Len(t, clusters, 1)
		assert.Equal(t, used.Party.ID, clusters[0].Winner.Party.ID,
			"the most-referenced row must survive even with a noisier name")
	})

	t.Run("the oldest row breaks every other tie", func(t *testing.T) {
		t.Parallel()

		// Same length and both free of parentheses, so the name tiebreak
		// cannot separate them either.
		newer := catalogEntry("Acme Inc")
		newer.Party.CreatedAt = time.Now().UTC()

		older := catalogEntry("Acme Ltd")
		older.Party.CreatedAt = newer.Party.CreatedAt.Add(-48 * time.Hour)

		clusters := FindDuplicates([]*CatalogEntry{newer, older}, 0)

		require.Len(t, clusters, 1)
		assert.Equal(t, older.Party.ID, clusters[0].Winner.Party.ID)
	})
}

// TestFindDuplicates_IsDeterministic pins run-to-run stability. The output
// is piped into merge commands, so the same catalog must always produce the
// same winner and the same ordering regardless of input order.
func TestFindDuplicates_IsDeterministic(t *testing.T) {
	t.Parallel()

	a := catalogEntry("Acme")
	b := catalogEntry("Acme Inc")
	c := catalogEntry("Acme Ltd")

	first := FindDuplicates([]*CatalogEntry{a, b, c}, 0)
	second := FindDuplicates([]*CatalogEntry{c, b, a}, 0)

	require.Len(t, first, 1)
	require.Len(t, second, 1)
	assert.Equal(t, first[0].Winner.Party.ID, second[0].Winner.Party.ID)
}

// TestFindDuplicates_EmptyAndNilInputs pins that degenerate input is
// tolerated rather than panicking a diagnosis run.
func TestFindDuplicates_EmptyAndNilInputs(t *testing.T) {
	t.Parallel()

	assert.Empty(t, FindDuplicates(nil, 0))
	assert.Empty(t, FindDuplicates([]*CatalogEntry{}, 0))
	assert.Empty(t, FindDuplicates([]*CatalogEntry{nil, {}}, 0))
	assert.Empty(t, FindDuplicates([]*CatalogEntry{catalogEntry("Solo")}, 0))
}

// TestFindDuplicates_BlankNamesDoNotCluster pins that rows whose names
// normalize to nothing are not all declared duplicates of each other.
func TestFindDuplicates_BlankNamesDoNotCluster(t *testing.T) {
	t.Parallel()

	// Several rows whose names normalize to nothing. They must not bucket
	// together: a shared empty key would compare them all pairwise and, worse,
	// report them as duplicates of each other.
	entries := []*CatalogEntry{
		catalogEntry("!!!"),
		catalogEntry("???"),
		catalogEntry("---"),
		catalogEntry("***"),
	}

	assert.Empty(t, FindDuplicates(entries, 0))
}

// TestFindDuplicates_RejectsNaNThreshold pins that a NaN threshold falls back
// to the default. Every comparison against NaN is false, so `score < minScore`
// would admit zero-score pairs and report unrelated vendors as duplicates.
func TestFindDuplicates_RejectsNaNThreshold(t *testing.T) {
	t.Parallel()

	unrelated := []*CatalogEntry{
		catalogEntry("Hotjar", "hotjar.com"),
		catalogEntry("Mixpanel", "mixpanel.com"),
	}

	assert.Empty(t, FindDuplicates(unrelated, math.NaN()))
}
