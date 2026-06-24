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

// UpdateMapping writes only the columns the tracker-mapping worker
// resolves — common_tracker_pattern_id, third_party_id and an enriched
// description — leaving the user-editable fields (display_name,
// excluded, cookie_category_id, max_age_seconds, source, last_matched_at)
// untouched.
//
// The worker loads the pattern in its claim transaction and commits the
// resolution in a separate, later transaction, so a full-row Update
// would write back stale values and clobber any concurrent edit made in
// between. The description is only filled when still empty in the
// database, so a concurrently set description is never overwritten.
func (tp *TrackerPattern) UpdateMapping(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE tracker_patterns
SET
	common_tracker_pattern_id = @common_tracker_pattern_id,
	third_party_id = @third_party_id,
	description = CASE
		WHEN description = '' THEN @description
		ELSE description
	END,
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
		"description":               tp.Description,
		"updated_at":                tp.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update tracker pattern mapping: %w", err)
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

// LinkThirdPartyByCommonThirdPartyID points the organization's unlinked
// tracker patterns at an org ThirdParty when their catalog row resolves
// to the given common third party. It is the backfill the explicit
// import action runs so patterns that previously surfaced the catalog
// (CommonThirdParty) entry now surface the managed org ThirdParty. Only
// patterns with no third_party_id are touched, so it is idempotent and
// never overrides an existing link. The common_tracker_patterns
// subquery only narrows the WHERE clause, keeping the resolution in the
// database.
func (tps *TrackerPatterns) LinkThirdPartyByCommonThirdPartyID(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	organizationID gid.GID,
	commonThirdPartyID gid.GID,
	thirdPartyID gid.GID,
) error {
	q := `
UPDATE tracker_patterns
SET
	third_party_id = @third_party_id,
	updated_at = NOW()
WHERE
	%s
	AND organization_id = @organization_id
	AND third_party_id IS NULL
	AND common_tracker_pattern_id IN (
		SELECT id FROM common_tracker_patterns
		WHERE common_third_party_id = @common_third_party_id
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":       organizationID,
		"common_third_party_id": commonThirdPartyID,
		"third_party_id":        thirdPartyID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot link tracker patterns to third party: %w", err)
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

// ClearMappingRequestedAt removes the row from the mapping queue. It
// bumps updated_at so the stale-recovery clock starts at claim time,
// keeping ResetStaleMappings from re-arming a row that is still being
// processed.
func (tp *TrackerPattern) ClearMappingRequestedAt(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE tracker_patterns
SET
    mapping_requested_at = NULL,
    updated_at = NOW()
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

// ResetStaleMappings re-arms mapping_requested_at on rows whose mapping
// was claimed but never completed (no common_tracker_pattern_id) and
// have been idle longer than staleAfter, so a crashed or timed-out
// mapping run is retried. A successful Process always assigns a catalog
// row (the unmatched fallback in createUnmatchedPattern), so a missing
// common_tracker_pattern_id on a dequeued row marks an interrupted run.
//
// Like the claim query, this sweep is intentionally cross-tenant: the
// mapping worker is a system worker that drains the queue regardless of
// tenant.
func ResetStaleMappings(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
) error {
	q := `
UPDATE tracker_patterns
SET
    mapping_requested_at = NOW(),
    updated_at = NOW()
WHERE
    mapping_requested_at IS NULL
    AND common_tracker_pattern_id IS NULL
    AND updated_at < @stale_before
`

	args := pgx.StrictNamedArgs{"stale_before": time.Now().Add(-staleAfter)}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reset stale tracker pattern mappings: %w", err)
	}

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

// RequestMappingForUnmappedSiblings re-arms mapping_requested_at on
// sibling tracker patterns of the same banner that share an initiator
// domain with the just-mapped pattern but are still unpromoted. It is
// the backward-propagation counterpart to the mapping worker's
// sibling-origin matching: when a pattern newly resolves a vendor, its
// siblings that were processed earlier and left unmatched can now be
// re-evaluated against it.
//
// Only siblings still genuinely unresolved are touched: not promoted to
// an org party (third_party_id IS NULL), not already linked to a catalog
// row that carries a common third party, and not marked FIRST_PARTY
// (a terminal verdict). third_party_id IS NULL alone is no longer a
// sufficient guard: since org-party auto-creation was dropped a pattern
// can resolve a common third party yet stay third_party_id IS NULL, and
// re-enqueueing those (or first-party siblings) on every cascade step is
// what amplified reprocessing to O(N^2) per banner. The siblings must
// also be not-already-queued (mapping_requested_at IS NULL) and
// non-extension. A fully mapped banner re-enqueues nothing.
// common_tracker_patterns and detected_trackers are used only as
// filtering subqueries. Returns the number of siblings re-enqueued.
func (tps *TrackerPatterns) RequestMappingForUnmappedSiblings(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieBannerID gid.GID,
	excludePatternID gid.GID,
	domains []string,
) (int64, error) {
	if len(domains) == 0 {
		return 0, nil
	}

	// The target rows are locked through an ORDER BY id ... FOR UPDATE
	// subquery so concurrent re-enqueues over overlapping sibling sets
	// always acquire their row locks in the same ascending id order. Two
	// workers mapping sibling patterns on the same banner would otherwise
	// lock the shared rows in opposite orders and deadlock (40P01).
	q := `
UPDATE tracker_patterns
SET
	mapping_requested_at = NOW(),
	updated_at = NOW()
WHERE id IN (
	SELECT id
	FROM tracker_patterns
	WHERE
		%[1]s
		AND cookie_banner_id = @cookie_banner_id
		AND id != @exclude_pattern_id
		AND third_party_id IS NULL
		AND mapping_requested_at IS NULL
		AND (source IS NULL OR source != @extension_source)
		AND NOT EXISTS (
			SELECT 1
			FROM common_tracker_patterns ctp
			WHERE ctp.id = tracker_patterns.common_tracker_pattern_id
				AND (
					ctp.common_third_party_id IS NOT NULL
					OR ctp.attribution = 'FIRST_PARTY'
				)
		)
		AND id IN (
			SELECT DISTINCT tracker_pattern_id
			FROM detected_trackers
			WHERE %[1]s
				AND cookie_banner_id = @cookie_banner_id
				AND initiator_domain = ANY(@domains)
				AND tracker_pattern_id IS NOT NULL
		)
	ORDER BY id
	FOR UPDATE
)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id":   cookieBannerID,
		"exclude_pattern_id": excludePatternID,
		"extension_source":   CookieSourceExtension,
		"domains":            domains,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot request mapping for unmapped siblings: %w", err)
	}

	return result.RowsAffected(), nil
}

