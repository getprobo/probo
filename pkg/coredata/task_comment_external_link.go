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
	"go.probo.inc/probo/pkg/page"
)

type (
	TaskCommentExternalLink struct {
		TaskCommentID   gid.GID           `db:"task_comment_id"`
		OrganizationID  gid.GID           `db:"organization_id"`
		TaskID          gid.GID           `db:"task_id"`
		ConnectorID     gid.GID           `db:"connector_id"`
		Provider        ConnectorProvider `db:"provider"`
		ExternalID      string            `db:"external_id"`
		RemoteUpdatedAt *time.Time        `db:"remote_updated_at"`
		ContentHash     *string           `db:"content_hash"`
		CreatedAt       time.Time         `db:"created_at"`
		UpdatedAt       time.Time         `db:"updated_at"`
	}

	TaskCommentExternalLinks []*TaskCommentExternalLink
)

func (l TaskCommentExternalLink) CursorKey(orderBy TaskCommentExternalLinkOrderField) page.CursorKey {
	switch orderBy {
	case TaskCommentExternalLinkOrderFieldTaskCommentID:
		return page.NewCursorKey(l.TaskCommentID, l.TaskCommentID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (l *TaskCommentExternalLink) LoadByCommentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	commentID gid.GID,
) error {
	q := `
SELECT
    task_comment_id,
    organization_id,
    task_id,
    connector_id,
    provider,
    external_id,
    remote_updated_at,
    content_hash,
    created_at,
    updated_at
FROM
    task_comment_external_links
WHERE
    %s
    AND task_comment_id = @task_comment_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_comment_id": commentID}
	maps.Copy(args, scope.SQLArguments())

	return l.loadExactlyOne(ctx, conn, q, args)
}

func (l *TaskCommentExternalLinks) LoadByTaskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
	cursor *page.Cursor[TaskCommentExternalLinkOrderField],
) error {
	q := `
SELECT
    task_comment_id,
    organization_id,
    task_id,
    connector_id,
    provider,
    external_id,
    remote_updated_at,
    content_hash,
    created_at,
    updated_at
FROM
    task_comment_external_links
WHERE
    %s
    AND task_id = @task_id
    AND (
        CAST(@cursor_id AS text) IS NULL
        OR task_comment_id >= @cursor_id
    )
ORDER BY
    task_comment_id ASC
LIMIT @cursor_limit
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	var cursorID *gid.GID
	if cursor.Key != nil {
		cursorID = &cursor.Key.ID
	}

	limit := cursor.Size
	if cursor.Key == nil {
		limit++
	} else {
		limit += 2
	}

	args := pgx.StrictNamedArgs{
		"task_id":      taskID,
		"cursor_id":    cursorID,
		"cursor_limit": limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task comment external links: %w", err)
	}

	links, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TaskCommentExternalLink])
	if err != nil {
		return fmt.Errorf("cannot collect task comment external links: %w", err)
	}

	*l = links

	return nil
}

func (l *TaskCommentExternalLinks) LoadByExternalID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	provider ConnectorProvider,
	externalID string,
) error {
	q := `
SELECT
    task_comment_id,
    organization_id,
    task_id,
    connector_id,
    provider,
    external_id,
    remote_updated_at,
    content_hash,
    created_at,
    updated_at
FROM
    task_comment_external_links
WHERE
    %s
    AND provider = @provider
    AND external_id = @external_id
ORDER BY
    created_at ASC,
    task_comment_id ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"provider":    provider,
		"external_id": externalID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task comment external links: %w", err)
	}

	links, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TaskCommentExternalLink])
	if err != nil {
		return fmt.Errorf("cannot collect task comment external links: %w", err)
	}

	*l = links

	return nil
}

func (l *TaskCommentExternalLink) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO task_comment_external_links (
    task_comment_id,
    tenant_id,
    organization_id,
    task_id,
    connector_id,
    provider,
    external_id,
    remote_updated_at,
    content_hash,
    created_at,
    updated_at
) VALUES (
    @task_comment_id,
    @tenant_id,
    @organization_id,
    @task_id,
    @connector_id,
    @provider,
    @external_id,
    @remote_updated_at,
    @content_hash,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"task_comment_id":   l.TaskCommentID,
		"tenant_id":         scope.GetTenantID(),
		"organization_id":   l.OrganizationID,
		"task_id":           l.TaskID,
		"connector_id":      l.ConnectorID,
		"provider":          l.Provider,
		"external_id":       l.ExternalID,
		"remote_updated_at": l.RemoteUpdatedAt,
		"content_hash":      l.ContentHash,
		"created_at":        l.CreatedAt,
		"updated_at":        l.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "task_comment_external_links_pkey",
					"task_comment_external_links_external_key":
					return ErrResourceAlreadyExists
				}
			}
		}

		return fmt.Errorf("cannot insert task comment external link: %w", err)
	}

	return nil
}

func (l *TaskCommentExternalLink) Update(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
UPDATE task_comment_external_links
SET
    external_id = @external_id,
    remote_updated_at = @remote_updated_at,
    content_hash = @content_hash,
    updated_at = @updated_at
WHERE
    %s
    AND task_comment_id = @task_comment_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"task_comment_id":   l.TaskCommentID,
		"external_id":       l.ExternalID,
		"remote_updated_at": l.RemoteUpdatedAt,
		"content_hash":      l.ContentHash,
		"updated_at":        l.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "task_comment_external_links_external_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot update task comment external link: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (l *TaskCommentExternalLink) Delete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
DELETE FROM task_comment_external_links
WHERE
    %s
    AND task_comment_id = @task_comment_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_comment_id": l.TaskCommentID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete task comment external link: %w", err)
	}

	return nil
}

func (l *TaskCommentExternalLinks) DeleteByTaskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
) error {
	q := `
DELETE FROM task_comment_external_links
WHERE
    %s
    AND task_id = @task_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete task comment external links: %w", err)
	}

	return nil
}

func (l *TaskCommentExternalLink) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query task comment external link: %w", err)
	}

	link, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TaskCommentExternalLink])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect task comment external link: %w", err)
	}

	*l = link

	return nil
}
