// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	TrustCenterReference struct {
		ID            gid.GID   `db:"id"`
		TrustCenterID gid.GID   `db:"trust_center_id"`
		Name          string    `db:"name"`
		Description   string    `db:"description"`
		WebsiteURL    string    `db:"website_url"`
		LogoFileID    gid.GID   `db:"logo_file_id"`
		CreatedAt     time.Time `db:"created_at"`
		UpdatedAt     time.Time `db:"updated_at"`
	}

	TrustCenterReferences []*TrustCenterReference
)

func (t TrustCenterReference) CursorKey(orderBy TrustCenterReferenceOrderField) page.CursorKey {
	switch orderBy {
	case TrustCenterReferenceOrderFieldName:
		return page.NewCursorKey(t.ID, t.Name)
	case TrustCenterReferenceOrderFieldCreatedAt:
		return page.NewCursorKey(t.ID, t.CreatedAt)
	case TrustCenterReferenceOrderFieldUpdatedAt:
		return page.NewCursorKey(t.ID, t.UpdatedAt)
	}
	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (t *TrustCenterReference) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterReferenceID gid.GID,
) error {
	q := `
SELECT
    id,
    trust_center_id,
    name,
    description,
    website_url,
    logo_file_id,
    created_at,
    updated_at
FROM
    trust_center_references
WHERE
    %s
    AND id = @trust_center_reference_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_reference_id": trustCenterReferenceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust_center_references: %w", err)
	}

	reference, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterReference])
	if err != nil {
		return fmt.Errorf("cannot collect trust center reference: %w", err)
	}

	*t = reference

	return nil
}

func (t TrustCenterReference) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    trust_center_references (
        tenant_id,
        id,
        trust_center_id,
        name,
        description,
        website_url,
        logo_file_id,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @trust_center_id,
    @name,
    @description,
    @website_url,
    @logo_file_id,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"id":              t.ID,
		"trust_center_id": t.TrustCenterID,
		"name":            t.Name,
		"description":     t.Description,
		"website_url":     t.WebsiteURL,
		"logo_file_id":    t.LogoFileID,
		"created_at":      t.CreatedAt,
		"updated_at":      t.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert trust center reference: %w", err)
	}

	return nil
}

func (t *TrustCenterReference) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE trust_center_references
SET
    name = @name,
    description = @description,
    website_url = @website_url,
    logo_file_id = @logo_file_id,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
RETURNING
    id,
    trust_center_id,
    name,
    description,
    website_url,
    logo_file_id,
    created_at,
    updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           t.ID,
		"name":         t.Name,
		"description":  t.Description,
		"website_url":  t.WebsiteURL,
		"logo_file_id": t.LogoFileID,
		"updated_at":   t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update trust center reference: %w", err)
	}

	reference, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterReference])
	if err != nil {
		return fmt.Errorf("cannot collect updated trust center reference: %w", err)
	}

	*t = reference

	return nil
}

func (t *TrustCenterReference) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM
    trust_center_references
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": t.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center reference: %w", err)
	}

	return nil
}

func (t *TrustCenterReferences) LoadByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[TrustCenterReferenceOrderField],
) error {
	q := `
SELECT
    id,
    trust_center_id,
    name,
    description,
    website_url,
    logo_file_id,
    created_at,
    updated_at
FROM
    trust_center_references
WHERE
    %s
    AND trust_center_id = @trust_center_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_id": trustCenterID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust_center_references: %w", err)
	}

	references, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrustCenterReference])
	if err != nil {
		return fmt.Errorf("cannot collect trust center references: %w", err)
	}

	*t = references

	return nil
}

func (t *TrustCenterReferences) CountByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    trust_center_references
WHERE
    %s
    AND trust_center_id = @trust_center_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_id": trustCenterID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count trust center references: %w", err)
	}

	return count, nil
}
