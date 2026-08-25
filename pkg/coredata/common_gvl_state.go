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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	CommonGVLState struct {
		LatestVendorListVersion *int      `db:"latest_vendor_list_version"`
		ETag                    *string   `db:"etag"`
		CacheMaxAgeSeconds      *int      `db:"cache_max_age_seconds"`
		LastFetchedAt           time.Time `db:"last_fetched_at"`
	}
)

func (s *CommonGVLState) Load(
	ctx context.Context,
	conn pg.Querier,
) error {
	q := `
SELECT
    latest_vendor_list_version,
    etag,
    cache_max_age_seconds,
    last_fetched_at
FROM
    common_gvl_state
WHERE
    singleton
LIMIT 1;
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query common gvl state: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonGVLState])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common gvl state: %w", err)
	}

	*s = row

	return nil
}

func (s *CommonGVLState) Upsert(
	ctx context.Context,
	conn pg.Querier,
) error {
	q := `
INSERT INTO common_gvl_state (
    singleton,
    latest_vendor_list_version,
    etag,
    cache_max_age_seconds,
    last_fetched_at
) VALUES (
    TRUE,
    @latest_vendor_list_version,
    @etag,
    @cache_max_age_seconds,
    @last_fetched_at
)
ON CONFLICT (singleton) DO UPDATE
SET
    latest_vendor_list_version = EXCLUDED.latest_vendor_list_version,
    etag                       = EXCLUDED.etag,
    cache_max_age_seconds      = EXCLUDED.cache_max_age_seconds,
    last_fetched_at            = EXCLUDED.last_fetched_at
`

	args := pgx.StrictNamedArgs{
		"latest_vendor_list_version": s.LatestVendorListVersion,
		"etag":                       s.ETag,
		"cache_max_age_seconds":      s.CacheMaxAgeSeconds,
		"last_fetched_at":            s.LastFetchedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot upsert common gvl state: %w", err)
	}

	return nil
}

// IsFresh reports whether a previous fetch is still inside IAB's Cache-Control
// max-age, so a new HTTP request should not be made.
func (s CommonGVLState) IsFresh(now time.Time) bool {
	if s.LastFetchedAt.IsZero() || s.CacheMaxAgeSeconds == nil || *s.CacheMaxAgeSeconds <= 0 {
		return false
	}

	return now.Before(s.LastFetchedAt.Add(time.Duration(*s.CacheMaxAgeSeconds) * time.Second))
}
