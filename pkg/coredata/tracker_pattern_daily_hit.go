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

// DailyPatternHit leaves Presentation and ConsentAction empty when the
// report omitted them.
type DailyPatternHit struct {
	TrackerPatternID gid.GID                  `db:"tracker_pattern_id"`
	BrowserFamily    BrowserFamily            `db:"browser_family"`
	Presentation     CookieBannerPresentation `db:"presentation"`
	ConsentAction    CookieConsentAction      `db:"consent_action"`
	CategoryGranted  bool                     `db:"category_granted"`
	Count            int                      `db:"count"`
}

// RecordDailyPatternHit adds count to the bucket of day, which must be a
// UTC midnight. A pattern deleted since the count was buffered is skipped.
func RecordDailyPatternHit(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieBannerID gid.GID,
	trackerPatternID gid.GID,
	day time.Time,
	family BrowserFamily,
	presentation CookieBannerPresentation,
	consentAction CookieConsentAction,
	categoryGranted bool,
	count int64,
) error {
	if count == 0 {
		return nil
	}

	now := time.Now()

	q := `
INSERT INTO tracker_pattern_daily_hits (
	tracker_pattern_id,
	day,
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
	@tracker_pattern_id,
	@day::date,
	@browser_family,
	@presentation,
	@consent_action,
	@category_granted,
	@tenant_id,
	@cookie_banner_id,
	@count,
	@now,
	@now
WHERE EXISTS (
	SELECT 1
	FROM tracker_patterns
	WHERE id = @tracker_pattern_id
)
ON CONFLICT (
	tracker_pattern_id,
	day,
	browser_family,
	presentation,
	consent_action,
	category_granted
) DO UPDATE
SET
	count = tracker_pattern_daily_hits.count + EXCLUDED.count,
	updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"tracker_pattern_id": trackerPatternID,
		"day":                day,
		"browser_family":     family,
		"presentation":       presentation,
		"consent_action":     consentAction,
		"category_granted":   categoryGranted,
		"tenant_id":          scope.GetTenantID(),
		"cookie_banner_id":   cookieBannerID,
		"count":              count,
		"now":                now,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot record daily pattern hit: %w", err)
	}

	return nil
}

// AddDailyPatternHits adds every daily bucket of the exact pattern onto
// the glob pattern that replaces it.
func AddDailyPatternHits(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	exactPatternID gid.GID,
	globPatternID gid.GID,
) error {
	now := time.Now()

	q := `
INSERT INTO tracker_pattern_daily_hits (
	tracker_pattern_id,
	day,
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
	@glob_id,
	exact.day,
	exact.browser_family,
	exact.presentation,
	exact.consent_action,
	exact.category_granted,
	exact.tenant_id,
	exact.cookie_banner_id,
	exact.count,
	@now,
	@now
FROM tracker_pattern_daily_hits exact
WHERE
	%s
	AND exact.tracker_pattern_id = @exact_id
ON CONFLICT (
	tracker_pattern_id,
	day,
	browser_family,
	presentation,
	consent_action,
	category_granted
) DO UPDATE
SET
	count = tracker_pattern_daily_hits.count + EXCLUDED.count,
	updated_at = EXCLUDED.updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"exact_id": exactPatternID,
		"glob_id":  globPatternID,
		"now":      now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot add daily pattern hits: %w", err)
	}

	return nil
}

// LoadDailyPatternHitsByTrackerPatternIDs sums hits per pattern and
// dimension over the days from since, a UTC midnight, onward.
func LoadDailyPatternHitsByTrackerPatternIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	trackerPatternIDs []gid.GID,
	since time.Time,
) ([]DailyPatternHit, error) {
	if len(trackerPatternIDs) == 0 {
		return nil, nil
	}

	q := `
SELECT
	tracker_pattern_id,
	browser_family,
	presentation,
	consent_action,
	category_granted,
	SUM(count) AS count
FROM tracker_pattern_daily_hits
WHERE
	%s
	AND tracker_pattern_id = ANY(@tracker_pattern_ids)
	AND day >= @since::date
GROUP BY
	tracker_pattern_id,
	browser_family,
	presentation,
	consent_action,
	category_granted
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tracker_pattern_ids": trackerPatternIDs,
		"since":               since,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query daily pattern hits: %w", err)
	}

	hits, err := pgx.CollectRows(rows, pgx.RowToStructByName[DailyPatternHit])
	if err != nil {
		return nil, fmt.Errorf("cannot collect daily pattern hits: %w", err)
	}

	return hits, nil
}

func DeleteDailyPatternHitsBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
DELETE FROM tracker_pattern_daily_hits
WHERE (
	tracker_pattern_id,
	day,
	browser_family,
	presentation,
	consent_action,
	category_granted
) IN (
	SELECT
		tracker_pattern_id,
		day,
		browser_family,
		presentation,
		consent_action,
		category_granted
	FROM tracker_pattern_daily_hits
	WHERE day < @before::date
	LIMIT @limit
)
`

	args := pgx.StrictNamedArgs{
		"before": before,
		"limit":  limit,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot delete daily pattern hits: %w", err)
	}

	return result.RowsAffected(), nil
}
