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
	Cookie struct {
		ID               gid.GID   `db:"id"`
		OrganizationID   gid.GID   `db:"organization_id"`
		CookieBannerID   gid.GID   `db:"cookie_banner_id"`
		CookieCategoryID gid.GID   `db:"cookie_category_id"`
		Name             string    `db:"name"`
		Duration         string    `db:"duration"`
		Description      string    `db:"description"`
		CreatedAt        time.Time `db:"created_at"`
		UpdatedAt        time.Time `db:"updated_at"`
	}

	Cookies []*Cookie
)

func (c *Cookie) CursorKey(field CookieOrderField) page.CursorKey {
	switch field {
	case CookieOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (c *Cookie) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM cookies WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, c.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot query cookie authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (c *Cookie) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	name,
	duration,
	description,
	created_at,
	updated_at
FROM
	cookies
WHERE
	%s
	AND id = @cookie_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_id": cookieID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookies: %w", err)
	}

	cookie, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Cookie])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect cookie: %w", err)
	}

	*c = cookie

	return nil
}

func (c *Cookies) LoadByCookieCategoryID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieCategoryID gid.GID,
	cursor *page.Cursor[CookieOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	name,
	duration,
	description,
	created_at,
	updated_at
FROM
	cookies
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
		return fmt.Errorf("cannot query cookies: %w", err)
	}

	cookies, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Cookie])
	if err != nil {
		return fmt.Errorf("cannot collect cookies: %w", err)
	}

	*c = cookies

	return nil
}

func (c *Cookies) CountByCookieCategoryID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieCategoryID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	cookies
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

func (c *Cookies) LoadAllByCookieBannerID(
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
	cookie_category_id,
	name,
	duration,
	description,
	created_at,
	updated_at
FROM
	cookies
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
ORDER BY
	created_at ASC, id ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookies: %w", err)
	}

	cookies, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Cookie])
	if err != nil {
		return fmt.Errorf("cannot collect cookies: %w", err)
	}

	*c = cookies

	return nil
}

func (c *Cookie) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cookies (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	name,
	duration,
	description,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_category_id,
	@name,
	@duration,
	@description,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                 c.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    c.OrganizationID,
		"cookie_banner_id":   c.CookieBannerID,
		"cookie_category_id": c.CookieCategoryID,
		"name":               c.Name,
		"duration":           c.Duration,
		"description":        c.Description,
		"created_at":         c.CreatedAt,
		"updated_at":         c.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_cookies_unique_name_per_banner" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert cookie: %w", err)
	}

	return nil
}

func (c *Cookie) InsertIfNotExists(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO cookies (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	cookie_category_id,
	name,
	duration,
	description,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_category_id,
	@name,
	@duration,
	@description,
	@created_at,
	@updated_at
)
ON CONFLICT (cookie_banner_id, name) DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"id":                 c.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    c.OrganizationID,
		"cookie_banner_id":   c.CookieBannerID,
		"cookie_category_id": c.CookieCategoryID,
		"name":               c.Name,
		"duration":           c.Duration,
		"description":        c.Description,
		"created_at":         c.CreatedAt,
		"updated_at":         c.UpdatedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot insert cookie: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (c *Cookie) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE cookies
SET
	cookie_category_id = @cookie_category_id,
	name = @name,
	duration = @duration,
	description = @description,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                 c.ID,
		"cookie_category_id": c.CookieCategoryID,
		"name":               c.Name,
		"duration":           c.Duration,
		"description":        c.Description,
		"updated_at":         c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_cookies_unique_name_per_banner" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot update cookie: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (c *Cookie) Delete(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM cookies
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete cookie: %w", err)
	}

	return nil
}

func (c *Cookies) MoveToCategoryByCookieCategoryID(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	sourceCategoryID gid.GID,
	targetCategoryID gid.GID,
) error {
	q := `
UPDATE cookies
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
		return fmt.Errorf("cannot move cookies to category: %w", err)
	}

	return nil
}
