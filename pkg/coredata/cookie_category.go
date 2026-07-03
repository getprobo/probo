// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"encoding/json"
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
	CookieItem struct {
		Name          string      `json:"name"`
		TrackerType   TrackerType `json:"tracker_type"`
		MaxAgeSeconds *int        `json:"max_age_seconds"`
		Description   string      `json:"description"`
	}

	CookieItems []CookieItem

	cookieDurationUnit struct {
		secs int
		name string
		snap int
	}

	CookieCategory struct {
		ID              gid.GID            `db:"id"`
		OrganizationID  gid.GID            `db:"organization_id"`
		CookieBannerID  gid.GID            `db:"cookie_banner_id"`
		Name            string             `db:"name"`
		Slug            string             `db:"slug"`
		Description     string             `db:"description"`
		Kind            CookieCategoryKind `db:"kind"`
		Rank            int                `db:"rank"`
		GCMConsentTypes []string           `db:"gcm_consent_types"`
		PostHogConsent  bool               `db:"posthog_consent"`
		CreatedAt       time.Time          `db:"created_at"`
		UpdatedAt       time.Time          `db:"updated_at"`
	}

	CookieCategories []*CookieCategory
)

// cookieDurationUnits mirrors DURATION_UNITS in
// packages/cookie-banner/src/cookie-utils.ts (and the snap table used by the
// pattern-analysis worker) so durations rendered server-side match exactly what
// visitors see in the consent banner. snap is the per-unit buffer: when the
// remainder is within snap seconds of the next whole unit, round up instead of
// carrying into smaller units.
var cookieDurationUnits = [...]cookieDurationUnit{
	{365 * 24 * 3600, "year", 21 * 24 * 3600},
	{30 * 24 * 3600, "month", 2 * 24 * 3600},
	{7 * 24 * 3600, "week", 12 * 3600},
	{24 * 3600, "day", 2 * 3600},
	{3600, "hour", 5 * 60},
	{60, "minute", 5},
	{1, "second", 0},
}

// HumanizedDuration renders the tracker's max-age into a human-readable
// lifetime using the same snapping and composition rules as the banner's
// humanizeDuration helper. A nil or non-positive max-age has no fixed
// expiry, so the lifetime is described by the tracker type: session cookies
// are cleared when the browser closes, session storage is cleared when the
// tab closes, and the remaining storage technologies persist until cleared.
func (c CookieItem) HumanizedDuration() string {
	if c.MaxAgeSeconds == nil || *c.MaxAgeSeconds <= 0 {
		switch c.TrackerType {
		case TrackerTypeSessionStorage:
			return "Until the tab is closed"
		case TrackerTypeLocalStorage, TrackerTypeIndexedDB, TrackerTypeCacheStorage:
			return "Persistent"
		default:
			return "Session"
		}
	}

	remaining := *c.MaxAgeSeconds

	var parts []string

	for _, u := range cookieDurationUnits {
		if remaining < u.secs-u.snap {
			continue
		}

		count := remaining / u.secs

		leftover := remaining - count*u.secs
		switch {
		case leftover >= u.secs-u.snap:
			count++
			remaining = 0
		case leftover <= u.snap:
			remaining = 0
		default:
			remaining = leftover
		}

		if count == 1 {
			parts = append(parts, fmt.Sprintf("1 %s", u.name))
		} else {
			parts = append(parts, fmt.Sprintf("%d %ss", count, u.name))
		}
	}

	if len(parts) == 0 {
		return "Session"
	}

	return strings.Join(parts, ", ")
}

func (c CookieItems) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("[]"), nil
	}

	return json.Marshal([]CookieItem(c))
}

func (c *CookieItems) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*c = CookieItems{}
		return nil
	}

	return json.Unmarshal(data, (*[]CookieItem)(c))
}

