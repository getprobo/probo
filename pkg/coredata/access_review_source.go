// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	AccessReviewSource struct {
		ID                    gid.GID    `db:"id"`
		OrganizationID        gid.GID    `db:"organization_id"`
		ConnectorID           *gid.GID   `db:"connector_id"`
		Name                  string     `db:"name"`
		CsvData               *string    `db:"csv_data"`
		NameSyncedAt          *time.Time `db:"name_synced_at"`
		NameSyncAttempts      int        `db:"name_sync_attempts"`
		NameSyncNextAttemptAt *time.Time `db:"name_sync_next_attempt_at"`
		CreatedAt             time.Time  `db:"created_at"`
		UpdatedAt             time.Time  `db:"updated_at"`
	}

	AccessReviewSources []*AccessReviewSource
)

func (as AccessReviewSource) CursorKey(orderBy AccessReviewSourceOrderField) page.CursorKey {
	switch orderBy {
	case AccessReviewSourceOrderFieldCreatedAt:
		return page.NewCursorKey(as.ID, as.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (as *AccessReviewSource) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_review_sources WHERE id = ANY(@resource_ids::text[])`

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

func (as *AccessReviewSource) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    csv_data,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
FROM
    access_review_sources
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
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	source, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access source: %w", err)
	}

	*as = source

	return nil
}

// LoadByIDForUpdate is LoadByID under FOR UPDATE, so the caller's
// connector handoff reads connector_id under the row lock.
func (as *AccessReviewSource) LoadByIDForUpdate(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    csv_data,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
FROM
    access_review_sources
WHERE
    %s
    AND id = @id
LIMIT 1
FOR UPDATE;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	source, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access source: %w", err)
	}

	*as = source

	return nil
}

func (sources *AccessReviewSources) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ids []gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    csv_data,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
FROM
    access_review_sources
WHERE
    %s
    AND id = ANY(@ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewSource])
	if err != nil {
		return fmt.Errorf("cannot collect access_review_sources: %w", err)
	}

	*sources = result

	if len(result) != len(gid.NewSet(ids...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (sources *AccessReviewSources) CountByConnectorID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_review_sources
WHERE
    %s
    AND connector_id = @connector_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"connector_id": connectorID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_review_sources by connector ID: %w", err)
	}

	return count, nil
}

// Insert reports whether a row was inserted. The partial unique index
// on connector_id arbitrates: a source referencing an already-taken
// connector is skipped, making creation idempotent per connector. CSV
// sources (nil connector) always insert.
func (as *AccessReviewSource) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO
    access_review_sources (
        id,
        tenant_id,
        organization_id,
        connector_id,
        name,
        csv_data,
        name_synced_at,
        name_sync_attempts,
        name_sync_next_attempt_at,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @connector_id,
    @name,
    @csv_data,
    @name_synced_at,
    @name_sync_attempts,
    @name_sync_next_attempt_at,
    @created_at,
    @updated_at
)
ON CONFLICT (connector_id) WHERE connector_id IS NOT NULL DO NOTHING
RETURNING id;
`

	args := pgx.StrictNamedArgs{
		"id":                        as.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           as.OrganizationID,
		"connector_id":              as.ConnectorID,
		"name":                      as.Name,
		"csv_data":                  as.CsvData,
		"name_synced_at":            as.NameSyncedAt,
		"name_sync_attempts":        as.NameSyncAttempts,
		"name_sync_next_attempt_at": as.NameSyncNextAttemptAt,
		"created_at":                as.CreatedAt,
		"updated_at":                as.UpdatedAt,
	}

	var insertedID gid.GID

	err := conn.QueryRow(ctx, q, args).Scan(&insertedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("cannot insert access_source: %w", err)
	}

	return true, nil
}

func (as *AccessReviewSource) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE access_review_sources
SET
    name = @name,
    connector_id = @connector_id,
    csv_data = @csv_data,
    name_synced_at = @name_synced_at,
    name_sync_attempts = @name_sync_attempts,
    name_sync_next_attempt_at = @name_sync_next_attempt_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                        as.ID,
		"name":                      as.Name,
		"connector_id":              as.ConnectorID,
		"csv_data":                  as.CsvData,
		"name_synced_at":            as.NameSyncedAt,
		"name_sync_attempts":        as.NameSyncAttempts,
		"name_sync_next_attempt_at": as.NameSyncNextAttemptAt,
		"updated_at":                as.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update access_source: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// DeleteReturningConnectorID deletes the source and returns the
// connector it referenced at delete time, read under the DELETE's own
// row lock. Nil for CSV sources; ErrResourceNotFound when the source
// does not exist.
func (as *AccessReviewSource) DeleteReturningConnectorID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (*gid.GID, error) {
	q := `
DELETE FROM access_review_sources
WHERE %s AND id = @id
RETURNING connector_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": as.ID}
	maps.Copy(args, scope.SQLArguments())

	var connectorID *gid.GID

	err := conn.QueryRow(ctx, q, args).Scan(&connectorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot delete access_source: %w", err)
	}

	return connectorID, nil
}

// LoadByConnectorID loads the access source referencing the connector.
// Returns ErrResourceNotFound when none references it.
func (as *AccessReviewSource) LoadByConnectorID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    csv_data,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
FROM
    access_review_sources
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
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	source, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access source: %w", err)
	}

	*as = source

	return nil
}

func (sources *AccessReviewSources) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AccessReviewSourceOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    csv_data,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
FROM
    access_review_sources
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
		return fmt.Errorf("cannot query access_review_sources: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewSource])
	if err != nil {
		return fmt.Errorf("cannot collect access_review_sources: %w", err)
	}

	*sources = result

	return nil
}

func (sources *AccessReviewSources) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_review_sources
WHERE
    %s
    AND organization_id = @organization_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_review_sources: %w", err)
	}

	return count, nil
}

// RecordNameSyncAttempt charges the claimed source for one name-resolution
// attempt and holds it out of the queue for backoff. It must run in the
// transaction that claimed the row: the claim predicate does not otherwise
// change on failure, so an uncharged attempt is re-claimed immediately.
//
// The deadline is computed from the database clock, which is the clock the
// claim predicate compares against.
func (as *AccessReviewSource) RecordNameSyncAttempt(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	backoff time.Duration,
) error {
	q := `
UPDATE access_review_sources
SET
    name_sync_attempts = name_sync_attempts + 1,
    name_sync_next_attempt_at = NOW() + @backoff::interval
WHERE
    %s
    AND id = @id
RETURNING name_sync_attempts, name_sync_next_attempt_at
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":      as.ID,
		"backoff": backoff,
	}
	maps.Copy(args, scope.SQLArguments())

	err := conn.QueryRow(ctx, q, args).Scan(&as.NameSyncAttempts, &as.NameSyncNextAttemptAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot record access source name sync attempt: %w", err)
	}

	return nil
}

// MarkNameSynced retires the source from the name-sync queue, writing the
// resolved name alongside. It touches only the name-sync columns and applies
// only while the row still carries the deadline and attempt count the caller
// claimed: a reconnect or a relink that lands during the resolve wins, and
// this reports no error so the worker treats the task as done.
//
// The deadline is what makes the guard a claim generation. The attempt count
// alone is not unique — a reconnect resets it to zero and the next claim
// charges it straight back to the same number, so a resolve still in flight
// from before the reconnect would match again and overwrite the fresh claim
// with a stale name. RecordNameSyncAttempt stamps a new deadline from the
// database clock on every claim, so no two claims of a row carry the same one.
func (as *AccessReviewSource) MarkNameSynced(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	syncedAt time.Time,
) error {
	q := `
UPDATE access_review_sources
SET
    name = @name,
    name_synced_at = @name_synced_at,
    name_sync_next_attempt_at = NULL,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
    AND name_synced_at IS NULL
    AND name_sync_attempts = @name_sync_attempts
    AND name_sync_next_attempt_at IS NOT DISTINCT FROM @name_sync_next_attempt_at
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                        as.ID,
		"name":                      as.Name,
		"name_synced_at":            syncedAt,
		"name_sync_attempts":        as.NameSyncAttempts,
		"name_sync_next_attempt_at": as.NameSyncNextAttemptAt,
		"updated_at":                syncedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot mark access source name synced: %w", err)
	}

	as.NameSyncedAt = &syncedAt
	as.NameSyncNextAttemptAt = nil
	as.UpdatedAt = syncedAt

	return nil
}

// ResetNameSync puts the source back at the head of the name-sync queue with a
// clean retry budget. updated_at is left to the caller, which owns the rest of
// the mutation.
func (as *AccessReviewSource) ResetNameSync() {
	as.NameSyncedAt = nil
	as.NameSyncAttempts = 0
	as.NameSyncNextAttemptAt = nil
}

// ResetNameSyncByConnectorID clears the synced name and the retry budget for
// every access source backed by connectorID, so the source-name worker
// re-resolves their display name. No-op when no source references it.
//
// Rows already at NULL are included: a source still inside its backoff, or one
// that exhausted its attempts against a misconfigured connector, is exactly
// what the reconnect that fixed it needs to release.
func (sources *AccessReviewSources) ResetNameSyncByConnectorID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	connectorID gid.GID,
) error {
	q := `
UPDATE access_review_sources
SET
    name_synced_at = NULL,
    name_sync_attempts = 0,
    name_sync_next_attempt_at = NULL,
    updated_at = @updated_at
WHERE
    %s
    AND connector_id = @connector_id
    AND (name_synced_at IS NOT NULL OR name_sync_attempts > 0)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": connectorID,
		"updated_at":   time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot clear access source name sync: %w", err)
	}

	return nil
}

