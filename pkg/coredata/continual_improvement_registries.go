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
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	ContinualImprovementRegistry struct {
		ID             gid.GID                                `db:"id"`
		OrganizationID gid.GID                                `db:"organization_id"`
		ReferenceID    string                                 `db:"reference_id"`
		Description    *string                                `db:"description"`
		AuditID        gid.GID                                `db:"audit_id"`
		Source         *string                                `db:"source"`
		OwnerID        gid.GID                                `db:"owner_id"`
		TargetDate     *time.Time                             `db:"target_date"`
		Status         ContinualImprovementRegistriesStatus   `db:"status"`
		Priority       ContinualImprovementRegistriesPriority `db:"priority"`
		CreatedAt      time.Time                              `db:"created_at"`
		UpdatedAt      time.Time                              `db:"updated_at"`
	}

	ContinualImprovementRegistries []*ContinualImprovementRegistry
)

func (cir *ContinualImprovementRegistry) CursorKey(field ContinualImprovementRegistriesOrderField) page.CursorKey {
	switch field {
	case ContinualImprovementRegistriesOrderFieldCreatedAt:
		return page.NewCursorKey(cir.ID, cir.CreatedAt)
	case ContinualImprovementRegistriesOrderFieldTargetDate:
		return page.NewCursorKey(cir.ID, cir.TargetDate)
	case ContinualImprovementRegistriesOrderFieldStatus:
		return page.NewCursorKey(cir.ID, cir.Status)
	case ContinualImprovementRegistriesOrderFieldPriority:
		return page.NewCursorKey(cir.ID, cir.Priority)
	case ContinualImprovementRegistriesOrderFieldReferenceId:
		return page.NewCursorKey(cir.ID, cir.ReferenceID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (cir *ContinualImprovementRegistry) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	continualImprovementRegistryID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	description,
	audit_id,
	source,
	owner_id,
	target_date,
	status,
	priority,
	created_at,
	updated_at
FROM
	continual_improvement_registries
WHERE
	%s
	AND id = @continual_improvement_registry_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"continual_improvement_registry_id": continualImprovementRegistryID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query continual improvement registry: %w", err)
	}

	registry, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ContinualImprovementRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect continual improvement registry: %w", err)
	}

	*cir = registry

	return nil
}

func (cirs *ContinualImprovementRegistries) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	continual_improvement_registries
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count continual improvement registries: %w", err)
	}

	return count, nil
}

func (cirs *ContinualImprovementRegistries) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ContinualImprovementRegistriesOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	description,
	audit_id,
	source,
	owner_id,
	target_date,
	status,
	priority,
	created_at,
	updated_at
FROM
	continual_improvement_registries
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
		return fmt.Errorf("cannot query continual improvement registries: %w", err)
	}

	registries, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ContinualImprovementRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect continual improvement registries: %w", err)
	}

	*cirs = registries

	return nil
}

func (cir *ContinualImprovementRegistry) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO continual_improvement_registries (
	id,
	tenant_id,
	organization_id,
	reference_id,
	description,
	audit_id,
	source,
	owner_id,
	target_date,
	status,
	priority,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@reference_id,
	@description,
	@audit_id,
	@source,
	@owner_id,
	@target_date,
	@status,
	@priority,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              cir.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": cir.OrganizationID,
		"reference_id":    cir.ReferenceID,
		"description":     cir.Description,
		"audit_id":        cir.AuditID,
		"source":          cir.Source,
		"owner_id":        cir.OwnerID,
		"target_date":     cir.TargetDate,
		"status":          cir.Status,
		"priority":        cir.Priority,
		"created_at":      cir.CreatedAt,
		"updated_at":      cir.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert continual improvement registry: %w", err)
	}

	return nil
}

func (cir *ContinualImprovementRegistry) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE continual_improvement_registries SET
	reference_id = @reference_id,
	description = @description,
	audit_id = @audit_id,
	source = @source,
	owner_id = @owner_id,
	target_date = @target_date,
	status = @status,
	priority = @priority,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           cir.ID,
		"reference_id": cir.ReferenceID,
		"description":  cir.Description,
		"audit_id":     cir.AuditID,
		"source":       cir.Source,
		"owner_id":     cir.OwnerID,
		"target_date":  cir.TargetDate,
		"status":       cir.Status,
		"priority":     cir.Priority,
		"updated_at":   cir.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update continual improvement registry: %w", err)
	}

	return nil
}

func (cir *ContinualImprovementRegistry) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM continual_improvement_registries
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": cir.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete continual improvement registry: %w", err)
	}

	return nil
}
