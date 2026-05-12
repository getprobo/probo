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
	Asset struct {
		ID              gid.GID   `db:"id"`
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
func (a *Asset) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
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
	conn pg.Querier,
	scope Scoper,
	assetID gid.GID,
) error {
	q := `
SELECT
	id,
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
	AND snapshot_id IS NULL
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
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
SELECT
	id,
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
	AND snapshot_id IS NULL
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
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	assets
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
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (a *Assets) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AssetOrderField],
) error {
	q := `
SELECT
	id,
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
	AND snapshot_id IS NULL
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
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

func (a *Assets) LoadAllByOrganizationID(
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
	AND snapshot_id IS NULL
ORDER BY
	name ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

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
	conn pg.Tx,
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
	conn pg.Tx,
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
	conn pg.Tx,
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

func (a Asset) GetGeneratedDocumentID(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := conn.QueryRow(
		ctx,
		`
SELECT
	asset_list_document_id
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
		return nil, fmt.Errorf("cannot get asset list document ID: %w", err)
	}

	return documentID, nil
}

func (a Asset) UpsertGeneratedDocumentID(
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
	asset_list_document_id,
	created_at,
	updated_at
) VALUES (
	@organization_id,
	@tenant_id,
	@asset_list_document_id,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id) DO UPDATE
SET
	asset_list_document_id = @asset_list_document_id,
	updated_at = @updated_at
`,
		pgx.NamedArgs{
			"organization_id":        organizationID,
			"tenant_id":              tenantID,
			"asset_list_document_id": documentID,
			"created_at":             now,
			"updated_at":             now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot upsert asset list document ID: %w", err)
	}

	return nil
}

func (a Asset) ClearGeneratedDocumentID(
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
	asset_list_document_id = NULL,
	updated_at = @now
WHERE
	asset_list_document_id = ANY(@ids)
`,
		pgx.NamedArgs{
			"ids": ids,
			"now": time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot clear asset list document references: %w", err)
	}

	return nil
}
