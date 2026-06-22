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
	TrustCenterAlias struct {
		OrganizationID gid.GID   `db:"organization_id"`
		Alias          string    `db:"alias"`
		ResourceID     gid.GID   `db:"resource_id"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	TrustCenterAliases []*TrustCenterAlias
)

func (t *TrustCenterAlias) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceID gid.GID,
	alias string,
) error {
	now := time.Now()

	q := `
INSERT INTO trust_center_aliases (
	tenant_id,
	organization_id,
	alias,
	resource_id,
	created_at,
	updated_at
)
VALUES (
	@tenant_id,
	COALESCE(
		(SELECT organization_id FROM documents WHERE id = @resource_id AND tenant_id = @tenant_id),
		(SELECT organization_id FROM trust_center_files WHERE id = @resource_id AND tenant_id = @tenant_id),
		(SELECT organization_id FROM files WHERE id = @resource_id AND tenant_id = @tenant_id)
	),
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
	organization_id,
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
			if pgErr.Code == "23505" && pgErr.ConstraintName == "trust_center_aliases_organization_id_alias_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot upsert trust center alias: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterAlias])
	if err != nil {
		return fmt.Errorf("cannot collect trust center alias: %w", err)
	}

	*t = row

	return nil
}

func (t *TrustCenterAlias) LoadByAlias(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	alias string,
) error {
	q := `
SELECT
	organization_id,
	alias,
	resource_id,
	created_at,
	updated_at
FROM
	trust_center_aliases
WHERE
	%s
	AND organization_id = @organization_id
	AND alias = @alias
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"alias":           alias,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center alias: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterAlias])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center alias: %w", err)
	}

	*t = row

	return nil
}

func (t *TrustCenterAlias) LoadByResourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	resourceID gid.GID,
) error {
	q := `
SELECT
	organization_id,
	alias,
	resource_id,
	created_at,
	updated_at
FROM
	trust_center_aliases
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
		return fmt.Errorf("cannot query trust center alias: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TrustCenterAlias])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect trust center alias: %w", err)
	}

	*t = row

	return nil
}

func (ts *TrustCenterAliases) LoadByResourceIDs(
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
	organization_id,
	alias,
	resource_id,
	created_at,
	updated_at
FROM
	trust_center_aliases
WHERE
	%s
	AND resource_id = ANY(@resource_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query trust center aliases: %w", err)
	}

	aliases, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TrustCenterAlias])
	if err != nil {
		return fmt.Errorf("cannot collect trust center aliases: %w", err)
	}

	*ts = aliases

	return nil
}

func (t *TrustCenterAlias) Delete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
DELETE FROM trust_center_aliases
WHERE
	%s
	AND resource_id = @resource_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"resource_id": t.ResourceID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete trust center alias: %w", err)
	}

	return nil
}
