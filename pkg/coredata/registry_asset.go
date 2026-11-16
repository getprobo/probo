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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

var (
	_ Searchable = (*RegistryAsset)(nil)
)

type (
	RegistryAsset struct {
		ID                gid.GID           `db:"id"`
		Name              string            `db:"name"`
		Amount            int               `db:"amount"`
		OwnerID           gid.GID           `db:"owner_profile_id"`
		OrganizationID    gid.GID           `db:"organization_id"`
		RegistryAssetType RegistryAssetType `db:"registry_asset_type"`
		DataTypesStored   string            `db:"data_types_stored"`
		CreatedAt         time.Time         `db:"created_at"`
		UpdatedAt         time.Time         `db:"updated_at"`
	}

	RegistryAssets []*RegistryAsset
)

func (a *RegistryAsset) CursorKey(field RegistryAssetOrderField) page.CursorKey {
	switch field {
	case RegistryAssetOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	case RegistryAssetOrderFieldAmount:
		return page.NewCursorKey(a.ID, a.Amount)
	case RegistryAssetOrderFieldName:
		return page.NewCursorKey(a.ID, a.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (a RegistryAsset) SearchText() []string {
	return []string{a.Name, a.DataTypesStored}
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (a *RegistryAsset) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM assets_registry WHERE id = ANY(@resource_ids::text[])`

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

func (a *RegistryAsset) LoadByID(
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
	registry_asset_type,
	data_types_stored,
	created_at,
	updated_at
FROM
	assets_registry
WHERE
	%s
	AND id = @registry_asset_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"registry_asset_id": assetID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query assets: %w", err)
	}

	asset, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RegistryAsset])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect asset: %w", err)
	}

	*a = asset

	return nil
}

func (a *RegistryAsset) LoadByOwnerID(
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
	registry_asset_type,
	data_types_stored,
	created_at,
	updated_at
FROM
	assets_registry
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

	asset, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RegistryAsset])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect asset: %w", err)
	}

	*a = asset

	return nil
}

func (a *RegistryAssets) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	assets_registry
WHERE
	%s
	AND organization_id = @organization_id
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

func (a *RegistryAssets) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[RegistryAssetOrderField],
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	registry_asset_type,
	data_types_stored,
	created_at,
	updated_at
FROM
	assets_registry
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
		return fmt.Errorf("cannot query assets: %w", err)
	}

	assets, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RegistryAsset])
	if err != nil {
		return fmt.Errorf("cannot collect assets: %w", err)
	}

	*a = assets

	return nil
}

func (a *RegistryAsset) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO assets_registry (
	id,
	tenant_id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	registry_asset_type,
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
	@registry_asset_type,
	@data_types_stored,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                  a.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     a.OrganizationID,
		"name":                a.Name,
		"owner_profile_id":    a.OwnerID,
		"amount":              a.Amount,
		"registry_asset_type": a.RegistryAssetType,
		"data_types_stored":   a.DataTypesStored,
		"created_at":          a.CreatedAt,
		"updated_at":          a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert asset: %w", err)
	}

	return nil
}

func (a *RegistryAsset) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE assets_registry
SET
	name = @name,
	owner_profile_id = @owner_profile_id,
	amount = @amount,
	registry_asset_type = @registry_asset_type,
	data_types_stored = @data_types_stored,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
RETURNING
	id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	registry_asset_type,
	data_types_stored,
	created_at,
	updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  a.ID,
		"name":                a.Name,
		"owner_profile_id":    a.OwnerID,
		"amount":              a.Amount,
		"registry_asset_type": a.RegistryAssetType,
		"data_types_stored":   a.DataTypesStored,
		"updated_at":          time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update asset: %w", err)
	}

	asset, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RegistryAsset])
	if err != nil {
		return fmt.Errorf("cannot collect updated asset: %w", err)
	}

	*a = asset

	return nil
}

func (a *RegistryAsset) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM assets_registry
WHERE
	%s
	AND id = @id
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

func (a *RegistryAssets) CountByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetIDs []gid.GID,
) (int, error) {
	if len(assetIDs) == 0 {
		return 0, nil
	}

	q := `
SELECT
	COUNT(id)
FROM
	assets_registry
WHERE
	%s
	AND id = ANY(@registry_asset_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"registry_asset_ids": assetIDs}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (a *RegistryAssets) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetIDs []gid.GID,
) error {
	if len(assetIDs) == 0 {
		*a = nil
		return nil
	}

	q := `
SELECT
	id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	registry_asset_type,
	data_types_stored,
	created_at,
	updated_at
FROM
	assets_registry
WHERE
	%s
	AND id = ANY(@registry_asset_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"registry_asset_ids": assetIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query assets: %w", err)
	}

	assets, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RegistryAsset])
	if err != nil {
		return fmt.Errorf("cannot collect assets: %w", err)
	}

	*a = assets

	if len(assets) != len(gid.NewSet(assetIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (a *RegistryAssets) LoadByIDsWithCursor(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetIDs []gid.GID,
	cursor *page.Cursor[RegistryAssetOrderField],
) error {
	if len(assetIDs) == 0 {
		*a = nil
		return nil
	}

	q := `
SELECT
	id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	registry_asset_type,
	data_types_stored,
	created_at,
	updated_at
FROM
	assets_registry
WHERE
	%s
	AND id = ANY(@registry_asset_ids)
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"registry_asset_ids": assetIDs}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query assets: %w", err)
	}

	assets, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RegistryAsset])
	if err != nil {
		return fmt.Errorf("cannot collect assets: %w", err)
	}

	*a = assets

	return nil
}

func (a RegistryAsset) GetGeneratedDocumentID(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := conn.QueryRow(
		ctx,
		`
SELECT
	registry_asset_list_document_id
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

func (a RegistryAsset) UpsertGeneratedDocumentID(
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
	registry_asset_list_document_id,
	created_at,
	updated_at
) VALUES (
	@organization_id,
	@tenant_id,
	@registry_asset_list_document_id,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id) DO UPDATE
SET
	registry_asset_list_document_id = @registry_asset_list_document_id,
	updated_at = @updated_at
`,
		pgx.NamedArgs{
			"organization_id":                 organizationID,
			"tenant_id":                       tenantID,
			"registry_asset_list_document_id": documentID,
			"created_at":                      now,
			"updated_at":                      now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot upsert asset list document ID: %w", err)
	}

	return nil
}

func (a RegistryAsset) ClearGeneratedDocumentID(
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
	registry_asset_list_document_id = NULL,
	updated_at = @now
WHERE
	registry_asset_list_document_id = ANY(@ids)
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

func (c *RegistryAssets) DeleteByOrganizationID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
DELETE FROM assets_registry
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete assets: %w", err)
	}

	return nil
}
