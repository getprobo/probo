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

package coredata

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type SamplePatternHit struct {
	BrowserFamily   BrowserFamily `db:"browser_family"`
	Presentation    string        `db:"presentation"`
	ConsentAction   string        `db:"consent_action"`
	CategoryGranted bool          `db:"category_granted"`
	Count           int           `db:"count"`
}

func RecordSamplePatternHit(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	sampleID gid.GID,
	cookieBannerID gid.GID,
	trackerPatternID gid.GID,
	family BrowserFamily,
	presentation string,
	consentAction string,
	categoryGranted bool,
	count int64,
) error {
	if count == 0 {
		return nil
	}

	now := time.Now()

	q := `
WITH open_sample AS (
	SELECT id
	FROM cookie_banner_samples
	WHERE id = @sample_id
	  AND frozen_at IS NULL
	FOR UPDATE
)
INSERT INTO tracker_pattern_sample_hits (
	sample_id,
	tracker_pattern_id,
	browser_family,
	presentation,
	consent_action,
	category_granted,
	tenant_id,
	cookie_banner_id,
	count,
	created_at,
	updated_at
)
SELECT
	@sample_id,
	@tracker_pattern_id,
	@browser_family,
	@presentation,
	@consent_action,
	@category_granted,
	@tenant_id,
	@cookie_banner_id,
	@count,
	@now,
	@now
FROM open_sample
WHERE EXISTS (
	SELECT 1
	FROM tracker_patterns
	WHERE id = @tracker_pattern_id
)
ON CONFLICT (
	sample_id,
	tracker_pattern_id,
	browser_family,
	presentation,
	consent_action,
	category_granted
) DO UPDATE
SET
	count = tracker_pattern_sample_hits.count + EXCLUDED.count,
	updated_at = EXCLUDED.updated_at
WHERE EXISTS (
	SELECT 1
	FROM cookie_banner_samples
	WHERE id = tracker_pattern_sample_hits.sample_id
	  AND frozen_at IS NULL
)
`

	args := pgx.StrictNamedArgs{
		"sample_id":          sampleID,
		"tracker_pattern_id": trackerPatternID,
		"cookie_banner_id":   cookieBannerID,
		"tenant_id":          scope.GetTenantID(),
		"browser_family":     family,
		"presentation":       presentation,
		"consent_action":     consentAction,
		"category_granted":   categoryGranted,
		"count":              count,
		"now":                now,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot record sample pattern hit: %w", err)
	}

	return nil
}

func AddSampleHits(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	exactPatternID gid.GID,
	globPatternID gid.GID,
) error {
	now := time.Now()

	q := `
INSERT INTO tracker_pattern_sample_hits (
	sample_id,
	tracker_pattern_id,
	browser_family,
	presentation,
	consent_action,
	category_granted,
	tenant_id,
	cookie_banner_id,
	count,
	created_at,
	updated_at
)
SELECT
	exact.sample_id,
	@glob_id,
	exact.browser_family,
	exact.presentation,
	exact.consent_action,
	exact.category_granted,
	exact.tenant_id,
	exact.cookie_banner_id,
	exact.count,
	@now,
	@now
FROM tracker_pattern_sample_hits exact
WHERE exact.tracker_pattern_id = @exact_id
  AND exact.tenant_id = @tenant_id
ON CONFLICT (
	sample_id,
	tracker_pattern_id,
	browser_family,
	presentation,
	consent_action,
	category_granted
) DO UPDATE
SET
	count = tracker_pattern_sample_hits.count + EXCLUDED.count,
	updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"exact_id":  exactPatternID,
		"glob_id":   globPatternID,
		"tenant_id": scope.GetTenantID(),
		"now":       now,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot add sample hits: %w", err)
	}

	return nil
}

func LoadSamplePatternHits(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sampleID gid.GID,
	trackerPatternID gid.GID,
) ([]SamplePatternHit, error) {
	q := `
SELECT
	browser_family,
	presentation,
	consent_action,
	category_granted,
	count
FROM tracker_pattern_sample_hits
WHERE
	%s
	AND sample_id = @sample_id
	AND tracker_pattern_id = @tracker_pattern_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"sample_id":          sampleID,
		"tracker_pattern_id": trackerPatternID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query tracker pattern sample hits: %w", err)
	}

	hits, err := pgx.CollectRows(rows, pgx.RowToStructByName[SamplePatternHit])
	if err != nil {
		return nil, fmt.Errorf("cannot collect tracker pattern sample hits: %w", err)
	}

	return hits, nil
}

type TrackerPatternBannerRef struct {
	ID             gid.GID `db:"id"`
	CookieBannerID gid.GID `db:"cookie_banner_id"`
}

func LoadTrackerPatternBannerRefs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	patternIDs []gid.GID,
) ([]TrackerPatternBannerRef, error) {
	if len(patternIDs) == 0 {
		return nil, nil
	}

	q := `
SELECT
	id,
	cookie_banner_id
FROM tracker_patterns
WHERE
	%s
	AND id = ANY(@pattern_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"pattern_ids": patternIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query tracker patterns: %w", err)
	}

	refs, err := pgx.CollectRows(rows, pgx.RowToStructByName[TrackerPatternBannerRef])
	if err != nil {
		return nil, fmt.Errorf("cannot collect tracker patterns: %w", err)
	}

	return refs, nil
}

type SamplePatternHitForPattern struct {
	SampleID         gid.GID       `db:"sample_id"`
	TrackerPatternID gid.GID       `db:"tracker_pattern_id"`
	BrowserFamily    BrowserFamily `db:"browser_family"`
	Presentation     string        `db:"presentation"`
	ConsentAction    string        `db:"consent_action"`
	CategoryGranted  bool          `db:"category_granted"`
	Count            int           `db:"count"`
}

func LoadSamplePatternHitsForPatterns(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sampleIDs []gid.GID,
	patternIDs []gid.GID,
) ([]SamplePatternHitForPattern, error) {
	if len(sampleIDs) == 0 || len(patternIDs) == 0 {
		return nil, nil
	}

	q := `
SELECT
	sample_id,
	tracker_pattern_id,
	browser_family,
	presentation,
	consent_action,
	category_granted,
	count
FROM tracker_pattern_sample_hits
WHERE
	%s
	AND sample_id = ANY(@sample_ids)
	AND tracker_pattern_id = ANY(@pattern_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"sample_ids":  sampleIDs,
		"pattern_ids": patternIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query tracker pattern sample hits: %w", err)
	}

	hits, err := pgx.CollectRows(rows, pgx.RowToStructByName[SamplePatternHitForPattern])
	if err != nil {
		return nil, fmt.Errorf("cannot collect tracker pattern sample hits: %w", err)
	}

	return hits, nil
}
