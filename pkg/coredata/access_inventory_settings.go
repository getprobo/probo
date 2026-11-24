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
	"go.probo.inc/probo/pkg/timespan"
)

type (
	// AccessInventorySettings is org-level live inventory config: hide-after-X
	// and (stub) retention. Missing row means product defaults apply.
	AccessInventorySettings struct {
		ID                    gid.GID            `db:"id"`
		OrganizationID        gid.GID            `db:"organization_id"`
		HideAfterMissingSyncs int                `db:"hide_after_missing_syncs"`
		ChangeEventRetention  *timespan.TimeSpan `db:"change_event_retention"`
		SnapshotRetention     *timespan.TimeSpan `db:"snapshot_retention"`
		CreatedAt             time.Time          `db:"created_at"`
		UpdatedAt             time.Time          `db:"updated_at"`
	}
)

func (s *AccessInventorySettings) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_inventory_settings WHERE id = ANY(@resource_ids::text[])`

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

func (s *AccessInventorySettings) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    hide_after_missing_syncs,
    change_event_retention,
    snapshot_retention,
    created_at,
    updated_at
FROM
    access_inventory_settings
WHERE
    %s
    AND id = @id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_settings: %w", err)
	}

	settings, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventorySettings])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access inventory settings: %w", err)
	}

	*s = settings

	return nil
}

func (s *AccessInventorySettings) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    hide_after_missing_syncs,
    change_event_retention,
    snapshot_retention,
    created_at,
    updated_at
FROM
    access_inventory_settings
WHERE
    %s
    AND organization_id = @organization_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_settings: %w", err)
	}

	settings, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventorySettings])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access inventory settings: %w", err)
	}

	*s = settings

	return nil
}

func (s *AccessInventorySettings) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO access_inventory_settings (
    id,
    tenant_id,
    organization_id,
    hide_after_missing_syncs,
    change_event_retention,
    snapshot_retention,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @hide_after_missing_syncs,
    @change_event_retention,
    @snapshot_retention,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                       s.ID,
		"tenant_id":                scope.GetTenantID(),
		"organization_id":          s.OrganizationID,
		"hide_after_missing_syncs": s.HideAfterMissingSyncs,
		"change_event_retention":   s.ChangeEventRetention,
		"snapshot_retention":       s.SnapshotRetention,
		"created_at":               s.CreatedAt,
		"updated_at":               s.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert access inventory settings: %w", err)
	}

	return nil
}

func (s *AccessInventorySettings) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (inserted bool, err error) {
	q := `
INSERT INTO access_inventory_settings (
    id,
    tenant_id,
    organization_id,
    hide_after_missing_syncs,
    change_event_retention,
    snapshot_retention,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @hide_after_missing_syncs,
    @change_event_retention,
    @snapshot_retention,
    @created_at,
    @updated_at
)
ON CONFLICT (organization_id) DO UPDATE SET
    hide_after_missing_syncs = EXCLUDED.hide_after_missing_syncs,
    change_event_retention   = EXCLUDED.change_event_retention,
    snapshot_retention       = EXCLUDED.snapshot_retention,
    updated_at               = EXCLUDED.updated_at
RETURNING
    id,
    organization_id,
    hide_after_missing_syncs,
    change_event_retention,
    snapshot_retention,
    created_at,
    updated_at
`
	originalID := s.ID

	args := pgx.StrictNamedArgs{
		"id":                       s.ID,
		"tenant_id":                scope.GetTenantID(),
		"organization_id":          s.OrganizationID,
		"hide_after_missing_syncs": s.HideAfterMissingSyncs,
		"change_event_retention":   s.ChangeEventRetention,
		"snapshot_retention":       s.SnapshotRetention,
		"created_at":               s.CreatedAt,
		"updated_at":               s.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert access inventory settings: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventorySettings])
	if err != nil {
		return false, fmt.Errorf("cannot collect access inventory settings upsert: %w", err)
	}

	*s = row

	return originalID == s.ID, nil
}

func (s *AccessInventorySettings) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE access_inventory_settings
SET
    hide_after_missing_syncs = @hide_after_missing_syncs,
    change_event_retention = @change_event_retention,
    snapshot_retention = @snapshot_retention,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                       s.ID,
		"hide_after_missing_syncs": s.HideAfterMissingSyncs,
		"change_event_retention":   s.ChangeEventRetention,
		"snapshot_retention":       s.SnapshotRetention,
		"updated_at":               s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update access inventory settings: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *AccessInventorySettings) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM access_inventory_settings
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": s.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete access inventory settings: %w", err)
	}

	return nil
}
