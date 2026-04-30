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
)

type (
	Cookie struct {
		ID              gid.GID      `db:"id"`
		OrganizationID  gid.GID      `db:"organization_id"`
		CookieBannerID  gid.GID      `db:"cookie_banner_id"`
		CookiePatternID gid.GID      `db:"cookie_pattern_id"`
		Name            string       `db:"name"`
		MaxAgeSeconds   *int         `db:"max_age_seconds"`
		Source          CookieSource `db:"source"`
		CreatedAt       time.Time    `db:"created_at"`
		UpdatedAt       time.Time    `db:"updated_at"`
	}

	Cookies []*Cookie
)

func (c *Cookies) CountByCookiePatternID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookiePatternID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	cookies
WHERE
	%s
	AND cookie_pattern_id = @cookie_pattern_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_pattern_id": cookiePatternID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
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
	cookie_pattern_id,
	name,
	max_age_seconds,
	source,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_pattern_id,
	@name,
	@max_age_seconds,
	@source,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                c.ID,
		"tenant_id":         scope.GetTenantID(),
		"organization_id":   c.OrganizationID,
		"cookie_banner_id":  c.CookieBannerID,
		"cookie_pattern_id": c.CookiePatternID,
		"name":              c.Name,
		"max_age_seconds":   c.MaxAgeSeconds,
		"source":            c.Source,
		"created_at":        c.CreatedAt,
		"updated_at":        c.UpdatedAt,
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
	cookie_pattern_id,
	name,
	max_age_seconds,
	source,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@cookie_pattern_id,
	@name,
	@max_age_seconds,
	@source,
	@created_at,
	@updated_at
)
ON CONFLICT (cookie_banner_id, name) DO UPDATE
	SET source = EXCLUDED.source, updated_at = EXCLUDED.updated_at
	WHERE cookies.source != @source_script AND EXCLUDED.source = @source_script
`

	args := pgx.StrictNamedArgs{
		"id":                c.ID,
		"tenant_id":         scope.GetTenantID(),
		"organization_id":   c.OrganizationID,
		"cookie_banner_id":  c.CookieBannerID,
		"cookie_pattern_id": c.CookiePatternID,
		"name":              c.Name,
		"max_age_seconds":   c.MaxAgeSeconds,
		"source":            c.Source,
		"source_script":     CookieSourceScript,
		"created_at":        c.CreatedAt,
		"updated_at":        c.UpdatedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot insert cookie: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (c *Cookies) RelinkByCookiePatternID(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	sourcePatternID gid.GID,
	targetPatternID gid.GID,
) error {
	q := `
UPDATE cookies
SET
	cookie_pattern_id = @target_pattern_id,
	updated_at = @updated_at
WHERE
	%s
	AND cookie_pattern_id = @source_pattern_id
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
		return fmt.Errorf("cannot relink cookies to pattern: %w", err)
	}

	return nil
}
