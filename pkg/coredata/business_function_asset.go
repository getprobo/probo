// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
)

type (
	BusinessFunctionAsset struct {
		BusinessFunctionID gid.GID   `db:"business_function_id"`
		AssetID            gid.GID   `db:"asset_id"`
		OrganizationID     gid.GID   `db:"organization_id"`
		CreatedAt          time.Time `db:"created_at"`
	}

	BusinessFunctionAssets []*BusinessFunctionAsset
)

func (bfa BusinessFunctionAssets) Merge(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	businessFunctionID gid.GID,
	organizationID gid.GID,
	assetIDs []gid.GID,
) error {
	q := `
WITH asset_ids AS (
	SELECT DISTINCT
		unnest(@asset_ids::text[]) AS asset_id,
		@tenant_id AS tenant_id,
		@business_function_id AS business_function_id,
		@organization_id AS organization_id,
		@created_at::timestamptz AS created_at
)
MERGE INTO business_function_assets AS tgt
USING asset_ids AS src
ON tgt.tenant_id = src.tenant_id
	AND tgt.business_function_id = src.business_function_id
	AND tgt.asset_id = src.asset_id
WHEN NOT MATCHED
	THEN INSERT (tenant_id, business_function_id, asset_id, organization_id, created_at)
		VALUES (src.tenant_id, src.business_function_id, src.asset_id, src.organization_id, src.created_at)
	WHEN NOT MATCHED BY SOURCE
		AND tgt.tenant_id = @tenant_id AND tgt.business_function_id = @business_function_id
		THEN DELETE
	`

	args := pgx.StrictNamedArgs{
		"tenant_id":            scope.GetTenantID(),
		"business_function_id": businessFunctionID,
		"organization_id":      organizationID,
		"created_at":           time.Now(),
		"asset_ids":            assetIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot merge business function assets: %w", err)
	}

	return nil
}

func (bfa BusinessFunctionAssets) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	businessFunctionID gid.GID,
	organizationID gid.GID,
	assetIDs []gid.GID,
) error {
	q := `
WITH asset_ids AS (
	SELECT DISTINCT unnest(@asset_ids::text[]) AS asset_id
)
INSERT INTO business_function_assets (tenant_id, business_function_id, asset_id, organization_id, created_at)
SELECT
	@tenant_id AS tenant_id,
	@business_function_id AS business_function_id,
	asset_id,
	@organization_id AS organization_id,
	@created_at AS created_at
FROM asset_ids
`

	args := pgx.StrictNamedArgs{
		"tenant_id":            scope.GetTenantID(),
		"business_function_id": businessFunctionID,
		"organization_id":      organizationID,
		"created_at":           time.Now(),
		"asset_ids":            assetIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert business function assets: %w", err)
	}

	return nil
}

func (bfa *BusinessFunctionAssets) LoadAssetIDsByBusinessFunctionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	businessFunctionID gid.GID,
) ([]gid.GID, error) {
	q := `
SELECT
	asset_id
FROM
	business_function_assets
WHERE
	%s
	AND business_function_id = @business_function_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"business_function_id": businessFunctionID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query business function asset ids: %w", err)
	}

	assetIDs, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect business function asset ids: %w", err)
	}

	return assetIDs, nil
}

func (bfa *BusinessFunctionAssets) LoadAssetIDsByBusinessFunctionIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	businessFunctionIDs []gid.GID,
) (map[gid.GID][]gid.GID, error) {
	result := make(map[gid.GID][]gid.GID)
	if len(businessFunctionIDs) == 0 {
		return result, nil
	}

	q := `
SELECT
	business_function_id,
	asset_id
FROM
	business_function_assets
WHERE
	%s
	AND business_function_id = ANY(@business_function_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"business_function_ids": businessFunctionIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query business function assets: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var businessFunctionID, assetID gid.GID
		if err := rows.Scan(&businessFunctionID, &assetID); err != nil {
			return nil, fmt.Errorf("cannot scan business function asset: %w", err)
		}

		result[businessFunctionID] = append(result[businessFunctionID], assetID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate business function assets: %w", err)
	}

	return result, nil
}
