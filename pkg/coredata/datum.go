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
	Datum struct {
		ID                 gid.GID            `db:"id"`
		Name               string             `db:"name"`
		OrganizationID     gid.GID            `db:"organization_id"`
		OwnerID            gid.GID            `db:"owner_profile_id"`
		DataClassification DataClassification `db:"data_classification"`
		CreatedAt          time.Time          `db:"created_at"`
		UpdatedAt          time.Time          `db:"updated_at"`
	}

	Data []*Datum
)

func (d *Datum) CursorKey(field DatumOrderField) page.CursorKey {
	switch field {
	case DatumOrderFieldCreatedAt:
		return page.NewCursorKey(d.ID, d.CreatedAt)
	case DatumOrderFieldName:
		return page.NewCursorKey(d.ID, d.Name)
	case DatumOrderFieldDataClassification:
		return page.NewCursorKey(d.ID, d.DataClassification)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (d *Datum) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM data WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, d.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query datum authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (d *Datum) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	dataID gid.GID,
) error {
	q := `
SELECT
	id,
	name,
	owner_profile_id,
	organization_id,
	data_classification,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND id = @data_id
	AND snapshot_id IS NULL
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"data_id": dataID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data: %w", err)
	}

	datum, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Datum])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*d = datum

	return nil
}

func (d *Datum) LoadByOwnerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	name,
	owner_profile_id,
	organization_id,
	data_classification,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND owner_profile_id = @owner_profile_id
	AND snapshot_id IS NULL
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"owner_profile_id": d.OwnerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data: %w", err)
	}

	data, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Datum])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*d = data

	return nil
}

func (d *Data) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	data
WHERE
	%s
	AND organization_id = @organization_id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count data: %w", err)
	}

	return count, nil
}

func (d *Data) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DatumOrderField],
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	owner_profile_id,
	data_classification,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND organization_id = @organization_id
	AND snapshot_id IS NULL
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data: %w", err)
	}

	data, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Datum])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*d = data

	return nil
}

func (d *Data) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	owner_profile_id,
	data_classification,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND organization_id = @organization_id
	AND snapshot_id IS NULL
ORDER BY
	name ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data: %w", err)
	}

	data, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Datum])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*d = data

	return nil
}

func (d *Datum) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO data (
	id,
	tenant_id,
	name,
	owner_profile_id,
	organization_id,
	data_classification,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@name,
	@owner_profile_id,
	@organization_id,
	@data_classification,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                  d.ID,
		"tenant_id":           scope.GetTenantID(),
		"name":                d.Name,
		"owner_profile_id":    d.OwnerID,
		"organization_id":     d.OrganizationID,
		"data_classification": d.DataClassification,
		"created_at":          d.CreatedAt,
		"updated_at":          d.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert data: %w", err)
	}

	return nil
}

func (d *Datum) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE data
SET
	name = @name,
	owner_profile_id = @owner_profile_id,
	data_classification = @data_classification,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
RETURNING
	id,
	name,
	owner_profile_id,
	organization_id,
	data_classification,
	created_at,
	updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  d.ID,
		"name":                d.Name,
		"owner_profile_id":    d.OwnerID,
		"data_classification": d.DataClassification,
		"updated_at":          d.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update data: %w", err)
	}

	datum, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Datum])
	if err != nil {
		return fmt.Errorf("cannot collect updated data: %w", err)
	}

	*d = datum

	return nil
}

func (d *Datum) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM data
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": d.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete data: %w", err)
	}

	return nil
}

func (d Datum) GetGeneratedDocumentID(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := conn.QueryRow(
		ctx,
		`
SELECT
	data_document_id
FROM
	generated_documents
WHERE
	organization_id = @organization_id
`,
		pgx.NamedArgs{"organization_id": organizationID},
	).Scan(&documentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot get data document ID: %w", err)
	}

	return documentID, nil
}

func (d Datum) UpsertGeneratedDocumentID(
	ctx context.Context,
	conn pg.Tx,
	organizationID gid.GID,
	tenantID gid.TenantID,
	documentID gid.GID,
) error {
	now := time.Now()

	_, err := conn.Exec(
		ctx,
		`
INSERT INTO generated_documents (
	organization_id,
	tenant_id,
	data_document_id,
	created_at,
	updated_at
) VALUES (
	@organization_id,
	@tenant_id,
	@data_document_id,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id) DO UPDATE
SET
	data_document_id = @data_document_id,
	updated_at = @updated_at
`,
		pgx.NamedArgs{
			"organization_id":  organizationID,
			"tenant_id":        tenantID,
			"data_document_id": documentID,
			"created_at":       now,
			"updated_at":       now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot upsert data document ID: %w", err)
	}

	return nil
}

func (d Datum) ClearGeneratedDocumentID(
	ctx context.Context,
	conn pg.Tx,
	documentIDs []gid.GID,
) error {
	ids := make([]string, len(documentIDs))
	for i, id := range documentIDs {
		ids[i] = id.String()
	}

	_, err := conn.Exec(
		ctx,
		`
UPDATE
	generated_documents
SET
	data_document_id = NULL,
	updated_at = @now
WHERE
	data_document_id = ANY(@ids)
`,
		pgx.NamedArgs{
			"ids": ids,
			"now": time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot clear data document references: %w", err)
	}

	return nil
}
