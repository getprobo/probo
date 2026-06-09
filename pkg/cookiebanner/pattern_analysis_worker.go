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
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	patternMergeThreshold = 3

	// primarySeparators are the structural delimiters that always
	// start a new token in splitTokens. `-` is intentionally excluded
	// because it appears inside UUIDs, which splitTokens preserves as a
	// single token; a value like "done:ecdd43d7-0193-4d24-b6ed-..."
	// must split on ":" so the trailing UUID is isolated and collapsed
	// to a wildcard rather than shredded into fixed hex anchors.
	primarySeparators = "_:."
)

// durationUnits mirrors the snap table from cookie-utils.ts. The same
// tracker observed across different clients can have jitter in its
// max-age (e.g. an "Expires" header computed from Date.now() yields
// slightly different seconds each time). Snapping to the nearest
// human-meaningful unit absorbs that jitter so the patterns still
// merge. This is compliant because the resulting bucket matches the
// duration shown to end users in the cookie banner — two cookies that
// display the same human-readable lifetime will merge, two that
// display differently will not.
var durationUnits = [...]struct {
	seconds int
	snap    int
}{
	{365 * 24 * 3600, 21 * 24 * 3600}, // years, snap +-21 days
	{30 * 24 * 3600, 2 * 24 * 3600},   // months, snap +-2 days
	{7 * 24 * 3600, 12 * 3600},        // weeks, snap +-12 hours
	{24 * 3600, 2 * 3600},             // days, snap +-2 hours
	{3600, 5 * 60},                    // hours, snap +-5 minutes
	{60, 5},                           // minutes, snap +-5 seconds
	{1, 0},                            // seconds, no snap
}

func durationBucket(maxAge *int) int {
	if maxAge == nil || *maxAge <= 0 {
		return -1
	}

	remaining := *maxAge
	total := 0

	for _, u := range durationUnits {
		if remaining >= u.seconds-u.snap {
			count := remaining / u.seconds

			leftover := remaining - count*u.seconds
			if leftover >= u.seconds-u.snap {
				count++
				remaining = 0
			} else if leftover <= u.snap {
				remaining = 0
			} else {
				remaining = leftover
			}

			total += count * u.seconds
		}
	}

	return total
}

type patternAnalysisHandler struct {
	svc    *Service
	pg     *pg.Client
	logger *log.Logger
}

func NewPatternAnalysisWorker(
	svc *Service,
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.CookieBanner] {
	h := &patternAnalysisHandler{
		svc:    svc,
		pg:     pgClient,
		logger: logger,
	}

	return worker.New(
		"tracker-pattern-analysis-worker",
		h,
		logger,
		opts...,
	)
}

func (h *patternAnalysisHandler) Claim(ctx context.Context) (coredata.CookieBanner, error) {
	var banner coredata.CookieBanner

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := banner.LoadNextForPatternAnalysisForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			return banner.ClearPatternAnalysisRequestedAt(ctx, tx)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.CookieBanner{}, worker.ErrNoTask
		}

		return coredata.CookieBanner{}, fmt.Errorf("cannot claim pattern analysis task: %w", err)
	}

	return banner, nil
}

