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
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
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

func (t CookieBannerTranslation) CursorKey(field CookieBannerTranslationOrderField) page.CursorKey {
	switch field {
	case CookieBannerTranslationOrderFieldLanguage:
		return page.NewCursorKey(t.ID, t.Language)
	case CookieBannerTranslationOrderFieldCreatedAt:
		return page.NewCursorKey(t.ID, t.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (t *CookieBannerTranslation) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cookie_banner_translations WHERE id = ANY(@resource_ids::text[])`

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

func (t *CookieBannerTranslations) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[CookieBannerTranslationOrderField],
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
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

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
