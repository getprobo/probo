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
	EmployeePortal struct {
		ID             gid.GID                    `db:"id"`
		OrganizationID gid.GID                    `db:"organization_id"`
		Name           string                     `db:"name"`
		Active         bool                       `db:"active"`
		Capabilities   EmployeePortalCapabilities `db:"capabilities"`
		LogoFileID     *gid.GID                   `db:"logo_file_id"`
		DarkLogoFileID *gid.GID                   `db:"dark_logo_file_id"`
		CreatedAt      time.Time                  `db:"created_at"`
		UpdatedAt      time.Time                  `db:"updated_at"`
	}

	EmployeePortals []*EmployeePortal
)

func (p *EmployeePortal) CursorKey(orderBy EmployeePortalOrderField) page.CursorKey {
	switch orderBy {
	case EmployeePortalOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p *EmployeePortal) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM employee_portals WHERE id = ANY(@resource_ids::text[])`

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

func (p *EmployeePortal) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	employeePortalID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	active,
	capabilities,
	logo_file_id,
	dark_logo_file_id,
	created_at,
	updated_at
FROM
	employee_portals
WHERE
	%s
	AND id = @employee_portal_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"employee_portal_id": employeePortalID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query employee portal: %w", err)
	}

	portal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[EmployeePortal])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect employee portal: %w", err)
	}

	*p = portal

	return nil
}

func (p *EmployeePortal) LoadOldestByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	active,
	capabilities,
	logo_file_id,
	dark_logo_file_id,
	created_at,
	updated_at
FROM
	employee_portals
WHERE
	%s
	AND organization_id = @organization_id
	AND active = TRUE
ORDER BY
	created_at ASC,
	id ASC
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query employee portal: %w", err)
	}

	portal, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[EmployeePortal])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect employee portal: %w", err)
	}

	*p = portal

	return nil
}

func (ps *EmployeePortals) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[EmployeePortalOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	active,
	capabilities,
	logo_file_id,
	dark_logo_file_id,
	created_at,
	updated_at
FROM
	employee_portals
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query employee portals: %w", err)
	}

	portals, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[EmployeePortal])
	if err != nil {
		return fmt.Errorf("cannot collect employee portals: %w", err)
	}

	*ps = portals

	return nil
}

func (ps *EmployeePortals) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	employee_portals
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int

	err := conn.QueryRow(ctx, q, args).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count employee portals: %w", err)
	}

	return count, nil
}

func (p *EmployeePortal) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO employee_portals (
	id,
	organization_id,
	tenant_id,
	name,
	active,
	capabilities,
	logo_file_id,
	dark_logo_file_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@organization_id,
	@tenant_id,
	@name,
	@active,
	@capabilities,
	@logo_file_id,
	@dark_logo_file_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                p.ID,
		"organization_id":   p.OrganizationID,
		"tenant_id":         scope.GetTenantID(),
		"name":              p.Name,
		"active":            p.Active,
		"capabilities":      p.Capabilities,
		"logo_file_id":      p.LogoFileID,
		"dark_logo_file_id": p.DarkLogoFileID,
		"created_at":        p.CreatedAt,
		"updated_at":        p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_employee_portals_organization_id" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert employee portal: %w", err)
	}

	return nil
}

func (p *EmployeePortal) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE employee_portals
SET
	name = @name,
	active = @active,
	capabilities = @capabilities,
	logo_file_id = @logo_file_id,
	dark_logo_file_id = @dark_logo_file_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                p.ID,
		"name":              p.Name,
		"active":            p.Active,
		"capabilities":      p.Capabilities,
		"logo_file_id":      p.LogoFileID,
		"dark_logo_file_id": p.DarkLogoFileID,
		"updated_at":        p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update employee portal: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