func (h *patternAnalysisHandler) Process(ctx context.Context, banner coredata.CookieBanner) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(banner.ID)

			var exactPatterns coredata.TrackerPatterns
			if err := exactPatterns.LoadAllByCookieBannerID(
				ctx,
				tx,
				scope,
				banner.ID,
				coredata.NewTrackerPatternFilter(new(coredata.TrackerPatternMatchTypeExact), nil, new(false)),
				nil,
			); err != nil {
				return fmt.Errorf("cannot load exact patterns: %w", err)
			}

			mergeGroups := findMergeGroups(exactPatterns, patternMergeThreshold)

			for key, group := range mergeGroups {
				var maxAge *int

				if key.durationBucket >= 0 {
					v := key.durationBucket
					maxAge = &v
				}

				source := bestSource(group)

				// Carry over the resolved org ThirdParty (and its
				// description) from the merged exacts when they agree,
				// so the glob is seeded rather than re-mapped from
				// scratch. Mapping is still re-armed below so the glob
				// derives its own catalog row; the pre-set third party
				// lets the mapping worker skip the expensive org
				// resolution. Only the insert path consumes these: an
				// existing glob is reloaded and keeps its own mapping.
				inheritedThirdPartyID, inheritedDescription := inheritedMapping(group)

				now := time.Now()
				globPattern := &coredata.TrackerPattern{
					ID:                 gid.New(banner.ID.TenantID(), coredata.TrackerPatternEntityType),
					OrganizationID:     group[0].OrganizationID,
					CookieBannerID:     banner.ID,
					CookieCategoryID:   key.categoryID,
					TrackerType:        key.trackerType,
					Pattern:            key.template,
					MatchType:          coredata.TrackerPatternMatchTypeGlob,
					DisplayName:        key.template,
					MaxAgeSeconds:      maxAge,
					ThirdPartyID:       inheritedThirdPartyID,
					Description:        inheritedDescription,
					Source:             source,
					MappingRequestedAt: &now,
					CreatedAt:          now,
					UpdatedAt:          now,
				}

				inserted, err := globPattern.InsertIfNotExists(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert glob pattern %q: %w", key.template, err)
				}

				if !inserted {
					if err := globPattern.LoadByBannerIDTypeAndPattern(ctx, tx, scope, banner.ID, key.trackerType, key.template, maxAge); err != nil {
						return fmt.Errorf("cannot load existing glob pattern %q: %w", key.template, err)
					}

					if globPattern.MatchType != coredata.TrackerPatternMatchTypeGlob || globPattern.CookieCategoryID != key.categoryID {
						// The slot is occupied by an exact pattern or
						// a user-recategorised glob. Skip the relink
						// here so we don't overwrite the user's
						// categorisation; the exact patterns in
						// `group` will be picked up below by
						// adoptUncategorisedPatterns if they live in
						// the uncategorised category and globMatch
						// the existing glob.
						continue
					}

					if shouldPromoteSource(globPattern.Source, source) {
						globPattern.Source = source
						globPattern.UpdatedAt = now

						if err := globPattern.Update(ctx, tx, scope); err != nil {
							return fmt.Errorf("cannot promote source on glob pattern %q: %w", key.template, err)
						}

						// A stronger source can unblock mapping (e.g.
						// EXTENSION->SCRIPT lifts the creationAllowed
						// gate), so re-arm mapping on the existing glob.
						if err := globPattern.SetMappingRequested(ctx, tx); err != nil {
							return fmt.Errorf("cannot request mapping after source promotion on glob pattern %q: %w", key.template, err)
						}
					}
				}

				for _, exactPattern := range group {
					var trackers coredata.DetectedTrackers
					if err := trackers.RelinkByTrackerPatternID(ctx, tx, scope, exactPattern.ID, globPattern.ID); err != nil {
						return fmt.Errorf("cannot relink detected trackers from pattern %q: %w", exactPattern.Pattern, err)
					}

					if err := exactPattern.Delete(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot delete orphaned exact pattern %q: %w", exactPattern.Pattern, err)
					}
				}

				h.logger.InfoCtx(
					ctx,
					"merged exact patterns into glob pattern",
					log.String("template", key.template),
					log.Int("count", len(group)),
					log.Bool("inserted", inserted),
					log.String("banner_id", banner.ID.String()),
				)
			}

			adopted, err := h.adoptUncategorisedPatterns(ctx, tx, scope, banner)
			if err != nil {
				return fmt.Errorf("cannot adopt uncategorised patterns: %w", err)
			}

			var patterns coredata.TrackerPatterns
			if err := patterns.RefreshLastMatchedAtByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
				return fmt.Errorf("cannot refresh last_matched_at: %w", err)
			}

			// Merging exact patterns into a glob in the same category
			// does not change visitor consent for those identifiers
			// (findMergeGroups keys on category, so every member of a
			// group is already under key.categoryID). Adoption is the
			// only operation in this worker that moves trackers
			// between categories and therefore changes consent.
			if adopted {
				if _, err := h.svc.ensureDraftVersionForBanner(ctx, tx, scope, banner.ID); err != nil {
					return fmt.Errorf("cannot ensure draft version: %w", err)
				}
			}

			return nil
		},
	)
}

type mergeGroupKey struct {
	categoryID     gid.GID
	trackerType    coredata.TrackerType
	template       string
	durationBucket int
}