func (c *CookieCategory) CursorKey(field CookieCategoryOrderField) page.CursorKey {
	switch field {
	case CookieCategoryOrderFieldRank:
		return page.NewCursorKey(c.ID, c.Rank)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (c *CookieCategory) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cookie_categories WHERE id = ANY(@resource_ids::text[])`

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

func (c *CookieCategory) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	categoryID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	name,
	slug,
	description,
	kind,
	rank,
	gcm_consent_types,
	posthog_consent,
	created_at,
	updated_at
FROM
	cookie_categories
WHERE
	%s
	AND id = @category_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"category_id": categoryID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie categories: %w", err)
	}

	category, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieCategory])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cookie category: %w", err)
	}

	*c = category

	return nil
}

func (c *CookieCategories) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	categoryIDs []gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	name,
	slug,
	description,
	kind,
	rank,
	gcm_consent_types,
	posthog_consent,
	created_at,
	updated_at
FROM
	cookie_categories
WHERE
	%s
	AND id = ANY(@category_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"category_ids": categoryIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie categories: %w", err)
	}

	categories, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CookieCategory])
	if err != nil {
		return fmt.Errorf("cannot collect cookie categories: %w", err)
	}

	*c = categories

	if len(categories) != len(gid.NewSet(categoryIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (c *CookieCategories) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[CookieCategoryOrderField],
	filter *CookieCategoryFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	name,
	slug,
	description,
	kind,
	rank,
	gcm_consent_types,
	posthog_consent,
	created_at,
	updated_at
FROM
	cookie_categories
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie categories: %w", err)
	}

	categories, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CookieCategory])
	if err != nil {
		return fmt.Errorf("cannot collect cookie categories: %w", err)
	}

	*c = categories

	return nil
}

func (c *CookieCategories) CountByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	filter *CookieCategoryFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	cookie_categories
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (c *CookieCategory) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cookie_categories (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	name,
	slug,
	description,
	kind,
	rank,
	gcm_consent_types,
	posthog_consent,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@name,
	@slug,
	@description,
	@kind,
	@rank,
	@gcm_consent_types,
	@posthog_consent,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                c.ID,
		"tenant_id":         scope.GetTenantID(),
		"organization_id":   c.OrganizationID,
		"cookie_banner_id":  c.CookieBannerID,
		"name":              c.Name,
		"slug":              c.Slug,
		"description":       c.Description,
		"kind":              c.Kind,
		"rank":              c.Rank,
		"gcm_consent_types": c.GCMConsentTypes,
		"posthog_consent":   c.PostHogConsent,
		"created_at":        c.CreatedAt,
		"updated_at":        c.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_cookie_categories_unique_slug_per_banner" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert cookie category: %w", err)
	}

	return nil
}

func (c *CookieCategory) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE cookie_categories
SET
	name = @name,
	slug = @slug,
	description = @description,
	gcm_consent_types = @gcm_consent_types,
	posthog_consent = @posthog_consent,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                c.ID,
		"name":              c.Name,
		"slug":              c.Slug,
		"description":       c.Description,
		"gcm_consent_types": c.GCMConsentTypes,
		"posthog_consent":   c.PostHogConsent,
		"updated_at":        c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_cookie_categories_unique_slug_per_banner" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot update cookie category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (c *CookieCategory) UpdateRank(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
WITH old AS (
	SELECT rank AS old_rank
	FROM cookie_categories
	WHERE %s AND id = @id AND cookie_banner_id = @cookie_banner_id
)
UPDATE cookie_categories
SET
	rank = CASE
		WHEN id = @id THEN @new_rank
		ELSE rank + CASE
			WHEN @new_rank < old.old_rank THEN 1
			WHEN @new_rank > old.old_rank THEN -1
		END
	END,
	updated_at = @updated_at
FROM old
WHERE %s
	AND cookie_banner_id = @cookie_banner_id
	AND (
		id = @id
		OR (rank BETWEEN LEAST(old.old_rank, @new_rank) AND GREATEST(old.old_rank, @new_rank))
	);
`

	scopeFragment := scope.SQLFragment()
	q = fmt.Sprintf(q, scopeFragment, scopeFragment)

	args := pgx.StrictNamedArgs{
		"id":               c.ID,
		"new_rank":         c.Rank,
		"cookie_banner_id": c.CookieBannerID,
		"updated_at":       c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update cookie category rank: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (c *CookieCategory) Delete(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM cookie_categories
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete cookie category: %w", err)
	}

	return nil
}

func (c *CookieCategories) ClearPostHogConsentByBannerID(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	q := `
UPDATE cookie_categories
SET
	posthog_consent = false,
	updated_at = @updated_at
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND posthog_consent = true
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"updated_at":       time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot clear posthog consent: %w", err)
	}

	return nil
}

func (c *CookieCategory) LoadUncategorisedByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	name,
	slug,
	description,
	kind,
	rank,
	gcm_consent_types,
	posthog_consent,
	created_at,
	updated_at
FROM
	cookie_categories
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND kind = @kind
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"kind":             CookieCategoryKindUncategorised,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query uncategorised cookie category: %w", err)
	}

	category, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieCategory])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect uncategorised cookie category: %w", err)
	}

	*c = category

	return nil
}