// BackfillDescriptionByCommonTrackerPatternID copies an enriched
// description onto every tracker pattern linked to the given common
// pattern that does not yet have one. It is invoked by the common-pattern
// enrichment worker, a global system process, so it is intentionally not
// tenant-scoped: a single catalog enrichment fans out to all tenants'
// linked patterns. The description = ” guard guarantees a pattern that
// already carries a description is never overwritten. Returns the number
// of patterns backfilled.
func (tps *TrackerPatterns) BackfillDescriptionByCommonTrackerPatternID(
	ctx context.Context,
	tx pg.Tx,
	commonTrackerPatternID gid.GID,
	description string,
) (int64, error) {
	q := `
UPDATE tracker_patterns
SET
	description = @description,
	updated_at = NOW()
WHERE
	common_tracker_pattern_id = @common_tracker_pattern_id
	AND description = ''
`

	args := pgx.StrictNamedArgs{
		"common_tracker_pattern_id": commonTrackerPatternID,
		"description":               description,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot backfill tracker pattern descriptions: %w", err)
	}

	return result.RowsAffected(), nil
}

// RequestMappingForUncategorisedByCommonTrackerPatternIDs re-arms mapping
// on the uncategorised org tracker patterns linked to the given common
// tracker patterns: it clears their resolved org third party and stamps
// mapping_requested_at so the mapping worker re-resolves the vendor from
// the catalog row's (now changed) common third party. Like the
// description backfill it is a global catalog operation, so it is
// intentionally not tenant-scoped. Excluded patterns and patterns in
// user-chosen categories are left untouched - only the uncategorised
// category is remapped, matching the reset-trackers philosophy. The
// cookie_categories subquery is used only for filtering. Returns the
// number of org patterns re-armed.
func (tps *TrackerPatterns) RequestMappingForUncategorisedByCommonTrackerPatternIDs(
	ctx context.Context,
	tx pg.Tx,
	commonIDs []gid.GID,
) (int64, error) {
	q := `
UPDATE tracker_patterns
SET
	third_party_id = NULL,
	mapping_requested_at = NOW(),
	updated_at = NOW()
WHERE
	common_tracker_pattern_id = ANY(@common_ids)
	AND excluded = false
	AND cookie_category_id IN (
		SELECT id FROM cookie_categories WHERE kind = @uncategorised_kind
	)
`

	args := pgx.StrictNamedArgs{
		"common_ids":         commonIDs,
		"uncategorised_kind": CookieCategoryKindUncategorised,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot request mapping for uncategorised tracker patterns: %w", err)
	}

	return result.RowsAffected(), nil
}

