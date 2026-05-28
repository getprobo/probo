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
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	TrackerPattern struct {
		ID                     gid.GID                 `db:"id"`
		OrganizationID         gid.GID                 `db:"organization_id"`
		CookieBannerID         gid.GID                 `db:"cookie_banner_id"`
		CookieCategoryID       gid.GID                 `db:"cookie_category_id"`
		CommonTrackerPatternID *gid.GID                `db:"common_tracker_pattern_id"`
		ThirdPartyID           *gid.GID                `db:"third_party_id"`
		TrackerType            TrackerType             `db:"tracker_type"`
		Pattern                string                  `db:"pattern"`
		MatchType              TrackerPatternMatchType `db:"match_type"`
		DisplayName            string                  `db:"display_name"`
		Description            string                  `db:"description"`
		Excluded               bool                    `db:"excluded"`
		MaxAgeSeconds          *int                    `db:"max_age_seconds"`
		Source                 *CookieSource           `db:"source"`
		LastMatchedAt          *time.Time              `db:"last_matched_at"`
		MappingRequestedAt     *time.Time              `db:"mapping_requested_at"`
		CreatedAt              time.Time               `db:"created_at"`
		UpdatedAt              time.Time               `db:"updated_at"`
	}

	TrackerPatterns []*TrackerPattern
)

