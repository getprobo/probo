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
	TrustCenterFile struct {
		ID                    gid.GID               `db:"id"`
		OrganizationID        gid.GID               `db:"organization_id"`
		Name                  string                `db:"name"`
		Category              string                `db:"category"`
		FileID                gid.GID               `db:"file_id"`
		TrustCenterVisibility TrustCenterVisibility `db:"trust_center_visibility"`
		CreatedAt             time.Time             `db:"created_at"`
		UpdatedAt             time.Time             `db:"updated_at"`
	}

	TrustCenterFiles []*TrustCenterFile
)

func (t TrustCenterFile) CursorKey(orderBy TrustCenterFileOrderField) page.CursorKey {
	switch orderBy {
	case TrustCenterFileOrderFieldName:
		return page.NewCursorKey(t.ID, t.Name)
	case TrustCenterFileOrderFieldCreatedAt:
		return page.NewCursorKey(t.ID, t.CreatedAt)
	case TrustCenterFileOrderFieldUpdatedAt:
		return page.NewCursorKey(t.ID, t.UpdatedAt)
	}
	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (t *TrustCenterFile) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM trust_center_files WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, t.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query trust center file authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (t *TrustCenterFile) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterFileID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    category,
    file_id,
    trust_center_visibility,
    created_at,
    updated_at
FROM
    trust_center_files
WHERE
    %s
    AND id = @trust_center_file_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"trust_center_file_id": trustCenterFileID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust_center_files: %w", err)
	}

	file, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterFile])
	if err != nil {
		return fmt.Errorf("cannot collect trust center file: %w", err)
	}

	*t = file

	return nil
}

func (f *TrustCenterFiles) LoadByIDs(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterFileIDs []gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    category,
    file_id,
    trust_center_visibility,
    created_at,
    updated_at
FROM
    trust_center_files
WHERE
    %s
    AND id = ANY(@ids);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"ids": trustCenterFileIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query file: %w", err)
	}
	defer rows.Close()

	files, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrustCenterFile])
	if err != nil {
		return fmt.Errorf("cannot collect file: %w", err)
	}

	*f = files

	return nil
}

func (t TrustCenterFile) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    trust_center_files (
        tenant_id,
        id,
        organization_id,
        name,
        category,
        file_id,
        trust_center_visibility,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @organization_id,
    @name,
    @category,
    @file_id,
    @trust_center_visibility,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":               scope.GetTenantID(),
		"id":                      t.ID,
		"organization_id":         t.OrganizationID,
		"name":                    t.Name,
		"category":                t.Category,
		"file_id":                 t.FileID,
		"trust_center_visibility": t.TrustCenterVisibility,
		"created_at":              t.CreatedAt,
		"updated_at":              t.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert trust center file: %w", err)
	}

	return nil
}

func (t *TrustCenterFile) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE trust_center_files
SET
    name = @name,
    category = @category,
    trust_center_visibility = @trust_center_visibility,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
RETURNING
    id,
    organization_id,
    name,
    category,
    file_id,
    trust_center_visibility,
    created_at,
    updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                      t.ID,
		"name":                    t.Name,
		"category":                t.Category,
		"trust_center_visibility": t.TrustCenterVisibility,
		"updated_at":              t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update trust center file: %w", err)
	}

	file, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterFile])
	if err != nil {
		return fmt.Errorf("cannot collect updated trust center file: %w", err)
	}

	*t = file

	return nil
}

func (t *TrustCenterFile) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM
    trust_center_files
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": t.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center file: %w", err)
	}

	return nil
}

func (t *TrustCenterFiles) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[TrustCenterFileOrderField],
	filter *TrustCenterFileFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    category,
    file_id,
    trust_center_visibility,
    created_at,
    updated_at
FROM
    trust_center_files
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust_center_files: %w", err)
	}

	files, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrustCenterFile])
	if err != nil {
		return fmt.Errorf("cannot collect trust center files: %w", err)
	}

	*t = files

	return nil
}

func (t *TrustCenterFiles) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    trust_center_files
WHERE
    %s
    AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count trust center files: %w", err)
	}

	return count, nil
}

func (t *TrustCenterFiles) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *TrustCenterFileFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    category,
    file_id,
    trust_center_visibility,
    created_at,
    updated_at
FROM
    trust_center_files
WHERE
    %s
    AND %s
    AND organization_id = @organization_id
ORDER BY
    created_at DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center files: %w", err)
	}

	files, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrustCenterFile])
	if err != nil {
		return fmt.Errorf("cannot collect trust center files: %w", err)
	}

	*t = files

	return nil
}
