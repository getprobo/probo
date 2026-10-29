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
	DocumentComment struct {
		ID              gid.GID   `db:"id"`
		OrganizationID  gid.GID   `db:"organization_id"`
		ThreadID        gid.GID   `db:"thread_id"`
		AuthorProfileID gid.GID   `db:"author_profile_id"`
		Content         string    `db:"content"`
		CreatedAt       time.Time `db:"created_at"`
		UpdatedAt       time.Time `db:"updated_at"`
	}

	DocumentComments []*DocumentComment
)

func (c DocumentComment) CursorKey(orderBy DocumentCommentOrderField) page.CursorKey {
	switch orderBy {
	case DocumentCommentOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *DocumentComment) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    organization_id
FROM
    document_comments
WHERE
    id = ANY(@resource_ids)
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot query comment authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan comment authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate comment authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (c *DocumentComment) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	commentID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    thread_id,
    author_profile_id,
    content,
    created_at,
    updated_at
FROM
    document_comments
WHERE
    %s
    AND id = @comment_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"comment_id": commentID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document comment: %w", err)
	}

	comment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentComment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document comment: %w", err)
	}

	*c = comment

	return nil
}

func (c *DocumentComments) LoadByThreadID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	threadID gid.GID,
	cursor *page.Cursor[DocumentCommentOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    thread_id,
    author_profile_id,
    content,
    created_at,
    updated_at
FROM
    document_comments
WHERE
    %s
    AND thread_id = @thread_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"thread_id": threadID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document comments: %w", err)
	}

	comments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentComment])
	if err != nil {
		return fmt.Errorf("cannot collect document comments: %w", err)
	}

	*c = comments

	return nil
}

func (c *DocumentComments) LoadByThreadIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	threadIDs []gid.GID,
) error {
	if len(threadIDs) == 0 {
		*c = nil
		return nil
	}

	q := `
SELECT
    id,
    organization_id,
    thread_id,
    author_profile_id,
    content,
    created_at,
    updated_at
FROM
    document_comments
WHERE
    %s
    AND thread_id = ANY(@thread_ids)
ORDER BY
    created_at ASC,
    id ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"thread_ids": threadIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document comments: %w", err)
	}

	comments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentComment])
	if err != nil {
		return fmt.Errorf("cannot collect document comments: %w", err)
	}

	*c = comments

	return nil
}

func (c *DocumentComments) CountByThreadID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	threadID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    document_comments
WHERE
    %s
    AND thread_id = @thread_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"thread_id": threadID}
	maps.Copy(args, scope.SQLArguments())

	var count int

	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count document comments: %w", err)
	}

	return count, nil
}

func (c DocumentComment) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    document_comments (
        tenant_id,
        id,
        organization_id,
        thread_id,
        author_profile_id,
        content,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @comment_id,
    @organization_id,
    @thread_id,
    @author_profile_id,
    @content,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":         scope.GetTenantID(),
		"comment_id":        c.ID,
		"organization_id":   c.OrganizationID,
		"thread_id":         c.ThreadID,
		"author_profile_id": c.AuthorProfileID,
		"content":           c.Content,
		"created_at":        c.CreatedAt,
		"updated_at":        c.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert document comment: %w", err)
	}

	return nil
}

func (c DocumentComment) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
    document_comments
SET
    content = @content,
    updated_at = @updated_at
WHERE
    %s
    AND id = @comment_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"comment_id": c.ID,
		"content":    c.Content,
		"updated_at": c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document comment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (c DocumentComment) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
    document_comments
WHERE
    %s
    AND id = @comment_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"comment_id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete document comment: %w", err)
	}

	return nil
}
