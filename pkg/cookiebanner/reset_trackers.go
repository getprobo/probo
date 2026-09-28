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

package cookiebanner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

const decomposeGlobAttempts = 3

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
// (proboctl), tenant-scoped via the provided Scoper.
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
// from detections. Each uncategorised, non-excluded glob is decomposed
// in its own transaction — every detection it covers becomes (or
// rejoins) an exact pattern keyed by its identifier — and the now-empty
// glob is deleted. Per-glob commits keep lock lifetime short so a
// long-running reset does not deadlock with live detection reporting.
// User-categorised and excluded patterns are never touched.
//
// When keyword is non-nil and non-empty, the reset is scoped to patterns
// whose pattern or display name contains it (case-insensitive): only
// matching globs are decomposed and only matching patterns are re-armed
// for mapping. The banner-wide pattern-analysis re-arm is unaffected.
func ResetBannerTrackers(
	ctx context.Context,
	pgClient *pg.Client,
	scope coredata.Scoper,
	bannerID gid.GID,
	mappingOnly bool,
	keyword *string,
) (ResetTrackersResult, error) {
	var (
		result          ResetTrackersResult
		uncategorisedID gid.GID
	)

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var uncategorised coredata.CookieCategory
			if err := uncategorised.LoadUncategorisedByCookieBannerID(ctx, conn, scope, bannerID); err != nil {
				return fmt.Errorf("cannot load uncategorised category: %w", err)
			}

			uncategorisedID = uncategorised.ID

			return nil
		},
	)
	if err != nil {
		return ResetTrackersResult{}, err
	}

	if !mappingOnly {
		if err := decomposeGlobs(ctx, pgClient, scope, bannerID, uncategorisedID, keyword, &result); err != nil {
			return ResetTrackersResult{}, err
		}
	}

	err = pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var patterns coredata.TrackerPatterns

			reset, err := patterns.ResetAndRequestMappingByCookieCategoryID(ctx, tx, scope, uncategorisedID, keyword)
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
// Listing is a read; each glob is locked and rewritten in its own
// transaction.
func decomposeGlobs(
	ctx context.Context,
	pgClient *pg.Client,
	scope coredata.Scoper,
	bannerID gid.GID,
	uncategorisedID gid.GID,
	keyword *string,
	result *ResetTrackersResult,
) error {
	globMatchType := coredata.TrackerPatternMatchTypeGlob
	notExcluded := false

	return page.WalkAll(
		ctx,
		page.OrderBy[coredata.TrackerPatternOrderField]{
			Field:     coredata.TrackerPatternOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.TrackerPatternOrderField]) ([]*coredata.TrackerPattern, error) {
			var batch coredata.TrackerPatterns

			err := pgClient.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					if err := batch.LoadByCookieBannerID(ctx, conn, scope, bannerID, cursor, coredata.NewTrackerPatternFilter(&globMatchType, &uncategorisedID, &notExcluded).WithPatternKeyword(keyword)); err != nil {
						return fmt.Errorf("cannot load glob patterns: %w", err)
					}

					return nil
				},
			)
			if err != nil {
				return nil, err
			}

			return batch, nil
		},
		func(globs []*coredata.TrackerPattern) error {
			for _, glob := range globs {
				if err := decomposeGlob(ctx, pgClient, scope, uncategorisedID, glob, result); err != nil {
					return err
				}
			}

			return nil
		},
	)
}

func decomposeGlob(
	ctx context.Context,
	pgClient *pg.Client,
	scope coredata.Scoper,
	uncategorisedID gid.GID,
	glob *coredata.TrackerPattern,
	result *ResetTrackersResult,
) error {
	for attempt := 1; attempt <= decomposeGlobAttempts; attempt++ {
		var (
			exactsCreated      int
			detectionsRelinked int
			decomposed         bool
		)

		err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				locked := coredata.TrackerPattern{ID: glob.ID}
				if err := locked.LoadByIDForUpdate(ctx, tx, scope, glob.ID); err != nil {
					if errors.Is(err, coredata.ErrResourceNotFound) {
						return nil
					}

					return fmt.Errorf("cannot lock glob pattern %q: %w", glob.Pattern, err)
				}

				if locked.MatchType != coredata.TrackerPatternMatchTypeGlob ||
					locked.Excluded ||
					locked.CookieCategoryID != uncategorisedID {
					return nil
				}

				err := page.WalkAll(
					ctx,
					page.OrderBy[coredata.DetectedTrackerOrderField]{
						Field:     coredata.DetectedTrackerOrderFieldLastDetectedAt,
						Direction: page.OrderDirectionAsc,
					},
					func(ctx context.Context, cursor *page.Cursor[coredata.DetectedTrackerOrderField]) ([]*coredata.DetectedTracker, error) {
						var batch coredata.DetectedTrackers
						if err := batch.LoadByTrackerPatternID(ctx, tx, scope, locked.ID, cursor); err != nil {
							return nil, fmt.Errorf("cannot load detections for glob %q: %w", locked.Pattern, err)
						}

						return batch, nil
					},
					func(detections []*coredata.DetectedTracker) error {
						for _, detection := range detections {
							exactID, created, err := ensureExactPattern(ctx, tx, scope, &locked, uncategorisedID, detection)
							if err != nil {
								return err
							}

							if created {
								exactsCreated++
							}

							detection.TrackerPatternID = &exactID
							if err := detection.UpdateTrackerPatternID(ctx, tx, scope); err != nil {
								return fmt.Errorf("cannot relink detection %s: %w", detection.ID, err)
							}

							detectionsRelinked++
						}

						return nil
					},
				)
				if err != nil {
					return err
				}

				if err := locked.Delete(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot delete glob pattern %q: %w", locked.Pattern, err)
				}

				decomposed = true

				return nil
			},
		)
		if err != nil {
			if attempt < decomposeGlobAttempts && isDeadlock(err) {
				continue
			}

			return err
		}

		if decomposed {
			result.GlobsDecomposed++
			result.ExactsCreated += exactsCreated
			result.DetectionsRelinked += detectionsRelinked
		}

		return nil
	}

	return nil
}

func isDeadlock(err error) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)

	return ok && pgErr.Code == "40P01"
}

// ensureExactPattern finds or creates the exact pattern for a detection
// (keyed by banner, tracker type, identifier, and max-age) in the
// uncategorised category, returning its id and whether it was created.
func ensureExactPattern(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
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

	created, err := exact.InsertIfNotExists(ctx, tx, scope)
	if err != nil {
		return gid.GID{}, false, fmt.Errorf("cannot insert exact pattern %q: %w", detection.Identifier, err)
	}

	if !created {
		if err := exact.LoadByBannerIDTypeAndPattern(ctx, tx, scope, glob.CookieBannerID, detection.TrackerType, detection.Identifier, detection.MaxAgeSeconds); err != nil {
			return gid.GID{}, false, fmt.Errorf("cannot load existing exact pattern %q: %w", detection.Identifier, err)
		}
	}

	return exact.ID, created, nil
}
