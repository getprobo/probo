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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	Tag struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		Key            string    `db:"key"`
		Value          string    `db:"value"`
		Color          *string   `db:"color"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	Tags []*Tag
)

func (t Tag) CursorKey(orderBy TagOrderField) page.CursorKey {
	switch orderBy {
	case TagOrderFieldCreatedAt:
		return page.NewCursorKey(t.ID, t.CreatedAt)
	case TagOrderFieldKey:
		return page.NewCursorKey(t.ID, t.Key)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (t *Tag) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM tags WHERE id = ANY(@resource_ids::text[])`

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

func (t *Tag) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	tagID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	key,
	value,
	color,
	created_at,
	updated_at
FROM
	tags
WHERE
	%s
	AND id = @tag_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"tag_id": tagID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tag: %w", err)
	}

	tag, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Tag])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect tag: %w", err)
	}

	*t = tag

	return nil
}

func (ts *Tags) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[TagOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	key,
	value,
	color,
	created_at,
	updated_at
FROM
	tags
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tags: %w", err)
	}

	tags, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Tag])
	if err != nil {
		return fmt.Errorf("cannot collect tags: %w", err)
	}

	*ts = tags

	return nil
}

func (ts *Tags) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(*)
FROM
	tags
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count tags: %w", err)
	}

	return count, nil
}

func (ts *Tags) LoadByResourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	key,
	value,
	color,
	created_at,
	updated_at
FROM
	tags
WHERE
	%s
	AND id IN (
		SELECT tag_id
		FROM tag_assignments
		WHERE resource_id = @resource_id
			AND tenant_id = @tenant_id
	)
ORDER BY
	key ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"resource_id": resourceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tags by resource: %w", err)
	}

	tags, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Tag])
	if err != nil {
		return fmt.Errorf("cannot collect tags by resource: %w", err)
	}

	*ts = tags

	return nil
}

func (ts *Tags) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	tagIDs []gid.GID,
) error {
	if len(tagIDs) == 0 {
		*ts = nil

		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	key,
	value,
	color,
	created_at,
	updated_at
FROM
	tags
WHERE
	%s
	AND id = ANY(@tag_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"tag_ids": tagIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query tags by ids: %w", err)
	}

	tags, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Tag])
	if err != nil {
		return fmt.Errorf("cannot collect tags by ids: %w", err)
	}

	*ts = tags

	return nil
}

func (t *Tag) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO tags (
	tenant_id,
	id,
	organization_id,
	key,
	value,
	color,
	created_at,
	updated_at
)
VALUES (
	@tenant_id,
	@id,
	@organization_id,
	@key,
	@value,
	@color,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"id":              t.ID,
		"organization_id": t.OrganizationID,
		"key":             t.Key,
		"value":           t.Value,
		"color":           t.Color,
		"created_at":      t.CreatedAt,
		"updated_at":      t.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "tags_organization_id_key_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert tag: %w", err)
	}

	return nil
}

func (t *Tag) Update(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
UPDATE tags
SET
	key = @key,
	value = @value,
	color = @color,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         t.ID,
		"key":        t.Key,
		"value":      t.Value,
		"color":      t.Color,
		"updated_at": t.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "tags_organization_id_key_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot update tag: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (t *Tag) Delete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
DELETE FROM tags
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": t.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete tag: %w", err)
	}

	return nil
}
