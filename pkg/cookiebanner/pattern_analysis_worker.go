// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

const patternMergeThreshold = 3

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

			var uncategorised coredata.CookieCategory
			hasUncategorised := true
			if err := uncategorised.LoadUncategorisedByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load uncategorised category: %w", err)
				}
				hasUncategorised = false
			}

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

			consentChanged := false
			for key, group := range mergeGroups {
				var maxAge *int
				if key.durationBucket >= 0 {
					v := key.durationBucket
					maxAge = &v
				}

				source := bestSource(group)

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
					Description:        "",
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

					if globPattern.CookieCategoryID != key.categoryID || globPattern.MatchType != coredata.TrackerPatternMatchTypeGlob {
						continue
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

				if !hasUncategorised || key.categoryID != uncategorised.ID {
					consentChanged = true
				}

				h.logger.InfoCtx(
					ctx,
					"merged exact patterns into glob pattern",
					log.String("template", key.template),
					log.Int("count", len(group)),
					log.String("banner_id", banner.ID.String()),
				)
			}

			if _, err := h.adoptUncategorisedPatterns(ctx, tx, scope, banner); err != nil {
				return fmt.Errorf("cannot adopt uncategorised patterns: %w", err)
			}

			var patterns coredata.TrackerPatterns
			if err := patterns.RefreshLastMatchedAtByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
				return fmt.Errorf("cannot refresh last_matched_at: %w", err)
			}

			if consentChanged {
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
	sort.Slice(candidates, func(i, j int) bool {
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
	})

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
	tokens, sep := splitTokens(name)
	if sep == 0 {
		return "", false
	}

	s := string(sep)
	changed := false
	var result []string
	for _, t := range tokens {
		if looksVariable(t) {
			changed = true
			if len(result) == 0 || result[len(result)-1] != "*" {
				result = append(result, "*")
			}
		} else {
			result = append(result, t)
		}
	}

	if !changed {
		return "", false
	}

	return strings.Join(result, s), true
}

func templateCandidates(name string) []string {
	var candidates []string

	for i, ch := range name {
		if ch == '_' || ch == '-' {
			candidates = append(candidates, name[:i+1]+"*")
		}
	}

	tokens, sep := splitTokens(name)
	if len(tokens) >= 3 && sep != 0 {
		s := string(sep)
		for pos := 1; pos < len(tokens)-1; pos++ {
			tmpl := strings.Join(tokens[:pos], s) + s + "*" + s + strings.Join(tokens[pos+1:], s)
			candidates = append(candidates, tmpl)
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

func splitTokens(name string) ([]string, byte) {
	if found := strings.Contains(name, "_"); found {
		return strings.Split(name, "_"), '_'
	}
	if found := strings.Contains(name, "-"); found {
		return strings.Split(name, "-"), '-'
	}
	return []string{name}, 0
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

func bestSource(patterns []*coredata.TrackerPattern) *coredata.CookieSource {
	for _, p := range patterns {
		if p.Source != nil && *p.Source == coredata.CookieSourceScript {
			return p.Source
		}
	}
	src := coredata.CookieSourcePreExisting
	return &src
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

	sort.Slice(globPatterns, func(i, j int) bool {
		return len(globPatterns[i].Pattern) > len(globPatterns[j].Pattern)
	})

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
