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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	TrackerPatternDiscoveryHits struct {
		TrackerPatternID gid.GID   `db:"tracker_pattern_id"`
		CookieBannerID   gid.GID   `db:"cookie_banner_id"`
		ChromeHits       int       `db:"chrome_hits"`
		EdgeHits         int       `db:"edge_hits"`
		FirefoxHits      int       `db:"firefox_hits"`
		SafariHits       int       `db:"safari_hits"`
		OtherHits        int       `db:"other_hits"`
		CreatedAt        time.Time `db:"created_at"`
		UpdatedAt        time.Time `db:"updated_at"`
	}
)

// RecordDiscoveryPatternHit counts one first sighting of a pattern.
// It no-ops until a page-load row exists, and after that row is frozen
// or a version is published, so every pattern shares one window.
func RecordDiscoveryPatternHit(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieBannerID gid.GID,
	trackerPatternID gid.GID,
	family DiscoveryBrowserFamily,
) error {
	if !family.IsValid() {
		family = DiscoveryBrowserFamilyOther
	}

	now := time.Now()
	args := pgx.StrictNamedArgs{
		"tracker_pattern_id": trackerPatternID,
		"cookie_banner_id":   cookieBannerID,
		"tenant_id":          scope.GetTenantID(),
		"family":             family,
		"chrome":             DiscoveryBrowserFamilyChrome,
		"edge":               DiscoveryBrowserFamilyEdge,
		"firefox":            DiscoveryBrowserFamilyFirefox,
		"safari":             DiscoveryBrowserFamilySafari,
		"other":              DiscoveryBrowserFamilyOther,
		"published":          CookieBannerVersionStatePublished,
		"now":                now,
	}

	q := `
INSERT INTO tracker_pattern_discovery_hits (
    tracker_pattern_id,
    tenant_id,
    cookie_banner_id,
    chrome_hits,
    edge_hits,
    firefox_hits,
    safari_hits,
    other_hits,
    created_at,
    updated_at
)
SELECT
    @tracker_pattern_id,
    @tenant_id,
    @cookie_banner_id,
    CASE WHEN @family = @chrome THEN 1 ELSE 0 END,
    CASE WHEN @family = @edge THEN 1 ELSE 0 END,
    CASE WHEN @family = @firefox THEN 1 ELSE 0 END,
    CASE WHEN @family = @safari THEN 1 ELSE 0 END,
    CASE WHEN @family = @other THEN 1 ELSE 0 END,
    @now,
    @now
FROM cookie_banner_discovery_stats
WHERE cookie_banner_id = @cookie_banner_id
  AND frozen_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM cookie_banner_versions
      WHERE cookie_banner_id = @cookie_banner_id
        AND state = @published
  )
ON CONFLICT (tracker_pattern_id) DO UPDATE
SET
    chrome_hits = tracker_pattern_discovery_hits.chrome_hits + EXCLUDED.chrome_hits,
    edge_hits = tracker_pattern_discovery_hits.edge_hits + EXCLUDED.edge_hits,
    firefox_hits = tracker_pattern_discovery_hits.firefox_hits + EXCLUDED.firefox_hits,
    safari_hits = tracker_pattern_discovery_hits.safari_hits + EXCLUDED.safari_hits,
    other_hits = tracker_pattern_discovery_hits.other_hits + EXCLUDED.other_hits,
    updated_at = EXCLUDED.updated_at
`

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot record discovery pattern hit: %w", err)
	}

	return nil
}

// AddDiscoveryHits folds an exact pattern's discovery hits into a glob
// before the exact row is deleted.
func AddDiscoveryHits(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	exactPatternID gid.GID,
	globPatternID gid.GID,
) error {
	now := time.Now()

	q := `
INSERT INTO tracker_pattern_discovery_hits (
    tracker_pattern_id,
    tenant_id,
    cookie_banner_id,
    chrome_hits,
    edge_hits,
    firefox_hits,
    safari_hits,
    other_hits,
    created_at,
    updated_at
)
SELECT
    @glob_id,
    exact.tenant_id,
    exact.cookie_banner_id,
    exact.chrome_hits + COALESCE(glob.chrome_hits, 0),
    exact.edge_hits + COALESCE(glob.edge_hits, 0),
    exact.firefox_hits + COALESCE(glob.firefox_hits, 0),
    exact.safari_hits + COALESCE(glob.safari_hits, 0),
    exact.other_hits + COALESCE(glob.other_hits, 0),
    @now,
    @now
FROM tracker_pattern_discovery_hits exact
LEFT JOIN tracker_pattern_discovery_hits glob
    ON glob.tracker_pattern_id = @glob_id
WHERE exact.tracker_pattern_id = @exact_id
  AND exact.tenant_id = @tenant_id
ON CONFLICT (tracker_pattern_id) DO UPDATE
SET
    chrome_hits = EXCLUDED.chrome_hits,
    edge_hits = EXCLUDED.edge_hits,
    firefox_hits = EXCLUDED.firefox_hits,
    safari_hits = EXCLUDED.safari_hits,
    other_hits = EXCLUDED.other_hits,
    updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"exact_id":  exactPatternID,
		"glob_id":   globPatternID,
		"tenant_id": scope.GetTenantID(),
		"now":       now,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot add discovery hits: %w", err)
	}

	return nil
}

func (h *TrackerPatternDiscoveryHits) LoadByTrackerPatternID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	trackerPatternID gid.GID,
) error {
	q := `
SELECT
    tracker_pattern_id,
    cookie_banner_id,
    chrome_hits,
    edge_hits,
    firefox_hits,
    safari_hits,
    other_hits,
    created_at,
    updated_at
FROM tracker_pattern_discovery_hits
WHERE %s
  AND tracker_pattern_id = @tracker_pattern_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"tracker_pattern_id": trackerPatternID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tracker pattern discovery hits: %w", err)
	}

	hits, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrackerPatternDiscoveryHits])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect tracker pattern discovery hits: %w", err)
	}

	*h = hits

	return nil
}
