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
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

const TaskSyncJobDefaultMaxAttempts = 5

type (
	TaskSyncJob struct {
		ID                   gid.GID              `db:"id"`
		OrganizationID       gid.GID              `db:"organization_id"`
		Direction            TaskSyncJobDirection `db:"direction"`
		Status               TaskSyncJobStatus    `db:"status"`
		Payload              json.RawMessage      `db:"payload"`
		Error                *string              `db:"error"`
		CreatedAt            time.Time            `db:"created_at"`
		UpdatedAt            time.Time            `db:"updated_at"`
		StartedAt            *time.Time           `db:"started_at"`
		CompletedAt          *time.Time           `db:"completed_at"`
		AttemptCount         int                  `db:"attempt_count"`
		NextAttemptAt        *time.Time           `db:"next_attempt_at"`
		ProcessingOwnerToken *string              `db:"processing_owner_token"`
	}

	TaskSyncJobs []*TaskSyncJob
)

func (j *TaskSyncJob) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO task_sync_jobs (
    id,
    tenant_id,
    organization_id,
    direction,
    status,
    payload,
    error,
    created_at,
    updated_at,
    started_at,
    completed_at,
    attempt_count,
    next_attempt_at,
    processing_owner_token
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @direction,
    @status,
    @payload,
    @error,
    @created_at,
    @updated_at,
    @started_at,
    @completed_at,
    @attempt_count,
    @next_attempt_at,
    @processing_owner_token
)
`

	args := pgx.StrictNamedArgs{
		"id":                     j.ID,
		"tenant_id":              scope.GetTenantID(),
		"organization_id":        j.OrganizationID,
		"direction":              j.Direction,
		"status":                 j.Status,
		"payload":                j.Payload,
		"error":                  j.Error,
		"created_at":             j.CreatedAt,
		"updated_at":             j.UpdatedAt,
		"started_at":             j.StartedAt,
		"completed_at":           j.CompletedAt,
		"attempt_count":          j.AttemptCount,
		"next_attempt_at":        j.NextAttemptAt,
		"processing_owner_token": j.ProcessingOwnerToken,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert task sync job: %w", err)
	}

	return nil
}

func (j *TaskSyncJob) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    direction,
    status,
    payload,
    error,
    created_at,
    updated_at,
    started_at,
    completed_at,
    attempt_count,
    next_attempt_at,
    processing_owner_token
FROM
    task_sync_jobs
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
		return fmt.Errorf("cannot query task sync job: %w", err)
	}

	job, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TaskSyncJob])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect task sync job: %w", err)
	}

	*j = job

	return nil
}

func (j *TaskSyncJob) ClaimNextForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
SELECT id, payload->>'task_id'
FROM task_sync_jobs
WHERE status = @pending_status
    AND direction = @direction
    AND processing_owner_token IS NULL
    AND (next_attempt_at IS NULL OR next_attempt_at <= @now)
    AND (
        payload->>'task_id' IS NULL
        OR NOT EXISTS (
            SELECT 1
            FROM task_sync_jobs active
            WHERE active.direction = @direction
                AND active.status = @processing_status
                AND active.payload->>'task_id' = task_sync_jobs.payload->>'task_id'
        )
    )
ORDER BY created_at ASC, id ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	args := pgx.StrictNamedArgs{
		"pending_status":    TaskSyncJobStatusPending,
		"processing_status": TaskSyncJobStatusProcessing,
		"direction":         TaskSyncJobDirectionOutbound,
		"now":               now,
	}

	var (
		id     gid.GID
		taskID *string
	)

	if err := conn.QueryRow(ctx, q, args).Scan(&id, &taskID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot select task sync job: %w", err)
	}

	if taskID != nil && *taskID != "" {
		if err := lockOutboundTaskSyncJob(ctx, conn, *taskID); err != nil {
			return err
		}

		busy, err := hasProcessingOutboundTaskSyncJob(ctx, conn, *taskID)
		if err != nil {
			return err
		}

		if busy {
			return ErrResourceNotFound
		}
	}

	return j.claimSelected(ctx, conn, id, now)
}

func lockOutboundTaskSyncJob(
	ctx context.Context,
	conn pg.Querier,
	taskID string,
) error {
	q := `
