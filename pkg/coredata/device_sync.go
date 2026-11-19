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

// ErrDeviceSyncSuperseded is returned by LockClaim when the connector was
// saved again after the sync was claimed: what the sync fetched used the
// previous credentials or settings and must not be applied.
var ErrDeviceSyncSuperseded = errors.New("device sync superseded")

type (
	// DeviceSync is the sync schedule of one connector whose provider syncs
	// devices. It has no ID of its own: a connector has at most one.
	//
	// A worker claims a sync by charging an attempt and taking its lease.
	// While the lease runs no other worker can claim it, even if a save
	// makes it due again. The lease and the generation it was claimed at
	// identify the claim: completion only applies while the lease is still
	// the one taken, and only reschedules while the generation is unchanged,
	// so a save that lands mid-sync keeps its own schedule.
	DeviceSync struct {
		ConnectorID    gid.GID    `db:"connector_id"`
		OrganizationID gid.GID    `db:"organization_id"`
		Attempts       int        `db:"attempts"`
		NextSyncAt     time.Time  `db:"next_sync_at"`
		LastSyncedAt   *time.Time `db:"last_synced_at"`
		LastFailedAt   *time.Time `db:"last_failed_at"`
		LeaseUntil     *time.Time `db:"lease_until"`
		Generation     int64      `db:"generation"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}
)

const deviceSyncColumns = `
    connector_id,
    organization_id,
    attempts,
    next_sync_at,
    last_synced_at,
    last_failed_at,
    lease_until,
    generation,
    created_at,
    updated_at`

// Schedule creates the sync of a connector, or resets an existing one, so
// that it runs at NextSyncAt with a fresh retry budget and a new generation.
// Saving a connector goes through here, in the same transaction, so a
// credential that was failing is retried at once instead of after its
// backoff, and a sync still running with the old one is not applied. A lease
// held by a running sync is kept: the sync runs again once that one is done.
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
    generation,
    created_at,
    updated_at
) VALUES (
    @connector_id,
    @tenant_id,
    @organization_id,
    0,
    @next_sync_at,
    0,
    @created_at,
    @updated_at
)
ON CONFLICT (connector_id) DO UPDATE
SET
    attempts = 0,
    next_sync_at = EXCLUDED.next_sync_at,
    generation = device_syncs.generation + 1,
    updated_at = EXCLUDED.updated_at
RETURNING
    generation
`

	args := pgx.StrictNamedArgs{
		"connector_id":    s.ConnectorID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": s.OrganizationID,
		"next_sync_at":    s.NextSyncAt,
		"created_at":      s.CreatedAt,
		"updated_at":      s.UpdatedAt,
	}

	if err := conn.QueryRow(ctx, q, args).Scan(&s.Generation); err != nil {
		return fmt.Errorf("cannot schedule device sync: %w", err)
	}

	s.Attempts = 0

	return nil
}

// ScheduleMissingDeviceSyncs creates a sync, due at now, for every connector
// of the given providers that has none, across every tenant. Saving a
// connector schedules its sync in the same transaction; this covers
// connectors saved before their provider synced devices.
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
    generation,
    created_at,
    updated_at
)
SELECT
    c.id,
    c.tenant_id,
    c.organization_id,
    0,
    @now,
    0,
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
SELECT` + deviceSyncColumns + `
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
// its next_sync_at and that no worker holds a lease on, across every tenant.
func (s *DeviceSync) LoadNextDueForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
SELECT` + deviceSyncColumns + `
FROM
    device_syncs
WHERE
    next_sync_at <= @now
    AND (lease_until IS NULL OR lease_until <= @now)
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

// RecordAttempt claims the sync: it charges one attempt and takes the lease
// until leaseUntil, before the provider is called, so a crash mid-sync
// cannot leave the row claimable on the next poll. A worker that crashed
// leaves it claimable again once the lease runs out.
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
    lease_until = @lease_until,
    updated_at = @now
