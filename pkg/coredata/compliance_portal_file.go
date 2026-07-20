// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	CompliancePortalFile struct {
		ID                         gid.GID                    `db:"id"`
		OrganizationID             gid.GID                    `db:"organization_id"`
		Name                       string                     `db:"name"`
		Category                   string                     `db:"category"`
		FileID                     gid.GID                    `db:"file_id"`
		CompliancePortalVisibility CompliancePortalVisibility `db:"trust_center_visibility"`
		CreatedAt                  time.Time                  `db:"created_at"`
		UpdatedAt                  time.Time                  `db:"updated_at"`
	}

	CompliancePortalFiles []*CompliancePortalFile
)

func (t CompliancePortalFile) CursorKey(orderBy CompliancePortalFileOrderField) page.CursorKey {
	switch orderBy {
	case CompliancePortalFileOrderFieldName:
		return page.NewCursorKey(t.ID, t.Name)
	case CompliancePortalFileOrderFieldCreatedAt:
		return page.NewCursorKey(t.ID, t.CreatedAt)
	case CompliancePortalFileOrderFieldUpdatedAt:
		return page.NewCursorKey(t.ID, t.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (t *CompliancePortalFile) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM trust_center_files WHERE id = ANY(@resource_ids::text[])`

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

func (t *CompliancePortalFile) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalFileID gid.GID,
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

	args := pgx.StrictNamedArgs{"trust_center_file_id": compliancePortalFileID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust_center_files: %w", err)
	}

	file, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalFile])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal file: %w", err)
	}

	*t = file

	return nil
}

func (f *CompliancePortalFiles) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	compliancePortalFileIDs []gid.GID,
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

	args := pgx.StrictNamedArgs{"ids": compliancePortalFileIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query file: %w", err)
	}
	defer rows.Close()

	files, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalFile])
	if err != nil {
		return fmt.Errorf("cannot collect file: %w", err)
	}

	*f = files

	if len(files) != len(gid.NewSet(compliancePortalFileIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (t CompliancePortalFile) Insert(
	ctx context.Context,
	conn pg.Tx,
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
		"trust_center_visibility": t.CompliancePortalVisibility,
		"created_at":              t.CreatedAt,
		"updated_at":              t.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert compliance portal file: %w", err)
	}

	return nil
}

func (t *CompliancePortalFile) Update(
	ctx context.Context,
	conn pg.Tx,
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
		"trust_center_visibility": t.CompliancePortalVisibility,
		"updated_at":              t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance portal file: %w", err)
	}

	file, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CompliancePortalFile])
	if err != nil {
		return fmt.Errorf("cannot collect updated compliance portal file: %w", err)
	}

	*t = file

	return nil
}

func (t *CompliancePortalFile) Delete(
	ctx context.Context,
	conn pg.Tx,
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
		return fmt.Errorf("cannot delete compliance portal file: %w", err)
	}

	return nil
}

func (t *CompliancePortalFiles) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[CompliancePortalFileOrderField],
	filter *CompliancePortalFileFilter,
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

	files, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CompliancePortalFile])
	if err != nil {
		return fmt.Errorf("cannot collect compliance portal files: %w", err)
	}

	*t = files

	return nil
}

func (t *CompliancePortalFiles) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
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
		return 0, fmt.Errorf("cannot count compliance portal files: %w", err)
	}

	return count, nil
}
