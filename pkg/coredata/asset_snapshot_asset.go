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
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AssetSnapshotAsset struct {
		ID                  gid.GID         `db:"id"`
		SnapshotID          gid.GID         `db:"snapshot_id"`
		OrganizationID      gid.GID         `db:"organization_id"`
		AssetID             gid.GID         `db:"asset_id"`
		Name                string          `db:"name"`
		Amount              int             `db:"amount"`
		OwnerID             *gid.GID        `db:"owner_profile_id"`
		AssetType           AssetType       `db:"asset_type"`
		DataTypesStored     *string         `db:"data_types_stored"`
		Kind                *string         `db:"kind"`
		ExternalID          *string         `db:"external_id"`
		Region              *string         `db:"region"`
		Configuration       json.RawMessage `db:"configuration"`
		SchemaVersion       *int            `db:"schema_version"`
		LastSeenAt          *time.Time      `db:"last_seen_at"`
		Status              AssetStatus     `db:"status"`
		ConnectorID         *gid.GID        `db:"connector_id"`
		AssetCreatedAt      time.Time       `db:"asset_created_at"`
		AssetUpdatedAt      time.Time       `db:"asset_updated_at"`
		ThirdPartyIDs       []string        `db:"third_party_ids"`
		BusinessFunctionIDs []string        `db:"business_function_ids"`
		TagIDs              []string        `db:"tag_ids"`
	}

	AssetSnapshotAssets []*AssetSnapshotAsset
)

func (a AssetSnapshotAsset) CursorKey(orderBy AssetSnapshotAssetOrderField) page.CursorKey {
	switch orderBy {
	case AssetSnapshotAssetOrderFieldName:
		return page.NewCursorKey(a.ID, a.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (a *AssetSnapshotAsset) ThirdParties() ([]gid.GID, error) {
	return parseStoredGIDs(a.ThirdPartyIDs)
}

func (a *AssetSnapshotAsset) BusinessFunctions() ([]gid.GID, error) {
	return parseStoredGIDs(a.BusinessFunctionIDs)
}

func (a *AssetSnapshotAsset) Tags() ([]gid.GID, error) {
	return parseStoredGIDs(a.TagIDs)
}

func parseStoredGIDs(values []string) ([]gid.GID, error) {
	if len(values) == 0 {
		return []gid.GID{}, nil
	}

	ids := make([]gid.GID, len(values))
	for i, value := range values {
		id, err := gid.ParseGID(value)
		if err != nil {
			return nil, fmt.Errorf("cannot parse stored gid: %w", err)
		}

		ids[i] = id
	}

	return ids, nil
}

func (a *AssetSnapshotAssets) LoadBySnapshotID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	snapshotID gid.GID,
	cursor *page.Cursor[AssetSnapshotAssetOrderField],
) error {
	q := `
SELECT
	id,
	snapshot_id,
	organization_id,
	asset_id,
	name,
	amount,
	owner_profile_id,
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
	asset_created_at,
	asset_updated_at,
	third_party_ids,
	business_function_ids,
	tag_ids
FROM
	asset_snapshot_assets
WHERE
	%s
	AND snapshot_id = @snapshot_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"snapshot_id": snapshotID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query asset snapshot assets: %w", err)
	}

	assets, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AssetSnapshotAsset])
	if err != nil {
		return fmt.Errorf("cannot collect asset snapshot assets: %w", err)
	}

	for _, asset := range assets {
		if asset.ThirdPartyIDs == nil {
			asset.ThirdPartyIDs = []string{}
		}

		if asset.BusinessFunctionIDs == nil {
			asset.BusinessFunctionIDs = []string{}
		}

		if asset.TagIDs == nil {
			asset.TagIDs = []string{}
		}
	}

	*a = assets

	return nil
}

func (a *AssetSnapshotAssets) CountBySnapshotID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	snapshotID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	asset_snapshot_assets
WHERE
	%s
	AND snapshot_id = @snapshot_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"snapshot_id": snapshotID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count asset snapshot assets: %w", err)
	}

	return count, nil
}
