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

type (
	DetectedTracker struct {
		ID               gid.GID       `db:"id"`
		CookieBannerID   gid.GID       `db:"cookie_banner_id"`
		TrackerPatternID *gid.GID      `db:"tracker_pattern_id"`
		TrackerType      TrackerType   `db:"tracker_type"`
		Identifier       string        `db:"identifier"`
		MaxAgeSeconds    *int          `db:"max_age_seconds"`
		Source           *CookieSource `db:"source"`
		ValueSize        *int          `db:"value_size"`
		LastDetectedAt   time.Time     `db:"last_detected_at"`
		CreatedAt        time.Time     `db:"created_at"`
		UpdatedAt        time.Time     `db:"updated_at"`
	}

	DetectedTrackers []*DetectedTracker
)

func (dt *DetectedTracker) InsertIfNotExists(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO detected_trackers (
	id,
	tenant_id,
	cookie_banner_id,
	tracker_pattern_id,
	tracker_type,
	identifier,
	max_age_seconds,
	source,
	value_size,
	last_detected_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@cookie_banner_id,
	@tracker_pattern_id,
	@tracker_type,
	@identifier,
	@max_age_seconds,
	@source,
	@value_size,
	@last_detected_at,
	@created_at,
	@updated_at
)
ON CONFLICT (cookie_banner_id, tracker_type, identifier) DO UPDATE
	SET last_detected_at = EXCLUDED.last_detected_at,
		source = CASE WHEN detected_trackers.source IS NULL OR (detected_trackers.source != @source_script AND EXCLUDED.source = @source_script) THEN EXCLUDED.source ELSE detected_trackers.source END,
		updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"id":                 dt.ID,
		"tenant_id":          scope.GetTenantID(),
		"cookie_banner_id":   dt.CookieBannerID,
		"tracker_pattern_id": dt.TrackerPatternID,
		"tracker_type":       dt.TrackerType,
		"identifier":         dt.Identifier,
		"max_age_seconds":    dt.MaxAgeSeconds,
		"source":             dt.Source,
		"source_script":      CookieSourceScript,
		"value_size":         dt.ValueSize,
		"last_detected_at":   dt.LastDetectedAt,
		"created_at":         dt.CreatedAt,
		"updated_at":         dt.UpdatedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot insert detected tracker: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (dts *DetectedTrackers) CountByTrackerPatternID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	trackerPatternID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	detected_trackers
WHERE
	%s
	AND tracker_pattern_id = @tracker_pattern_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"tracker_pattern_id": trackerPatternID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (dts *DetectedTrackers) RelinkByTrackerPatternID(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	sourcePatternID gid.GID,
	targetPatternID gid.GID,
) error {
	q := `
UPDATE detected_trackers
SET
	tracker_pattern_id = @target_pattern_id,
	updated_at = @updated_at
WHERE
	%s
	AND tracker_pattern_id = @source_pattern_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"source_pattern_id": sourcePatternID,
		"target_pattern_id": targetPatternID,
		"updated_at":        time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot relink detected trackers to pattern: %w", err)
	}

	return nil
}