func (tp *TrackerPattern) CursorKey(field TrackerPatternOrderField) page.CursorKey {
	switch field {
	case TrackerPatternOrderFieldCreatedAt:
		return page.NewCursorKey(tp.ID, tp.CreatedAt)
	case TrackerPatternOrderFieldName:
		return page.NewCursorKey(tp.ID, tp.DisplayName)
	case TrackerPatternOrderFieldLastMatchedAt:
		if tp.LastMatchedAt == nil {
			return page.NewCursorKey(tp.ID, time.Time{})
		}

		return page.NewCursorKey(tp.ID, *tp.LastMatchedAt)
	case TrackerPatternOrderFieldUpdatedAt:
		return page.NewCursorKey(tp.ID, tp.UpdatedAt)
	case TrackerPatternOrderFieldSource:
		if tp.Source == nil {
			return page.NewCursorKey(tp.ID, "")
		}

		return page.NewCursorKey(tp.ID, string(*tp.Source))
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (tp *TrackerPattern) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM tracker_patterns WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
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
	common_tracker_pattern_id,
	third_party_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	mapping_requested_at,
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
	maxAgeSeconds *int,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	common_tracker_pattern_id,
	third_party_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	mapping_requested_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND tracker_type = @tracker_type
	AND pattern = @pattern
	AND COALESCE(max_age_seconds, -1) = COALESCE(@max_age_seconds, -1)
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"tracker_type":     trackerType,
		"pattern":          pattern,
		"max_age_seconds":  maxAgeSeconds,
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
	common_tracker_pattern_id,
	third_party_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	mapping_requested_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND tracker_type = @tracker_type
	AND (
		(match_type = @match_type_glob
		 AND @identifier LIKE
		     replace(replace(replace(replace(
		         pattern, E'\\', E'\\\\'), '%', E'\\%'), '_', E'\\_'), '*', '%')
		     ESCAPE E'\\')
		OR (match_type = @match_type_exact AND pattern = @identifier)
	)
ORDER BY
	CASE WHEN match_type = @match_type_exact AND pattern = @identifier THEN 0
	     ELSE 1
	END,
	length(replace(pattern, '*', '')) DESC
LIMIT 1;
`

	q = strings.Replace(q, "%s", scope.SQLFragment(), 1)

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"tracker_type":     trackerType,
		"identifier":       identifier,
		"match_type_glob":  TrackerPatternMatchTypeGlob,
		"match_type_exact": TrackerPatternMatchTypeExact,
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
	common_tracker_pattern_id,
	third_party_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	mapping_requested_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_category_id,
	@common_tracker_pattern_id,
	@third_party_id,
	@tracker_type,
	@pattern,
	@match_type,
	@display_name,
	@description,
	@excluded,
	@max_age_seconds,
	@source,
	@last_matched_at,
	@mapping_requested_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                        tp.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           tp.OrganizationID,
		"cookie_banner_id":          tp.CookieBannerID,
		"cookie_category_id":        tp.CookieCategoryID,
		"common_tracker_pattern_id": tp.CommonTrackerPatternID,
		"third_party_id":            tp.ThirdPartyID,
		"tracker_type":              tp.TrackerType,
		"pattern":                   tp.Pattern,
		"match_type":                tp.MatchType,
		"display_name":              tp.DisplayName,
		"description":               tp.Description,
		"excluded":                  tp.Excluded,
		"max_age_seconds":           tp.MaxAgeSeconds,
		"source":                    tp.Source,
		"last_matched_at":           tp.LastMatchedAt,
		"mapping_requested_at":      tp.MappingRequestedAt,
		"created_at":                tp.CreatedAt,
		"updated_at":                tp.UpdatedAt,
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
	common_tracker_pattern_id,
	third_party_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	mapping_requested_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_category_id,
	@common_tracker_pattern_id,
	@third_party_id,
	@tracker_type,
	@pattern,
	@match_type,
	@display_name,
	@description,
	@excluded,
	@max_age_seconds,
	@source,
	@last_matched_at,
	@mapping_requested_at,
	@created_at,
	@updated_at
)
ON CONFLICT (cookie_banner_id, tracker_type, pattern, COALESCE(max_age_seconds, -1)) DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"id":                        tp.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           tp.OrganizationID,
		"cookie_banner_id":          tp.CookieBannerID,
		"cookie_category_id":        tp.CookieCategoryID,
		"common_tracker_pattern_id": tp.CommonTrackerPatternID,
		"third_party_id":            tp.ThirdPartyID,
		"tracker_type":              tp.TrackerType,
		"pattern":                   tp.Pattern,
		"match_type":                tp.MatchType,
		"display_name":              tp.DisplayName,
		"description":               tp.Description,
		"excluded":                  tp.Excluded,
		"max_age_seconds":           tp.MaxAgeSeconds,
		"source":                    tp.Source,
		"last_matched_at":           tp.LastMatchedAt,
		"mapping_requested_at":      tp.MappingRequestedAt,
		"created_at":                tp.CreatedAt,
		"updated_at":                tp.UpdatedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot insert tracker pattern: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

// Update rewrites the editable columns of the receiver's row,
// including `source`. Callers MUST load the pattern under the same
// transaction before mutating fields and calling Update, otherwise
// stale local values will clobber concurrent writes. To advance
// `source`, gate the assignment behind shouldPromoteSource in
// pkg/cookiebanner — there is no DB-side ranking.
func (tp *TrackerPattern) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE tracker_patterns
SET
	common_tracker_pattern_id = @common_tracker_pattern_id,
	third_party_id = @third_party_id,
	cookie_category_id = @cookie_category_id,
	display_name = @display_name,
	max_age_seconds = @max_age_seconds,
	description = @description,
	excluded = @excluded,
	source = @source,
	last_matched_at = @last_matched_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                        tp.ID,
		"common_tracker_pattern_id": tp.CommonTrackerPatternID,
		"third_party_id":            tp.ThirdPartyID,
		"cookie_category_id":        tp.CookieCategoryID,
		"display_name":              tp.DisplayName,
		"max_age_seconds":           tp.MaxAgeSeconds,
		"description":               tp.Description,
		"excluded":                  tp.Excluded,
		"source":                    tp.Source,
		"last_matched_at":           tp.LastMatchedAt,
		"updated_at":                tp.UpdatedAt,
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
	filter *TrackerPatternFilter,
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
	common_tracker_pattern_id,
	third_party_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	mapping_requested_at,
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

func (tps *TrackerPatterns) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[TrackerPatternOrderField],
	filter *TrackerPatternFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	common_tracker_pattern_id,
	third_party_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	mapping_requested_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

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

func (tps *TrackerPatterns) CountByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	filter *TrackerPatternFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count tracker patterns: %w", err)
	}

	return count, nil
}

func (tps *TrackerPatterns) LoadByCookieCategoryID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieCategoryID gid.GID,
	cursor *page.Cursor[TrackerPatternOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	common_tracker_pattern_id,
	third_party_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	mapping_requested_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_category_id = @cookie_category_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_category_id": cookieCategoryID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

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

func (tps *TrackerPatterns) CountByCookieCategoryID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieCategoryID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	tracker_patterns
WHERE
	%s
	AND cookie_category_id = @cookie_category_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_category_id": cookieCategoryID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

// LoadDistinctThirdPartyIDsByCookieBannerID returns the distinct non-null
// `third_party_id` values referenced by tracker patterns of the given
// banner. Callers feed it to ThirdParty.GetByIDs to power per-banner
// pickers without crossing the entity boundary.
func (tps *TrackerPatterns) LoadDistinctThirdPartyIDsByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) ([]gid.GID, error) {
	q := `
SELECT DISTINCT third_party_id
FROM tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND third_party_id IS NOT NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query distinct third party ids: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect distinct third party ids: %w", err)
	}

	return ids, nil
}

