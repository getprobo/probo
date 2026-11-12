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
	// DeviceSync is the sync schedule of one connector whose provider syncs
	// devices. It has no ID of its own: a connector has at most one.
	DeviceSync struct {
		ConnectorID    gid.GID    `db:"connector_id"`
		OrganizationID gid.GID    `db:"organization_id"`
		Attempts       int        `db:"attempts"`
		NextSyncAt     time.Time  `db:"next_sync_at"`
		LastSyncedAt   *time.Time `db:"last_synced_at"`
		LastFailedAt   *time.Time `db:"last_failed_at"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}
)

// Schedule creates the sync of a connector, or resets an existing one, so
// that it runs at NextSyncAt with a fresh retry budget. Reconnecting a
// connector goes through here, so a credential that was failing is retried
// at once instead of after its backoff.
func (s *DeviceSync) Schedule(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO device_syncs (
    connector_id,
    tenant_id,
    organization_id,
    attempts,
    next_sync_at,
    created_at,
    updated_at
) VALUES (
    @connector_id,
    @tenant_id,
    @organization_id,
    0,
    @next_sync_at,
    @created_at,
    @updated_at
)
ON CONFLICT (connector_id) DO UPDATE
SET
    attempts = 0,
    next_sync_at = EXCLUDED.next_sync_at,
    updated_at = EXCLUDED.updated_at
`

	args := pgx.StrictNamedArgs{
		"connector_id":    s.ConnectorID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": s.OrganizationID,
		"next_sync_at":    s.NextSyncAt,
		"created_at":      s.CreatedAt,
		"updated_at":      s.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot schedule device sync: %w", err)
	}

	s.Attempts = 0

	return nil
}

// LoadByConnectorID loads the sync of a connector.
func (s *DeviceSync) LoadByConnectorID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
) error {
	q := `
SELECT
    connector_id,
    organization_id,
    attempts,
    next_sync_at,
    last_synced_at,
    last_failed_at,
    created_at,
    updated_at
FROM
    device_syncs
WHERE
    %s
    AND connector_id = @connector_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"connector_id": connectorID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device sync: %w", err)
	}

	sync, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DeviceSync])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect device sync: %w", err)
	}

	*s = sync

	return nil
}

// LoadNextDueForUpdateSkipLocked locks the sync that has waited longest past
// its next_sync_at, across every tenant.
func (s *DeviceSync) LoadNextDueForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
SELECT
    connector_id,
    organization_id,
    attempts,
    next_sync_at,
    last_synced_at,
    last_failed_at,
    created_at,
    updated_at
FROM
    device_syncs
WHERE
    next_sync_at <= @now
ORDER BY
    next_sync_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"now": now})
	if err != nil {
		return fmt.Errorf("cannot query due device sync: %w", err)
	}

	sync, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DeviceSync])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect due device sync: %w", err)
	}

	*s = sync

	return nil
}

// RecordAttempt charges one attempt and pushes next_sync_at out before the
// provider is called, so a crash mid-sync cannot leave the row due again on
// the next poll.
func (s *DeviceSync) RecordAttempt(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	now time.Time,
	nextSyncAt time.Time,
) error {
	q := fmt.Sprintf(`
UPDATE device_syncs
SET
    attempts = attempts + 1,
    next_sync_at = @next_sync_at,
    updated_at = @now
WHERE %s
    AND connector_id = @connector_id
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": s.ConnectorID,
		"next_sync_at": nextSyncAt,
		"now":          now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot record device sync attempt: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.Attempts++
	s.NextSyncAt = nextSyncAt
	s.UpdatedAt = now

	return nil
}

// RecordSuccess resets the retry budget and schedules the next regular sync.
func (s *DeviceSync) RecordSuccess(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	now time.Time,
	nextSyncAt time.Time,
) error {
	q := fmt.Sprintf(`
UPDATE device_syncs
SET
    attempts = 0,
    last_synced_at = @now,
    next_sync_at = @next_sync_at,
    updated_at = @now
WHERE %s
    AND connector_id = @connector_id
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": s.ConnectorID,
		"next_sync_at": nextSyncAt,
		"now":          now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot record device sync success: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.Attempts = 0
	s.LastSyncedAt = &now
	s.NextSyncAt = nextSyncAt
	s.UpdatedAt = now

	return nil
}

// RecordFailure stamps the failure. The backoff was already charged by
// RecordAttempt, so next_sync_at is left as it is.
func (s *DeviceSync) RecordFailure(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	now time.Time,
) error {
	q := fmt.Sprintf(`
UPDATE device_syncs
SET
    last_failed_at = @now,
    updated_at = @now
WHERE %s
    AND connector_id = @connector_id
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": s.ConnectorID,
		"now":          now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot record device sync failure: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.LastFailedAt = &now
	s.UpdatedAt = now

	return nil
}
