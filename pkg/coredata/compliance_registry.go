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
	ComplianceRegistry struct {
		ID                     gid.GID                  `db:"id"`
		OrganizationID         gid.GID                  `db:"organization_id"`
		ReferenceID            string                   `db:"reference_id"`
		Area                   *string                  `db:"area"`
		Source                 *string                  `db:"source"`
		Requirement            *string                  `db:"requirement"`
		ActionsToBeImplemented *string                  `db:"actions_to_be_implemented"`
		Regulator              *string                  `db:"regulator"`
		OwnerID                gid.GID                  `db:"owner_id"`
		LastReviewDate         *time.Time               `db:"last_review_date"`
		DueDate                *time.Time               `db:"due_date"`
		Status                 ComplianceRegistryStatus `db:"status"`
		CreatedAt              time.Time                `db:"created_at"`
		UpdatedAt              time.Time                `db:"updated_at"`
	}

	ComplianceRegistries []*ComplianceRegistry
)

func (cr *ComplianceRegistry) CursorKey(field ComplianceRegistryOrderField) page.CursorKey {
	switch field {
	case ComplianceRegistryOrderFieldCreatedAt:
		return page.NewCursorKey(cr.ID, cr.CreatedAt)
	case ComplianceRegistryOrderFieldLastReviewDate:
		return page.NewCursorKey(cr.ID, cr.LastReviewDate)
	case ComplianceRegistryOrderFieldDueDate:
		return page.NewCursorKey(cr.ID, cr.DueDate)
	case ComplianceRegistryOrderFieldStatus:
		return page.NewCursorKey(cr.ID, cr.Status)
	case ComplianceRegistryOrderFieldReferenceId:
		return page.NewCursorKey(cr.ID, cr.ReferenceID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (cr *ComplianceRegistry) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	complianceRegistryID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	area,
	source,
	requirement,
	actions_to_be_implemented,
	regulator,
	owner_id,
	last_review_date,
	due_date,
	status,
	created_at,
	updated_at
FROM
	compliance_registries
WHERE
	%s
	AND id = @compliance_registry_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"compliance_registry_id": complianceRegistryID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance registry: %w", err)
	}

	registry, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ComplianceRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect compliance registry: %w", err)
	}

	*cr = registry

	return nil
}

func (crs *ComplianceRegistries) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	compliance_registries
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
		return 0, fmt.Errorf("cannot count compliance registries: %w", err)
	}

	return count, nil
}

func (crs *ComplianceRegistries) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ComplianceRegistryOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	reference_id,
	area,
	source,
	requirement,
	actions_to_be_implemented,
	regulator,
	owner_id,
	last_review_date,
	due_date,
	status,
	created_at,
	updated_at
FROM
	compliance_registries
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
		return fmt.Errorf("cannot query compliance registries: %w", err)
	}

	registries, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ComplianceRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect compliance registries: %w", err)
	}

	*crs = registries

	return nil
}

func (cr *ComplianceRegistry) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO compliance_registries (
	id,
	tenant_id,
	organization_id,
	reference_id,
	area,
	source,
	requirement,
	actions_to_be_implemented,
	regulator,
	owner_id,
	last_review_date,
	due_date,
	status,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@reference_id,
	@area,
	@source,
	@requirement,
	@actions_to_be_implemented,
	@regulator,
	@owner_id,
	@last_review_date,
	@due_date,
	@status,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                        cr.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           cr.OrganizationID,
		"reference_id":              cr.ReferenceID,
		"area":                      cr.Area,
		"source":                    cr.Source,
		"requirement":               cr.Requirement,
		"actions_to_be_implemented": cr.ActionsToBeImplemented,
		"regulator":                 cr.Regulator,
		"owner_id":                  cr.OwnerID,
		"last_review_date":          cr.LastReviewDate,
		"due_date":                  cr.DueDate,
		"status":                    cr.Status,
		"created_at":                cr.CreatedAt,
		"updated_at":                cr.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert compliance registry: %w", err)
	}

	return nil
}

func (cr *ComplianceRegistry) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE compliance_registries SET
	reference_id = @reference_id,
	area = @area,
	source = @source,
	requirement = @requirement,
	actions_to_be_implemented = @actions_to_be_implemented,
	regulator = @regulator,
	owner_id = @owner_id,
	last_review_date = @last_review_date,
	due_date = @due_date,
	status = @status,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                        cr.ID,
		"reference_id":              cr.ReferenceID,
		"area":                      cr.Area,
		"source":                    cr.Source,
		"requirement":               cr.Requirement,
		"actions_to_be_implemented": cr.ActionsToBeImplemented,
		"regulator":                 cr.Regulator,
		"owner_id":                  cr.OwnerID,
		"last_review_date":          cr.LastReviewDate,
		"due_date":                  cr.DueDate,
		"status":                    cr.Status,
		"updated_at":                cr.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance registry: %w", err)
	}

	return nil
}

func (cr *ComplianceRegistry) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM compliance_registries
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": cr.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance registry: %w", err)
	}

	return nil
}
