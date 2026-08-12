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
	"regexp"
	"sort"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/slug"
)

// Duplicate-pair scores. Higher means more certainly the same vendor.
//
// The thresholds encode an asymmetry: folding two distinct vendors into one
// row destroys which of them each reference meant and cannot be undone,
// while leaving two rows for one vendor is repaired by merging them. So a
// signal is scored high only when the difference between the two names
// carries no identity.
const (
	// DuplicateScoreExactName is a case-insensitive name match: the state
	// the catalog's original name uniqueness used to prevent.
	DuplicateScoreExactName = 1.0

	// DuplicateScoreNormalizedName is a match after removing legal forms
	// and geographic qualifiers ("Acme Inc" / "Acme France" / "Acme").
	DuplicateScoreNormalizedName = 0.95

	// DuplicateScoreNormalizedSlug is a match once punctuation and spacing
	// are folded away too.
	DuplicateScoreNormalizedSlug = 0.9

	// DuplicateScoreBrandPrefixWithDomain is one name being a leading
	// whole-token prefix of the other AND a shared domain.
	//
	// Deliberately below DefaultDuplicateMinScore, because this is precisely
	// the shape a vendor's product family takes: "Google" prefixes "Google
	// Analytics", "Google Tag Manager", and "Google Workspace", and all of
	// them carry google.com, yet each is a distinct catalog entry with its
	// own category, cookie family, and DPA. Merging them would be an
	// irreversible loss. It does also catch true duplicates ("Tawk.to" and
	// "Tawk.to Chat"), so it is reported one threshold step down for a human
	// to adjudicate.
	DuplicateScoreBrandPrefixWithDomain = 0.75

	// DuplicateScoreSharedDomain is a shared owned domain and nothing else.
	//
	// Also below the default: the catalog routinely gives a whole product
	// family the parent's marketing domain, so "Adobe Analytics" and "Adobe
	// ColdFusion" share adobe.com and "Github" and "TimescaleDB" share
	// github.com while being entirely distinct entries.
	DuplicateScoreSharedDomain = 0.7

	// DuplicateScoreBrandPrefix is a leading whole-token prefix alone.
	//
	// Deliberately below DefaultDuplicateMinScore: distinct products of one
	// parent are legitimately separate catalog entries with their own
	// categories and cookie families, so "Google Analytics" and "Google"
	// must never be suggested for merging on their names alone. It is
	// reported only when the operator explicitly lowers the threshold.
	DuplicateScoreBrandPrefix = 0.55

	// DefaultDuplicateMinScore is the reporting threshold. It admits only
	// the name-identity signals — two rows whose names are the same once
	// noise that carries no identity is removed — because those are the
	// merges an operator can accept without adjudicating each one.
	//
	// Every domain- and prefix-based signal sits below it. Those find real
	// duplicates too, but they equally find a vendor's legitimate product
	// family, so they are opt-in via a lower threshold.
	DefaultDuplicateMinScore = 0.9
)

// minBrandPrefixTokens and minBrandPrefixRunes bound the prefix signal. A
// short single token prefixes a large share of any catalog ("Go", "In"), so
// a prefix only counts when the shorter side is either multi-token or long
// enough to be distinctive on its own.
const (
	minBrandPrefixTokens = 2
	minBrandPrefixRunes  = 5
)

// countryQualifiers are trailing geographic qualifiers removed when
// comparing catalog names.
//
// Unlike corporateSuffixes these are NOT removed on the write path: a
// country often marks a separately incorporated entity with its own data
// residency and DPA, so collapsing it there would irreversibly erase a
// jurisdictional distinction. Here the operator confirms each cluster
// before merging, so surfacing them as candidates is safe and useful.
//
// Order matters: stripCountryQualifier returns on the first match, so
// comma-prefixed and longer forms come before their shorter siblings.
var countryQualifiers = []string{
	" deutschland",
	" nederland",
	" singapore",
	" australia",
	" netherlands",
	" switzerland",
	" germany",
	" ireland",
	" belgium",
	" portugal",
	" benelux",
	" nordics",
	" espana",
	" france",
	" canada",
	" brasil",
	" brazil",
	" mexico",
	" italia",
	" japan",
	" india",
	" china",
	" korea",
	" spain",
	" italy",
	" latam",
	" emea",
	" apac",
	" usa",
	" uk",
	" us",
	" eu",
	" de",
	" fr",
	" jp",
	" kk",
}

