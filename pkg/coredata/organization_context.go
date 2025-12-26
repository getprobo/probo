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

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	OrganizationContext struct {
		OrganizationID gid.GID   `db:"organization_id"`
		Summary        *string   `db:"summary"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}
)

func (oc *OrganizationContext) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    organization_id,
    summary,
    created_at,
    updated_at
FROM
    organization_contexts
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
		return fmt.Errorf("cannot query organization context: %w", err)
	}

	orgContext, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OrganizationContext])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect organization context: %w", err)
	}

	*oc = orgContext

	return nil
}

func (oc *OrganizationContext) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO organization_contexts (
    organization_id,
    tenant_id,
    summary,
    created_at,
    updated_at
) VALUES (
    @organization_id,
    @tenant_id,
    @summary,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"organization_id": oc.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"summary":         oc.Summary,
		"created_at":      oc.CreatedAt,
		"updated_at":      oc.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert organization context: %w", err)
	}

	return nil
}

func (oc *OrganizationContext) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE organization_contexts
SET
    summary = @summary,
    updated_at = @updated_at
WHERE
    %s
    AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": oc.OrganizationID,
		"summary":         oc.Summary,
		"updated_at":      oc.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update organization context: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