WHERE %s
    AND connector_id = @connector_id
RETURNING
    attempts,
    lease_until,
    generation
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": s.ConnectorID,
		"lease_until":  leaseUntil,
		"now":          now,
	}
	maps.Copy(args, scope.SQLArguments())

	// The stored lease is read back rather than kept from Go: the database
	// rounds timestamps to the microsecond, and the claim is recognized by
	// comparing against what it stored.
	err := conn.QueryRow(ctx, q, args).Scan(&s.Attempts, &s.LeaseUntil, &s.Generation)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot record device sync attempt: %w", err)
	}

	s.UpdatedAt = now

	return nil
}

// LockClaim locks the sync row for the rest of tx and checks that this claim
// still holds it, so devices are only written by the claim that owns the
// sync. It returns ErrResourceNotFound when the lease ran out and another
// worker claimed the sync, or the connector was deleted, and
// ErrDeviceSyncSuperseded when the connector was saved since the claim.
func (s *DeviceSync) LockClaim(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := fmt.Sprintf(`
SELECT
    generation
FROM
    device_syncs
WHERE %s
    AND connector_id = @connector_id
    AND lease_until = @lease_until
FOR UPDATE
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": s.ConnectorID,
		"lease_until":  s.LeaseUntil,
	}
	maps.Copy(args, scope.SQLArguments())

	var generation int64
	if err := conn.QueryRow(ctx, q, args).Scan(&generation); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot lock device sync claim: %w", err)
	}

	if generation != s.Generation {
		return ErrDeviceSyncSuperseded
	}

	return nil
}

// RecordSuccess releases the claim and stamps the sync. Unless the connector
// was saved since the claim, it also resets the retry budget and schedules
// the next regular sync; otherwise the save's own schedule stands. It returns
// ErrResourceNotFound, and changes nothing, when this claim no longer holds
// the lease.
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
    attempts = CASE WHEN generation = @generation THEN 0 ELSE attempts END,
    next_sync_at = CASE WHEN generation = @generation THEN @next_sync_at ELSE next_sync_at END,
    last_synced_at = @now,
    lease_until = NULL,
    updated_at = @now
WHERE %s
    AND connector_id = @connector_id
    AND lease_until = @lease_until
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": s.ConnectorID,
		"next_sync_at": nextSyncAt,
		"now":          now,
		"generation":   s.Generation,
		"lease_until":  s.LeaseUntil,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot record device sync success: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.LastSyncedAt = &now
	s.LeaseUntil = nil
	s.UpdatedAt = now

	return nil
}

// RecordFailure releases the claim, stamps the failure and, unless the
// connector was saved since the claim, schedules the retry. Like
// RecordSuccess, it changes nothing once this claim lost the lease.
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
    next_sync_at = CASE WHEN generation = @generation THEN @next_sync_at ELSE next_sync_at END,
    last_failed_at = @now,
    lease_until = NULL,
    updated_at = @now
WHERE %s
    AND connector_id = @connector_id
    AND lease_until = @lease_until
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": s.ConnectorID,
		"next_sync_at": nextSyncAt,
		"now":          now,
		"generation":   s.Generation,
		"lease_until":  s.LeaseUntil,
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
	s.LeaseUntil = nil
	s.UpdatedAt = now

	return nil
}

// Release gives up the claim without recording an outcome, for a sync that
// was superseded before it could apply anything. The schedule the save set
// is left as it is, so the sync runs again with the new credentials.
func (s *DeviceSync) Release(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	now time.Time,
) error {
	q := fmt.Sprintf(`
UPDATE device_syncs
SET
    lease_until = NULL,
    updated_at = @now
WHERE %s
    AND connector_id = @connector_id
    AND lease_until = @lease_until
`, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": s.ConnectorID,
		"now":          now,
		"lease_until":  s.LeaseUntil,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot release device sync: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.LeaseUntil = nil
	s.UpdatedAt = now

	return nil
}
