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

type Data struct {
	ID              gid.GID         `db:"id"`
	Name            string          `db:"name"`
	OrganizationID  gid.GID         `db:"organization_id"`
	OwnerID         gid.GID         `db:"owner_id"`
	DataSensitivity DataSensitivity `db:"data_sensitivity"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
}

func (d *Data) CursorKey(field DatumOrderField) page.CursorKey {
	switch field {
	case DatumOrderFieldCreatedAt:
		return page.NewCursorKey(d.ID, d.CreatedAt)
	case DatumOrderFieldName:
		return page.NewCursorKey(d.ID, d.Name)
	case DatumOrderFieldDataSensitivity:
		return page.NewCursorKey(d.ID, d.DataSensitivity)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

type DataList []*Data

func (d *Data) LoadByID(
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
	data_sensitivity,
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

	data, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Data])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*d = data

	return nil
}

func (d *Data) LoadByOwnerID(
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
	data_sensitivity,
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

	data, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Data])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*d = data

	return nil
}

func (dl *DataList) LoadByOwnerID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	ownerID gid.GID,
	cursor *page.Cursor[DatumOrderField],
) error {
	q := `
SELECT
	id,
	name,
	owner_id,
	data_sensitivity,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND owner_id = @owner_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"owner_id": ownerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data: %w", err)
	}

	data, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Data])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*dl = data

	return nil
}

func (dl *DataList) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DatumOrderField],
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	owner_id,
	data_sensitivity,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND organization_id = @organization_id
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

	data, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Data])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*dl = data

	return nil
}

func (d *Data) Insert(
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
	data_sensitivity,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@name,
	@owner_id,
	@organization_id,
	@data_sensitivity,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               d.ID,
		"tenant_id":        scope.GetTenantID(),
		"name":             d.Name,
		"owner_id":         d.OwnerID,
		"organization_id":  d.OrganizationID,
		"data_sensitivity": d.DataSensitivity,
		"created_at":       d.CreatedAt,
		"updated_at":       d.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert data: %w", err)
	}

	return nil
}

func (d *Data) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE data
SET
	name = @name,
	owner_id = @owner_id,
	data_sensitivity = @data_sensitivity,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
RETURNING
	id,
	name,
	owner_id,
	organization_id,
	data_sensitivity,
	created_at,
	updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":               d.ID,
		"name":             d.Name,
		"owner_id":         d.OwnerID,
		"data_sensitivity": d.DataSensitivity,
		"updated_at":       time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update data: %w", err)
	}

	data, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Data])
	if err != nil {
		return fmt.Errorf("cannot collect updated data: %w", err)
	}

	*d = data

	return nil
}

func (d *Data) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM data
WHERE
	%s
	AND id = @id
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

type DataVendor struct {
	DataID    gid.GID   `db:"data_id"`
	VendorID  gid.GID   `db:"vendor_id"`
	CreatedAt time.Time `db:"created_at"`
}

func (d *Data) CreateWithVendors(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	vendorIDs []gid.GID,
	now time.Time,
) error {
	if err := d.Insert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot insert data: %w", err)
	}

	if len(vendorIDs) > 0 {
		for _, vendorID := range vendorIDs {
			_, err := conn.Exec(ctx, `
				INSERT INTO data_vendors (tenant_id, datum_id, vendor_id, created_at)
				VALUES ($1, $2, $3, $4)
			`, scope.GetTenantID(), d.ID, vendorID, now)
			if err != nil {
				return fmt.Errorf("cannot insert data vendor: %w", err)
			}
		}
	}

	return nil
}

func (d *Data) UpdateWithVendors(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	vendorIDs []gid.GID,
	now time.Time,
) error {
	existing := &Data{}
	if err := existing.LoadByID(ctx, conn, scope, d.ID); err != nil {
		return fmt.Errorf("cannot load data: %w", err)
	}

	d.CreatedAt = existing.CreatedAt
	d.UpdatedAt = now

	if err := d.Update(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot update data: %w", err)
	}

	_, err := conn.Exec(ctx, `
		DELETE FROM data_vendors
		WHERE tenant_id = $1 AND datum_id = $2
	`, scope.GetTenantID(), d.ID)
	if err != nil {
		return fmt.Errorf("cannot delete data vendors: %w", err)
	}

	if len(vendorIDs) > 0 {
		for _, vendorID := range vendorIDs {
			_, err := conn.Exec(ctx, `
				INSERT INTO data_vendors (tenant_id, datum_id, vendor_id, created_at)
				VALUES ($1, $2, $3, $4)
			`, scope.GetTenantID(), d.ID, vendorID, now)
			if err != nil {
				return fmt.Errorf("cannot insert data vendor: %w", err)
			}
		}
	}

	return nil
}

// UpdateWithVendorsTx updates a data entry and its vendor relationships in a single transaction
func (d *Data) UpdateWithVendorsTx(
	ctx context.Context,
	db *pg.Client,
	scope Scoper,
	vendorIDs []gid.GID,
	now time.Time,
) error {
	return db.WithTx(ctx, func(conn pg.Conn) error {
		return d.UpdateWithVendors(ctx, conn, scope, vendorIDs, now)
	})
}
