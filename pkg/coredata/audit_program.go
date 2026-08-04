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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	AuditProgram struct {
		ID             gid.GID    `db:"id"`
		OrganizationID gid.GID    `db:"organization_id"`
		FrameworkID    gid.GID    `db:"framework_id"`
		Name           string     `db:"name"`
		ValidFrom      *time.Time `db:"valid_from"`
		ValidUntil     *time.Time `db:"valid_until"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}

	AuditPrograms []*AuditProgram
)

func (a *AuditProgram) CursorKey(field AuditProgramOrderField) page.CursorKey {
	switch field {
	case AuditProgramOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	case AuditProgramOrderFieldValidFrom:
		return page.NewCursorKey(a.ID, a.ValidFrom)
	case AuditProgramOrderFieldValidUntil:
		return page.NewCursorKey(a.ID, a.ValidUntil)
	case AuditProgramOrderFieldName:
		return page.NewCursorKey(a.ID, a.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (a *AuditProgram) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM audit_programs WHERE id = ANY(@resource_ids::text[])`

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

func (a *AuditProgram) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	auditProgramID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	framework_id,
	name,
	valid_from,
	valid_until,
	created_at,
	updated_at
FROM
	audit_programs
WHERE
	%s
	AND id = @audit_program_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"audit_program_id": auditProgramID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query audit program: %w", err)
	}

	auditProgram, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuditProgram])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect audit program: %w", err)
	}

	*a = auditProgram

	return nil
}

func (a *AuditPrograms) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	audit_programs
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
		return 0, fmt.Errorf("cannot count audit programs: %w", err)
	}

	return count, nil
}

func (a *AuditPrograms) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AuditProgramOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	framework_id,
	name,
	valid_from,
	valid_until,
	created_at,
	updated_at
FROM
	audit_programs
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
		return fmt.Errorf("cannot query audit programs: %w", err)
	}

	auditPrograms, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AuditProgram])
	if err != nil {
		return fmt.Errorf("cannot collect audit programs: %w", err)
	}

	*a = auditPrograms

	return nil
}

func (a *AuditPrograms) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	auditProgramIDs []gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	framework_id,
	name,
	valid_from,
	valid_until,
	created_at,
	updated_at
FROM
	audit_programs
WHERE
	%s
	AND id = ANY(@audit_program_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"audit_program_ids": auditProgramIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query audit programs: %w", err)
	}

	auditPrograms, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AuditProgram])
	if err != nil {
		return fmt.Errorf("cannot collect audit programs: %w", err)
	}

	*a = auditPrograms

	if len(auditPrograms) != len(gid.NewSet(auditProgramIDs...)) {
		return ErrResourceNotFound
	}

	return nil
}

func (a *AuditProgram) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO audit_programs (
	id,
	tenant_id,
	organization_id,
	framework_id,
	name,
	valid_from,
	valid_until,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@framework_id,
	@name,
	@valid_from,
	@valid_until,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              a.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": a.OrganizationID,
		"framework_id":    a.FrameworkID,
		"name":            a.Name,
		"valid_from":      a.ValidFrom,
		"valid_until":     a.ValidUntil,
		"created_at":      a.CreatedAt,
		"updated_at":      a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert audit program: %w", err)
	}

	return nil
}

func (a *AuditProgram) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE audit_programs
SET
	name = @name,
	valid_from = @valid_from,
	valid_until = @valid_until,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          a.ID,
		"name":        a.Name,
		"valid_from":  a.ValidFrom,
		"valid_until": a.ValidUntil,
		"updated_at":  a.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update audit program: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (a *AuditProgram) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM audit_programs
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": a.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete audit program: %w", err)
	}

	return nil
}
