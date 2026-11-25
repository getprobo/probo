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
	AccessSourceSyncRun struct {
		ID                   gid.GID                   `db:"id"`
		OrganizationID       gid.GID                   `db:"organization_id"`
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

	return nil
}

// Update writes lifecycle fields on a non-terminal run. Terminal rows
// (SUCCESS / FAILED) are immutable: the WHERE clause rejects them so a stale
// writer cannot overwrite a completed sync.
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
    AND status IN (@queued_status, @running_status)
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
		"queued_status":         AccessSourceSyncRunStatusQueued,
		"running_status":        AccessSourceSyncRunStatusRunning,
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

func (r *AccessSourceSyncRun) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM access_source_sync_runs
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": r.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete access source sync run: %w", err)
	}

	return nil
}

// LoadNextQueuedForUpdateSkipLocked is intentionally cross-tenant: the
// background worker claims the next available sync run regardless of tenant.
// The caller builds a Scope with NewScope(run.ID.TenantID(), ...) for later
// scoped writes.
//
// Only one run may be in flight per access_review_source_id: a queued row is
// skipped while another run for that source is RUNNING, or while an older
// QUEUED row for the same source still exists (so concurrent workers cannot
// claim two queued runs for one source).
func (r *AccessSourceSyncRun) LoadNextQueuedForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
SELECT
    id,
    organization_id,
    access_review_source_id,
    status,
    grants_upserted_count,
    error,
    started_at,
    completed_at,
    created_at,
    updated_at
FROM access_source_sync_runs candidate
WHERE
    candidate.status = @queued_status
    AND NOT EXISTS (
        SELECT 1
        FROM access_source_sync_runs other
        WHERE
            other.access_review_source_id = candidate.access_review_source_id
            AND (
                other.status = @running_status
                OR (
                    other.status = @queued_status
                    AND (other.created_at, other.id) < (candidate.created_at, candidate.id)
                )
            )
    )
ORDER BY candidate.created_at ASC, candidate.id ASC
LIMIT 1
FOR UPDATE OF candidate SKIP LOCKED
`
	args := pgx.StrictNamedArgs{
		"queued_status":  AccessSourceSyncRunStatusQueued,
		"running_status": AccessSourceSyncRunStatusRunning,
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

// LoadLatestBySourceID loads the most recent sync run for a source, if any.
func (r *AccessSourceSyncRun) LoadLatestBySourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
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
ORDER BY created_at DESC, id DESC
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_source_id": sourceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query latest access source sync run: %w", err)
	}

	run, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessSourceSyncRun])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect latest access source sync run: %w", err)
	}

	*r = run

	return nil
}

// LoadLatestBySourceIDs loads the most recent sync run for each source ID.
// Sources with no runs are absent from the result.
func (rs *AccessSourceSyncRuns) LoadLatestBySourceIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceIDs []gid.GID,
) error {
	if len(sourceIDs) == 0 {
		*rs = nil

		return nil
	}

	q := `
SELECT DISTINCT ON (access_review_source_id)
    id,
    organization_id,
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
    AND access_review_source_id = ANY(@access_review_source_ids::text[])
ORDER BY access_review_source_id, created_at DESC, id DESC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_source_ids": sourceIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query latest access source sync runs: %w", err)
	}

	runs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessSourceSyncRun])
	if err != nil {
		return fmt.Errorf("cannot collect latest access source sync runs: %w", err)
	}

	*rs = runs

	return nil
}

// LoadLatestSuccessfulBySourceID loads the most recent SUCCESS sync run for a
// source, if any. Used to detect that live inventory was already initialized
// even when a later run failed.
func (r *AccessSourceSyncRun) LoadLatestSuccessfulBySourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
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
    AND status = @success_status
ORDER BY created_at DESC, id DESC
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"access_review_source_id": sourceID,
		"success_status":          AccessSourceSyncRunStatusSuccess,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query latest successful access source sync run: %w", err)
	}

	run, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessSourceSyncRun])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect latest successful access source sync run: %w", err)
	}

	*r = run

	return nil
}

// LoadInFlightBySourceID loads the QUEUED or RUNNING run for a source, if any.
func (r *AccessSourceSyncRun) LoadInFlightBySourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
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
    AND status IN (@queued_status, @running_status)
ORDER BY created_at ASC, id ASC
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"access_review_source_id": sourceID,
		"queued_status":           AccessSourceSyncRunStatusQueued,
		"running_status":          AccessSourceSyncRunStatusRunning,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query in-flight access source sync run: %w", err)
	}

	run, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessSourceSyncRun])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect in-flight access source sync run: %w", err)
	}

	*r = run

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