func findMergeGroups(
	patterns coredata.TrackerPatterns,
	threshold int,
) map[mergeGroupKey][]*coredata.TrackerPattern {
	type memberKey struct {
		groupKey mergeGroupKey
		pattern  *coredata.TrackerPattern
	}

	templateCounts := make(map[mergeGroupKey][]*coredata.TrackerPattern)
	heuristicKeys := make(map[mergeGroupKey]bool)
	seen := make(map[memberKey]bool)

	for _, p := range patterns {
		bucket := durationBucket(p.MaxAgeSeconds)

		if tmpl, ok := heuristicTemplate(p.Pattern); ok {
			key := mergeGroupKey{categoryID: p.CookieCategoryID, trackerType: p.TrackerType, template: tmpl, durationBucket: bucket}

			mk := memberKey{key, p}
			if !seen[mk] {
				seen[mk] = true

				templateCounts[key] = append(templateCounts[key], p)
			}

			heuristicKeys[key] = true
		}

		for _, tmpl := range templateCandidates(p.Pattern) {
			key := mergeGroupKey{categoryID: p.CookieCategoryID, trackerType: p.TrackerType, template: tmpl, durationBucket: bucket}

			mk := memberKey{key, p}
			if !seen[mk] {
				seen[mk] = true

				templateCounts[key] = append(templateCounts[key], p)
			}
		}
	}

	type candidate struct {
		key         mergeGroupKey
		fixedChars  int
		isHeuristic bool
		patterns    []*coredata.TrackerPattern
	}

	var candidates []candidate

	for key, pats := range templateCounts {
		isH := heuristicKeys[key]

		effectiveThreshold := threshold
		if isH {
			effectiveThreshold = 1
		}

		if len(pats) >= effectiveThreshold {
			candidates = append(candidates, candidate{key, len(strings.ReplaceAll(key.template, "*", "")), isH, pats})
		}
	}

	// Sort: heuristic first, then descending specificity (more fixed
	// characters), then descending coverage, then template name for a
	// fully deterministic order.
	sort.Slice(
		candidates,
		func(i, j int) bool {
			if candidates[i].isHeuristic != candidates[j].isHeuristic {
				return candidates[i].isHeuristic
			}

			if candidates[i].fixedChars != candidates[j].fixedChars {
				return candidates[i].fixedChars > candidates[j].fixedChars
			}

			if len(candidates[i].patterns) != len(candidates[j].patterns) {
				return len(candidates[i].patterns) > len(candidates[j].patterns)
			}

			return candidates[i].key.template < candidates[j].key.template
		},
	)

	assigned := make(map[*coredata.TrackerPattern]bool)
	groups := make(map[mergeGroupKey][]*coredata.TrackerPattern)

	for _, c := range candidates {
		effectiveThreshold := threshold
		if c.isHeuristic {
			effectiveThreshold = 1
		}

		var unassigned []*coredata.TrackerPattern

		for _, p := range c.patterns {
			if !assigned[p] {
				unassigned = append(unassigned, p)
			}
		}

		if len(unassigned) < effectiveThreshold {
			continue
		}

		groups[c.key] = unassigned
		for _, p := range unassigned {
			assigned[p] = true
		}
	}

	return groups
}

func heuristicTemplate(name string) (string, bool) {
	tokens, seps := splitTokens(name)
	if len(seps) == 0 {
		return "", false
	}

	// Trim leading empty tokens (e.g. "__Secure-..." yields ["", "", ...]).
	var prefix strings.Builder
	for len(tokens) > 1 && tokens[0] == "" {
		prefix.WriteString(string(seps[0]))

		tokens = tokens[1:]
		seps = seps[1:]
	}

	// Trim trailing empty tokens.
	var suffix string
	for len(tokens) > 1 && tokens[len(tokens)-1] == "" {
		suffix = string(seps[len(seps)-1]) + suffix
		tokens = tokens[:len(tokens)-1]
		seps = seps[:len(seps)-1]
	}

	if len(seps) == 0 {
		return "", false
	}

	changed := false

	var (
		resultTokens []string
		resultSeps   []byte
	)

	for i, t := range tokens {
		if looksVariable(t) {
			changed = true

			if len(resultTokens) == 0 || resultTokens[len(resultTokens)-1] != "*" {
				if i > 0 {
					resultSeps = append(resultSeps, seps[i-1])
				}

				resultTokens = append(resultTokens, "*")
			}
		} else {
			if i > 0 {
				resultSeps = append(resultSeps, seps[i-1])
			}

			resultTokens = append(resultTokens, t)
		}
	}

	if !changed {
		return "", false
	}

	tmpl := prefix.String() + joinTokens(resultTokens, resultSeps) + suffix
	if !templateHasFixedAnchor(tmpl) {
		return "", false
	}

	return tmpl, true
}