// parentheticalSuffix matches a trailing parenthesised group. Any content
// counts: "(UK)", "(EU)", and "(legacy)" are all the same kind of
// disambiguating noise appended to make a second row for one vendor.
var parentheticalSuffix = regexp.MustCompile(`\s*\([^)]*\)\s*$`)

// whitespaceRun collapses internal whitespace so spacing differences do not
// prevent a match.
var whitespaceRun = regexp.MustCompile(`\s+`)

type (
	// CatalogEntry is a catalog row plus the signals duplicate detection
	// ranks on. Domains come from the row's owned-domain set; the counts are
	// how many patterns and organization third parties reference it.
	CatalogEntry struct {
		Party           *coredata.CommonThirdParty
		Domains         []string
		TrackerPatterns int
		OrgThirdParties int

		// Seeded marks a row whose slug appears in the curated seed data.
		// Such a row is authoritative and, decisively, is recreated by the
		// next seed run — so it must win its cluster.
		Seeded bool
	}

	// DuplicatePair is one scored relation between two catalog entries.
	DuplicatePair struct {
		A      gid.GID
		B      gid.GID
		Score  float64
		Reason string
	}

	// DuplicateCluster is a set of catalog entries connected by pairs,
	// with the entry that should absorb the others.
	DuplicateCluster struct {
		Winner *CatalogEntry
		Losers []*CatalogEntry
		Pairs  []DuplicatePair
	}
)

// normalized is an entry with its comparison keys precomputed. Normalizing
// costs two regexes and two suffix scans, so it is done once per entry
// rather than once per pair.
type normalized struct {
	entry      *CatalogEntry
	lowerName  string
	normName   string
	normSlug   string
	nameTokens []string
	domains    map[string]struct{}
}

// normalizeCatalogName reduces a catalog name to the form used for
// duplicate comparison: lowercased, with a trailing parenthesised group,
// legal form, and geographic qualifier removed.
//
// Legal-form removal runs on both sides of the country removal so both
// "Hotjar Ltd (UK)" (parentheses, then suffix) and "Acme France Inc"
// (suffix, then country) reduce to the same brand.
func normalizeCatalogName(name string) string {
	out := strings.ToLower(strings.TrimSpace(name))
	out = parentheticalSuffix.ReplaceAllString(out, "")
	out = stripCorporateSuffixes(out)
	out = stripCountryQualifier(out)
	out = stripCorporateSuffixes(out)
	out = whitespaceRun.ReplaceAllString(out, " ")

	return strings.TrimSpace(out)
}

// stripCountryQualifier removes a single trailing geographic qualifier from
// a lowercased name. Only one is removed, matching stripCorporateSuffixes,
// so a name ending in two such words is not mangled.
func stripCountryQualifier(lowerName string) string {
	for _, q := range countryQualifiers {
		if before, ok := strings.CutSuffix(lowerName, q); ok {
			return strings.TrimSpace(before)
		}
	}

	return lowerName
}

// FindDuplicates groups catalog entries that are likely the same vendor.
//
// Pairs scoring at least minScore are grouped transitively, so three
// spellings of one vendor come back as one cluster rather than three
// separate pairs to reassemble. Pass zero for minScore to use
// DefaultDuplicateMinScore.
//
// Pure function: deterministic on its inputs, no I/O.
func FindDuplicates(entries []*CatalogEntry, minScore float64) []DuplicateCluster {
	if minScore <= 0 {
		minScore = DefaultDuplicateMinScore
	}

	items := make([]normalized, 0, len(entries))

	for _, e := range entries {
		if e == nil || e.Party == nil {
			continue
		}

		normName := normalizeCatalogName(e.Party.Name)

		domains := make(map[string]struct{}, len(e.Domains))
		for _, d := range e.Domains {
			if d = strings.ToLower(strings.TrimSpace(d)); d != "" {
				domains[d] = struct{}{}
			}
		}

		items = append(items, normalized{
			entry:      e,
			lowerName:  strings.ToLower(strings.TrimSpace(e.Party.Name)),
			normName:   normName,
			normSlug:   slug.Make(normName),
			nameTokens: strings.Fields(normName),
			domains:    domains,
		})
	}

	pairs := scorePairs(items, minScore)
	if len(pairs) == 0 {
		return nil
	}

	return buildClusters(items, pairs)
}

