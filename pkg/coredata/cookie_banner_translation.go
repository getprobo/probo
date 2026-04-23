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
	"encoding/json"
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
	CookieBannerTranslation struct {
		ID             gid.GID         `db:"id"`
		OrganizationID gid.GID         `db:"organization_id"`
		CookieBannerID gid.GID         `db:"cookie_banner_id"`
		Language       string          `db:"language"`
		Translations   json.RawMessage `db:"translations"`
		CreatedAt      time.Time       `db:"created_at"`
		UpdatedAt      time.Time       `db:"updated_at"`
	}

	CookieBannerTranslations []*CookieBannerTranslation
)

func (t *CookieBannerTranslation) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM cookie_banner_translations WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, t.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot query cookie banner translation authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (t *CookieBannerTranslation) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	translationID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	language,
	translations,
	created_at,
	updated_at
FROM
	cookie_banner_translations
WHERE
	%s
	AND id = @translation_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"translation_id": translationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner translations: %w", err)
	}

	translation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerTranslation])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect cookie banner translation: %w", err)
	}

	*t = translation

	return nil
}

func (t *CookieBannerTranslation) LoadByCookieBannerIDAndLanguage(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	language string,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	language,
	translations,
	created_at,
	updated_at
FROM
	cookie_banner_translations
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND language = @language
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"language":         language,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner translations: %w", err)
	}

	translation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerTranslation])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect cookie banner translation: %w", err)
	}

	*t = translation

	return nil
}

func (t *CookieBannerTranslations) LoadAllByCookieBannerID(
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
	language,
	translations,
	created_at,
	updated_at
FROM
	cookie_banner_translations
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
ORDER BY
	language ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner translations: %w", err)
	}

	translations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CookieBannerTranslation])
	if err != nil {
		return fmt.Errorf("cannot collect cookie banner translations: %w", err)
	}

	*t = translations

	return nil
}

func (t *CookieBannerTranslation) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cookie_banner_translations (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	language,
	translations,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@language,
	@translations,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               t.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  t.OrganizationID,
		"cookie_banner_id": t.CookieBannerID,
		"language":         t.Language,
		"translations":     t.Translations,
		"created_at":       t.CreatedAt,
		"updated_at":       t.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_cookie_banner_translations_unique_language_per_banner" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert cookie banner translation: %w", err)
	}

	return nil
}

func (t *CookieBannerTranslation) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE cookie_banner_translations
SET
	translations = @translations,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           t.ID,
		"translations": t.Translations,
		"updated_at":   t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update cookie banner translation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (t *CookieBannerTranslation) Delete(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM cookie_banner_translations
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": t.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete cookie banner translation: %w", err)
	}

	return nil
}