func templateCandidates(name string) []string {
	var candidates []string

	for i, ch := range name {
		if ch == '_' || ch == '-' {
			tmpl := name[:i+1] + "*"
			if templateHasFixedAnchor(tmpl) {
				candidates = append(candidates, tmpl)
			}
		}
	}

	tokens, seps := splitTokens(name)
	if len(tokens) >= 3 && len(seps) > 0 {
		for pos := 1; pos < len(tokens)-1; pos++ {
			left := joinTokens(tokens[:pos], seps[:pos-1])
			right := joinTokens(tokens[pos+1:], seps[pos+1:])

			tmpl := left + string(seps[pos-1]) + "*" + string(seps[pos]) + right
			if templateHasFixedAnchor(tmpl) {
				candidates = append(candidates, tmpl)
			}
		}
	}

	return candidates
}

func looksVariable(token string) bool {
	if len(token) == 0 {
		return false
	}

	hasDigit := false
	hasLetter := false
	allHex := true
	allDigits := true

	for _, ch := range token {
		switch {
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case ch >= 'a' && ch <= 'f', ch >= 'A' && ch <= 'F':
			hasLetter = true
			allDigits = false
		case ch >= 'g' && ch <= 'z', ch >= 'G' && ch <= 'Z':
			hasLetter = true
			allHex = false
			allDigits = false
		case ch == '-':
			allHex = false
			allDigits = false
		default:
			allHex = false
			allDigits = false
		}
	}

	if len(token) >= 8 && hasDigit && hasLetter {
		return true
	}

	if len(token) >= 16 && allHex && hasDigit {
		return true
	}

	if isUUIDShape(token) {
		return true
	}

	if len(token) >= 8 && allDigits {
		return true
	}

	return false
}

func isUUIDShape(s string) bool {
	if len(s) != 36 {
		return false
	}

	for i, ch := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if ch != '-' {
				return false
			}

			continue
		}

		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') && (ch < 'A' || ch > 'F') {
			return false
		}
	}

	return true
}

func splitTokens(name string) ([]string, []byte) {
	primaryParts, primarySeps := splitOnAny(name, primarySeparators)

	var (
		tokens []string
		seps   []byte
	)

	for i, part := range primaryParts {
		if i > 0 {
			seps = append(seps, primarySeps[i-1])
		}

		if isUUIDShape(part) || !strings.Contains(part, "-") {
			tokens = append(tokens, part)
		} else {
			for j, sub := range strings.Split(part, "-") {
				if j > 0 {
					seps = append(seps, '-')
				}

				tokens = append(tokens, sub)
			}
		}
	}

	if len(seps) == 0 {
		return []string{name}, nil
	}

	return tokens, seps
}

// splitOnAny splits s on every byte found in separators, returning the
// parts and the separator byte that preceded each part after the first.
// len(seps) == len(parts)-1.
func splitOnAny(s, separators string) ([]string, []byte) {
	var (
		parts []string
		seps  []byte
		start int
	)

	for i := 0; i < len(s); i++ {
		if strings.IndexByte(separators, s[i]) >= 0 {
			parts = append(parts, s[start:i])
			seps = append(seps, s[i])
			start = i + 1
		}
	}

	parts = append(parts, s[start:])

	return parts, seps
}

func joinTokens(tokens []string, seps []byte) string {
	var b strings.Builder

	for i, t := range tokens {
		if i > 0 {
			b.WriteByte(seps[i-1])
		}

		b.WriteString(t)
	}

	return b.String()
}

