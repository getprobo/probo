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
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const patternMergeThreshold = 3

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
		"cookie-pattern-analysis-worker",
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

			var patterns coredata.CookiePatterns
			if err := patterns.LoadAllByCookieBannerID(ctx, tx, scope, banner.ID); err != nil {
				return fmt.Errorf("cannot load patterns: %w", err)
			}

			mergeGroups := findMergeGroups(patterns, patternMergeThreshold)

			merged := false
			for key, group := range mergeGroups {

				maxAge := mostCommonMaxAge(group)
				source := bestSource(group)

				prefixPattern := &coredata.CookiePattern{
					ID:               gid.New(banner.ID.TenantID(), coredata.CookiePatternEntityType),
					OrganizationID:   group[0].OrganizationID,
					CookieBannerID:   banner.ID,
					CookieCategoryID: key.categoryID,
					Pattern:          key.prefix,
					MatchType:        coredata.CookiePatternMatchTypePrefix,
					DisplayName:      key.prefix + "*",
					MaxAgeSeconds:    maxAge,
					Description:      "",
					Source:           source,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				}

				inserted, err := prefixPattern.InsertIfNotExists(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert prefix pattern %q: %w", key.prefix, err)
				}
				if !inserted {
					if err := prefixPattern.LoadByBannerIDAndPattern(ctx, tx, scope, banner.ID, key.prefix); err != nil {
						return fmt.Errorf("cannot load existing prefix pattern %q: %w", key.prefix, err)
					}

					if prefixPattern.CookieCategoryID != key.categoryID || prefixPattern.MatchType != coredata.CookiePatternMatchTypePrefix {
						continue
					}
				}

				for _, exactPattern := range group {
					var cookies coredata.Cookies
					if err := cookies.RelinkByCookiePatternID(ctx, tx, scope, exactPattern.ID, prefixPattern.ID); err != nil {
						return fmt.Errorf("cannot relink cookies from pattern %q: %w", exactPattern.Pattern, err)
					}

					if err := exactPattern.Delete(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot delete orphaned exact pattern %q: %w", exactPattern.Pattern, err)
					}
				}

				merged = true
				h.logger.InfoCtx(
					ctx,
					"merged exact patterns into prefix pattern",
					log.String("prefix", key.prefix),
					log.Int("count", len(group)),
					log.String("banner_id", banner.ID.String()),
				)
			}

			if merged {
				if _, err := h.svc.ensureDraftVersionForBanner(ctx, tx, scope, banner.ID); err != nil {
					return fmt.Errorf("cannot ensure draft version: %w", err)
				}
			}

			return nil
		},
	)
}

type mergeGroupKey struct {
	categoryID gid.GID
	prefix     string
}

func findMergeGroups(
	patterns coredata.CookiePatterns,
	threshold int,
) map[mergeGroupKey][]*coredata.CookiePattern {
	var exact []*coredata.CookiePattern
	for _, p := range patterns {
		if p.MatchType == coredata.CookiePatternMatchTypeExact {
			exact = append(exact, p)
		}
	}

	prefixCounts := make(map[mergeGroupKey][]*coredata.CookiePattern)
	for _, p := range exact {
		for _, pfx := range separatorPrefixes(p.Pattern) {
			key := mergeGroupKey{categoryID: p.CookieCategoryID, prefix: pfx}
			prefixCounts[key] = append(prefixCounts[key], p)
		}
	}

	type candidate struct {
		key      mergeGroupKey
		patterns []*coredata.CookiePattern
	}

	var candidates []candidate
	for key, pats := range prefixCounts {
		if len(pats) >= threshold {
			candidates = append(candidates, candidate{key, pats})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return len(candidates[i].key.prefix) > len(candidates[j].key.prefix)
	})

	assigned := make(map[*coredata.CookiePattern]bool)
	groups := make(map[mergeGroupKey][]*coredata.CookiePattern)

	for _, c := range candidates {
		var unassigned []*coredata.CookiePattern
		for _, p := range c.patterns {
			if !assigned[p] {
				unassigned = append(unassigned, p)
			}
		}

		if len(unassigned) < threshold {
			continue
		}

		groups[c.key] = unassigned
		for _, p := range unassigned {
			assigned[p] = true
		}
	}

	return groups
}

func separatorPrefixes(name string) []string {
	var prefixes []string
	for i, ch := range name {
		if ch == '_' || ch == '-' {
			prefixes = append(prefixes, name[:i+1])
		}
	}
	return prefixes
}

func bestSource(patterns []*coredata.CookiePattern) coredata.CookieSource {
	for _, p := range patterns {
		if p.Source == coredata.CookieSourceScript {
			return coredata.CookieSourceScript
		}
	}
	return coredata.CookieSourcePreExisting
}

func mostCommonMaxAge(patterns []*coredata.CookiePattern) *int {
	type key struct {
		valid bool
		val   int
	}
	counts := make(map[key]int)
	for _, p := range patterns {
		k := key{}
		if p.MaxAgeSeconds != nil {
			k = key{valid: true, val: *p.MaxAgeSeconds}
		}
		counts[k]++
	}

	type entry struct {
		k     key
		count int
	}
	entries := make([]entry, 0, len(counts))
	for k, c := range counts {
		entries = append(entries, entry{k, c})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].count > entries[j].count
	})

	if !entries[0].k.valid {
		return nil
	}
	v := entries[0].k.val
	return &v
}
