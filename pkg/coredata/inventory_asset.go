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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	// InventoryAsset is one discovered thing in an external account.
	// The table is assets. The register type Asset reads assets_registry.
	InventoryAsset struct {
		ID             gid.GID         `db:"id"`
		TenantID       gid.TenantID    `db:"tenant_id"`
		OrganizationID gid.GID         `db:"organization_id"`
		AssetSourceID  gid.GID         `db:"assets_source_id"`
		Kind           string          `db:"kind"`
		ExternalID     string          `db:"external_id"`
		Name           string          `db:"name"`
		Region         *string         `db:"region"`
		Configuration  json.RawMessage `db:"configuration"`
		SchemaVersion  int             `db:"schema_version"`
		CreatedAt      time.Time       `db:"created_at"`
		UpdatedAt      time.Time       `db:"updated_at"`
		LastSeenAt     time.Time       `db:"last_seen_at"`
	}
)

func (r *InventoryAsset) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	if len(r.Configuration) == 0 {
		r.Configuration = json.RawMessage(`{}`)
	}

	q := `
INSERT INTO assets (
    id,
    tenant_id,
    organization_id,
    assets_source_id,
    kind,
    external_id,
    name,
    region,
    configuration,
    schema_version,
    created_at,
    updated_at,
    last_seen_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @assets_source_id,
    @kind,
    @external_id,
    @name,
    @region,
    @configuration::jsonb,
    @schema_version,
    @created_at,
    @updated_at,
    @last_seen_at
)
ON CONFLICT ON CONSTRAINT assets_identity_key DO UPDATE
SET
    name = EXCLUDED.name,
    region = EXCLUDED.region,
    configuration = EXCLUDED.configuration,
    schema_version = EXCLUDED.schema_version,
    updated_at = EXCLUDED.updated_at,
    last_seen_at = EXCLUDED.last_seen_at
RETURNING
    id,
    tenant_id,
    organization_id,
    assets_source_id,
    kind,
    external_id,
    name,
    region,
    configuration,
    schema_version,
    created_at,
    updated_at,
    last_seen_at
`

	args := pgx.StrictNamedArgs{
		"id":               r.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  r.OrganizationID,
		"assets_source_id": r.AssetSourceID,
		"kind":             r.Kind,
		"external_id":      r.ExternalID,
		"name":             r.Name,
		"region":           r.Region,
		"configuration":    r.Configuration,
		"schema_version":   r.SchemaVersion,
		"created_at":       r.CreatedAt,
		"updated_at":       r.UpdatedAt,
		"last_seen_at":     r.LastSeenAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query inventory asset: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[InventoryAsset])
	if err != nil {
		return fmt.Errorf("cannot collect inventory asset: %w", err)
	}

	*r = row

	return nil
}

func (r *InventoryAsset) LoadBySourceKindExternalID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
	kind string,
	externalID string,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    assets_source_id,
    kind,
    external_id,
    name,
    region,
    configuration,
    schema_version,
    created_at,
    updated_at,
    last_seen_at
FROM
    assets
WHERE
    %s
    AND assets_source_id = @assets_source_id
    AND kind = @kind
    AND external_id = @external_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"assets_source_id": sourceID,
		"kind":             kind,
		"external_id":      externalID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query inventory asset by identity: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[InventoryAsset])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect inventory asset by identity: %w", err)
	}

	*r = row

	return nil
}
