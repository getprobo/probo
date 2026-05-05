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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	TrackerPattern struct {
		ID               gid.GID                `db:"id"`
		OrganizationID   gid.GID                `db:"organization_id"`
		CookieBannerID   gid.GID                `db:"cookie_banner_id"`
		CookieCategoryID gid.GID                `db:"cookie_category_id"`
		TrackerType      TrackerType            `db:"tracker_type"`
		Pattern          string                 `db:"pattern"`
		MatchType        CookiePatternMatchType `db:"match_type"`
		DisplayName      string                 `db:"display_name"`
		Description      string                 `db:"description"`
		Excluded         bool                   `db:"excluded"`
		MaxAgeSeconds    *int                   `db:"max_age_seconds"`
		Source           *CookieSource          `db:"source"`
		LastMatchedAt    *time.Time             `db:"last_matched_at"`
		CreatedAt        time.Time              `db:"created_at"`
		UpdatedAt        time.Time              `db:"updated_at"`
	}

	TrackerPatterns []*TrackerPattern
)

func (tp *TrackerPattern) CursorKey(field CookiePatternOrderField) page.CursorKey {
	switch field {
	case CookiePatternOrderFieldCreatedAt:
		return page.NewCursorKey(tp.ID, tp.CreatedAt)
	case CookiePatternOrderFieldName:
		return page.NewCursorKey(tp.ID, tp.DisplayName)
	case CookiePatternOrderFieldLastMatchedAt:
		if tp.LastMatchedAt == nil {
			return page.NewCursorKey(tp.ID, time.Time{})
		}
		return page.NewCursorKey(tp.ID, *tp.LastMatchedAt)
	case CookiePatternOrderFieldUpdatedAt:
		return page.NewCursorKey(tp.ID, tp.UpdatedAt)
	case CookiePatternOrderFieldSource:
		if tp.Source == nil {
			return page.NewCursorKey(tp.ID, "")
		}
		return page.NewCursorKey(tp.ID, string(*tp.Source))
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (tp *TrackerPattern) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM tracker_patterns WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, tp.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot query tracker pattern authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (tp *TrackerPattern) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	trackerPatternID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	%s
	AND id = @tracker_pattern_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"tracker_pattern_id": trackerPatternID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tracker patterns: %w", err)
	}

	pattern, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrackerPattern])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect tracker pattern: %w", err)
	}

	*tp = pattern

	return nil
}

func (tp *TrackerPattern) LoadByBannerIDTypeAndPattern(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	trackerType TrackerType,
	pattern string,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND tracker_type = @tracker_type
	AND pattern = @pattern
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"tracker_type":     trackerType,
		"pattern":          pattern,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tracker patterns: %w", err)
	}

	p, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrackerPattern])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect tracker pattern: %w", err)
	}

	*tp = p

	return nil
}

func (tp *TrackerPattern) FindMatchingPattern(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	trackerType TrackerType,
	identifier string,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND tracker_type = @tracker_type
	AND (
		(match_type = @match_type_prefix AND starts_with(@identifier, pattern))
		OR (match_type = @match_type_exact AND pattern = @identifier)
	)
ORDER BY
	CASE WHEN match_type = @match_type_exact AND pattern = @identifier THEN 0
	     WHEN match_type = @match_type_prefix THEN 1
	     ELSE 2
	END,
	LENGTH(pattern) DESC
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id":  cookieBannerID,
		"tracker_type":      trackerType,
		"identifier":        identifier,
		"match_type_prefix": CookiePatternMatchTypePrefix,
		"match_type_exact":  CookiePatternMatchTypeExact,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tracker patterns: %w", err)
	}

	pattern, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrackerPattern])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect tracker pattern: %w", err)
	}

	*tp = pattern

	return nil
}

func (tp *TrackerPattern) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO tracker_patterns (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_category_id,
	@tracker_type,
	@pattern,
	@match_type,
	@display_name,
	@description,
	@excluded,
	@max_age_seconds,
	@source,
	@last_matched_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                 tp.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    tp.OrganizationID,
		"cookie_banner_id":   tp.CookieBannerID,
		"cookie_category_id": tp.CookieCategoryID,
		"tracker_type":       tp.TrackerType,
		"pattern":            tp.Pattern,
		"match_type":         tp.MatchType,
		"display_name":       tp.DisplayName,
		"description":        tp.Description,
		"excluded":           tp.Excluded,
		"max_age_seconds":    tp.MaxAgeSeconds,
		"source":             tp.Source,
		"last_matched_at":    tp.LastMatchedAt,
		"created_at":         tp.CreatedAt,
		"updated_at":         tp.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_tracker_patterns_unique_pattern_per_banner" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert tracker pattern: %w", err)
	}

	return nil
}