// scorePairs returns every pair scoring at least minScore.
//
// Candidates come from three buckets — normalized name, normalized slug,
// and each domain — so the exact, normalized, slug, and domain signals cost
// one pass over the entries rather than one over every pair. The unqualified
// prefix signal cannot be indexed that way, so it runs as a full pairwise
// pass only when the caller lowers the threshold enough to report it.
func scorePairs(items []normalized, minScore float64) []DuplicatePair {
	best := make(map[[2]int]DuplicatePair)

	consider := func(i, j int) {
		if i == j {
			return
		}

		if j < i {
			i, j = j, i
		}

		score, reason := scorePair(items[i], items[j])
		if score < minScore {
			return
		}

		key := [2]int{i, j}
		if prev, ok := best[key]; ok && prev.Score >= score {
			return
		}

		best[key] = DuplicatePair{
			A:      items[i].entry.Party.ID,
			B:      items[j].entry.Party.ID,
			Score:  score,
			Reason: reason,
		}
	}

	buckets := make(map[string][]int)

	addBucket := func(key string, idx int) {
		if key == "" {
			return
		}

		buckets[key] = append(buckets[key], idx)
	}

	for i, it := range items {
		addBucket("name:"+it.normName, i)
		addBucket("slug:"+it.normSlug, i)

		for d := range it.domains {
			addBucket("domain:"+d, i)
		}
	}

	for _, idxs := range buckets {
		for a := range idxs {
			for b := a + 1; b < len(idxs); b++ {
				consider(idxs[a], idxs[b])
			}
		}
	}

	// The bare prefix signal is the only one the buckets above cannot index,
	// and it scores below the default threshold, so this pass only happens
	// when the operator lowered the threshold far enough to see it.
	if minScore <= DuplicateScoreBrandPrefix {
		for i := range items {
			for j := i + 1; j < len(items); j++ {
				consider(i, j)
			}
		}
	}

	pairs := make([]DuplicatePair, 0, len(best))
	for _, p := range best {
		pairs = append(pairs, p)
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Score != pairs[j].Score {
			return pairs[i].Score > pairs[j].Score
		}

		if a, b := pairs[i].A.String(), pairs[j].A.String(); a != b {
			return a < b
		}

		return pairs[i].B.String() < pairs[j].B.String()
	})

	return pairs
}

// scorePair scores one candidate pair, returning the highest applicable
// signal and a human-readable reason.
func scorePair(a, b normalized) (float64, string) {
	sharedDomain := ""

	for d := range a.domains {
		if _, ok := b.domains[d]; ok {
			if sharedDomain == "" || d < sharedDomain {
				sharedDomain = d
			}
		}
	}

	switch {
	case a.lowerName != "" && a.lowerName == b.lowerName:
		return DuplicateScoreExactName, "identical name"
	case a.normName != "" && a.normName == b.normName:
		return DuplicateScoreNormalizedName, "same name once legal form and country are removed: " + a.normName
	case a.normSlug != "" && a.normSlug == b.normSlug:
		return DuplicateScoreNormalizedSlug, "same normalized slug: " + a.normSlug
	case sharedDomain != "":
		if isBrandPrefix(a.nameTokens, b.nameTokens) {
			return DuplicateScoreBrandPrefixWithDomain, "one name prefixes the other and they share the domain " + sharedDomain
		}

		return DuplicateScoreSharedDomain, "shared domain " + sharedDomain
	case isBrandPrefix(a.nameTokens, b.nameTokens):
		return DuplicateScoreBrandPrefix, "one name prefixes the other, no shared domain"
	default:
		return 0, ""
	}
}

