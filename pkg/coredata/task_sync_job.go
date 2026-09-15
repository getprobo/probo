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

const TaskSyncJobDefaultMaxAttempts = 5

type (
	TaskSyncJob struct {
		ID             gid.GID              `db:"id"`
		OrganizationID gid.GID              `db:"organization_id"`
		Direction      TaskSyncJobDirection `db:"direction"`
		Status         TaskSyncJobStatus    `db:"status"`
		Payload        json.RawMessage      `db:"payload"`
		Error          *string              `db:"error"`
		CreatedAt      time.Time            `db:"created_at"`
		UpdatedAt      time.Time            `db:"updated_at"`
		StartedAt      *time.Time           `db:"started_at"`
		CompletedAt    *time.Time           `db:"completed_at"`
		AttemptCount   int                  `db:"attempt_count"`
		NextAttemptAt  *time.Time           `db:"next_attempt_at"`
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
    next_attempt_at
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
    @next_attempt_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              j.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": j.OrganizationID,
		"direction":       j.Direction,
		"status":          j.Status,
		"payload":         j.Payload,
		"error":           j.Error,
		"created_at":      j.CreatedAt,
		"updated_at":      j.UpdatedAt,
		"started_at":      j.StartedAt,
		"completed_at":    j.CompletedAt,
		"attempt_count":   j.AttemptCount,
		"next_attempt_at": j.NextAttemptAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert task sync job: %w", err)
	}

	return nil
}

func (j *TaskSyncJob) Update(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
UPDATE task_sync_jobs
SET
    status = @status,
    error = @error,
    updated_at = @updated_at,
    started_at = @started_at,
    completed_at = @completed_at,
    attempt_count = @attempt_count,
    next_attempt_at = @next_attempt_at
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":              j.ID,
		"status":          j.Status,
		"error":           j.Error,
		"updated_at":      j.UpdatedAt,
		"started_at":      j.StartedAt,
		"completed_at":    j.CompletedAt,
		"attempt_count":   j.AttemptCount,
		"next_attempt_at": j.NextAttemptAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update task sync job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (j *TaskSyncJob) LoadNextPendingForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
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
    next_attempt_at
FROM
    task_sync_jobs
WHERE
    status = @status
    AND direction = @direction
    AND (next_attempt_at IS NULL OR next_attempt_at <= @now)
ORDER BY
    created_at ASC,
    id ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"status":    TaskSyncJobStatusPending,
			"direction": TaskSyncJobDirectionOutbound,
			"now":       time.Now(),
		},
	)
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

func ResetStaleTaskSyncJobs(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
) error {
	now := time.Now()

	q := `
UPDATE task_sync_jobs
SET
    status = @pending_status,
    error = @stale_error,
    updated_at = @now,
    started_at = NULL
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
			"stale_error":       "task sync job processing lease expired",
			"now":               now,
			"stale_threshold":   now.Add(-staleAfter),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot reset stale task sync jobs: %w", err)
	}

	return nil
}