func (tp *TrackerPattern) InsertIfNotExists(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO tracker_patterns (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_category_id,
	@tracker_type,
	@pattern,
	@match_type,
	@display_name,
	@description,
	@excluded,
	@max_age_seconds,
	@source,
	@last_matched_at,
	@created_at,
	@updated_at
)
ON CONFLICT (cookie_banner_id, tracker_type, pattern) DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"id":                 tp.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    tp.OrganizationID,
		"cookie_banner_id":   tp.CookieBannerID,
		"cookie_category_id": tp.CookieCategoryID,
		"tracker_type":       tp.TrackerType,
		"pattern":            tp.Pattern,
		"match_type":         tp.MatchType,
		"display_name":       tp.DisplayName,
		"description":        tp.Description,
		"excluded":           tp.Excluded,
		"max_age_seconds":    tp.MaxAgeSeconds,
		"source":             tp.Source,
		"last_matched_at":    tp.LastMatchedAt,
		"created_at":         tp.CreatedAt,
		"updated_at":         tp.UpdatedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot insert tracker pattern: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (tp *TrackerPattern) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE tracker_patterns
SET
	cookie_category_id = @cookie_category_id,
	display_name = @display_name,
	max_age_seconds = @max_age_seconds,
	description = @description,
	excluded = @excluded,
	last_matched_at = @last_matched_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                 tp.ID,
		"cookie_category_id": tp.CookieCategoryID,
		"display_name":       tp.DisplayName,
		"max_age_seconds":    tp.MaxAgeSeconds,
		"description":        tp.Description,
		"excluded":           tp.Excluded,
		"last_matched_at":    tp.LastMatchedAt,
		"updated_at":         tp.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_tracker_patterns_unique_pattern_per_banner" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot update tracker pattern: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (tp *TrackerPattern) Delete(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM tracker_patterns
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": tp.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete tracker pattern: %w", err)
	}

	return nil
}

func (tps *TrackerPatterns) LoadAllByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	filter *CookiePatternFilter,
	trackerType *TrackerType,
) error {
	trackerTypeFragment := "TRUE"
	if trackerType != nil {
		trackerTypeFragment = "tracker_type = @tracker_type"
	}

	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
	AND %s
ORDER BY
	created_at ASC, id ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment(), trackerTypeFragment, filter.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	if trackerType != nil {
		args["tracker_type"] = *trackerType
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tracker patterns: %w", err)
	}

	patterns, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrackerPattern])
	if err != nil {
		return fmt.Errorf("cannot collect tracker patterns: %w", err)
	}

	*tps = patterns

	return nil
}

func (tps *TrackerPatterns) RefreshLastMatchedAtByCookieBannerID(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	q := `
UPDATE tracker_patterns
SET
	last_matched_at = sub.max_detected
FROM (
	SELECT tracker_pattern_id, MAX(last_detected_at) AS max_detected
	FROM detected_trackers
	WHERE %[1]s AND cookie_banner_id = @cookie_banner_id
	GROUP BY tracker_pattern_id
) sub
WHERE
	tracker_patterns.id = sub.tracker_pattern_id
	AND %[1]s
	AND tracker_patterns.cookie_banner_id = @cookie_banner_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot refresh last_matched_at for banner tracker patterns: %w", err)
	}

	return nil
}

func (tps *TrackerPatterns) LoadUncategorisedByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[CookiePatternOrderField],
	filter *CookiePatternFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND cookie_category_id = (
		SELECT id FROM cookie_categories
		WHERE cookie_banner_id = @cookie_banner_id
			AND kind = @category_kind
			AND %s
		LIMIT 1
	)
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"category_kind":    CookieCategoryKindUncategorised,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query uncategorised tracker patterns: %w", err)
	}

	patterns, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrackerPattern])
	if err != nil {
		return fmt.Errorf("cannot collect uncategorised tracker patterns: %w", err)
	}

	*tps = patterns

	return nil
}

func (tps *TrackerPatterns) CountUncategorisedByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	filter *CookiePatternFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND cookie_category_id = (
		SELECT id FROM cookie_categories
		WHERE cookie_banner_id = @cookie_banner_id
			AND kind = @category_kind
			AND %s
		LIMIT 1
	)
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"category_kind":    CookieCategoryKindUncategorised,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}