// ErrNoAccessReviewSourceNameSyncAvailable is returned when no access source
// needs its name synced from its connector.
var ErrNoAccessReviewSourceNameSyncAvailable = fmt.Errorf("no access source name sync available")

// LoadNextUnsyncedNameForUpdateSkipLocked claims the next access source that
// has a connector, has not yet had its name synced, and is not backing off
// from a previous failure. The row is locked with FOR UPDATE SKIP LOCKED so
// concurrent workers do not pick the same row.
//
// The query is intentionally cross-tenant: the worker serves every tenant and
// re-scopes from the claimed row's own GID.
//
// The backoff gate is what bounds the claim. Nothing else transitions the
// predicate on failure -- name_synced_at is untouched and the row lock ends
// with this transaction -- so without it the same row is re-claimed as fast as
// it can be processed, and being ordered by created_at it also blocks every
// source behind it. Callers must charge the attempt with RecordNameSyncAttempt
// in this same transaction.
func (as *AccessReviewSource) LoadNextUnsyncedNameForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    name,
    csv_data,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
FROM
    access_review_sources
WHERE
    connector_id IS NOT NULL
    AND name_synced_at IS NULL
    AND (name_sync_next_attempt_at IS NULL OR name_sync_next_attempt_at <= NOW())
ORDER BY
    created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query unsynced access_review_sources: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoAccessReviewSourceNameSyncAvailable
		}

		return fmt.Errorf("cannot collect unsynced access source: %w", err)
	}

	*as = row

	return nil
}
