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
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	DatumVendor struct {
		DatumID    gid.GID      `db:"datum_id"`
		VendorID   gid.GID      `db:"vendor_id"`
		SnapshotID *gid.GID     `db:"snapshot_id"`
		TenantID   gid.TenantID `db:"tenant_id"`
		CreatedAt  time.Time    `db:"created_at"`
	}

	DatumVendors []*DatumVendor
)

func (dv DatumVendors) Merge(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	datumID gid.GID,
	vendorIDs []gid.GID,
) error {
	q := `
WITH vendor_ids AS (
	SELECT
		unnest(@vendor_ids::text[]) AS vendor_id,
		@tenant_id AS tenant_id,
		@datum_id AS datum_id,
		@created_at::timestamptz AS created_at
)
MERGE INTO data_vendors AS tgt
USING vendor_ids AS src
ON tgt.tenant_id = src.tenant_id
	AND tgt.datum_id = src.datum_id
	AND tgt.vendor_id = src.vendor_id
WHEN NOT MATCHED THEN
	INSERT (tenant_id, datum_id, vendor_id, created_at)
	VALUES (src.tenant_id, src.datum_id, src.vendor_id, src.created_at)
WHEN NOT MATCHED BY SOURCE
	AND tgt.tenant_id = @tenant_id AND tgt.datum_id = @datum_id
	THEN DELETE
	`

	args := pgx.StrictNamedArgs{
		"tenant_id":  scope.GetTenantID(),
		"datum_id":   datumID,
		"created_at": time.Now(),
		"vendor_ids": vendorIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot merge data vendors: %w", err)
	}

	return nil
}

func (dv DatumVendors) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	datumID gid.GID,
	vendorIDs []gid.GID,
) error {
	q := `
WITH vendor_ids AS (
	SELECT unnest(@vendor_ids::text[]) AS vendor_id
)
INSERT INTO data_vendors (tenant_id, datum_id, vendor_id, created_at)
SELECT
	@tenant_id::text AS tenant_id,
	@datum_id::text AS datum_id,
	vendor_id,
	@created_at::timestamptz AS created_at
FROM vendor_ids
`

	args := pgx.StrictNamedArgs{
		"tenant_id":  scope.GetTenantID(),
		"datum_id":   datumID,
		"created_at": time.Now(),
		"vendor_ids": vendorIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert data vendors: %w", err)
	}

	return nil
}

func (dv *DatumVendors) LoadByDatumIDs(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	datumIDs []gid.GID,
) error {
	if len(datumIDs) == 0 {
		return nil
	}

	q := `
SELECT
	datum_id,
	vendor_id,
	tenant_id,
	snapshot_id,
	created_at
FROM
	data_vendors
WHERE
	tenant_id = @tenant_id
	AND datum_id = ANY(@datum_ids::text[])
ORDER BY datum_id, vendor_id
`

	args := pgx.StrictNamedArgs{
		"tenant_id": scope.GetTenantID(),
		"datum_ids": datumIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query datum vendors: %w", err)
	}

	datumVendors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DatumVendor])
	if err != nil {
		return fmt.Errorf("cannot collect datum vendors: %w", err)
	}

	*dv = datumVendors

	return nil
}

func (dv DatumVendors) BulkInsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	columnNames := []string{
		"tenant_id",
		"datum_id",
		"vendor_id",
		"snapshot_id",
		"created_at",
	}

	copyFromSource := &datumVendorCopy{
		datumVendors: dv,
		scope:        scope,
		position:     0,
	}

	_, err := conn.CopyFrom(ctx, pgx.Identifier{"data_vendors"}, columnNames, copyFromSource)
	if err != nil {
		return fmt.Errorf("cannot bulk insert datum vendors: %w", err)
	}

	return nil
}
