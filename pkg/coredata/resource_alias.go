// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	ResourceAlias struct {
		Alias      string    `db:"alias"`
		ResourceID gid.GID   `db:"resource_id"`
		CreatedAt  time.Time `db:"created_at"`
		UpdatedAt  time.Time `db:"updated_at"`
	}

	ResourceAliases []*ResourceAlias
)

func (t *ResourceAlias) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceID gid.GID,
	alias string,
) error {
	now := time.Now()

	q := `
INSERT INTO resource_aliases (
	tenant_id,
	alias,
	resource_id,
	created_at,
	updated_at
)
VALUES (
	@tenant_id,
	@alias,
	@resource_id,
	@created_at,
	@updated_at
)
ON CONFLICT (resource_id) DO UPDATE
SET
	alias = EXCLUDED.alias,
	updated_at = EXCLUDED.updated_at
RETURNING
	alias,
	resource_id,
	created_at,
	updated_at
`

	args := pgx.StrictNamedArgs{
		"tenant_id":   scope.GetTenantID(),
		"alias":       alias,
		"resource_id": resourceID,
		"created_at":  now,
		"updated_at":  now,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "resource_aliases_tenant_id_alias_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot upsert resource alias: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ResourceAlias])
	if err != nil {
		return fmt.Errorf("cannot collect resource alias: %w", err)
	}

	*t = row

	return nil
}

func (t *ResourceAlias) LoadByAlias(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	alias string,
) error {
	q := `
SELECT
	alias,
	resource_id,
	created_at,
	updated_at
FROM
	resource_aliases
WHERE
	%s
	AND alias = @alias
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"alias": alias,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query resource alias: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ResourceAlias])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect resource alias: %w", err)
	}

	*t = row

	return nil
}

func (t *ResourceAlias) LoadByResourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceID gid.GID,
) error {
	q := `
SELECT
	alias,
	resource_id,
	created_at,
	updated_at
FROM
	resource_aliases
WHERE
	%s
	AND resource_id = @resource_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"resource_id": resourceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query resource alias: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ResourceAlias])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect resource alias: %w", err)
	}

	*t = row

	return nil
}

func (ts *ResourceAliases) LoadByResourceIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceIDs []gid.GID,
) error {
	if len(resourceIDs) == 0 {
		*ts = nil

		return nil
	}

	q := `
SELECT
	alias,
	resource_id,
	created_at,
	updated_at
FROM
	resource_aliases
WHERE
	%s
	AND resource_id = ANY(@resource_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query resource aliases: %w", err)
	}

	aliases, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ResourceAlias])
	if err != nil {
		return fmt.Errorf("cannot collect resource aliases: %w", err)
	}

	*ts = aliases

	return nil
}

func (t *ResourceAlias) Delete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
DELETE FROM resource_aliases
WHERE
	%s
	AND resource_id = @resource_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"resource_id": t.ResourceID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete resource alias: %w", err)
	}

	return nil
}
