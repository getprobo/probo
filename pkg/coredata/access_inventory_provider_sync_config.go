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
	"go.probo.inc/probo/pkg/timespan"
)

type (
	// AccessInventoryProviderSyncConfig is an optional org override of the
	// built-in sync cadence for one provider key.
	AccessInventoryProviderSyncConfig struct {
		ID             gid.GID                    `db:"id"`
		OrganizationID gid.GID                    `db:"organization_id"`
		ProviderKey    AccessInventoryProviderKey `db:"provider_key"`
		SyncCadence    timespan.TimeSpan          `db:"sync_cadence"`
		CreatedAt      time.Time                  `db:"created_at"`
		UpdatedAt      time.Time                  `db:"updated_at"`
	}

	AccessInventoryProviderSyncConfigs []*AccessInventoryProviderSyncConfig
)

func (c AccessInventoryProviderSyncConfig) CursorKey(
	orderBy AccessInventoryProviderSyncConfigOrderField,
) page.CursorKey {
	switch orderBy {
	case AccessInventoryProviderSyncConfigOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *AccessInventoryProviderSyncConfig) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_inventory_provider_sync_configs WHERE id = ANY(@resource_ids::text[])`

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

func (c *AccessInventoryProviderSyncConfig) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    provider_key,
    sync_cadence,
    created_at,
    updated_at
FROM
    access_inventory_provider_sync_configs
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
		return fmt.Errorf("cannot query access_inventory_provider_sync_configs: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryProviderSyncConfig])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access inventory provider sync config: %w", err)
	}

	*c = config

	return nil
}

func (c *AccessInventoryProviderSyncConfig) LoadByOrganizationIDAndProviderKey(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	providerKey AccessInventoryProviderKey,
) error {
	q := `
SELECT
    id,
    organization_id,
    provider_key,
    sync_cadence,
    created_at,
    updated_at
FROM
    access_inventory_provider_sync_configs
WHERE
    %s
    AND organization_id = @organization_id
    AND provider_key = @provider_key
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"provider_key":    providerKey,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_inventory_provider_sync_configs: %w", err)
	}

	config, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryProviderSyncConfig])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access inventory provider sync config: %w", err)
	}

	*c = config

	return nil
}

func (cs *AccessInventoryProviderSyncConfigs) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AccessInventoryProviderSyncConfigOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    provider_key,
    sync_cadence,
    created_at,
    updated_at
FROM
    access_inventory_provider_sync_configs
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
		return fmt.Errorf("cannot query access_inventory_provider_sync_configs: %w", err)
	}

	configs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessInventoryProviderSyncConfig])
	if err != nil {
		return fmt.Errorf("cannot collect access_inventory_provider_sync_configs: %w", err)
	}

	*cs = configs

	return nil
}

func (c *AccessInventoryProviderSyncConfig) validateForWrite() error {
	if !c.ProviderKey.IsValid() {
		return fmt.Errorf("invalid access inventory provider key: %q", c.ProviderKey)
	}

	if c.SyncCadence.Compare(timespan.Zero) < 0 {
		return fmt.Errorf("sync cadence must be non-negative")
	}

	if _, ok := c.SyncCadence.ClockDuration(); !ok {
		return fmt.Errorf("sync cadence must be representable as a fixed clock interval")
	}

	return nil
}

func (c *AccessInventoryProviderSyncConfig) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	if err := c.validateForWrite(); err != nil {
		return fmt.Errorf("cannot insert access inventory provider sync config: %w", err)
	}

	q := `
INSERT INTO access_inventory_provider_sync_configs (
    id,
    tenant_id,
    organization_id,
    provider_key,
    sync_cadence,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @provider_key,
    @sync_cadence,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              c.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": c.OrganizationID,
		"provider_key":    c.ProviderKey,
		"sync_cadence":    c.SyncCadence,
		"created_at":      c.CreatedAt,
		"updated_at":      c.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert access inventory provider sync config: %w", err)
	}

	return nil
}

func (c *AccessInventoryProviderSyncConfig) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (inserted bool, err error) {
	if err := c.validateForWrite(); err != nil {
		return false, fmt.Errorf("cannot upsert access inventory provider sync config: %w", err)
	}

	q := `
INSERT INTO access_inventory_provider_sync_configs (
    id,
    tenant_id,
    organization_id,
    provider_key,
    sync_cadence,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @provider_key,
    @sync_cadence,
    @created_at,
    @updated_at
)
ON CONFLICT (organization_id, provider_key) DO UPDATE SET
    sync_cadence = EXCLUDED.sync_cadence,
    updated_at   = EXCLUDED.updated_at
RETURNING
    id,
    organization_id,
    provider_key,
    sync_cadence,
    created_at,
    updated_at
`
	originalID := c.ID

	args := pgx.StrictNamedArgs{
		"id":              c.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": c.OrganizationID,
		"provider_key":    c.ProviderKey,
		"sync_cadence":    c.SyncCadence,
		"created_at":      c.CreatedAt,
		"updated_at":      c.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert access inventory provider sync config: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessInventoryProviderSyncConfig])
	if err != nil {
		return false, fmt.Errorf("cannot collect access inventory provider sync config upsert: %w", err)
	}

	*c = row

	return originalID == c.ID, nil
}

func (c *AccessInventoryProviderSyncConfig) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	if c.SyncCadence.Compare(timespan.Zero) < 0 {
		return fmt.Errorf("cannot update access inventory provider sync config: sync cadence must be non-negative")
	}

	q := `
UPDATE access_inventory_provider_sync_configs
SET
    sync_cadence = @sync_cadence,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           c.ID,
		"sync_cadence": c.SyncCadence,
		"updated_at":   c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update access inventory provider sync config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (c *AccessInventoryProviderSyncConfig) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM access_inventory_provider_sync_configs
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete access inventory provider sync config: %w", err)
	}

	return nil
}