// templateHasFixedAnchor reports whether tmpl contains at least one
// character beyond separators and wildcards. Templates like "_*",
// "__*", "-*", "--*", "__*__" would merge unrelated third parties
// (e.g. __support__, __darkreader__wasEnabledForHost,
// __EXT_APP_REFRESH_BLACK_SUB_DOMAINS__) under a single glob, so
// candidates without any fixed alphanumeric anchor are rejected. The
// separators recognised here mirror primarySeparators (plus the
// UUID-internal "-") so a separator-only template such as ":.*" is
// rejected too.
func templateHasFixedAnchor(tmpl string) bool {
	for _, ch := range tmpl {
		if ch != '*' && ch != '-' && !strings.ContainsRune(primarySeparators, ch) {
			return true
		}
	}

	return false
}

func globMatch(pattern, name string) bool {
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == name
	}

	if !strings.HasPrefix(name, parts[0]) {
		return false
	}

	name = name[len(parts[0]):]

	last := parts[len(parts)-1]
	if !strings.HasSuffix(name, last) {
		return false
	}

	name = name[:len(name)-len(last)]

	for _, part := range parts[1 : len(parts)-1] {
		idx := strings.Index(name, part)
		if idx == -1 {
			return false
		}

		name = name[idx+len(part):]
	}

	return true
}

// sourceRank converts a CookieSource into a comparable rank that
// reflects signal strength: SCRIPT > EXTENSION > PRE_EXISTING. HTTP
// and nil collapse into the PRE_EXISTING rank because bestSource
// already normalises them; if a future caller hands us either, the
// ranking still produces a sane "no promotion" outcome against
// PRE_EXISTING/EXTENSION/SCRIPT existing values.
func sourceRank(s *coredata.CookieSource) int {
	if s == nil {
		return 0
	}

	switch *s {
	case coredata.CookieSourceScript:
		return 2
	case coredata.CookieSourceExtension:
		return 1
	default:
		return 0
	}
}

// shouldPromoteSource reports whether candidate represents a stronger
// signal than existing under the SCRIPT > EXTENSION > PRE_EXISTING
// precedence used across the cookie-banner pipeline. Equal ranks do
// not promote so we avoid pointless writes.
func shouldPromoteSource(existing, candidate *coredata.CookieSource) bool {
	return sourceRank(candidate) > sourceRank(existing)
}

// bestSource rolls up the source values of a group of exact patterns
// being merged into a single glob. Precedence is SCRIPT > EXTENSION
// > PRE_EXISTING, mirroring both the page-script-wins rule in
// detected_trackers and the asymmetric signal strength of each
// bucket: SCRIPT is high-confidence page evidence (a real page
// tracker), EXTENSION is high-confidence extension evidence, and
// PRE_EXISTING is the catch-all that may include extension state
// injected before SDK load. HTTP and nil collapse into PRE_EXISTING
// here, preserving the original two-value rollup behaviour for
// non-script values.
func bestSource(patterns []*coredata.TrackerPattern) *coredata.CookieSource {
	var hasExtension bool

	for _, p := range patterns {
		if p.Source == nil {
			continue
		}

		switch *p.Source {
		case coredata.CookieSourceScript:
			return p.Source
		case coredata.CookieSourceExtension:
			hasExtension = true
		}
	}

	if hasExtension {
		src := coredata.CookieSourceExtension
		return &src
	}

	src := coredata.CookieSourcePreExisting

	return &src
}

// inheritedMapping rolls up the resolved org ThirdParty of a group of
// exact patterns being merged into a glob, so the glob can be seeded
// instead of re-mapped from scratch. It returns a third party only when
// every resolved member agrees on a single id: a conflicting group (or
// one with no resolved member) returns nil, leaving the glob blank for a
// fresh mapping pass. When a third party is chosen, the description of
// the first member carrying that same id with non-empty text is returned
// too; the catalog link (common_tracker_pattern_id) is deliberately not
// inherited, since it is keyed on the exact pattern string rather than
// the glob template and the mapping worker derives the right row itself.
func inheritedMapping(patterns []*coredata.TrackerPattern) (*gid.GID, string) {
	var thirdPartyID *gid.GID

	for _, p := range patterns {
		if p.ThirdPartyID == nil {
			continue
		}

		if thirdPartyID == nil {
			thirdPartyID = p.ThirdPartyID
			continue
		}

		if *thirdPartyID != *p.ThirdPartyID {
			return nil, ""
		}
	}

	if thirdPartyID == nil {
		return nil, ""
	}

	for _, p := range patterns {
		if p.ThirdPartyID != nil && *p.ThirdPartyID == *thirdPartyID && p.Description != "" {
			return thirdPartyID, p.Description
		}
	}

	return thirdPartyID, ""
}

