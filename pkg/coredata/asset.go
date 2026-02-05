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
	Asset struct {
		ID              gid.GID   `db:"id"`
		SnapshotID      *gid.GID  `db:"snapshot_id"`
		SourceID        *gid.GID  `db:"source_id"`
		Name            string    `db:"name"`
		Amount          int       `db:"amount"`
		OwnerID         gid.GID   `db:"owner_profile_id"`
		OrganizationID  gid.GID   `db:"organization_id"`
		AssetType       AssetType `db:"asset_type"`
		DataTypesStored string    `db:"data_types_stored"`
		CreatedAt       time.Time `db:"created_at"`
		UpdatedAt       time.Time `db:"updated_at"`
	}

	Assets []*Asset
)

func (a *Asset) CursorKey(field AssetOrderField) page.CursorKey {
	switch field {
	case AssetOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	case AssetOrderFieldAmount:
		return page.NewCursorKey(a.ID, a.Amount)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (a *Asset) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM assets WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, a.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query asset authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (a *Asset) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	assetID gid.GID,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	asset_type,
	data_types_stored,
	created_at,
	updated_at
FROM
	assets
WHERE
	%s
	AND id = @asset_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"asset_id": assetID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query assets: %w", err)
	}

	asset, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Asset])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect asset: %w", err)
	}

	*a = asset

	return nil
}

func (a *Asset) LoadByOwnerID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	asset_type,
	data_types_stored,
	created_at,
	updated_at
FROM
	assets
WHERE
	%s
	AND owner_profile_id = @owner_profile_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"owner_profile_id": a.OwnerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query assets: %w", err)
	}

	asset, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Asset])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect asset: %w", err)
	}

	*a = asset

	return nil
}

func (a *Assets) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *AssetFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	assets
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (a *Assets) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AssetOrderField],
	filter *AssetFilter,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	asset_type,
	data_types_stored,
	created_at,
	updated_at
FROM
	assets
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
		return fmt.Errorf("cannot query assets: %w", err)
	}

	assets, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Asset])
	if err != nil {
		return fmt.Errorf("cannot collect assets: %w", err)
	}

	*a = assets

	return nil
}

func (a *Asset) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO assets (
	id,
	tenant_id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	asset_type,
	data_types_stored,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@name,
	@organization_id,
	@owner_profile_id,
	@amount,
	@asset_type,
	@data_types_stored,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                a.ID,
		"tenant_id":         scope.GetTenantID(),
		"organization_id":   a.OrganizationID,
		"name":              a.Name,
		"owner_profile_id":  a.OwnerID,
		"amount":            a.Amount,
		"asset_type":        a.AssetType,
		"data_types_stored": a.DataTypesStored,
		"created_at":        a.CreatedAt,
		"updated_at":        a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert asset: %w", err)
	}

	return nil
}

func (a *Asset) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE assets
SET
	name = @name,
	owner_profile_id = @owner_profile_id,
	amount = @amount,
	asset_type = @asset_type,
	data_types_stored = @data_types_stored,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
RETURNING
	id,
	snapshot_id,
	source_id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	asset_type,
	data_types_stored,
	created_at,
	updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                a.ID,
		"name":              a.Name,
		"owner_profile_id":  a.OwnerID,
		"amount":            a.Amount,
		"asset_type":        a.AssetType,
		"data_types_stored": a.DataTypesStored,
		"updated_at":        time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update asset: %w", err)
	}

	asset, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Asset])
	if err != nil {
		return fmt.Errorf("cannot collect updated asset: %w", err)
	}

	*a = asset

	return nil
}

func (a *Asset) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM assets
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": a.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete asset: %w", err)
	}

	return nil
}

func (assets Assets) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	snapshotters := []AssetSnapshotter{Assets{}, Vendors{}, AssetVendors{}}

	for _, snapshotter := range snapshotters {
		if err := snapshotter.InsertAssetSnapshots(ctx, conn, scope, organizationID, snapshotID); err != nil {
			return fmt.Errorf("cannot create asset snapshots: (%T) %w", snapshotter, err)
		}
	}

	return nil
}

func (assets Assets) InsertAssetSnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	source_assets AS (
		SELECT *
		FROM assets
		WHERE %s AND organization_id = @organization_id AND snapshot_id IS NULL
	)
INSERT INTO assets (
	tenant_id,
	id,
	snapshot_id,
	source_id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	asset_type,
	data_types_stored,
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @asset_entity_type),
	@snapshot_id,
	a.id,
	a.name,
	a.organization_id,
	a.owner_profile_id,
	a.amount,
	a.asset_type,
	a.data_types_stored,
	a.created_at,
	a.updated_at
FROM source_assets a
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":         scope.GetTenantID(),
		"snapshot_id":       snapshotID,
		"organization_id":   organizationID,
		"asset_entity_type": AssetEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert asset snapshots: %w", err)
	}

	return nil
}
