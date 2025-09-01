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
	IncidentRegistry struct {
		ID             gid.GID                 `db:"id"`
		OrganizationID gid.GID                 `db:"organization_id"`
		ReferenceID    string                  `db:"reference_id"`
		Title          string                  `db:"title"`
		Description    *string                 `db:"description"`
		Source         *string                 `db:"source"`
		OwnerID        gid.GID                 `db:"owner_id"`
		IncidentDate   *time.Time              `db:"incident_date"`
		ResolvedDate   *time.Time              `db:"resolved_date"`
		Status         IncidentRegistryStatus  `db:"status"`
		Priority       IncidentRegistryPriority `db:"priority"`
		Severity       IncidentRegistrySeverity `db:"severity"`
		Category       *string                 `db:"category"`
		CreatedAt      time.Time               `db:"created_at"`
		UpdatedAt      time.Time               `db:"updated_at"`
	}

	IncidentRegistries []*IncidentRegistry
)

func (ir *IncidentRegistry) CursorKey(field IncidentRegistryOrderField) page.CursorKey {
	switch field {
	case IncidentRegistryOrderFieldCreatedAt:
		return page.NewCursorKey(ir.ID, ir.CreatedAt)
	case IncidentRegistryOrderFieldIncidentDate:
		return page.NewCursorKey(ir.ID, ir.IncidentDate)
	case IncidentRegistryOrderFieldStatus:
		return page.NewCursorKey(ir.ID, ir.Status)
	case IncidentRegistryOrderFieldPriority:
		return page.NewCursorKey(ir.ID, ir.Priority)
	case IncidentRegistryOrderFieldSeverity:
		return page.NewCursorKey(ir.ID, ir.Severity)
	case IncidentRegistryOrderFieldReferenceId:
		return page.NewCursorKey(ir.ID, ir.ReferenceID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (ir *IncidentRegistry) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	incidentRegistryID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	title,
	description,
	source,
	owner_id,
	incident_date,
	resolved_date,
	status,
	priority,
	severity,
	category,
	created_at,
	updated_at
FROM
	incident_registries
WHERE
	%s
	AND id = @incident_registry_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"incident_registry_id": incidentRegistryID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query incident registry: %w", err)
	}

	registry, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[IncidentRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect incident registry: %w", err)
	}

	*ir = registry

	return nil
}

func (irs *IncidentRegistries) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	incident_registries
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
		return 0, fmt.Errorf("cannot count incident registries: %w", err)
	}

	return count, nil
}

func (irs *IncidentRegistries) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[IncidentRegistryOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	title,
	description,
	source,
	owner_id,
	incident_date,
	resolved_date,
	status,
	priority,
	severity,
	category,
	created_at,
	updated_at
FROM
	incident_registries
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
		return fmt.Errorf("cannot query incident registries: %w", err)
	}

	registries, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[IncidentRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect incident registries: %w", err)
	}

	*irs = registries

	return nil
}

func (ir *IncidentRegistry) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO incident_registries (
	id,
	tenant_id,
	organization_id,
	reference_id,
	title,
	description,
	source,
	owner_id,
	incident_date,
	resolved_date,
	status,
	priority,
	severity,
	category,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@reference_id,
	@title,
	@description,
	@source,
	@owner_id,
	@incident_date,
	@resolved_date,
	@status,
	@priority,
	@severity,
	@category,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":             ir.ID,
		"tenant_id":      scope.GetTenantID(),
		"organization_id": ir.OrganizationID,
		"reference_id":   ir.ReferenceID,
		"title":          ir.Title,
		"description":    ir.Description,
		"source":         ir.Source,
		"owner_id":       ir.OwnerID,
		"incident_date":  ir.IncidentDate,
		"resolved_date":  ir.ResolvedDate,
		"status":         ir.Status,
		"priority":       ir.Priority,
		"severity":       ir.Severity,
		"category":       ir.Category,
		"created_at":     ir.CreatedAt,
		"updated_at":     ir.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert incident registry: %w", err)
	}

	return nil
}

func (ir *IncidentRegistry) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE incident_registries SET
	reference_id = @reference_id,
	title = @title,
	description = @description,
	source = @source,
	owner_id = @owner_id,
	incident_date = @incident_date,
	resolved_date = @resolved_date,
	status = @status,
	priority = @priority,
	severity = @severity,
	category = @category,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":            ir.ID,
		"reference_id":  ir.ReferenceID,
		"title":         ir.Title,
		"description":   ir.Description,
		"source":        ir.Source,
		"owner_id":      ir.OwnerID,
		"incident_date": ir.IncidentDate,
		"resolved_date": ir.ResolvedDate,
		"status":        ir.Status,
		"priority":      ir.Priority,
		"severity":      ir.Severity,
		"category":      ir.Category,
		"updated_at":    ir.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update incident registry: %w", err)
	}

	return nil
}

func (ir *IncidentRegistry) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM incident_registries
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": ir.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete incident registry: %w", err)
	}

	return nil
}