SELECT pg_advisory_xact_lock(
    hashtext('task-sync-outbound:' || @task_id::text)
)
`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"task_id": taskID})
	if err != nil {
		return fmt.Errorf("cannot lock outbound task sync job: %w", err)
	}

	return nil
}

func hasProcessingOutboundTaskSyncJob(
	ctx context.Context,
	conn pg.Querier,
	taskID string,
) (bool, error) {
	q := `
SELECT EXISTS (
    SELECT 1
    FROM task_sync_jobs
    WHERE direction = @direction
        AND status = @processing_status
        AND payload->>'task_id' = @task_id
)
`

	var exists bool

	err := conn.QueryRow(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"direction":         TaskSyncJobDirectionOutbound,
			"processing_status": TaskSyncJobStatusProcessing,
			"task_id":           taskID,
		},
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("cannot check processing outbound task sync job: %w", err)
	}

	return exists, nil
}

func (j *TaskSyncJob) claimSelected(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
	now time.Time,
) error {
	q := `
UPDATE task_sync_jobs
SET
    status = @processing_status,
    started_at = @now,
    updated_at = @now,
    processing_owner_token = @owner_token,
    next_attempt_at = NULL
WHERE id = @id
    AND status = @pending_status
    AND processing_owner_token IS NULL
RETURNING
    id,
    organization_id,
    direction,
    status,
    payload,
    error,
    created_at,
    updated_at,
    started_at,
    completed_at,
    attempt_count,
    next_attempt_at,
    processing_owner_token
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":                id,
			"pending_status":    TaskSyncJobStatusPending,
			"processing_status": TaskSyncJobStatusProcessing,
			"now":               now,
			"owner_token":       uuid.MustNewV4().String(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot claim task sync job: %w", err)
	}

	job, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TaskSyncJob])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect claimed task sync job: %w", err)
	}

	*j = job

	return nil
}

func (j *TaskSyncJob) UpdateProcessingState(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	if j.ProcessingOwnerToken == nil || *j.ProcessingOwnerToken == "" {
		panic("task sync job processing update requires an owner token")
	}

	q := `
UPDATE task_sync_jobs
SET
    status = @status,
    error = @error,
    updated_at = @updated_at,
    started_at = @started_at,
    completed_at = @completed_at,
    attempt_count = @attempt_count,
    next_attempt_at = @next_attempt_at,
    processing_owner_token = NULL
WHERE
    %s
    AND id = @id
    AND processing_owner_token = @processing_owner_token
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                     j.ID,
		"status":                 j.Status,
		"error":                  j.Error,
		"updated_at":             j.UpdatedAt,
		"started_at":             j.StartedAt,
		"completed_at":           j.CompletedAt,
		"attempt_count":          j.AttemptCount,
		"next_attempt_at":        j.NextAttemptAt,
		"processing_owner_token": j.ProcessingOwnerToken,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update task sync job processing state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProcessingLeaseLost
	}

	return nil
}

func (j *TaskSyncJob) TouchLease(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	now time.Time,
) error {
	if j.ProcessingOwnerToken == nil || *j.ProcessingOwnerToken == "" {
		panic("task sync job lease touch requires an owner token")
	}

	q := `
UPDATE task_sync_jobs
SET
    started_at = @now,
    updated_at = @now
WHERE
    %s
    AND id = @id
    AND processing_owner_token = @processing_owner_token
    AND status = @processing_status
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                     j.ID,
		"now":                    now,
		"processing_owner_token": j.ProcessingOwnerToken,
		"processing_status":      TaskSyncJobStatusProcessing,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot touch task sync job lease: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProcessingLeaseLost
	}

	j.StartedAt = &now
	j.UpdatedAt = now

	return nil
}

