// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"

	"go.gearno.de/kit/pg"
)

type (
	Task struct {
		ID             gid.GID        `db:"id"`
		OrganizationID gid.GID        `db:"organization_id"`
		MeasureID      *gid.GID       `db:"measure_id"`
		Name           string         `db:"name"`
		Description    *string        `db:"description"`
		State          TaskState      `db:"state"`
		ReferenceID    string         `db:"reference_id"`
		TimeEstimate   *time.Duration `db:"time_estimate"`
		AssignedToID   *gid.GID       `db:"assigned_to_profile_id"`
		Deadline       *time.Time     `db:"deadline"`
		CreatedAt      time.Time      `db:"created_at"`
		UpdatedAt      time.Time      `db:"updated_at"`
	}

	Tasks []*Task
)

func (c Task) CursorKey(orderBy TaskOrderField) page.CursorKey {
	switch orderBy {
	case TaskOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (t *Task) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM tasks WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, t.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query task authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (c *Task) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	taskID gid.GID,
) error {
	q := `
SELECT
    id,
	organization_id,
    measure_id,
    name,
    description,
    state,
    reference_id,
    time_estimate,
    assigned_to_profile_id,
    deadline,
    created_at,
    updated_at
FROM
    tasks
WHERE
    %s
    AND id = @task_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tasks: %w", err)
	}

	task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Task])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect tasks: %w", err)
	}

	*c = task

	return nil
}

func (c Task) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    tasks (
        tenant_id,
        id,
		organization_id,
        measure_id,
        name,
        description,
        reference_id,
        state,
        time_estimate,
        assigned_to_profile_id,
        deadline,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @task_id,
	@organization_id,
    @measure_id,
    @name,
    @description,
    @reference_id,
    @state,
    @time_estimate,
    @assigned_to_profile_id,
    @deadline,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"task_id":                c.ID,
		"organization_id":        c.OrganizationID,
		"measure_id":             c.MeasureID,
		"name":                   c.Name,
		"description":            c.Description,
		"reference_id":           c.ReferenceID,
		"state":                  c.State,
		"time_estimate":          c.TimeEstimate,
		"assigned_to_profile_id": c.AssignedToID,
		"deadline":               c.Deadline,
		"created_at":             c.CreatedAt,
		"updated_at":             c.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "tasks_reference_id_unique" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert task: %w", err)
	}

	return nil
}

func (c *Task) Upsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    tasks (
        tenant_id,
        id,
		organization_id,
        measure_id,
        name,
        description,
        reference_id,
        state,
        time_estimate,
        assigned_to_profile_id,
        deadline,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @task_id,
	@organization_id,
    @measure_id,
    @name,
    @description,
    @reference_id,
    @state,
    @time_estimate,
    @assigned_to_profile_id,
    @deadline,
    @created_at,
    @updated_at
)
ON CONFLICT (measure_id, reference_id) DO UPDATE SET
    name = @name,
    description = @description,
    updated_at = @updated_at,
    deadline = @deadline
RETURNING
    id,
    organization_id,
    measure_id,
    name,
    description,
    reference_id,
    state,
    time_estimate,
    assigned_to_profile_id,
    deadline,
    created_at,
    updated_at
`

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"task_id":                c.ID,
		"organization_id":        c.OrganizationID,
		"measure_id":             c.MeasureID,
		"name":                   c.Name,
		"description":            c.Description,
		"reference_id":           c.ReferenceID,
		"state":                  c.State,
		"time_estimate":          c.TimeEstimate,
		"assigned_to_profile_id": c.AssignedToID,
		"deadline":               c.Deadline,
		"created_at":             c.CreatedAt,
		"updated_at":             c.UpdatedAt,
	}
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert task: %w", err)
	}

	task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Task])
	if err != nil {
		return fmt.Errorf("cannot collect tasks: %w", err)
	}

	*c = task

	return nil
}

func (c *Tasks) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
	SELECT
		COUNT(id)
	FROM
		tasks
	WHERE
		%s
		AND organization_id = @organization_id
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot collect tasks: %w", err)
	}

	return count, nil
}

func (c *Tasks) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[TaskOrderField],
) error {
	q := `
	SELECT
		id,
		measure_id,
		organization_id,
		name,
		description,
		state,
		reference_id,
		time_estimate,
		assigned_to_profile_id,
		deadline,
		created_at,
		updated_at
	FROM
		tasks
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
		return fmt.Errorf("cannot query tasks: %w", err)
	}

	tasks, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Task])
	if err != nil {
		return fmt.Errorf("cannot collect tasks: %w", err)
	}

	*c = tasks

	return nil
}

func (c *Tasks) CountByMeasureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	measureID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    tasks
WHERE
    %s
    AND measure_id = @measure_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot collect tasks: %w", err)
	}

	return count, nil
}

func (c *Tasks) LoadByMeasureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	measureID gid.GID,
	cursor *page.Cursor[TaskOrderField],
) error {
	q := `
SELECT
    id,
    measure_id,
	organization_id,
    name,
    description,
    state,
    reference_id,
    time_estimate,
    assigned_to_profile_id,
    deadline,
    created_at,
    updated_at
FROM
    tasks
WHERE
    %s
    AND measure_id = @measure_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tasks: %w", err)
	}

	tasks, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Task])
	if err != nil {
		return fmt.Errorf("cannot collect tasks: %w", err)
	}

	*c = tasks

	return nil
}

func (c *Task) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE tasks
SET
  name = @name,
  description = @description,
  state = @state,
  time_estimate = @time_estimate,
  updated_at = @updated_at,
  assigned_to_profile_id = @assigned_to_profile_id,
  deadline = @deadline
WHERE %s
    AND id = @task_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"task_id":                c.ID,
		"name":                   c.Name,
		"description":            c.Description,
		"state":                  c.State,
		"time_estimate":          c.TimeEstimate,
		"updated_at":             c.UpdatedAt,
		"assigned_to_profile_id": c.AssignedToID,
		"deadline":               c.Deadline,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (c *Task) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM tasks
WHERE %s
    AND id = @task_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"task_id": c.ID,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete task: %w", err)
	}

	return nil
}
