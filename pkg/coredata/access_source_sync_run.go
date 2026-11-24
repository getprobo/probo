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
	// AccessSourceSyncRun is one inventory sync for a live AccessReviewSource.
	// Each enqueue produces a new row. Terminal rows (SUCCESS / FAILED) are
	// immutable; only the in-flight run is updated.
	//
	// TenantID is retained on the struct because the background worker claims
	// rows cross-tenant via LoadNextQueuedForUpdateSkipLocked and needs the
	// tenant to construct a Scope for subsequent operations.
	AccessSourceSyncRun struct {
		ID                   gid.GID                   `db:"id"`
		OrganizationID       gid.GID                   `db:"organization_id"`
		TenantID             gid.TenantID              `db:"tenant_id"`
		AccessReviewSourceID gid.GID                   `db:"access_review_source_id"`
		Status               AccessSourceSyncRunStatus `db:"status"`
		GrantsUpsertedCount  int                       `db:"grants_upserted_count"`
		Error                *string                   `db:"error"`
		StartedAt            *time.Time                `db:"started_at"`
		CompletedAt          *time.Time                `db:"completed_at"`
		CreatedAt            time.Time                 `db:"created_at"`
		UpdatedAt            time.Time                 `db:"updated_at"`
	}

	AccessSourceSyncRuns []*AccessSourceSyncRun
)

var (
	ErrNoAccessSourceSyncRunAvailable = errors.New("no access source sync run available")
)

func (r AccessSourceSyncRun) CursorKey(orderBy AccessSourceSyncRunOrderField) page.CursorKey {
	switch orderBy {
	case AccessSourceSyncRunOrderFieldCreatedAt:
		return page.NewCursorKey(r.ID, r.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (r *AccessSourceSyncRun) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_source_sync_runs WHERE id = ANY(@resource_ids::text[])`

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

func (r *AccessSourceSyncRun) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    tenant_id,
    access_review_source_id,
    status,
    grants_upserted_count,
    error,
    started_at,
    completed_at,
    created_at,
    updated_at
FROM
    access_source_sync_runs
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
		return fmt.Errorf("cannot query access_source_sync_runs: %w", err)
	}

	run, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessSourceSyncRun])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect access source sync run: %w", err)
	}

	*r = run

	return nil
}

func (r *AccessSourceSyncRun) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO access_source_sync_runs (
    id,
    tenant_id,
    organization_id,
    access_review_source_id,
    status,
    grants_upserted_count,
    error,
    started_at,
    completed_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @access_review_source_id,
    @status,
    @grants_upserted_count,
    @error,
    @started_at,
    @completed_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                      r.ID,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         r.OrganizationID,
		"access_review_source_id": r.AccessReviewSourceID,
		"status":                  r.Status,
		"grants_upserted_count":   r.GrantsUpsertedCount,
		"error":                   r.Error,
		"started_at":              r.StartedAt,
		"completed_at":            r.CompletedAt,
		"created_at":              r.CreatedAt,
		"updated_at":              r.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert access source sync run: %w", err)
	}

	r.TenantID = scope.GetTenantID()

	return nil
}

// Update writes the in-flight run's lifecycle fields. It must only be called
// on the run that the worker currently owns.
func (r *AccessSourceSyncRun) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE access_source_sync_runs
SET
    status = @status,
    grants_upserted_count = @grants_upserted_count,
    error = @error,
    started_at = @started_at,
    completed_at = @completed_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                    r.ID,
		"status":                r.Status,
		"grants_upserted_count": r.GrantsUpsertedCount,
		"error":                 r.Error,
		"started_at":            r.StartedAt,
		"completed_at":          r.CompletedAt,
		"updated_at":            r.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update access source sync run: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// LoadNextQueuedForUpdateSkipLocked is intentionally cross-tenant: the
// background worker claims the next available sync run regardless of tenant.
// The caller extracts TenantID from the returned struct to construct a Scope
// for subsequent operations.
func (r *AccessSourceSyncRun) LoadNextQueuedForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
SELECT
    id,
    organization_id,
    tenant_id,
    access_review_source_id,
    status,
    grants_upserted_count,
    error,
    started_at,
    completed_at,
    created_at,
    updated_at
FROM access_source_sync_runs
WHERE status = @status
ORDER BY created_at ASC, id ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`
	args := pgx.StrictNamedArgs{
		"status": AccessSourceSyncRunStatusQueued,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query next queued access source sync run: %w", err)
	}

	run, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessSourceSyncRun])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoAccessSourceSyncRunAvailable
		}

		return fmt.Errorf("cannot collect access source sync run: %w", err)
	}

	*r = run

	return nil
}

// LoadStaleRunningForUpdateSkipLocked locks RUNNING rows whose updated_at is
// older than staleThreshold, skipping rows locked elsewhere. Cross-tenant so
// a recovery worker can sweep all orgs. Limit caps the batch size.
func (rs *AccessSourceSyncRuns) LoadStaleRunningForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	staleThreshold time.Time,
	limit int,
) error {
	q := `
SELECT
    id,
    organization_id,
    tenant_id,
    access_review_source_id,
    status,
    grants_upserted_count,
    error,
    started_at,
    completed_at,
    created_at,
    updated_at
FROM access_source_sync_runs
WHERE
    status = @running_status
    AND updated_at < @stale_threshold
ORDER BY updated_at ASC, id ASC
LIMIT @limit
FOR UPDATE SKIP LOCKED
`
	args := pgx.StrictNamedArgs{
		"running_status":  AccessSourceSyncRunStatusRunning,
		"stale_threshold": staleThreshold,
		"limit":           limit,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query stale access source sync runs: %w", err)
	}

	runs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessSourceSyncRun])
	if err != nil {
		return fmt.Errorf("cannot collect stale access source sync runs: %w", err)
	}

	*rs = runs

	return nil
}

func (rs *AccessSourceSyncRuns) LoadBySourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
	cursor *page.Cursor[AccessSourceSyncRunOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    tenant_id,
    access_review_source_id,
    status,
    grants_upserted_count,
    error,
    started_at,
    completed_at,
    created_at,
    updated_at
FROM
    access_source_sync_runs
WHERE
    %s
    AND access_review_source_id = @access_review_source_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_source_id": sourceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_source_sync_runs: %w", err)
	}

	runs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessSourceSyncRun])
	if err != nil {
		return fmt.Errorf("cannot collect access_source_sync_runs: %w", err)
	}

	*rs = runs

	return nil
}
