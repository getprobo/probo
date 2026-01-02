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
	VendorService struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		VendorID       gid.GID   `db:"vendor_id"`
		Name           string    `db:"name"`
		Description    *string   `db:"description"`
		SnapshotID     *gid.GID  `db:"snapshot_id"`
		SourceID       *gid.GID  `db:"source_id"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	VendorServices []*VendorService
)

func (vs VendorService) CursorKey(orderBy VendorServiceOrderField) page.CursorKey {
	switch orderBy {
	case VendorServiceOrderFieldCreatedAt:
		return page.CursorKey{ID: vs.ID, Value: vs.CreatedAt}
	case VendorServiceOrderFieldName:
		return page.CursorKey{ID: vs.ID, Value: vs.Name}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (vs *VendorService) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM vendor_services WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, vs.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query vendor service authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (vs *VendorService) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	vendorServiceID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	vendor_id,
	name,
	description,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	vendor_services
WHERE
	%s
	AND id = @vendor_service_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"vendor_service_id": vendorServiceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query vendor service: %w", err)
	}
	defer rows.Close()

	vendorService, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[VendorService])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect vendor service: %w", err)
	}

	*vs = vendorService

	return nil
}

func (vs *VendorServices) LoadByVendorID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	vendorID gid.GID,
	cursor *page.Cursor[VendorServiceOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	vendor_id,
	name,
	description,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	vendor_services
WHERE
	%s
	AND vendor_id = @vendor_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"vendor_id": vendorID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query vendor services: %w", err)
	}
	defer rows.Close()

	vendorServices, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[VendorService])
	if err != nil {
		return fmt.Errorf("cannot collect vendor services: %w", err)
	}

	*vs = vendorServices

	return nil
}

func (vs VendorService) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
	vendor_services (
		tenant_id,
		id,
		organization_id,
		vendor_id,
		name,
		description,
		created_at,
		updated_at
	)
VALUES (
	@tenant_id,
	@vendor_service_id,
	@organization_id,
	@vendor_id,
	@name,
	@description,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":         scope.GetTenantID(),
		"vendor_service_id": vs.ID,
		"organization_id":   vs.OrganizationID,
		"vendor_id":         vs.VendorID,
		"name":              vs.Name,
		"description":       vs.Description,
		"created_at":        vs.CreatedAt,
		"updated_at":        vs.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert vendor service: %w", err)
	}

	return nil
}

func (vs VendorService) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE
	vendor_services
SET
	name = @name,
	description = @description,
	updated_at = @updated_at
WHERE
	%s
	AND id = @vendor_service_id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"vendor_service_id": vs.ID,
		"name":              vs.Name,
		"description":       vs.Description,
		"updated_at":        vs.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update vendor service: %w", err)
	}

	return nil
}

func (vs VendorService) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM
	vendor_services
WHERE
	%s
	AND id = @vendor_service_id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"vendor_service_id": vs.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete vendor service: %w", err)
	}

	return nil
}

func (vs VendorServices) InsertVendorSnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	snapshot_vendors AS (
		SELECT id, source_id
		FROM vendors
		WHERE organization_id = @organization_id AND snapshot_id = @snapshot_id
	)
INSERT INTO vendor_services (
	tenant_id,
	id,
	organization_id,
	snapshot_id,
	source_id,
	vendor_id,
	name,
	description,
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @vendor_service_entity_type),
	@organization_id,
	@snapshot_id,
	vs.id,
	sv.id,
	vs.name,
	vs.description,
	vs.created_at,
	vs.updated_at
FROM vendor_services vs
INNER JOIN snapshot_vendors sv ON sv.source_id = vs.vendor_id
WHERE %s AND vs.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":                  scope.GetTenantID(),
		"snapshot_id":                snapshotID,
		"organization_id":            organizationID,
		"vendor_service_entity_type": VendorServiceEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert vendor service snapshots: %w", err)
	}

	return nil
}
