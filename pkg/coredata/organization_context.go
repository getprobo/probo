// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	OrganizationContext struct {
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

func (oc *OrganizationContext) LoadByOrganizationID(
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
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO organization_contexts (
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
		"organization_id": oc.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"product":         oc.Product,
		"architecture":    oc.Architecture,
		"team":            oc.Team,
		"processes":       oc.Processes,
		"customers":       oc.Customers,
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
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE organization_contexts
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
		"organization_id": oc.OrganizationID,
		"product":         oc.Product,
		"architecture":    oc.Architecture,
		"team":            oc.Team,
		"processes":       oc.Processes,
		"customers":       oc.Customers,
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
