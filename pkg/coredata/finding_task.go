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

type FindingTask struct {
	FindingID      gid.GID   `db:"finding_id"`
	TaskID         gid.GID   `db:"task_id"`
	OrganizationID gid.GID   `db:"organization_id"`
	CreatedAt      time.Time `db:"created_at"`
}

func (ft *FindingTask) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    findings_tasks (
        finding_id,
        task_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @finding_id,
    @task_id,
    @organization_id,
    @tenant_id,
    @created_at
)
RETURNING
    finding_id,
    task_id,
    organization_id,
    created_at
`

	args := pgx.StrictNamedArgs{
		"finding_id":      ft.FindingID,
		"task_id":         ft.TaskID,
		"organization_id": ft.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      ft.CreatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "findings_tasks_pkey" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert finding task: %w", err)
	}

	findingTask, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[FindingTask])
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "findings_tasks_pkey" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot collect finding task: %w", err)
	}

	*ft = findingTask

	return nil
}

func (ft FindingTask) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	findingID gid.GID,
	taskID gid.GID,
) error {
	q := `
DELETE FROM
    findings_tasks
WHERE
    %s
    AND finding_id = @finding_id
    AND task_id = @task_id
`

	args := pgx.StrictNamedArgs{
		"finding_id": findingID,
		"task_id":    taskID,
	}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete finding task: %w", err)
	}

	return nil
}
