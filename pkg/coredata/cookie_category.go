// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	CookieItem struct {
		Name        string `json:"name"`
		Duration    string `json:"duration"`
		Description string `json:"description"`
	}

	CookieItems []CookieItem

	CookieCategory struct {
		ID             gid.GID     `db:"id"`
		CookieBannerID gid.GID     `db:"cookie_banner_id"`
		Name           string      `db:"name"`
		Description    string      `db:"description"`
		Required       bool        `db:"required"`
		Rank           int         `db:"rank"`
		Cookies        CookieItems `db:"cookies"`
		CreatedAt      time.Time   `db:"created_at"`
		UpdatedAt      time.Time   `db:"updated_at"`
	}

	CookieCategories []*CookieCategory
)

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

func (c *CookieCategory) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `
SELECT cb.organization_id
FROM cookie_categories cc
JOIN cookie_banners cb ON cc.cookie_banner_id = cb.id
WHERE cc.id = $1
LIMIT 1;
`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, c.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query cookie category authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
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
	cookie_banner_id,
	name,
	description,
	required,
	rank,
	cookies,
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

func (c *CookieCategories) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[CookieCategoryOrderField],
) error {
	q := `
SELECT
	id,
	cookie_banner_id,
	name,
	description,
	required,
	rank,
	cookies,
	created_at,
	updated_at
FROM
	cookie_categories
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
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
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	cookie_categories
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (c *CookieCategories) LoadAllByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	cookieBannerID gid.GID,
) error {
	q := `
SELECT
	id,
	cookie_banner_id,
	name,
	description,
	required,
	rank,
	cookies,
	created_at,
	updated_at
FROM
	cookie_categories
WHERE
	cookie_banner_id = @cookie_banner_id
ORDER BY
	rank ASC, id ASC;
`

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}

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

func (c *CookieCategory) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cookie_categories (
	id,
	tenant_id,
	cookie_banner_id,
	name,
	description,
	required,
	rank,
	cookies,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@cookie_banner_id,
	@name,
	@description,
	@required,
	@rank,
	@cookies,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               c.ID,
		"tenant_id":        scope.GetTenantID(),
		"cookie_banner_id": c.CookieBannerID,
		"name":             c.Name,
		"description":      c.Description,
		"required":         c.Required,
		"rank":             c.Rank,
		"cookies":          c.Cookies,
		"created_at":       c.CreatedAt,
		"updated_at":       c.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
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
	description = @description,
	rank = @rank,
	cookies = @cookies,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
RETURNING
	id,
	cookie_banner_id,
	name,
	description,
	required,
	rank,
	cookies,
	created_at,
	updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          c.ID,
		"name":        c.Name,
		"description": c.Description,
		"rank":        c.Rank,
		"cookies":     c.Cookies,
		"updated_at":  c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update cookie category: %w", err)
	}

	category, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieCategory])
	if err != nil {
		return fmt.Errorf("cannot collect updated cookie category: %w", err)
	}

	*c = category

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
