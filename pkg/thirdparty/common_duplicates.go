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

// Duplicate-pair scores, and the threshold that decides which are reported.
//
// Folding two distinct vendors into one row destroys which of them each
// reference meant and cannot be undone, while leaving two rows for one vendor
// is repaired by merging them. So only the name-identity signals — equal once
// noise that carries no identity is removed — sit at or above the default
// threshold.
//
// Everything domain- or prefix-based sits below it. Those find real
// duplicates ("Tawk.to" / "Tawk.to Chat") but equally find a vendor's
// legitimate product family: "Google" prefixes "Google Analytics" and "Google
// Workspace" and all carry google.com, while "Adobe Analytics" and "Adobe
// ColdFusion" share adobe.com, each a distinct entry with its own category
// and agreements. Those need a human, so they are opt-in via a lower
// threshold.
const (
	DuplicateScoreExactName      = 1.0  // equal lowercased
	DuplicateScoreNormalizedName = 0.95 // equal after legal form and parentheses
	DuplicateScoreNormalizedSlug = 0.9  // equal once punctuation folds away

	DuplicateScoreBrandPrefixWithDomain = 0.75 // token prefix and a shared domain
	DuplicateScoreSharedDomain          = 0.7  // shared domain alone
	DuplicateScoreBrandPrefix           = 0.55 // token prefix alone

	DefaultDuplicateMinScore = 0.9
)

// A short token prefixes a large share of any catalog ("Go", "In"), so the
// prefix signal requires the shorter side to be multi-token or long enough to
// be distinctive alone.
const (
	minBrandPrefixTokens = 2
	minBrandPrefixRunes  = 5
)

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

// normalizeCatalogName lowercases a name and removes a trailing parenthesised
// group and legal form, neither of which carries identity. Legal-form removal
// runs twice because the parentheses can hide one behind them: "Hotjar Ltd
// (UK)".
//
// Geographic qualifiers are deliberately kept. "OVHcloud US" is a separately
// incorporated entity with its own data residency, and unlike legal forms —
// a closed set that is never anything else — country codes collide with
// ordinary words, so stripping them would reduce "Fireworks AI" to
// "Fireworks". A country-suffixed duplicate still surfaces through the prefix
// signal at a lower threshold.
func normalizeCatalogName(name string) string {
	out := strings.ToLower(strings.TrimSpace(name))
	out = parentheticalSuffix.ReplaceAllString(out, "")
	out = stripCorporateSuffixes(out)
	out = stripCorporateSuffixes(out)
	out = whitespaceRun.ReplaceAllString(out, " ")

	return strings.TrimSpace(out)
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
		return DuplicateScoreNormalizedName, "same name once legal form and parentheses are removed: " + a.normName
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

// preferAsWinner reports whether a should absorb b, in the tier order below.
//
// Seeded wins decisively: the next seed run recreates that slug, so merging it
// away guarantees the duplicate returns. Organization links rank above pattern
// links because repointing fewer of them disturbs fewer tenants, and both rank
// above name shape so the most-referenced row always survives. Creation time
// and id break ties last, keeping the output stable enough to pipe into merge
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

	if ra, rb := enrichmentResolvedFields(a), enrichmentResolvedFields(b); ra != rb {
		return ra > rb
	}

	// Between rows that are otherwise equal, prefer the cleaner name. A
	// trailing parenthetical is disambiguating noise someone appended to get a
	// second row past the unique slug, so the bare name is the better survivor
	// — and suggesting a merge *into* the parenthesised one reads as a bug even
	// when the tiers above are sound.
	if qa, qb := nameNoise(a.Party.Name), nameNoise(b.Party.Name); qa != qb {
		return qa < qb
	}

	if !a.Party.CreatedAt.Equal(b.Party.CreatedAt) {
		return a.Party.CreatedAt.Before(b.Party.CreatedAt)
	}

	return a.Party.ID.String() < b.Party.ID.String()
}

// nameNoise scores how much disambiguating clutter a name carries, lower being
// cleaner. It only breaks ties between rows the reference counts could not
// separate, so it never overrides how widely an entry is actually used.
func nameNoise(name string) int {
	var noise int

	if parentheticalSuffix.MatchString(strings.ToLower(strings.TrimSpace(name))) {
		noise += 2
	}

	// A slash joins two brands into one row ("WooCommerce / Jetpack"), which is
	// the same kind of clutter as a parenthetical.
	if strings.Contains(name, "/") {
		noise += 2
	}

	// Length breaks the remaining ties, so "h265ify" beats a longer variant
	// that carries no parentheses either.
	return noise*100 + len(name)
}

// enrichmentResolvedFields counts how many fields the last enrichment run
// actually resolved, so a row that filled ten of them outranks one that filled
// one. Reading the run-level status instead would score both merely "partial"
// and let the emptier row win a tie.
//
// A row that was never enriched counts zero.
func enrichmentResolvedFields(e *CatalogEntry) int {
	var resolved int

	for _, field := range parseEnrichmentFields(e.Party.Enrichment) {
		if enrichmentFieldResolved(field.Status) {
			resolved++
		}
	}

	return resolved
}

// enrichmentFieldResolved reports whether a per-field status carries a value,
// as opposed to recording that the enricher looked and found nothing.
func enrichmentFieldResolved(status string) bool {
	switch status {
	case enrichmentFieldStatusFound,
		enrichmentFieldStatusExternal,
		enrichmentFieldStatusFallbackDisplayName:
		return true
	default:
		return false
	}
}
