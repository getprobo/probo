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
		Key            string    `db:"key"`
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

func (a *TagAssignment) LoadByResourceIDAndKey(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceID gid.GID,
	key string,
) error {
	q := `
SELECT
	id,
	organization_id,
	resource_id,
	tag_id,
	key,
	created_at
FROM
	tag_assignments
WHERE
	%s
	AND resource_id = @resource_id
	AND key = @key
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"resource_id": resourceID,
		"key":         key,
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
	tenant_id,
	id,
	organization_id,
	resource_id,
	tag_id,
	key,
	created_at
)
VALUES (
	@tenant_id,
	@id,
	@organization_id,
	@resource_id,
	@tag_id,
	@key,
	@created_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"id":              a.ID,
		"organization_id": a.OrganizationID,
		"resource_id":     a.ResourceID,
		"tag_id":          a.TagID,
		"key":             a.Key,
		"created_at":      a.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "tag_assignments_resource_id_key_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert tag assignment: %w", err)
	}

	return nil
}

func (a *TagAssignment) Delete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
DELETE FROM tag_assignments
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": a.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete tag assignment: %w", err)
	}

	return nil
}

// DeleteByResourceIDAndTagID removes the assignment and loads the deleted row
// into a. It returns ErrResourceNotFound when the tag was not attached.
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
RETURNING
	id,
	organization_id,
	resource_id,
	tag_id,
	key,
	created_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"resource_id": resourceID,
		"tag_id":      tagID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete tag assignment: %w", err)
	}

	assignment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TagAssignment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect deleted tag assignment: %w", err)
	}

	*a = assignment

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
	key,
	created_at
FROM
	tag_assignments
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
	key,
	created_at
FROM
	tag_assignments
WHERE
	%s
	AND resource_id = ANY(@resource_ids)
ORDER BY
	key ASC
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

// FilterResourceIDs returns the subset of candidateIDs that are assigned every
// tag in tagIDs (AND semantics). Empty tagIDs returns candidateIDs unchanged.
func (as *TagAssignments) FilterResourceIDs(
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
	tag_assignments
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
