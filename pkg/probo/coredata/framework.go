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

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	Framework struct {
		ID             gid.GID
		OrganizationID gid.GID
		Name           string
		Description    string
		ContentRef     string
		CreatedAt      time.Time
		UpdatedAt      time.Time
		Version        int
	}

	Frameworks []*Framework

	UpdateFrameworkParams struct {
		ExpectedVersion int
		Name            *string
		Description     *string
	}
)

func (f Framework) CursorKey() page.CursorKey {
	return page.NewCursorKey(f.ID, f.CreatedAt)
}

func (f *Framework) scan(r pgx.Row) error {
	return r.Scan(
		&f.ID,
		&f.OrganizationID,
		&f.Name,
		&f.Description,
		&f.ContentRef,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.Version,
	)
}

func (f *Frameworks) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope *Scope,
	organizationID gid.GID,
	cursor *page.Cursor,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    description,
    content_ref,
    created_at,
    updated_at,
    version
FROM
    frameworks
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, cursor.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	r, err := conn.Query(ctx, q, args)
	if err != nil {
		return err
	}
	defer r.Close()

	frameworks := Frameworks{}
	for r.Next() {
		framework := &Framework{}
		if err := framework.scan(r); err != nil {
			return err
		}

		frameworks = append(frameworks, framework)
	}

	if err := r.Err(); err != nil {
		return err
	}

	*f = frameworks

	return nil
}

func (f *Framework) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope *Scope,
	frameworkID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    description,
    content_ref,
    created_at,
    updated_at,
    version
FROM
    frameworks
WHERE
    %s
    AND id = @framework_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"framework_id": frameworkID}
	maps.Copy(args, scope.SQLArguments())
	r := conn.QueryRow(ctx, q, args)

	f2 := Framework{}
	if err := f2.scan(r); err != nil {
		return err
	}

	*f = f2

	return nil
}

func (f Framework) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    frameworks (
        id,
        organization_id,
        name,
        description,
        content_ref,
        created_at,
        updated_at,
        version
    )
VALUES (
    @framework_id,
    @organization_id,
    @name,
    @description,
    @content_ref,
    @created_at,
    @updated_at,
    @version
);
`

	args := pgx.NamedArgs{
		"framework_id":    f.ID,
		"organization_id": f.OrganizationID,
		"name":            f.Name,
		"description":     f.Description,
		"content_ref":     f.ContentRef,
		"created_at":      f.CreatedAt,
		"updated_at":      f.UpdatedAt,
		"version":         f.Version,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (f Framework) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope *Scope,
) error {
	q := `
DELETE
FROM
    frameworks
WHERE
    %s
    AND id = @framework_id;
`

	args := pgx.NamedArgs{"framework_id": f.ID}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (f *Framework) Update(
	ctx context.Context,
	conn pg.Conn,
	scope *Scope,
	params UpdateFrameworkParams,
) error {
	q := `
UPDATE frameworks SET
    name = COALESCE(@name, name),
    description = COALESCE(@description, description),
    updated_at = @updated_at,
    version = version + 1
WHERE %s
    AND id = @framework_id
    AND version = @expected_version
RETURNING 
    id,
    organization_id,
    name,
    description,
    content_ref,
    created_at,
    updated_at,
    version
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{
		"framework_id":     f.ID,
		"expected_version": params.ExpectedVersion,
		"updated_at":       time.Now(),
	}

	if params.Name != nil {
		args["name"] = *params.Name
	}
	if params.Description != nil {
		args["description"] = *params.Description
	}

	maps.Copy(args, scope.SQLArguments())

	r := conn.QueryRow(ctx, q, args)

	f2 := Framework{}
	if err := f2.scan(r); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConcurrentModification
		}
		return err
	}

	*f = f2
	return nil
}