// LoadDistinctCommonTrackerPatternIDsByCookieBannerID returns the
// distinct non-null `common_tracker_pattern_id` values referenced by
// tracker patterns of the given banner. Callers chain this with
// CommonTrackerPatterns.LoadByIDs and CommonThirdParties.LoadByIDs to
// resolve the linked common third parties without JOINs.
func (tps *TrackerPatterns) LoadDistinctCommonTrackerPatternIDsByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) ([]gid.GID, error) {
	q := `
SELECT DISTINCT common_tracker_pattern_id
FROM tracker_patterns
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND common_tracker_pattern_id IS NOT NULL
	AND third_party_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query distinct common tracker pattern ids: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect distinct common tracker pattern ids: %w", err)
	}

	return ids, nil
}

func (tps *TrackerPatterns) LoadDistinctThirdPartyIDsByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ids []gid.GID,
) ([]gid.GID, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	q := `
SELECT DISTINCT third_party_id
FROM tracker_patterns
WHERE
	%s
	AND id = ANY(@ids)
	AND third_party_id IS NOT NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query distinct third party ids by pattern ids: %w", err)
	}

	thirdPartyIDs, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect distinct third party ids by pattern ids: %w", err)
	}

	return thirdPartyIDs, nil
}

func (tps *TrackerPatterns) LoadDistinctCommonTrackerPatternIDsByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ids []gid.GID,
) ([]gid.GID, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	q := `
SELECT DISTINCT common_tracker_pattern_id
FROM tracker_patterns
WHERE
	%s
	AND id = ANY(@ids)
	AND common_tracker_pattern_id IS NOT NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query distinct common tracker pattern ids by pattern ids: %w", err)
	}

	commonPatternIDs, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect distinct common tracker pattern ids by pattern ids: %w", err)
	}

	return commonPatternIDs, nil
}

func (tps *TrackerPatterns) UpdateLastMatchedAt(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	patternIDs []gid.GID,
	matchedAt time.Time,
) error {
	if len(patternIDs) == 0 {
		return nil
	}

	q := `
UPDATE tracker_patterns
SET
	last_matched_at = @matched_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = ANY(@pattern_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"pattern_ids": patternIDs,
		"matched_at":  matchedAt,
		"updated_at":  matchedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update last_matched_at for tracker patterns: %w", err)
	}

	return nil
}

func (tps *TrackerPatterns) MoveToCategoryByCookieCategoryID(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	sourceCategoryID gid.GID,
	targetCategoryID gid.GID,
) error {
	q := `
UPDATE tracker_patterns
SET
	cookie_category_id = @target_category_id,
	updated_at = @updated_at
WHERE
	%s
	AND cookie_category_id = @source_category_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"source_category_id": sourceCategoryID,
		"target_category_id": targetCategoryID,
		"updated_at":         time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot move tracker patterns to category: %w", err)
	}

	return nil
}

func (tp *TrackerPattern) LoadNextForMappingForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	common_tracker_pattern_id,
	third_party_id,
	tracker_type,
	pattern,
	match_type,
	display_name,
	description,
	excluded,
	max_age_seconds,
	source,
	last_matched_at,
	mapping_requested_at,
	created_at,
	updated_at
FROM
	tracker_patterns
WHERE
	mapping_requested_at IS NOT NULL
ORDER BY
	mapping_requested_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 1;
`

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query tracker patterns for mapping: %w", err)
	}

	pattern, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrackerPattern])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect tracker pattern for mapping: %w", err)
	}

	*tp = pattern

	return nil
}

func (tp *TrackerPattern) ClearMappingRequestedAt(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE tracker_patterns
SET mapping_requested_at = NULL
WHERE id = @id
`

	args := pgx.StrictNamedArgs{"id": tp.ID}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot clear mapping requested at: %w", err)
	}

	tp.MappingRequestedAt = nil

	return nil
}

func (tp *TrackerPattern) SetMappingRequested(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE tracker_patterns
SET mapping_requested_at = NOW()
WHERE id = @id
  AND mapping_requested_at IS NULL
`

	args := pgx.StrictNamedArgs{"id": tp.ID}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot set mapping requested: %w", err)
	}

	return nil
}