func (h *patternAnalysisHandler) adoptUncategorisedPatterns(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	banner coredata.CookieBanner,
) (bool, error) {
	var uncategorised coredata.CookieCategory
	if err := uncategorised.LoadUncategorisedByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("cannot load uncategorised category: %w", err)
	}

	var globPatterns coredata.TrackerPatterns
	if err := globPatterns.LoadAllByCookieBannerID(
		ctx,
		tx,
		scope,
		banner.ID,
		coredata.NewTrackerPatternFilter(new(coredata.TrackerPatternMatchTypeGlob), nil, new(false)),
		nil,
	); err != nil {
		return false, fmt.Errorf("cannot load glob patterns: %w", err)
	}

	if len(globPatterns) == 0 {
		return false, nil
	}

	sort.Slice(
		globPatterns,
		func(i, j int) bool {
			return len(globPatterns[i].Pattern) > len(globPatterns[j].Pattern)
		},
	)

	exactMatchType := coredata.TrackerPatternMatchTypeExact

	var uncategorisedExact coredata.TrackerPatterns
	if err := uncategorisedExact.LoadAllByCookieBannerID(
		ctx,
		tx,
		scope,
		banner.ID,
		coredata.NewTrackerPatternFilter(&exactMatchType, &uncategorised.ID, new(false)),
		nil,
	); err != nil {
		return false, fmt.Errorf("cannot load uncategorised exact patterns: %w", err)
	}

	adopted := false

	for _, ep := range uncategorisedExact {
		var match *coredata.TrackerPattern

		epBucket := durationBucket(ep.MaxAgeSeconds)
		for _, gp := range globPatterns {
			if ep.TrackerType == gp.TrackerType && globMatch(gp.Pattern, ep.Pattern) && durationBucket(gp.MaxAgeSeconds) == epBucket {
				match = gp
				break
			}
		}

		if match == nil {
			continue
		}

		var trackers coredata.DetectedTrackers
		if err := trackers.RelinkByTrackerPatternID(ctx, tx, scope, ep.ID, match.ID); err != nil {
			return false, fmt.Errorf("cannot relink detected trackers from pattern %q: %w", ep.Pattern, err)
		}

		// Promote the glob's source if this exact carries a
		// stronger signal. The merge loop already handles
		// promotion when the glob shares the exact's category,
		// but adoption is the only path that can lift a glob
		// the user (or an earlier pass) has placed in a
		// non-uncategorised category. Without this, a
		// PRE_EXISTING glob never advances to SCRIPT/EXTENSION
		// even though new SDK-observed exacts confirm the
		// stronger signal. We mutate match.Source in place
		// before calling Update so subsequent adoptions against
		// the same glob ratchet correctly (PRE_EXISTING →
		// EXTENSION → SCRIPT) without redundant writes.
		if shouldPromoteSource(match.Source, ep.Source) {
			match.Source = ep.Source
			match.UpdatedAt = time.Now()

			if err := match.Update(ctx, tx, scope); err != nil {
				return false, fmt.Errorf("cannot promote source on glob pattern %q: %w", match.Pattern, err)
			}

			// A stronger source can unblock mapping (e.g.
			// EXTENSION->SCRIPT lifts the creationAllowed gate), so
			// re-arm mapping on the adopted glob.
			if err := match.SetMappingRequested(ctx, tx); err != nil {
				return false, fmt.Errorf("cannot request mapping after source promotion on glob pattern %q: %w", match.Pattern, err)
			}
		}

		if err := ep.Delete(ctx, tx, scope); err != nil {
			return false, fmt.Errorf("cannot delete adopted exact pattern %q: %w", ep.Pattern, err)
		}

		adopted = true

		h.logger.InfoCtx(
			ctx,
			"adopted uncategorised exact pattern into glob pattern",
			log.String("exact_pattern", ep.Pattern),
			log.String("glob_pattern", match.Pattern),
			log.String("banner_id", banner.ID.String()),
		)
	}

	return adopted, nil
}