func CompletePendingOutboundJobsForTask(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
	now time.Time,
) error {
	q := `
UPDATE task_sync_jobs
SET
    status = @succeeded_status,
    completed_at = @now,
    updated_at = @now,
    error = NULL
WHERE
    %s
    AND direction = @direction
    AND status = @pending_status
    AND processing_owner_token IS NULL
    AND payload->>'task_id' = @task_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"succeeded_status": TaskSyncJobStatusSucceeded,
		"pending_status":   TaskSyncJobStatusPending,
		"direction":        TaskSyncJobDirectionOutbound,
		"now":              now,
		"task_id":          taskID.String(),
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot complete pending outbound task sync jobs: %w", err)
	}

	return nil
}

func (j *TaskSyncJob) HasNewerOutboundJob(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
) (bool, error) {
	q := `
SELECT EXISTS (
    SELECT 1
    FROM task_sync_jobs
    WHERE
        %s
        AND direction = @direction
        AND status IN (@pending_status, @processing_status, @succeeded_status)
        AND payload->>'task_id' = @task_id
        AND (created_at, id) > (@created_at, @id)
)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"direction":         TaskSyncJobDirectionOutbound,
		"pending_status":    TaskSyncJobStatusPending,
		"processing_status": TaskSyncJobStatusProcessing,
		"succeeded_status":  TaskSyncJobStatusSucceeded,
		"task_id":           taskID.String(),
		"created_at":        j.CreatedAt,
		"id":                j.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	var exists bool

	if err := conn.QueryRow(ctx, q, args).Scan(&exists); err != nil {
		return false, fmt.Errorf("cannot check newer outbound task sync jobs: %w", err)
	}

	return exists, nil
}

func ResetStaleTaskSyncJobs(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
) error {
	now := time.Now()

	q := `
UPDATE task_sync_jobs
SET
    attempt_count = attempt_count + 1,
    status = CASE
        WHEN attempt_count + 1 >= @max_attempts THEN @failed_status
        ELSE @pending_status
    END,
    error = @stale_error,
    updated_at = @now,
    started_at = NULL,
    processing_owner_token = NULL,
    completed_at = CASE
        WHEN attempt_count + 1 >= @max_attempts THEN @now::timestamptz
        ELSE NULL
    END,
    next_attempt_at = CASE
        WHEN attempt_count + 1 >= @max_attempts THEN NULL
        ELSE @now::timestamptz
    END
WHERE
    status = @processing_status
    AND started_at IS NOT NULL
    AND started_at < @stale_threshold
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"pending_status":    TaskSyncJobStatusPending,
			"processing_status": TaskSyncJobStatusProcessing,
			"failed_status":     TaskSyncJobStatusFailed,
			"stale_error":       "task sync job processing lease expired",
			"now":               now,
			"stale_threshold":   now.Add(-staleAfter),
			"max_attempts":      TaskSyncJobDefaultMaxAttempts,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot reset stale task sync jobs: %w", err)
	}

	return nil
}

func DeleteSucceededTaskSyncJobsBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	return deleteTaskSyncJobsByStatusBeforeBatch(
		ctx,
		conn,
		TaskSyncJobStatusSucceeded,
		before,
		limit,
	)
}

func DeleteFailedTaskSyncJobsBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	return deleteTaskSyncJobsByStatusBeforeBatch(
		ctx,
		conn,
		TaskSyncJobStatusFailed,
		before,
		limit,
	)
}

func deleteTaskSyncJobsByStatusBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	status TaskSyncJobStatus,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
    SELECT id
    FROM task_sync_jobs
    WHERE status = @status
        AND COALESCE(completed_at, updated_at) < @before
    ORDER BY COALESCE(completed_at, updated_at) ASC, id ASC
    LIMIT @limit
)
DELETE FROM task_sync_jobs
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"status": status,
			"before": before,
			"limit":  limit,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete %s task sync jobs: %w", status, err)
	}

	return result.RowsAffected(), nil
}