// ClearDescriptionForUncategorisedByCommonTrackerPatternIDs blanks the
// description on the uncategorised org tracker patterns linked to the
// given common tracker patterns. It pairs with the first-party verdict on
// the catalog row: when the catalog description is cleared because its
// vendor attribution was wrong, the descriptions fanned out to org
// patterns named the same stale vendor and must be cleared too - the
// mapping worker only ever copies a description into an empty org row, it
// never clears one. Like the mapping re-arm it is a global catalog
// operation, so it is intentionally not tenant-scoped, and it leaves
// excluded and user-categorised patterns untouched. The cookie_categories
// subquery is used only for filtering. Returns the number of org patterns
// cleared.
func (tps *TrackerPatterns) ClearDescriptionForUncategorisedByCommonTrackerPatternIDs(
	ctx context.Context,
	tx pg.Tx,
	commonIDs []gid.GID,
) (int64, error) {
	q := `
UPDATE tracker_patterns
SET
	description = '',
	updated_at = NOW()
WHERE
	common_tracker_pattern_id = ANY(@common_ids)
	AND excluded = false
	AND cookie_category_id IN (
		SELECT id FROM cookie_categories WHERE kind = @uncategorised_kind
	)
`

	args := pgx.StrictNamedArgs{
		"common_ids":         commonIDs,
		"uncategorised_kind": CookieCategoryKindUncategorised,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot clear uncategorised tracker pattern descriptions: %w", err)
	}

	return result.RowsAffected(), nil
}

