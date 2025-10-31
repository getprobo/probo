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

	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
)

type (
	Framework struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		ReferenceID    string    `db:"reference_id"`
		Name           string    `db:"name"`
		Description    string    `db:"description"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	Frameworks []*Framework

	ErrFrameworkNotFound struct {
		Identifier string
	}

	ErrFrameworkAlreadyExists struct {
		message string
	}

	ErrFrameworkReferenceIDAlreadyExists struct {
		ReferenceID    string
		OrganizationID gid.GID
	}
)

func (e ErrFrameworkNotFound) Error() string {
	return fmt.Sprintf("framework not found: %q", e.Identifier)
}

func (e ErrFrameworkAlreadyExists) Error() string {
	return e.message
}

func (e ErrFrameworkReferenceIDAlreadyExists) Error() string {
	return fmt.Sprintf("framework with reference ID %q already exists for organization %s", e.ReferenceID, e.OrganizationID)
}

func (f *Framework) CursorKey(orderBy FrameworkOrderField) page.CursorKey {
	switch orderBy {
	case FrameworkOrderFieldCreatedAt:
		return page.CursorKey{ID: f.ID, Value: f.CreatedAt}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (f *Frameworks) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    frameworks
WHERE
    %s
    AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (f *Frameworks) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[FrameworkOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    reference_id,
    name,
    description,
    created_at,
    updated_at
FROM
    frameworks
WHERE
    %s
    AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query frameworks: %w", err)
	}

	frameworks, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Framework])
	if err != nil {
		return fmt.Errorf("cannot collect frameworks: %w", err)
	}

	*f = frameworks

	return nil
}

func (f *Framework) LoadByReferenceID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	referenceID string,
) error {
	q := `
SELECT
    id,
    organization_id,
    reference_id,
    name,
    description,
    created_at,
    updated_at
FROM
    frameworks
WHERE
    %s
    AND reference_id = @reference_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"reference_id": referenceID}
	maps.Copy(args, scope.SQLArguments())
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query frameworks: %w", err)
	}

	framework, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Framework])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrFrameworkNotFound{Identifier: referenceID}
		}

		return fmt.Errorf("cannot collect framework: %w", err)
	}

	*f = framework

	return nil
}

func (f *Framework) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	frameworkID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    reference_id,
    name,
    description,
    created_at,
    updated_at
FROM
    frameworks
WHERE
    %s
    AND id = @framework_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"framework_id": frameworkID}
	maps.Copy(args, scope.SQLArguments())
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query frameworks: %w", err)
	}

	framework, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Framework])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrFrameworkNotFound{Identifier: frameworkID.String()}
		}

		return fmt.Errorf("cannot collect framework: %w", err)
	}

	*f = framework

	return nil
}

func (f Framework) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    frameworks (
        tenant_id,
        id,
        organization_id,
        reference_id,
        name,
        description,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @framework_id,
    @organization_id,
    @reference_id,
    @name,
    @description,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"framework_id":    f.ID,
		"organization_id": f.OrganizationID,
		"reference_id":    f.ReferenceID,
		"name":            f.Name,
		"description":     f.Description,
		"created_at":      f.CreatedAt,
		"updated_at":      f.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "frameworks_org_ref_unique" {
				return &ErrFrameworkReferenceIDAlreadyExists{
					ReferenceID:    f.ReferenceID,
					OrganizationID: f.OrganizationID,
				}
			}
		}

		return err
	}

	return nil
}

func (f Framework) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	frameworkID gid.GID,
) error {
	q := `
DELETE
FROM
    frameworks
WHERE
    %s
    AND id = @framework_id;
`

	args := pgx.StrictNamedArgs{"framework_id": frameworkID}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (f *Framework) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE frameworks
SET
  name = @name,
  description = @description,
  updated_at = @updated_at
WHERE
  %s
  AND id = @framework_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"framework_id": f.ID,
		"updated_at":   f.UpdatedAt,
		"name":         f.Name,
		"description":  f.Description,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}
