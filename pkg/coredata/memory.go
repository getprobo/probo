// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	Memory struct {
		OrganizationID gid.GID   `db:"organization_id"`
		Product        *string   `db:"product"`
		Architecture   *string   `db:"architecture"`
		Team           *string   `db:"team"`
		Processes      *string   `db:"processes"`
		Customers      *string   `db:"customers"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}
)

func (m *Memory) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    organization_id,
    product,
    architecture,
    team,
    processes,
    customers,
    created_at,
    updated_at
FROM
    memories
WHERE
    %s
    AND organization_id = @organization_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query memory: %w", err)
	}

	memory, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Memory])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect memory: %w", err)
	}

	*m = memory

	return nil
}

func (m *Memory) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO memories (
    organization_id,
    tenant_id,
    product,
    architecture,
    team,
    processes,
    customers,
    created_at,
    updated_at
) VALUES (
    @organization_id,
    @tenant_id,
    @product,
    @architecture,
    @team,
    @processes,
    @customers,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"organization_id": m.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"product":         m.Product,
		"architecture":    m.Architecture,
		"team":            m.Team,
		"processes":       m.Processes,
		"customers":       m.Customers,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert memory: %w", err)
	}

	return nil
}

func (m *Memory) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE memories
SET
    product = @product,
    architecture = @architecture,
    team = @team,
    processes = @processes,
    customers = @customers,
    updated_at = @updated_at
WHERE
    %s
    AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": m.OrganizationID,
		"product":         m.Product,
		"architecture":    m.Architecture,
		"team":            m.Team,
		"processes":       m.Processes,
		"customers":       m.Customers,
		"updated_at":      m.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update memory: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
