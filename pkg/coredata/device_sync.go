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

// ScheduleMissingDeviceSyncs creates a sync, due at now, for every connector
// of the given providers that has none, across every tenant. The console
// schedules a connector's sync when it is saved, but that runs after the
// connector is committed; this is what makes a failed scheduling heal on the
// worker's next poll instead of leaving the connector never synced.
func ScheduleMissingDeviceSyncs(
	ctx context.Context,
	conn pg.Querier,
	providers []ConnectorProvider,
	now time.Time,
) (int64, error) {
	if len(providers) == 0 {
		return 0, nil
	}

	q := `
INSERT INTO device_syncs (
    connector_id,
    tenant_id,
    organization_id,
    attempts,
    next_sync_at,
    created_at,
    updated_at
)
SELECT
    c.id,
    c.tenant_id,
    c.organization_id,
    0,
    @now,
    @now,
    @now
FROM
    connectors c
WHERE
    c.provider = ANY(@providers::connector_provider[])
    AND NOT EXISTS (
        SELECT 1 FROM device_syncs s WHERE s.connector_id = c.id
    )
ON CONFLICT (connector_id) DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"providers": filterStrings(providers),
		"now":       now,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot schedule missing device syncs: %w", err)
	}

	return result.RowsAffected(), nil
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

// RecordAttempt charges one attempt and leases the row until leaseUntil
// before the provider is called, so a crash mid-sync cannot leave the row due
// again on the next poll, and no other worker claims it while the lease runs.
//
// The attempt count and lease it stores identify this claim: RecordSuccess
// and RecordFailure only apply while both are unchanged, so a Schedule that
// lands mid-sync is never overwritten.
func (s *DeviceSync) RecordAttempt(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	now time.Time,
	leaseUntil time.Time,
) error {
	q := fmt.Sprintf(`
UPDATE device_syncs
SET
    attempts = attempts + 1,
    next_sync_at = @next_sync_at,
    updated_at = @now
WHERE %s
    AND connector_id = @connector_id
RETURNING
    attempts,
    next_sync_at
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": s.ConnectorID,
		"next_sync_at": leaseUntil,
		"now":          now,
	}
	maps.Copy(args, scope.SQLArguments())

	// The stored values are read back rather than kept from Go: the
	// database rounds timestamps to the microsecond, and completion
	// compares against what it stored.
	err := conn.QueryRow(ctx, q, args).Scan(&s.Attempts, &s.NextSyncAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot record device sync attempt: %w", err)
	}

	s.UpdatedAt = now

	return nil
}

// RecordSuccess resets the retry budget and schedules the next regular sync.
// It returns ErrResourceNotFound, and changes nothing, when the claim was
// superseded: the connector was reconnected or deleted, or its lease ran out
// and another worker claimed it.
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
    AND attempts = @claimed_attempts
    AND next_sync_at = @claimed_next_sync_at
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id":         s.ConnectorID,
		"next_sync_at":         nextSyncAt,
		"now":                  now,
		"claimed_attempts":     s.Attempts,
		"claimed_next_sync_at": s.NextSyncAt,
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

// RecordFailure stamps the failure and schedules the retry. Like
// RecordSuccess, it changes nothing once the claim was superseded.
func (s *DeviceSync) RecordFailure(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	now time.Time,
	nextSyncAt time.Time,
) error {
	q := fmt.Sprintf(`
UPDATE device_syncs
SET
    last_failed_at = @now,
    next_sync_at = @next_sync_at,
    updated_at = @now
WHERE %s
    AND connector_id = @connector_id
    AND attempts = @claimed_attempts
    AND next_sync_at = @claimed_next_sync_at
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id":         s.ConnectorID,
		"next_sync_at":         nextSyncAt,
		"now":                  now,
		"claimed_attempts":     s.Attempts,
		"claimed_next_sync_at": s.NextSyncAt,
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
	s.NextSyncAt = nextSyncAt
	s.UpdatedAt = now

	return nil
}
