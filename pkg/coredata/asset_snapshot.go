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

type (
	AssetSnapshot struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CapturedAt     time.Time `db:"captured_at"`
		ActorID        *gid.GID  `db:"actor_profile_id"`
		Name           *string   `db:"name"`
		Note           *string   `db:"note"`
	}

	AssetSnapshots []*AssetSnapshot
)

func (s AssetSnapshot) CursorKey(orderBy AssetSnapshotOrderField) page.CursorKey {
	switch orderBy {
	case AssetSnapshotOrderFieldCapturedAt:
		return page.NewCursorKey(s.ID, s.CapturedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *AssetSnapshot) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
	id,
	organization_id
FROM
	asset_snapshots
WHERE
	id = ANY(@resource_ids)
`

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}

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

func (s *AssetSnapshot) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO asset_snapshots (
	id,
	tenant_id,
	organization_id,
	captured_at,
	actor_profile_id,
	name,
	note
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@captured_at,
	@actor_profile_id,
	@name,
	@note
)
`

	args := pgx.StrictNamedArgs{
		"id":               s.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  s.OrganizationID,
		"captured_at":      s.CapturedAt,
		"actor_profile_id": s.ActorID,
		"name":             s.Name,
		"note":             s.Note,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert asset snapshot: %w", err)
	}

	return nil
}

// CopyActiveAssets inserts one row for each ACTIVE asset in the snapshot's
// organization, including the link ids stored at this moment. Deactivated
// rows are left out. The statement does not read the assets again later.
func (s *AssetSnapshot) CopyActiveAssets(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO asset_snapshot_assets (
	id,
	tenant_id,
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
)
SELECT
	generate_gid(decode_base64_unpadded(assets.tenant_id), @snapshot_asset_entity_type),
	assets.tenant_id,
	@snapshot_id,
	assets.organization_id,
	assets.id,
	assets.name,
	assets.amount,
	assets.owner_profile_id,
	assets.asset_type,
	assets.data_types_stored,
	assets.kind,
	assets.external_id,
	assets.region,
	assets.configuration,
	assets.schema_version,
	assets.last_seen_at,
	assets.status,
	assets.connector_id,
	assets.created_at,
	assets.updated_at,
	COALESCE((
		SELECT array_agg(asset_third_parties.third_party_id ORDER BY asset_third_parties.third_party_id)
		FROM asset_third_parties
		WHERE asset_third_parties.asset_id = assets.id
	), '{}'::text[]),
	COALESCE((
		SELECT array_agg(business_function_assets.business_function_id ORDER BY business_function_assets.business_function_id)
		FROM business_function_assets
		WHERE business_function_assets.asset_id = assets.id
	), '{}'::text[]),
	COALESCE((
		SELECT array_agg(tag_assignments.tag_id ORDER BY tag_assignments.tag_id)
		FROM tag_assignments
		WHERE tag_assignments.resource_id = assets.id
	), '{}'::text[])
FROM
	assets
WHERE
	%s
	AND assets.organization_id = @organization_id
	AND assets.status = @status
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"snapshot_id":                s.ID,
		"organization_id":            s.OrganizationID,
		"status":                     AssetStatusActive,
		"snapshot_asset_entity_type": int(AssetSnapshotAssetEntityType),
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot copy active assets: %w", err)
	}

	return nil
}

func (s *AssetSnapshot) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	snapshotID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	captured_at,
	actor_profile_id,
	name,
	note
FROM
	asset_snapshots
WHERE
	%s
	AND id = @snapshot_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"snapshot_id": snapshotID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query asset snapshots: %w", err)
	}

	snapshot, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AssetSnapshot])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect asset snapshot: %w", err)
	}

	*s = snapshot

	return nil
}

func (s *AssetSnapshots) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AssetSnapshotOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	captured_at,
	actor_profile_id,
	name,
	note
FROM
	asset_snapshots
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
		return fmt.Errorf("cannot query asset snapshots: %w", err)
	}

	snapshots, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AssetSnapshot])
	if err != nil {
		return fmt.Errorf("cannot collect asset snapshots: %w", err)
	}

	*s = snapshots

	return nil
}

func (s *AssetSnapshots) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	asset_snapshots
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
		return 0, fmt.Errorf("cannot count asset snapshots: %w", err)
	}

	return count, nil
}

// DeleteAssetSnapshotsBeforeBatch deletes snapshots captured strictly before
// the cutoff. A name does not keep a snapshot. Copies cascade with the row.
//
// No ORDER BY captured_at. There is no index leading with captured_at, so
// each ordered batch would sort every expired row. Any expired row is
// eligible. SKIP LOCKED lets another replica take a different batch.
func DeleteAssetSnapshotsBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	batchSize int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM asset_snapshots
	WHERE captured_at < @before
	LIMIT @limit
	FOR UPDATE SKIP LOCKED
)
DELETE FROM asset_snapshots
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"before": before,
			"limit":  batchSize,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete asset snapshots: %w", err)
	}

	return result.RowsAffected(), nil
}
