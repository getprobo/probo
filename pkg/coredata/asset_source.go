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
)

type (
	AssetSource struct {
		ID                 gid.GID           `db:"id"`
		TenantID           gid.TenantID      `db:"tenant_id"`
		OrganizationID     gid.GID           `db:"organization_id"`
		Provider           ConnectorProvider `db:"provider"`
		ExternalAccountID  string            `db:"external_account_id"`
		ConnectorAccountID *gid.GID          `db:"connector_account_id"`
		NextSyncAt         time.Time         `db:"next_sync_at"`
		CreatedAt          time.Time         `db:"created_at"`
		UpdatedAt          time.Time         `db:"updated_at"`
	}
)

// Ensure inserts the source for this external account, or points the
// existing source at the account and marks it due. Another source that
// still names this account under a different external id loses that pointer.
func (s *AssetSource) Ensure(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	if s.ConnectorAccountID == nil {
		return fmt.Errorf("cannot ensure asset source: connector account is required")
	}

	if s.ExternalAccountID == "" {
		return fmt.Errorf("cannot ensure asset source: external account is required")
	}

	now := time.Now()
	s.ID = gid.New(scope.GetTenantID(), AssetSourceEntityType)
	s.CreatedAt = now
	s.UpdatedAt = now
	// A second in the past stays due after timestamp rounding.
	s.NextSyncAt = now.Add(-time.Second)

	q := `
INSERT INTO assets_sources (
    id,
    tenant_id,
    organization_id,
    provider,
    external_account_id,
    connector_account_id,
    next_sync_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @provider,
    @external_account_id,
    @connector_account_id,
    @next_sync_at,
    @created_at,
    @updated_at
)
ON CONFLICT ON CONSTRAINT assets_sources_identity_key DO UPDATE
SET
    connector_account_id = EXCLUDED.connector_account_id,
    next_sync_at = EXCLUDED.next_sync_at,
    updated_at = EXCLUDED.updated_at
RETURNING
    id,
    tenant_id,
    organization_id,
    provider,
    external_account_id,
    connector_account_id,
    next_sync_at,
    created_at,
    updated_at
`

	args := pgx.StrictNamedArgs{
		"id":                   s.ID,
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      s.OrganizationID,
		"provider":             s.Provider,
		"external_account_id":  s.ExternalAccountID,
		"connector_account_id": s.ConnectorAccountID,
		"next_sync_at":         s.NextSyncAt,
		"created_at":           s.CreatedAt,
		"updated_at":           s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query asset source: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AssetSource])
	if err != nil {
		return fmt.Errorf("cannot collect asset source: %w", err)
	}

	*s = row

	detach := `
UPDATE assets_sources
SET
    connector_account_id = NULL,
    updated_at = @updated_at
WHERE
    %s
    AND connector_account_id = @connector_account_id
    AND id <> @id
`
	detach = fmt.Sprintf(detach, scope.SQLFragment())

	detachArgs := pgx.StrictNamedArgs{
		"id":                   s.ID,
		"connector_account_id": s.ConnectorAccountID,
		"updated_at":           s.UpdatedAt,
	}
	maps.Copy(detachArgs, scope.SQLArguments())

	if _, err := conn.Exec(ctx, detach, detachArgs); err != nil {
		return fmt.Errorf("cannot detach asset source from connector account: %w", err)
	}

	return nil
}

func (s *AssetSource) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    provider,
    external_account_id,
    connector_account_id,
    next_sync_at,
    created_at,
    updated_at
FROM
    assets_sources
WHERE
    %s
    AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query asset source: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AssetSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect asset source: %w", err)
	}

	*s = row

	return nil
}

func (s *AssetSource) LoadByIdentity(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	provider ConnectorProvider,
	externalAccountID string,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    provider,
    external_account_id,
    connector_account_id,
    next_sync_at,
    created_at,
    updated_at
FROM
    assets_sources
WHERE
    %s
    AND organization_id = @organization_id
    AND provider = @provider
    AND external_account_id = @external_account_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":     organizationID,
		"provider":            provider,
		"external_account_id": externalAccountID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query asset source by identity: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AssetSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect asset source by identity: %w", err)
	}

	*s = row

	return nil
}

// LoadNextDueForUpdateSkipLocked locks one source that has an account and
// is due. It leaves next_sync_at alone so a crash can be retried before the
// daily interval. A running sync keeps the source out of this query.
// onlySourceID, when non-nil, restricts the lock to that source.
func (s *AssetSource) LoadNextDueForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	onlySourceID *gid.GID,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    provider,
    external_account_id,
    connector_account_id,
    next_sync_at,
    created_at,
    updated_at
FROM
    assets_sources
WHERE
    connector_account_id IS NOT NULL
    AND next_sync_at <= @now
    AND NOT EXISTS (
        SELECT 1
        FROM assets_syncs
        WHERE
            assets_syncs.assets_source_id = assets_sources.id
            AND assets_syncs.status = 'RUNNING'
    )
    AND (@only_source_id::text IS NULL OR id = @only_source_id)
ORDER BY
    next_sync_at
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	args := pgx.StrictNamedArgs{
		"now":            now,
		"only_source_id": onlySourceID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query due asset source: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AssetSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect due asset source: %w", err)
	}

	*s = row

	return nil
}

func (s *AssetSource) SetNextSyncAt(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	next time.Time,
) error {
	s.NextSyncAt = next
	s.UpdatedAt = time.Now()

	q := `
UPDATE assets_sources
SET
    next_sync_at = @next_sync_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           s.ID,
		"next_sync_at": s.NextSyncAt,
		"updated_at":   s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot schedule asset source: %w", err)
	}

	return nil
}