// RequestMappingForUnmappedByInitiatorDomains re-arms mapping on the
// still-unmapped org tracker patterns whose detected trackers share one
// of the given initiator domains, so the mapping worker re-resolves them
// via its domain-overlap signal. It is invoked by the common-third-party
// enrichment worker when a vendor gains new owned domains, a global
// catalog operation, so it is intentionally not tenant-scoped: a single
// catalog domain benefits all tenants' patterns.
//
// Only patterns with no resolved vendor are targeted (third_party_id IS
// NULL and an absent or unlinked catalog row), so a pattern already
// linked to the vendor that gained the domain - or to any other vendor -
// is never disturbed and never re-attributed. Extension-sourced patterns
// and patterns already queued for mapping are skipped. The
// common_tracker_patterns and detected_trackers subqueries are used only
// for filtering. Returns the number of patterns re-armed; an empty
// domains slice is a no-op.
func (tps *TrackerPatterns) RequestMappingForUnmappedByInitiatorDomains(
	ctx context.Context,
	tx pg.Tx,
	domains []string,
) (int64, error) {
	if len(domains) == 0 {
		return 0, nil
	}

	q := `
UPDATE tracker_patterns
SET
	mapping_requested_at = NOW(),
	updated_at = NOW()
WHERE
	third_party_id IS NULL
	AND mapping_requested_at IS NULL
	AND (source IS NULL OR source != @extension_source)
	AND (
		common_tracker_pattern_id IS NULL
		OR common_tracker_pattern_id IN (
			SELECT id FROM common_tracker_patterns
			WHERE common_third_party_id IS NULL
		)
	)
	AND id IN (
		SELECT DISTINCT tracker_pattern_id
		FROM detected_trackers
		WHERE initiator_domain = ANY(@domains)
			AND tracker_pattern_id IS NOT NULL
	)
`

	args := pgx.StrictNamedArgs{
		"extension_source": CookieSourceExtension,
		"domains":          domains,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot request mapping for unmapped tracker patterns by initiator domains: %w", err)
	}

	return result.RowsAffected(), nil
}

// ResetAndRequestMappingByCookieCategoryID detaches every pattern in the
// given category from its catalog row, org third party, and copied
// description, then re-arms mapping. Operators run this (via proboctl) on
// a banner's uncategorised category to force a clean re-map when
// iterating on the mapping agent. Excluded patterns are left untouched -
// exclusion is a deliberate suppression. The cookie_category_id key
// scopes the reset to the uncategorised category the caller resolves;
// the Scoper keeps it tenant-isolated. When keyword is non-nil and
// non-empty, the reset is further restricted to patterns whose pattern or
// display name contains it (case-insensitive). Returns the number of
// patterns reset.
func (tps *TrackerPatterns) ResetAndRequestMappingByCookieCategoryID(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieCategoryID gid.GID,
	keyword *string,
) (int64, error) {
	filter := NewTrackerPatternFilter(nil, nil, nil).WithPatternKeyword(keyword)

	q := `
UPDATE tracker_patterns
SET
	common_tracker_pattern_id = NULL,
	third_party_id = NULL,
	description = '',
	mapping_requested_at = NOW(),
	updated_at = NOW()
WHERE
	%s
	AND cookie_category_id = @cookie_category_id
	AND excluded = false
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_category_id": cookieCategoryID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot reset and request mapping by cookie category: %w", err)
	}

	return result.RowsAffected(), nil
}

// LoadAllLinkedCommonTrackerPatternIDsByCookieBannerID returns every
// distinct common_tracker_pattern_id referenced by the banner's patterns,
// regardless of mapping state. Unlike
// LoadDistinctCommonTrackerPatternIDsByCookieBannerID (which restricts to
// unmapped patterns for the mapping pipeline), this returns the full set
// of catalog rows the banner depends on, so an operator can re-describe
// exactly those before a reset.
func (tps *TrackerPatterns) LoadAllLinkedCommonTrackerPatternIDsByCookieBannerID(
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
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query linked common tracker pattern ids: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect linked common tracker pattern ids: %w", err)
	}

	return ids, nil
}

// LoadAllLinkedCommonTrackerPatternIDsByOrganizationID is the org-wide
// counterpart of LoadAllLinkedCommonTrackerPatternIDsByCookieBannerID:
// every distinct catalog row the organization's tracker patterns depend
// on, regardless of mapping state.
func (tps *TrackerPatterns) LoadAllLinkedCommonTrackerPatternIDsByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) ([]gid.GID, error) {
	q := `
SELECT DISTINCT common_tracker_pattern_id
FROM tracker_patterns
WHERE
	%s
	AND organization_id = @organization_id
	AND common_tracker_pattern_id IS NOT NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query linked common tracker pattern ids: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect linked common tracker pattern ids: %w", err)
	}

	return ids, nil
}
