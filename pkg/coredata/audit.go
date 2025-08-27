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
	Audit struct {
		ID                gid.GID    `db:"id"`
		Name              *string    `db:"name"`
		OrganizationID    gid.GID    `db:"organization_id"`
		FrameworkID       gid.GID    `db:"framework_id"`
		ValidFrom         *time.Time `db:"valid_from"`
		ValidUntil        *time.Time `db:"valid_until"`
		State             AuditState `db:"state"`
		ShowOnTrustCenter bool       `db:"show_on_trust_center"`
		CreatedAt         time.Time  `db:"created_at"`
		UpdatedAt         time.Time  `db:"updated_at"`
	}

	Audits []*Audit
)

func (a *Audit) CursorKey(field AuditOrderField) page.CursorKey {
	switch field {
	case AuditOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	case AuditOrderFieldValidFrom:
		return page.NewCursorKey(a.ID, a.ValidFrom)
	case AuditOrderFieldValidUntil:
		return page.NewCursorKey(a.ID, a.ValidUntil)
	case AuditOrderFieldState:
		return page.NewCursorKey(a.ID, a.State)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (a *Audit) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	auditID gid.GID,
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	framework_id,
	valid_from,
	valid_until,
	state,
	show_on_trust_center,
	created_at,
	updated_at
FROM
	audits
WHERE
	%s
	AND id = @audit_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"audit_id": auditID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query audit: %w", err)
	}

	audit, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Audit])
	if err != nil {
		return fmt.Errorf("cannot collect audit: %w", err)
	}

	*a = audit

	return nil
}

func (a *Audits) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	audits
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
		return 0, fmt.Errorf("cannot count audits: %w", err)
	}

	return count, nil
}

func (a *Audits) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AuditOrderField],
	filter *AuditFilter,
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	framework_id,
	valid_from,
	valid_until,
	state,
	show_on_trust_center,
	created_at,
	updated_at
FROM
	audits
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query audits: %w", err)
	}

	audits, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Audit])
	if err != nil {
		return fmt.Errorf("cannot collect audits: %w", err)
	}

	*a = audits

	return nil
}

func (a *Audit) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO audits (
	id,
	name,
	tenant_id,
	organization_id,
	framework_id,
	valid_from,
	valid_until,
	state,
	show_on_trust_center,
	created_at,
	updated_at
) VALUES (
	@id,
	@name,
	@tenant_id,
	@organization_id,
	@framework_id,
	@valid_from,
	@valid_until,
	@state,
	@show_on_trust_center,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                   a.ID,
		"name":                 a.Name,
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      a.OrganizationID,
		"framework_id":         a.FrameworkID,
		"valid_from":           a.ValidFrom,
		"valid_until":          a.ValidUntil,
		"state":                a.State,
		"show_on_trust_center": a.ShowOnTrustCenter,
		"created_at":           a.CreatedAt,
		"updated_at":           a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert audit: %w", err)
	}

	return nil
}

func (a *Audit) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE audits
SET
	name = @name,
	valid_from = @valid_from,
	valid_until = @valid_until,
	state = @state,
	show_on_trust_center = @show_on_trust_center,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                   a.ID,
		"name":                 a.Name,
		"valid_from":           a.ValidFrom,
		"valid_until":          a.ValidUntil,
		"state":                a.State,
		"show_on_trust_center": a.ShowOnTrustCenter,
		"updated_at":           a.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update audit: %w", err)
	}

	return nil
}

func (a *Audit) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM audits
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": a.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete audit: %w", err)
	}

	return nil
}

func (a *Audits) CountByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
) (int, error) {
	q := `
WITH audits_by_control AS (
		SELECT
			a.id,
			a.tenant_id
		FROM
			audits a
		INNER JOIN
			controls_audits ca ON a.id = ca.audit_id
		WHERE
			ca.control_id = @control_id
	)
	SELECT
		COUNT(id)
	FROM
		audits_by_control
	WHERE %s
	`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (a *Audits) LoadByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor[AuditOrderField],
) error {
	q := `
WITH audits_by_control AS (
	SELECT
		a.id,
		a.tenant_id,
		a.name,
		a.organization_id,
		a.framework_id,
		a.valid_from,
		a.valid_until,
		a.state,
		a.show_on_trust_center,
		a.created_at,
		a.updated_at
	FROM
		audits a
	INNER JOIN
		controls_audits ca ON a.id = ca.audit_id
	WHERE
		ca.control_id = @control_id
)
SELECT
	id,
	name,
	organization_id,
	framework_id,
	valid_from,
	valid_until,
	state,
	show_on_trust_center,
	created_at,
	updated_at
FROM
	audits_by_control
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query audits: %w", err)
	}

	audits, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Audit])
	if err != nil {
		return fmt.Errorf("cannot collect audits: %w", err)
	}

	*a = audits

	return nil
}
