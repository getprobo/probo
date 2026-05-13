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
	"errors"
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
		InitiatorURL     *string       `db:"initiator_url"`
		InitiatorDomain  *string       `db:"initiator_domain"`
		LastDetectedAt   time.Time     `db:"last_detected_at"`
		CreatedAt        time.Time     `db:"created_at"`
		UpdatedAt        time.Time     `db:"updated_at"`
	}

	DetectedTrackers []*DetectedTracker
)

func (dt *DetectedTracker) Upsert(
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
	initiator_url,
	initiator_domain,
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
	@initiator_url,
	@initiator_domain,
	@last_detected_at,
	@created_at,
	@updated_at
)
ON CONFLICT (cookie_banner_id, tracker_type, identifier) DO UPDATE
	SET last_detected_at = EXCLUDED.last_detected_at,
		source = CASE WHEN detected_trackers.source IS NULL OR (
				detected_trackers.source != @source_script AND EXCLUDED.source = @source_script
			) THEN EXCLUDED.source
			ELSE detected_trackers.source
		END,
		initiator_url = COALESCE(EXCLUDED.initiator_url, detected_trackers.initiator_url),
		initiator_domain = COALESCE(EXCLUDED.initiator_domain, detected_trackers.initiator_domain),
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
		"initiator_url":      dt.InitiatorURL,
		"initiator_domain":   dt.InitiatorDomain,
		"last_detected_at":   dt.LastDetectedAt,
		"created_at":         dt.CreatedAt,
		"updated_at":         dt.UpdatedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert detected tracker: %w", err)
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

func (dts *DetectedTrackers) LoadCommonThirdPartyIDByTrackerPatternID(
	ctx context.Context,
	conn pg.Querier,
	trackerPatternID gid.GID,
) (*gid.GID, error) {
	q := `
SELECT DISTINCT ctpd.common_third_party_id
FROM detected_trackers dt
JOIN common_third_party_domains ctpd ON ctpd.domain = dt.initiator_domain
WHERE dt.tracker_pattern_id = @tracker_pattern_id
  AND dt.initiator_domain IS NOT NULL
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"tracker_pattern_id": trackerPatternID}

	var commonThirdPartyID gid.GID
	if err := conn.QueryRow(ctx, q, args).Scan(&commonThirdPartyID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot load common third party ID by tracker pattern: %w", err)
	}

	return &commonThirdPartyID, nil
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
