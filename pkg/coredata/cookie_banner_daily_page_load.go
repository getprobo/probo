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

// DailyPageLoad leaves Presentation and ConsentAction empty when the report
// omitted them.
type DailyPageLoad struct {
	CookieBannerID gid.GID                  `db:"cookie_banner_id"`
	BrowserFamily  BrowserFamily            `db:"browser_family"`
	Presentation   CookieBannerPresentation `db:"presentation"`
	ConsentAction  CookieConsentAction      `db:"consent_action"`
	Count          int                      `db:"count"`
}

// RecordDailyPageLoad adds count to the bucket of day, which must be a UTC
// midnight. A banner deleted since the count was buffered is skipped.
func RecordDailyPageLoad(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieBannerID gid.GID,
	day time.Time,
	family BrowserFamily,
	presentation CookieBannerPresentation,
	consentAction CookieConsentAction,
	count int64,
) error {
	if count == 0 {
		return nil
	}

	now := time.Now()

	q := `
INSERT INTO cookie_banner_daily_page_loads (
	cookie_banner_id,
	day,
	browser_family,
	presentation,
	consent_action,
	tenant_id,
	count,
	created_at,
	updated_at
)
SELECT
	@cookie_banner_id,
	@day::date,
	@browser_family,
	@presentation,
	@consent_action,
	@tenant_id,
	@count,
	@now,
	@now
WHERE EXISTS (
	SELECT 1
	FROM cookie_banners
	WHERE id = @cookie_banner_id
)
ON CONFLICT (
	cookie_banner_id,
	day,
	browser_family,
	presentation,
	consent_action
) DO UPDATE
SET
	count = cookie_banner_daily_page_loads.count + EXCLUDED.count,
	updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"day":              day,
		"browser_family":   family,
		"presentation":     presentation,
		"consent_action":   consentAction,
		"tenant_id":        scope.GetTenantID(),
		"count":            count,
		"now":              now,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot record daily page load: %w", err)
	}

	return nil
}

// LoadDailyPageLoadsByCookieBannerIDs sums page loads per banner and
// dimension over the days from since, a UTC midnight, onward.
func LoadDailyPageLoadsByCookieBannerIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerIDs []gid.GID,
	since time.Time,
) ([]DailyPageLoad, error) {
	if len(cookieBannerIDs) == 0 {
		return nil, nil
	}

	q := `
SELECT
	cookie_banner_id,
	browser_family,
	presentation,
	consent_action,
	SUM(count) AS count
FROM cookie_banner_daily_page_loads
WHERE
	%s
	AND cookie_banner_id = ANY(@cookie_banner_ids)
	AND day >= @since::date
GROUP BY
	cookie_banner_id,
	browser_family,
	presentation,
	consent_action
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_ids": cookieBannerIDs,
		"since":             since,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query daily page loads: %w", err)
	}

	loads, err := pgx.CollectRows(rows, pgx.RowToStructByName[DailyPageLoad])
	if err != nil {
		return nil, fmt.Errorf("cannot collect daily page loads: %w", err)
	}

	return loads, nil
}

func DeleteDailyPageLoadsBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
DELETE FROM cookie_banner_daily_page_loads
WHERE (
	cookie_banner_id,
	day,
	browser_family,
	presentation,
	consent_action
) IN (
	SELECT
		cookie_banner_id,
		day,
		browser_family,
		presentation,
		consent_action
	FROM cookie_banner_daily_page_loads
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
		return 0, fmt.Errorf("cannot delete daily page loads: %w", err)
	}

	return result.RowsAffected(), nil
}
