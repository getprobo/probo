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
	TagAssignment struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		ResourceID     gid.GID   `db:"resource_id"`
		TagID          gid.GID   `db:"tag_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	TagAssignments []*TagAssignment
)

func (a TagAssignment) CursorKey(orderBy TagAssignmentOrderField) page.CursorKey {
	switch orderBy {
	case TagAssignmentOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (a *TagAssignment) LoadByResourceIDAndTagID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceID gid.GID,
	tagID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	resource_id,
	tag_id,
	created_at
FROM
	tag_assignments
WHERE
	%s
	AND resource_id = @resource_id
	AND tag_id = @tag_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"resource_id": resourceID,
		"tag_id":      tagID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tag assignment: %w", err)
	}

	assignment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TagAssignment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect tag assignment: %w", err)
	}

	*a = assignment

	return nil
}

func (a *TagAssignment) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO tag_assignments (
	id,
	tenant_id,
	organization_id,
	resource_id,
	tag_id,
	created_at
)
VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@resource_id,
	@tag_id,
	@created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              a.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": a.OrganizationID,
		"resource_id":     a.ResourceID,
		"tag_id":          a.TagID,
		"created_at":      a.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "tag_assignments_resource_id_tag_id_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert tag assignment: %w", err)
	}

	return nil
}

// DeleteByResourceIDAndTagID removes the assignment. It returns
// ErrResourceNotFound when the tag was not attached.
func (a *TagAssignment) DeleteByResourceIDAndTagID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceID gid.GID,
	tagID gid.GID,
) error {
	q := `
DELETE FROM tag_assignments
WHERE
	%s
	AND resource_id = @resource_id
	AND tag_id = @tag_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"resource_id": resourceID,
		"tag_id":      tagID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete tag assignment: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (as *TagAssignments) LoadByTagID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	tagID gid.GID,
	cursor *page.Cursor[TagAssignmentOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	resource_id,
	tag_id,
	created_at
FROM
	tag_assignments
WHERE
	%s
	AND tag_id = @tag_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"tag_id": tagID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tag assignments: %w", err)
	}

	assignments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TagAssignment])
	if err != nil {
		return fmt.Errorf("cannot collect tag assignments: %w", err)
	}

	*as = assignments

	return nil
}

func (as *TagAssignments) CountByTagID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	tagID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(*)
FROM
	tag_assignments
WHERE
	%s
	AND tag_id = @tag_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"tag_id": tagID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count tag assignments: %w", err)
	}

	return count, nil
}

func (as *TagAssignments) LoadByResourceIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceIDs []gid.GID,
) error {
	if len(resourceIDs) == 0 {
		*as = nil

		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	resource_id,
	tag_id,
	created_at
FROM
	tag_assignments
WHERE
	%s
	AND resource_id = ANY(@resource_ids::text[])
ORDER BY
	created_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tag assignments: %w", err)
	}

	assignments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TagAssignment])
	if err != nil {
		return fmt.Errorf("cannot collect tag assignments: %w", err)
	}

	*as = assignments

	return nil
}

func LoadResourceIDsByTagID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	tagID gid.GID,
) ([]gid.GID, error) {
	q := `
SELECT
	resource_id
FROM
	tag_assignments
WHERE
	%s
	AND tag_id = @tag_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"tag_id": tagID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query tag assignment resources: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect tag assignment resources: %w", err)
	}

	return ids, nil
}
