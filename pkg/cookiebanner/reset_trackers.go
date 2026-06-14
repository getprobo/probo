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
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// ResetTrackersResult summarizes what a banner reset changed.
type ResetTrackersResult struct {
	PatternsReset      int64
	GlobsDecomposed    int
	ExactsCreated      int
	DetectionsRelinked int
	AnalysisRequested  bool
}

// ResetBannerTrackers re-arms the tracker pipeline for a banner's
// uncategorised, non-excluded patterns. It is an operator action
// (proboctl), tenant-scoped via the provided coredata.Predicater.
//
// With mappingOnly, it only clears each pattern's catalog/vendor links
// and re-arms mapping, for iterating on the mapping agent without
// touching analysis.
//
// The full reset additionally rebuilds the raw exact patterns from the
// surviving detected_trackers and re-arms pattern analysis, so the
// analysis worker re-derives globs from scratch: the pattern-analysis
// worker consumes (deletes) exact patterns when it merges them into
// globs, so the only way to re-run analysis is to reconstruct the exacts
// from detections. Each uncategorised, non-excluded glob is decomposed -
// every detection it covers becomes (or rejoins) an exact pattern keyed
// by its identifier - and the now-empty glob is deleted. User-categorised
// and excluded patterns are never touched.
//
// When keyword is non-nil and non-empty, the reset is scoped to patterns
// whose pattern or display name contains it (case-insensitive): only
// matching globs are decomposed and only matching patterns are re-armed
// for mapping. The banner-wide pattern-analysis re-arm is unaffected.
func ResetBannerTrackers(
	ctx context.Context,
	pgClient *pg.Client,
	predicate coredata.Predicater,
	bannerID gid.GID,
	mappingOnly bool,
	keyword *string,
) (ResetTrackersResult, error) {
	var result ResetTrackersResult

	err := pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var uncategorised coredata.CookieCategory
			if err := uncategorised.LoadUncategorisedByCookieBannerID(ctx, tx, predicate, bannerID); err != nil {
				return fmt.Errorf("cannot load uncategorised category: %w", err)
			}

			if !mappingOnly {
				if err := decomposeGlobs(ctx, tx, predicate, bannerID, uncategorised.ID, keyword, &result); err != nil {
					return err
				}
			}

			var patterns coredata.TrackerPatterns

			reset, err := patterns.ResetAndRequestMappingByCookieCategoryID(ctx, tx, predicate, uncategorised.ID, keyword)
			if err != nil {
				return fmt.Errorf("cannot reset and request mapping: %w", err)
			}

			result.PatternsReset = reset

			if !mappingOnly {
				banner := coredata.CookieBanner{ID: bannerID}
				if err := banner.SetPatternAnalysisRequested(ctx, tx); err != nil {
					return fmt.Errorf("cannot request pattern analysis: %w", err)
				}

				result.AnalysisRequested = true
			}

			return nil
		},
	)
	if err != nil {
		return ResetTrackersResult{}, err
	}

	return result, nil
}

// decomposeGlobs turns every uncategorised, non-excluded glob pattern of
// the banner back into exact patterns derived from its detected trackers,
// relinking each detection to its exact and deleting the emptied glob.
func decomposeGlobs(
	ctx context.Context,
	tx pg.Tx,
	predicate coredata.Predicater,
	bannerID gid.GID,
	uncategorisedID gid.GID,
	keyword *string,
	result *ResetTrackersResult,
) error {
	globMatchType := coredata.TrackerPatternMatchTypeGlob
	notExcluded := false

	var globs coredata.TrackerPatterns
	if err := globs.LoadAllByCookieBannerID(
		ctx,
		tx, predicate, bannerID,
		coredata.NewTrackerPatternFilter(&globMatchType, &uncategorisedID, &notExcluded).WithPatternKeyword(keyword),
		nil,
	); err != nil {
		return fmt.Errorf("cannot load glob patterns: %w", err)
	}

	for _, glob := range globs {
		var detections coredata.DetectedTrackers
		if err := detections.LoadAllByTrackerPatternID(ctx, tx, predicate, glob.ID); err != nil {
			return fmt.Errorf("cannot load detections for glob %q: %w", glob.Pattern, err)
		}

		for _, detection := range detections {
			exactID, created, err := ensureExactPattern(ctx, tx, predicate, glob, uncategorisedID, detection)
			if err != nil {
				return err
			}

			if created {
				result.ExactsCreated++
			}

			detection.TrackerPatternID = &exactID
			if err := detection.UpdateTrackerPatternID(ctx, tx, predicate); err != nil {
				return fmt.Errorf("cannot relink detection %s: %w", detection.ID, err)
			}

			result.DetectionsRelinked++
		}

		if err := glob.Delete(ctx, tx, predicate); err != nil {
			return fmt.Errorf("cannot delete glob pattern %q: %w", glob.Pattern, err)
		}

		result.GlobsDecomposed++
	}

	return nil
}

// ensureExactPattern finds or creates the exact pattern for a detection
// (keyed by banner, tracker type, identifier, and max-age) in the
// uncategorised category, returning its id and whether it was created.
func ensureExactPattern(
	ctx context.Context,
	tx pg.Tx,
	predicate coredata.Predicater,
	glob *coredata.TrackerPattern,
	uncategorisedID gid.GID,
	detection *coredata.DetectedTracker,
) (gid.GID, bool, error) {
	now := time.Now()

	exact := &coredata.TrackerPattern{
		ID:                 gid.New(glob.CookieBannerID.TenantID(), coredata.TrackerPatternEntityType),
		OrganizationID:     glob.OrganizationID,
		CookieBannerID:     glob.CookieBannerID,
		CookieCategoryID:   uncategorisedID,
		TrackerType:        detection.TrackerType,
		Pattern:            detection.Identifier,
		MatchType:          coredata.TrackerPatternMatchTypeExact,
		DisplayName:        detection.Identifier,
		MaxAgeSeconds:      detection.MaxAgeSeconds,
		Source:             detection.Source,
		MappingRequestedAt: &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	created, err := exact.InsertIfNotExists(ctx, tx, predicate)
	if err != nil {
		return gid.GID{}, false, fmt.Errorf("cannot insert exact pattern %q: %w", detection.Identifier, err)
	}

	if !created {
		if err := exact.LoadByBannerIDTypeAndPattern(ctx, tx, predicate, glob.CookieBannerID, detection.TrackerType, detection.Identifier, detection.MaxAgeSeconds); err != nil {
			return gid.GID{}, false, fmt.Errorf("cannot load existing exact pattern %q: %w", detection.Identifier, err)
		}
	}

	return exact.ID, created, nil
}
