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
	ResourceTagAssignment struct {
		ID         gid.GID   `db:"id"`
		ResourceID gid.GID   `db:"resource_id"`
		TagID      gid.GID   `db:"tag_id"`
		CreatedAt  time.Time `db:"created_at"`
	}

	ResourceTagAssignments []*ResourceTagAssignment
)

func (a ResourceTagAssignment) CursorKey(orderBy ResourceTagAssignmentOrderField) page.CursorKey {
	switch orderBy {
	case ResourceTagAssignmentOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (a *ResourceTagAssignment) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO resource_tag_assignments (
	tenant_id,
	id,
	resource_id,
	tag_id,
	created_at
)
VALUES (
	@tenant_id,
	@id,
	@resource_id,
	@tag_id,
	@created_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":   scope.GetTenantID(),
		"id":          a.ID,
		"resource_id": a.ResourceID,
		"tag_id":      a.TagID,
		"created_at":  a.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "resource_tag_assignments_resource_id_tag_id_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert resource tag assignment: %w", err)
	}

	return nil
}

func (a *ResourceTagAssignment) Delete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
DELETE FROM resource_tag_assignments
WHERE
	%s
	AND resource_id = @resource_id
	AND tag_id = @tag_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"resource_id": a.ResourceID,
		"tag_id":      a.TagID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete resource tag assignment: %w", err)
	}

	return nil
}

func (as *ResourceTagAssignments) LoadByTagID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	tagID gid.GID,
	cursor *page.Cursor[ResourceTagAssignmentOrderField],
) error {
	q := `
SELECT
	id,
	resource_id,
	tag_id,
	created_at
FROM
	resource_tag_assignments
WHERE
	%s
	AND tag_id = @tag_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"tag_id": tagID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query resource tag assignments: %w", err)
	}

	assignments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ResourceTagAssignment])
	if err != nil {
		return fmt.Errorf("cannot collect resource tag assignments: %w", err)
	}

	*as = assignments

	return nil
}

func (as *ResourceTagAssignments) CountByTagID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	tagID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(*)
FROM
	resource_tag_assignments
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
		return 0, fmt.Errorf("cannot count resource tag assignments: %w", err)
	}

	return count, nil
}

func (as *ResourceTagAssignments) LoadByResourceIDs(
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
	resource_id,
	tag_id,
	created_at
FROM
	resource_tag_assignments
WHERE
	%s
	AND resource_id = ANY(@resource_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query resource tag assignments: %w", err)
	}

	assignments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ResourceTagAssignment])
	if err != nil {
		return fmt.Errorf("cannot collect resource tag assignments: %w", err)
	}

	*as = assignments

	return nil
}

// FilterResourceIDs returns the subset of candidateIDs that are assigned every
// tag in tagIDs (AND semantics). Empty tagIDs returns candidateIDs unchanged.
func (as *ResourceTagAssignments) FilterResourceIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	candidateIDs []gid.GID,
	tagIDs []gid.GID,
) ([]gid.GID, error) {
	if len(candidateIDs) == 0 {
		return nil, nil
	}

	if len(tagIDs) == 0 {
		return candidateIDs, nil
	}

	q := `
SELECT
	resource_id
FROM
	resource_tag_assignments
WHERE
	%s
	AND resource_id = ANY(@resource_ids)
	AND tag_id = ANY(@tag_ids)
GROUP BY
	resource_id
HAVING
	COUNT(DISTINCT tag_id) = @tag_count
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"resource_ids": candidateIDs,
		"tag_ids":      tagIDs,
		"tag_count":    len(tagIDs),
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot filter resource ids by tags: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect filtered resource ids: %w", err)
	}

	return ids, nil
}
