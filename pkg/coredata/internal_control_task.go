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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	InternalControlTask struct {
		InternalControlID gid.GID   `db:"internal_control_id"`
		TaskID            gid.GID   `db:"task_id"`
		OrganizationID    gid.GID   `db:"organization_id"`
		ReferenceID       string    `db:"reference_id"`
		CreatedAt         time.Time `db:"created_at"`
	}

	InternalControlTasks []*InternalControlTask
)

func (mt InternalControlTask) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    internal_controls_tasks (
        internal_control_id,
        task_id,
        organization_id,
        tenant_id,
        reference_id,
        created_at
    )
VALUES (
    @internal_control_id,
    @task_id,
    @organization_id,
    @tenant_id,
    @reference_id,
    @created_at
)
ON CONFLICT (internal_control_id, task_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"internal_control_id": mt.InternalControlID,
		"task_id":             mt.TaskID,
		"organization_id":     mt.OrganizationID,
		"tenant_id":           scope.GetTenantID(),
		"reference_id":        mt.ReferenceID,
		"created_at":          mt.CreatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "internal_controls_tasks_internal_control_reference_id_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot upsert internal control task: %w", err)
	}

	return nil
}

func (ict InternalControlTasks) Merge(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
	organizationID gid.GID,
	referenceID string,
	internalControlIDs []gid.GID,
) error {
	q := `
WITH internal_control_ids AS (
	SELECT DISTINCT
		unnest(@internal_control_ids::text[]) AS internal_control_id,
		@tenant_id AS tenant_id,
		@task_id AS task_id,
		@organization_id AS organization_id,
		@reference_id AS reference_id,
		@created_at::timestamptz AS created_at
)
MERGE INTO internal_controls_tasks AS tgt
USING internal_control_ids AS src
ON tgt.tenant_id = src.tenant_id
	AND tgt.task_id = src.task_id
	AND tgt.internal_control_id = src.internal_control_id
WHEN NOT MATCHED
	THEN INSERT (
		internal_control_id,
		task_id,
		organization_id,
		tenant_id,
		reference_id,
		created_at
	)
	VALUES (
		src.internal_control_id,
		src.task_id,
		src.organization_id,
		src.tenant_id,
		src.reference_id,
		src.created_at
	)
WHEN NOT MATCHED BY SOURCE
	AND tgt.tenant_id = @tenant_id
	AND tgt.task_id = @task_id
	THEN DELETE
`

	args := pgx.StrictNamedArgs{
		"tenant_id":            scope.GetTenantID(),
		"task_id":              taskID,
		"organization_id":      organizationID,
		"reference_id":         referenceID,
		"created_at":           time.Now(),
		"internal_control_ids": internalControlIDs,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "internal_controls_tasks_internal_control_reference_id_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot merge internal control tasks: %w", err)
	}

	return nil
}

func (mt InternalControlTask) CopyFromTask(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	fromTaskID gid.GID,
	toTaskID gid.GID,
	referenceID string,
	createdAt time.Time,
) error {
	q := `
INSERT INTO
    internal_controls_tasks (
        internal_control_id,
        task_id,
        organization_id,
        tenant_id,
        reference_id,
        created_at
    )
SELECT
    mt.internal_control_id,
    @to_task_id,
    mt.organization_id,
    mt.tenant_id,
    @reference_id,
    @created_at
FROM
    internal_controls_tasks mt
WHERE
    mt.tenant_id = @tenant_id
    AND mt.task_id = @from_task_id
ON CONFLICT (internal_control_id, task_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"from_task_id": fromTaskID,
		"to_task_id":   toTaskID,
		"reference_id": referenceID,
		"tenant_id":    scope.GetTenantID(),
		"created_at":   createdAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot copy internal control task mappings: %w", err)
	}

	return nil
}

func (mt InternalControlTask) LoadInternalControlIDsByTaskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
) ([]gid.GID, error) {
	idsByTaskID, err := mt.LoadInternalControlIDsByTaskIDs(ctx, conn, scope, []gid.GID{taskID})
	if err != nil {
		return nil, err
	}

	ids := idsByTaskID[taskID]
	if ids == nil {
		ids = []gid.GID{}
	}

	return ids, nil
}

func (mt InternalControlTask) LoadInternalControlIDsByTaskIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskIDs []gid.GID,
) (map[gid.GID][]gid.GID, error) {
	idsByTaskID := make(map[gid.GID][]gid.GID, len(taskIDs))
	if len(taskIDs) == 0 {
		return idsByTaskID, nil
	}

	q := `
SELECT
    task_id,
    internal_control_id
FROM
    internal_controls_tasks
WHERE
    %s
    AND task_id = ANY(@task_ids)
ORDER BY
    created_at ASC,
    internal_control_id ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_ids": taskIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query internal control task mappings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskID, internalControlID gid.GID
		if err := rows.Scan(&taskID, &internalControlID); err != nil {
			return nil, fmt.Errorf("cannot scan internal control task mapping: %w", err)
		}

		idsByTaskID[taskID] = append(idsByTaskID[taskID], internalControlID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate internal control task mappings: %w", err)
	}

	return idsByTaskID, nil
}
