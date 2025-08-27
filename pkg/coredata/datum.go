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
	Datum struct {
		ID                 gid.GID            `db:"id"`
		Name               string             `db:"name"`
		OrganizationID     gid.GID            `db:"organization_id"`
		OwnerID            gid.GID            `db:"owner_id"`
		DataClassification DataClassification `db:"data_classification"`
		SnapshotID         *gid.GID           `db:"snapshot_id"`
		SourceID           *gid.GID           `db:"source_id"`
		CreatedAt          time.Time          `db:"created_at"`
		UpdatedAt          time.Time          `db:"updated_at"`
	}

	Data []*Datum

	DataSnapshotter interface {
		InsertDataSnapshots(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error
	}
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

func (d *Datum) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	dataID gid.GID,
) error {
	q := `
SELECT
	id,
	name,
	owner_id,
	organization_id,
	data_classification,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND id = @data_id
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
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	name,
	owner_id,
	organization_id,
	data_classification,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND owner_id = @owner_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"owner_id": d.OwnerID}
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
	conn pg.Conn,
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
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DatumOrderField],
	filter *DatumFilter,
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	owner_id,
	data_classification,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	data
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
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO data (
	id,
	tenant_id,
	name,
	owner_id,
	organization_id,
	data_classification,
	snapshot_id,
	source_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@name,
	@owner_id,
	@organization_id,
	@data_classification,
	@snapshot_id,
	@source_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                  d.ID,
		"tenant_id":           scope.GetTenantID(),
		"name":                d.Name,
		"owner_id":            d.OwnerID,
		"organization_id":     d.OrganizationID,
		"data_classification": d.DataClassification,
		"snapshot_id":         d.SnapshotID,
		"source_id":           d.SourceID,
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
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE data
SET
	name = @name,
	owner_id = @owner_id,
	data_classification = @data_classification,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
RETURNING
	id,
	name,
	owner_id,
	organization_id,
	data_classification,
	snapshot_id,
	source_id,
	created_at,
	updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  d.ID,
		"name":                d.Name,
		"owner_id":            d.OwnerID,
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
	conn pg.Conn,
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

func (d Data) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	snapshotters := []DataSnapshotter{Data{}, Vendors{}, DatumVendors{}}

	for _, snapshotter := range snapshotters {
		if err := snapshotter.InsertDataSnapshots(ctx, conn, scope, organizationID, snapshotID); err != nil {
			return fmt.Errorf("cannot create data snapshots: (%T) %w", snapshotter, err)
		}
	}

	return nil
}

func (d Data) InsertDataSnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	source_data AS (
		SELECT *
		FROM data
		WHERE %s AND organization_id = @organization_id AND snapshot_id IS NULL
	)
INSERT INTO data (
	tenant_id,
	id,
	snapshot_id,
	source_id,
	name,
	organization_id,
	owner_id,
	data_classification,
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @datum_entity_type),
	@snapshot_id,
	d.id,
	d.name,
	d.organization_id,
	d.owner_id,
	d.data_classification,
	d.created_at,
	d.updated_at
FROM source_data d
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":         scope.GetTenantID(),
		"snapshot_id":       snapshotID,
		"organization_id":   organizationID,
		"datum_entity_type": DatumEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert data snapshots: %w", err)
	}

	return nil
}