// isBrandPrefix reports whether the shorter token sequence is a leading
// whole-token prefix of the longer one, e.g. ["google"] against
// ["google","analytics"]. Whole tokens only, so "Goo" does not prefix
// "Google", and the shorter side must clear the distinctiveness bounds.
func isBrandPrefix(a, b []string) bool {
	shorter, longer := a, b
	if len(shorter) > len(longer) {
		shorter, longer = longer, shorter
	}

	if len(shorter) == 0 || len(shorter) == len(longer) {
		return false
	}

	if len(shorter) < minBrandPrefixTokens &&
		len([]rune(strings.Join(shorter, " "))) < minBrandPrefixRunes {
		return false
	}

	for i, tok := range shorter {
		if longer[i] != tok {
			return false
		}
	}

	return true
}

// buildClusters groups pairs into connected components via union-find and
// picks each component's winner.
func buildClusters(items []normalized, pairs []DuplicatePair) []DuplicateCluster {
	indexByID := make(map[gid.GID]int, len(items))
	for i, it := range items {
		indexByID[it.entry.Party.ID] = i
	}

	parent := make([]int, len(items))
	for i := range parent {
		parent[i] = i
	}

	var find func(int) int

	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}

		return parent[i]
	}

	union := func(i, j int) {
		ri, rj := find(i), find(j)
		if ri != rj {
			parent[rj] = ri
		}
	}

	for _, p := range pairs {
		union(indexByID[p.A], indexByID[p.B])
	}

	members := make(map[int][]int)
	for i := range items {
		root := find(i)
		members[root] = append(members[root], i)
	}

	pairsByRoot := make(map[int][]DuplicatePair)
	for _, p := range pairs {
		root := find(indexByID[p.A])
		pairsByRoot[root] = append(pairsByRoot[root], p)
	}

	clusters := make([]DuplicateCluster, 0, len(pairsByRoot))

	for root, idxs := range members {
		clusterPairs := pairsByRoot[root]
		if len(clusterPairs) == 0 {
			continue
		}

		entries := make([]*CatalogEntry, 0, len(idxs))
		for _, i := range idxs {
			entries = append(entries, items[i].entry)
		}

		sort.SliceStable(entries, func(i, j int) bool {
			return preferAsWinner(entries[i], entries[j])
		})

		clusters = append(clusters, DuplicateCluster{
			Winner: entries[0],
			Losers: entries[1:],
			Pairs:  clusterPairs,
		})
	}

	sort.SliceStable(clusters, func(i, j int) bool {
		if a, b := clusters[i].Pairs[0].Score, clusters[j].Pairs[0].Score; a != b {
			return a > b
		}

		return clusters[i].Winner.Party.ID.String() < clusters[j].Winner.Party.ID.String()
	})

	return clusters
}

// preferAsWinner reports whether a should absorb b.
//
// Seeded wins first and decisively: the next seed run recreates that slug,
// so merging it away guarantees the duplicate returns. Organization links
// come next because repointing fewer of them disturbs fewer tenants and
// risks fewer same-organization collisions. Then pattern links, then how
// far enrichment got. Creation time and id break ties last so the ordering
// is stable across runs, which matters when the output is piped into merge
// commands.
func preferAsWinner(a, b *CatalogEntry) bool {
	if a.Seeded != b.Seeded {
		return a.Seeded
	}

	if a.OrgThirdParties != b.OrgThirdParties {
		return a.OrgThirdParties > b.OrgThirdParties
	}

	if a.TrackerPatterns != b.TrackerPatterns {
		return a.TrackerPatterns > b.TrackerPatterns
	}

	if ra, rb := enrichmentRank(a), enrichmentRank(b); ra != rb {
		return ra > rb
	}

	if !a.Party.CreatedAt.Equal(b.Party.CreatedAt) {
		return a.Party.CreatedAt.Before(b.Party.CreatedAt)
	}

	return a.Party.ID.String() < b.Party.ID.String()
}

// enrichmentRank scores how complete a row's profile is, so a fully
// enriched row outranks a bare stub carrying only name and category.
func enrichmentRank(e *CatalogEntry) int {
	if len(e.Party.Enrichment) == 0 {
		return 0
	}

	switch {
	case strings.Contains(string(e.Party.Enrichment), `"status":"done"`):
		return 3
	case strings.Contains(string(e.Party.Enrichment), `"status":"partial"`):
		return 2
	default:
		return 1
	}
}
