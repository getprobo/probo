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
	TaskActivity struct {
		ID             gid.GID            `db:"id"`
		OrganizationID gid.GID            `db:"organization_id"`
		TaskID         gid.GID            `db:"task_id"`
		ActorID        *gid.GID           `db:"actor_profile_id"`
		ActivityType   TaskActivityType   `db:"activity_type"`
		Field          *TaskActivityField `db:"field"`
		OldValue       *string            `db:"old_value"`
		NewValue       *string            `db:"new_value"`
		CreatedAt      time.Time          `db:"created_at"`
	}

	TaskActivities []*TaskActivity
)

func (ta TaskActivity) CursorKey(orderBy TaskActivityOrderField) page.CursorKey {
	switch orderBy {
	case TaskActivityOrderFieldCreatedAt:
		return page.NewCursorKey(ta.ID, ta.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (ta *TaskActivity) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM task_activities WHERE id = ANY(@resource_ids)`

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

func (ta *TaskActivity) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskActivityID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    task_id,
    actor_profile_id,
    activity_type,
    field,
    old_value,
    new_value,
    created_at
FROM
    task_activities
WHERE
    %s
    AND id = @task_activity_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_activity_id": taskActivityID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task activities: %w", err)
	}

	taskActivity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TaskActivity])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect task activities: %w", err)
	}

	*ta = taskActivity

	return nil
}

func (ta *TaskActivities) LoadByTaskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
	cursor *page.Cursor[TaskActivityOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    task_id,
    actor_profile_id,
    activity_type,
    field,
    old_value,
    new_value,
    created_at
FROM
    task_activities
WHERE
    %s
    AND task_id = @task_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"task_id": taskID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task activities: %w", err)
	}

	taskActivities, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TaskActivity])
	if err != nil {
		return fmt.Errorf("cannot collect task activities: %w", err)
	}

	*ta = taskActivities

	return nil
}

func (ta *TaskActivities) CountByTaskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    task_activities
WHERE
    %s
    AND task_id = @task_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count task activities: %w", err)
	}

	return count, nil
}

func (ta TaskActivity) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    task_activities (
        tenant_id,
        id,
        organization_id,
        task_id,
        actor_profile_id,
        activity_type,
        field,
        old_value,
        new_value,
        created_at
    )
VALUES (
    @tenant_id,
    @task_activity_id,
    @organization_id,
    @task_id,
    @actor_profile_id,
    @activity_type,
    @field,
    @old_value,
    @new_value,
    @created_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":        scope.GetTenantID(),
		"task_activity_id": ta.ID,
		"organization_id":  ta.OrganizationID,
		"task_id":          ta.TaskID,
		"actor_profile_id": ta.ActorID,
		"activity_type":    ta.ActivityType,
		"field":            ta.Field,
		"old_value":        ta.OldValue,
		"new_value":        ta.NewValue,
		"created_at":       ta.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert task activity: %w", err)
	}

	return nil
}
