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
	CookieBannerDiscoveryStats struct {
		CookieBannerID   gid.GID    `db:"cookie_banner_id"`
		ChromePageLoads  int        `db:"chrome_page_loads"`
		EdgePageLoads    int        `db:"edge_page_loads"`
		FirefoxPageLoads int        `db:"firefox_page_loads"`
		SafariPageLoads  int        `db:"safari_page_loads"`
		OtherPageLoads   int        `db:"other_page_loads"`
		FrozenAt         *time.Time `db:"frozen_at"`
		CreatedAt        time.Time  `db:"created_at"`
		UpdatedAt        time.Time  `db:"updated_at"`
	}
)

// RecordDiscoveryPageLoad counts one discovery page view for family.
// A published version or a frozen row makes this a no-op. The first
// successful call creates the shared window every pattern hit uses.
func RecordDiscoveryPageLoad(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieBannerID gid.GID,
	family DiscoveryBrowserFamily,
) error {
	if !family.IsValid() {
		family = DiscoveryBrowserFamilyOther
	}

	now := time.Now()

	q := `
INSERT INTO cookie_banner_discovery_stats (
    cookie_banner_id,
    tenant_id,
    chrome_page_loads,
    edge_page_loads,
    firefox_page_loads,
    safari_page_loads,
    other_page_loads,
    created_at,
    updated_at
)
SELECT
    @cookie_banner_id,
    @tenant_id,
    CASE WHEN @family = @chrome THEN 1 ELSE 0 END,
    CASE WHEN @family = @edge THEN 1 ELSE 0 END,
    CASE WHEN @family = @firefox THEN 1 ELSE 0 END,
    CASE WHEN @family = @safari THEN 1 ELSE 0 END,
    CASE WHEN @family = @other THEN 1 ELSE 0 END,
    @now,
    @now
WHERE NOT EXISTS (
    SELECT 1
    FROM cookie_banner_versions
    WHERE cookie_banner_id = @cookie_banner_id
      AND state = @published
)
ON CONFLICT (cookie_banner_id) DO UPDATE
SET
    chrome_page_loads = cookie_banner_discovery_stats.chrome_page_loads + EXCLUDED.chrome_page_loads,
    edge_page_loads = cookie_banner_discovery_stats.edge_page_loads + EXCLUDED.edge_page_loads,
    firefox_page_loads = cookie_banner_discovery_stats.firefox_page_loads + EXCLUDED.firefox_page_loads,
    safari_page_loads = cookie_banner_discovery_stats.safari_page_loads + EXCLUDED.safari_page_loads,
    other_page_loads = cookie_banner_discovery_stats.other_page_loads + EXCLUDED.other_page_loads,
    updated_at = EXCLUDED.updated_at
WHERE
    cookie_banner_discovery_stats.frozen_at IS NULL
`

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"tenant_id":        scope.GetTenantID(),
		"family":           family,
		"chrome":           DiscoveryBrowserFamilyChrome,
		"edge":             DiscoveryBrowserFamilyEdge,
		"firefox":          DiscoveryBrowserFamilyFirefox,
		"safari":           DiscoveryBrowserFamilySafari,
		"other":            DiscoveryBrowserFamilyOther,
		"published":        CookieBannerVersionStatePublished,
		"now":              now,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot record discovery page load: %w", err)
	}

	return nil
}

// FreezeDiscoveryStats stops page-load and pattern-hit counting for
// the banner. An empty frozen row is created when discovery saw no
// traffic, so a later deactivate cannot open a second window.
func FreezeDiscoveryStats(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	now := time.Now()

	q := `
INSERT INTO cookie_banner_discovery_stats (
    cookie_banner_id,
    tenant_id,
    created_at,
    updated_at,
    frozen_at
) VALUES (
    @cookie_banner_id,
    @tenant_id,
    @now,
    @now,
    @now
)
ON CONFLICT (cookie_banner_id) DO UPDATE
SET
    frozen_at = COALESCE(cookie_banner_discovery_stats.frozen_at, EXCLUDED.frozen_at),
    updated_at = EXCLUDED.updated_at
WHERE
    cookie_banner_discovery_stats.frozen_at IS NULL
`

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"tenant_id":        scope.GetTenantID(),
		"now":              now,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot freeze discovery stats: %w", err)
	}

	return nil
}

func (s *CookieBannerDiscoveryStats) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	q := `
SELECT
    cookie_banner_id,
    chrome_page_loads,
    edge_page_loads,
    firefox_page_loads,
    safari_page_loads,
    other_page_loads,
    frozen_at,
    created_at,
    updated_at
FROM cookie_banner_discovery_stats
WHERE %s
  AND cookie_banner_id = @cookie_banner_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner discovery stats: %w", err)
	}

	stats, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerDiscoveryStats])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cookie banner discovery stats: %w", err)
	}

	*s = stats

	return nil
}
