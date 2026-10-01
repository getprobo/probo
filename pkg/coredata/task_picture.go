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
	TaskPicture struct {
		ID                 gid.GID   `db:"id"`
		OrganizationID     gid.GID   `db:"organization_id"`
		TaskID             gid.GID   `db:"task_id"`
		FileID             gid.GID   `db:"file_id"`
		LinearAssetURL     *string   `db:"linear_asset_url"`
		LinearAttachmentID *string   `db:"linear_attachment_id"`
		CreatedAt          time.Time `db:"created_at"`
		UpdatedAt          time.Time `db:"updated_at"`
	}

	TaskPictures []*TaskPicture
)

func (p TaskPicture) CursorKey(orderBy TaskPictureOrderField) page.CursorKey {
	switch orderBy {
	case TaskPictureOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p *TaskPicture) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    organization_id
FROM
    task_pictures
WHERE
    id = ANY(@resource_ids)
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))

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

func (p *TaskPicture) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskPictureID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    task_id,
    file_id,
    linear_asset_url,
    linear_attachment_id,
    created_at,
    updated_at
FROM
    task_pictures
WHERE
    %s
    AND id = @task_picture_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_picture_id": taskPictureID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task pictures: %w", err)
	}

	picture, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TaskPicture])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect task pictures: %w", err)
	}

	*p = picture

	return nil
}

func (ps *TaskPictures) LoadByTaskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
	cursor *page.Cursor[TaskPictureOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    task_id,
    file_id,
    linear_asset_url,
    linear_attachment_id,
    created_at,
    updated_at
FROM
    task_pictures
WHERE
    %s
    AND task_id = @task_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task pictures: %w", err)
	}

	pictures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TaskPicture])
	if err != nil {
		return fmt.Errorf("cannot collect task pictures: %w", err)
	}

	*ps = pictures

	return nil
}

func (ps *TaskPictures) LoadUnsyncedByTaskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
	limit int,
) error {
	q := `
SELECT
    id,
    organization_id,
    task_id,
    file_id,
    linear_asset_url,
    linear_attachment_id,
    created_at,
    updated_at
FROM
    task_pictures
WHERE
    %s
    AND task_id = @task_id
    AND linear_attachment_id IS NULL
ORDER BY
    created_at ASC,
    id ASC
LIMIT %d
`

	q = fmt.Sprintf(q, scope.SQLFragment(), limit)

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query unsynced task pictures: %w", err)
	}

	pictures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TaskPicture])
	if err != nil {
		return fmt.Errorf("cannot collect unsynced task pictures: %w", err)
	}

	*ps = pictures

	return nil
}

func (ps *TaskPictures) CountByTaskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    task_pictures
WHERE
    %s
    AND task_id = @task_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())

	var count int

	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count task pictures: %w", err)
	}

	return count, nil
}

func (p TaskPicture) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    task_pictures (
        tenant_id,
        id,
        organization_id,
        task_id,
        file_id,
        linear_asset_url,
        linear_attachment_id,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @task_picture_id,
    @organization_id,
    @task_id,
    @file_id,
    @linear_asset_url,
    @linear_attachment_id,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":            scope.GetTenantID(),
		"task_picture_id":      p.ID,
		"organization_id":      p.OrganizationID,
		"task_id":              p.TaskID,
		"file_id":              p.FileID,
		"linear_asset_url":     p.LinearAssetURL,
		"linear_attachment_id": p.LinearAttachmentID,
		"created_at":           p.CreatedAt,
		"updated_at":           p.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert task picture: %w", err)
	}

	return nil
}

func (p TaskPicture) UpdateLinear(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
    task_pictures
SET
    linear_asset_url = @linear_asset_url,
    linear_attachment_id = @linear_attachment_id,
    updated_at = @updated_at
WHERE
    %s
    AND id = @task_picture_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"task_picture_id":      p.ID,
		"linear_asset_url":     p.LinearAssetURL,
		"linear_attachment_id": p.LinearAttachmentID,
		"updated_at":           p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update task picture: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (p TaskPicture) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
    task_pictures
WHERE
    %s
    AND id = @task_picture_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_picture_id": p.ID}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete task picture: %w", err)
	}

	return nil
}
