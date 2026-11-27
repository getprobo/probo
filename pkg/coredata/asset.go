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
	"encoding/json"
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
	_ Searchable = (*Asset)(nil)
)

type (
	Asset struct {
		ID              gid.GID         `db:"id"`
		Name            string          `db:"name"`
		Amount          int             `db:"amount"`
		OwnerID         *gid.GID        `db:"owner_profile_id"`
		OrganizationID  gid.GID         `db:"organization_id"`
		Type            AssetType       `db:"asset_type"`
		DataTypesStored *string         `db:"data_types_stored"`
		Kind            *string         `db:"kind"`
		ExternalID      *string         `db:"external_id"`
		Region          *string         `db:"region"`
		Configuration   json.RawMessage `db:"configuration"`
		SchemaVersion   *int            `db:"schema_version"`
		LastSeenAt      *time.Time      `db:"last_seen_at"`
		Status          AssetStatus     `db:"status"`
		ConnectorID     *gid.GID        `db:"connector_id"`
		CreatedAt       time.Time       `db:"created_at"`
		UpdatedAt       time.Time       `db:"updated_at"`
	}

	Assets []*Asset
)

func (a *Asset) CursorKey(field AssetOrderField) page.CursorKey {
	switch field {
	case AssetOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	case AssetOrderFieldAmount:
		return page.NewCursorKey(a.ID, a.Amount)
	case AssetOrderFieldName:
		return page.NewCursorKey(a.ID, a.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (a Asset) SearchText() []string {
	text := []string{a.Name}
	if a.DataTypesStored != nil && *a.DataTypesStored != "" {
		text = append(text, *a.DataTypesStored)
	}

	return text
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (a *Asset) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM assets WHERE id = ANY(@resource_ids::text[])`

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
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
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
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
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
	AND status = @status
	AND (
		connector_id IS NULL
		OR kind IS NULL
		OR external_id IS NULL
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"status":          AssetStatusActive,
	}
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
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
FROM
	assets
WHERE
	%s
	AND organization_id = @organization_id
	AND status = @status
	AND (
		connector_id IS NULL
		OR kind IS NULL
		OR external_id IS NULL
	)
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"status":          AssetStatusActive,
	}
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

func (a *Asset) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	if a.Status == "" {
		a.Status = AssetStatusActive
	}

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
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
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
	@kind,
	@external_id,
	@region,
	@configuration,
	@schema_version,
	@last_seen_at,
	@status,
	@connector_id,
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
		"asset_type":        a.Type,
		"data_types_stored": a.DataTypesStored,
		"kind":              a.Kind,
		"external_id":       a.ExternalID,
		"region":            a.Region,
		"configuration":     a.Configuration,
		"schema_version":    a.SchemaVersion,
		"last_seen_at":      a.LastSeenAt,
		"status":            a.Status,
		"connector_id":      a.ConnectorID,
		"created_at":        a.CreatedAt,
		"updated_at":        a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert asset: %w", err)
	}

	return nil
}

const discoveredUpsertBatchSize = 500

// UpsertDiscovered inserts or refreshes discovered rows in batches. Amount,
// owner, data types, status, and the connector stay as they are on a row
// that already exists. A new row stores amount 1. A repeated identity in
// the same batch keeps the last row, so one statement does not update the
// same asset twice.
func (assets Assets) UpsertDiscovered(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (Assets, error) {
	deduped := dedupeDiscovered(assets)
	if len(deduped) == 0 {
		return nil, nil
	}

	written := make(Assets, 0, len(deduped))

	for start := 0; start < len(deduped); start += discoveredUpsertBatchSize {
		end := min(start+discoveredUpsertBatchSize, len(deduped))

		batch, err := upsertDiscoveredBatch(ctx, conn, scope, deduped[start:end])
		if err != nil {
			return nil, err
		}

		written = append(written, batch...)
	}

	return written, nil
}

func (assets *Assets) LoadByDiscoveredIdentities(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	discovered Assets,
) error {
	deduped := dedupeDiscovered(discovered)
	if len(deduped) == 0 {
		*assets = nil
		return nil
	}

	organizationIDs := make([]string, len(deduped))
	connectorIDs := make([]string, len(deduped))
	kinds := make([]string, len(deduped))
	externalIDs := make([]string, len(deduped))

	for i, asset := range deduped {
		organizationIDs[i] = asset.OrganizationID.String()
		connectorIDs[i] = asset.ConnectorID.String()
		kinds[i] = *asset.Kind
		externalIDs[i] = *asset.ExternalID
	}

	q := `
SELECT
	id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	asset_type,
	data_types_stored,
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
FROM
	assets
WHERE
	%s
	AND (organization_id, connector_id, kind, external_id) IN (
		SELECT
			organization_id,
			connector_id,
			kind,
			external_id
		FROM unnest(
			@organization_ids::text[],
			@connector_ids::text[],
			@kinds::text[],
			@external_ids::text[]
		) AS identity(organization_id, connector_id, kind, external_id)
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_ids": organizationIDs,
		"connector_ids":    connectorIDs,
		"kinds":            kinds,
		"external_ids":     externalIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query discovered assets: %w", err)
	}

	loaded, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Asset])
	if err != nil {
		return fmt.Errorf("cannot collect discovered assets: %w", err)
	}

	*assets = loaded

	return nil
}

func discoveredIdentityKey(asset *Asset) (string, bool) {
	if asset == nil || asset.Kind == nil || asset.ExternalID == nil || asset.ConnectorID == nil {
		return "", false
	}

	return asset.OrganizationID.String() + "\x00" + asset.ConnectorID.String() + "\x00" + *asset.Kind + "\x00" + *asset.ExternalID, true
}

func dedupeDiscovered(assets Assets) Assets {
	indexByIdentity := make(map[string]int, len(assets))
	deduped := make(Assets, 0, len(assets))

	for _, asset := range assets {
		key, ok := discoveredIdentityKey(asset)
		if !ok {
			continue
		}

		if index, exists := indexByIdentity[key]; exists {
			deduped[index] = asset
			continue
		}

		indexByIdentity[key] = len(deduped)
		deduped = append(deduped, asset)
	}

	return deduped
}

func upsertDiscoveredBatch(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	assets Assets,
) (Assets, error) {
	ids := make([]string, len(assets))
	names := make([]string, len(assets))
	organizationIDs := make([]string, len(assets))
	assetTypes := make([]string, len(assets))
	kinds := make([]string, len(assets))
	externalIDs := make([]string, len(assets))
	regions := make([]*string, len(assets))
	configurations := make([]string, len(assets))
	schemaVersions := make([]int, len(assets))
	lastSeenAts := make([]time.Time, len(assets))
	statuses := make([]string, len(assets))
	connectorIDs := make([]string, len(assets))
	createdAts := make([]time.Time, len(assets))
	updatedAts := make([]time.Time, len(assets))

	for i, asset := range assets {
		configuration := string(asset.Configuration)
		if configuration == "" {
			configuration = "{}"
		}

		schemaVersion := 0
		if asset.SchemaVersion != nil {
			schemaVersion = *asset.SchemaVersion
		}

		lastSeenAt := asset.UpdatedAt
		if asset.LastSeenAt != nil {
			lastSeenAt = *asset.LastSeenAt
		}

		status := asset.Status
		if status == "" {
			status = AssetStatusActive
		}

		ids[i] = asset.ID.String()
		names[i] = asset.Name
		organizationIDs[i] = asset.OrganizationID.String()
		assetTypes[i] = string(asset.Type)
		kinds[i] = *asset.Kind
		externalIDs[i] = *asset.ExternalID
		regions[i] = asset.Region
		configurations[i] = configuration
		schemaVersions[i] = schemaVersion
		lastSeenAts[i] = lastSeenAt
		statuses[i] = string(status)
		connectorIDs[i] = asset.ConnectorID.String()
		createdAts[i] = asset.CreatedAt
		updatedAts[i] = asset.UpdatedAt
	}

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
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
)
SELECT
	id,
	@tenant_id,
	name,
	organization_id,
	NULL,
	1,
	asset_type::asset_type,
	NULL,
	kind,
	external_id,
	region,
	configuration::jsonb,
	schema_version,
	last_seen_at,
	status::asset_status,
	connector_id,
	created_at,
	updated_at
FROM unnest(
	@ids::text[],
	@names::text[],
	@organization_ids::text[],
	@asset_types::text[],
	@kinds::text[],
	@external_ids::text[],
	@regions::text[],
	@configurations::text[],
	@schema_versions::int[],
	@last_seen_ats::timestamptz[],
	@statuses::text[],
	@connector_ids::text[],
	@created_ats::timestamptz[],
	@updated_ats::timestamptz[]
) AS discovered(
	id,
	name,
	organization_id,
	asset_type,
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
)
ON CONFLICT (organization_id, connector_id, kind, external_id)
WHERE connector_id IS NOT NULL
	AND kind IS NOT NULL
	AND external_id IS NOT NULL
DO UPDATE
SET
	name = EXCLUDED.name,
	region = EXCLUDED.region,
	configuration = EXCLUDED.configuration,
	schema_version = EXCLUDED.schema_version,
	last_seen_at = EXCLUDED.last_seen_at,
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	name,
	organization_id,
	owner_profile_id,
	amount,
	asset_type,
	data_types_stored,
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
`

	args := pgx.StrictNamedArgs{
		"tenant_id":        scope.GetTenantID(),
		"ids":              ids,
		"names":            names,
		"organization_ids": organizationIDs,
		"asset_types":      assetTypes,
		"kinds":            kinds,
		"external_ids":     externalIDs,
		"regions":          regions,
		"configurations":   configurations,
		"schema_versions":  schemaVersions,
		"last_seen_ats":    lastSeenAts,
		"statuses":         statuses,
		"connector_ids":    connectorIDs,
		"created_ats":      createdAts,
		"updated_ats":      updatedAts,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot upsert assets: %w", err)
	}

	written, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Asset])
	if err != nil {
		return nil, fmt.Errorf("cannot collect upserted assets: %w", err)
	}

	return written, nil
}

func (a *Asset) LoadByDiscoveredIdentity(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
	kind string,
	externalID string,
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
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
FROM
	assets
WHERE
	%s
	AND connector_id = @connector_id
	AND kind = @kind
	AND external_id = @external_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": connectorID,
		"kind":         kind,
		"external_id":  externalID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query asset by discovered identity: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Asset])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect asset by discovered identity: %w", err)
	}

	*a = row

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
	status = @status,
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
	asset_type,
	data_types_stored,
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                a.ID,
		"name":              a.Name,
		"owner_profile_id":  a.OwnerID,
		"amount":            a.Amount,
		"asset_type":        a.Type,
		"data_types_stored": a.DataTypesStored,
		"status":            a.Status,
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

func (a *Asset) Deactivate(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	updatedAt time.Time,
) error {
	q := `
UPDATE assets
SET
	status = @status,
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
	asset_type,
	data_types_stored,
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         a.ID,
		"status":     AssetStatusDeactivated,
		"updated_at": updatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot deactivate asset: %w", err)
	}

	asset, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Asset])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect deactivated asset: %w", err)
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

func (a *Assets) CountByIDs(
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
	assets
WHERE
	%s
	AND id = ANY(@asset_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"asset_ids": assetIDs}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (a *Assets) LoadByIDs(
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
	asset_type,
	data_types_stored,
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
FROM
	assets
WHERE
	%s
	AND id = ANY(@asset_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"asset_ids": assetIDs}
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

	if len(assets) != len(gid.NewSet(assetIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (a *Assets) LoadByIDsWithCursor(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	assetIDs []gid.GID,
	cursor *page.Cursor[AssetOrderField],
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
	asset_type,
	data_types_stored,
	kind,
	external_id,
	region,
	configuration,
	schema_version,
	last_seen_at,
	status,
	connector_id,
	created_at,
	updated_at
FROM
	assets
WHERE
	%s
	AND id = ANY(@asset_ids)
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"asset_ids": assetIDs}
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

func (c *Assets) DeleteByOrganizationID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
DELETE FROM assets
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
