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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ApplicabilityStatement struct {
		ID                     gid.GID   `db:"id"`
		StateOfApplicabilityID gid.GID   `db:"state_of_applicability_id"`
		ControlID              gid.GID   `db:"control_id"`
		OrganizationID         gid.GID   `db:"organization_id"`
		SnapshotID             *gid.GID  `db:"snapshot_id"`
		Applicability          bool      `db:"applicability"`
		Justification          *string   `db:"justification"`
		CreatedAt              time.Time `db:"created_at"`
		UpdatedAt              time.Time `db:"updated_at"`

		// Ordering only.
		SectionTitle string `db:"section_title"`
	}

	ApplicabilityStatements []*ApplicabilityStatement
)

func (s ApplicabilityStatement) CursorKey(orderBy ApplicabilityStatementOrderField) page.CursorKey {
	switch orderBy {
	case ApplicabilityStatementOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	case ApplicabilityStatementOrderFieldControlSectionTitle:
		return page.NewCursorKey(s.ID, s.SectionTitle)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *ApplicabilityStatement) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM applicability_statements WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query applicability statement authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (sac *ApplicabilityStatement) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
WITH stmt AS (
    SELECT
        a.id,
        a.state_of_applicability_id,
        a.control_id,
        a.organization_id,
        a.snapshot_id,
        a.applicability,
        a.justification,
        a.created_at,
        a.updated_at,
        a.tenant_id,
        f.name || ' - ' || c.section_title AS section_title
    FROM
        applicability_statements a
    INNER JOIN
        controls c ON c.id = a.control_id
    INNER JOIN
        frameworks f ON f.id = c.framework_id
    WHERE
        a.%s
        AND a.id = @id
)
SELECT
    id,
    state_of_applicability_id,
    control_id,
    organization_id,
    snapshot_id,
    applicability,
    justification,
    created_at,
    updated_at,
    section_title
FROM
    stmt
WHERE
    %s
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query applicability_statements: %w", err)
	}

	applicabilityStatement, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ApplicabilityStatement])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect applicability statement: %w", err)
	}

	*sac = applicabilityStatement

	return nil
}

func (sac *ApplicabilityStatement) LoadByStateOfApplicabilityIDAndControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	stateOfApplicabilityID gid.GID,
	controlID gid.GID,
) error {
	q := `
WITH current_soa AS (
    SELECT id
    FROM states_of_applicability
    WHERE
        %s
        AND id = @state_of_applicability_id
        AND snapshot_id IS NULL
)
SELECT
    soac.id,
    soac.state_of_applicability_id,
    soac.control_id,
    soac.organization_id,
    soac.snapshot_id,
    soac.applicability,
    soac.justification,
    soac.created_at,
    soac.updated_at,
    f.name || ' - ' || c.section_title AS section_title
FROM
    applicability_statements soac
INNER JOIN
    current_soa ON soac.state_of_applicability_id = current_soa.id
INNER JOIN
    controls c ON c.id = soac.control_id
INNER JOIN
    frameworks f ON f.id = c.framework_id
WHERE
    soac.control_id = @control_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": stateOfApplicabilityID,
		"control_id":                controlID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query applicability_statements: %w", err)
	}

	control, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ApplicabilityStatement])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect applicability statement: %w", err)
	}

	*sac = control
	return nil
}

func (sac *ApplicabilityStatement) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    applicability_statements (
        id,
        state_of_applicability_id,
        control_id,
        organization_id,
        tenant_id,
        snapshot_id,
        applicability,
        justification,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @state_of_applicability_id,
    @control_id,
    @organization_id,
    @tenant_id,
    @snapshot_id,
    @applicability,
    @justification,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"id":                        sac.ID,
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
		"organization_id":           sac.OrganizationID,
		"tenant_id":                 scope.GetTenantID(),
		"snapshot_id":               sac.SnapshotID,
		"applicability":             sac.Applicability,
		"justification":             sac.Justification,
		"created_at":                sac.CreatedAt,
		"updated_at":                sac.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert applicability_statement: %w", err)
	}

	return nil
}

func (sac *ApplicabilityStatement) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE applicability_statements
SET
    applicability = @applicability,
    justification = @justification,
    updated_at = @updated_at
WHERE
    %s
    AND state_of_applicability_id = @state_of_applicability_id
    AND control_id = @control_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
		"applicability":             sac.Applicability,
		"justification":             sac.Justification,
		"updated_at":                sac.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update applicability_statement: %w", err)
	}

	return nil
}

func (sac *ApplicabilityStatement) UpdateByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE applicability_statements
SET
    applicability = @applicability,
    justification = @justification,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":            sac.ID,
		"applicability": sac.Applicability,
		"justification": sac.Justification,
		"updated_at":    sac.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update applicability_statement: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (sac *ApplicabilityStatement) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
WITH current_soa AS (
    SELECT id
    FROM states_of_applicability
    WHERE
        %s
        AND id = @state_of_applicability_id
        AND snapshot_id IS NULL
)
DELETE FROM applicability_statements
WHERE state_of_applicability_id IN (SELECT id FROM current_soa)
    AND control_id = @control_id;
`

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": sac.StateOfApplicabilityID,
		"control_id":                sac.ControlID,
	}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (sac *ApplicabilityStatement) DeleteByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	applicabilityStatementID gid.GID,
) error {
	q := `
DELETE FROM applicability_statements
WHERE
    %s
    AND id = @id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": applicabilityStatementID}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete applicability statement: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (sacs *ApplicabilityStatements) LoadByStateOfApplicabilityID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	stateOfApplicabilityID gid.GID,
	cursor *page.Cursor[ApplicabilityStatementOrderField],
) error {
	q := `
WITH stmt AS (
    SELECT
        a.id,
        a.state_of_applicability_id,
        a.control_id,
        a.organization_id,
        a.snapshot_id,
        a.applicability,
        a.justification,
        a.created_at,
        a.updated_at,
        a.tenant_id,
        f.name || ' - ' || c.section_title AS section_title
    FROM
        applicability_statements a
    INNER JOIN
        controls c ON c.id = a.control_id
    INNER JOIN
        frameworks f ON f.id = c.framework_id
    WHERE
        a.%[1]s
        AND a.state_of_applicability_id = @state_of_applicability_id
)
SELECT
    id,
    state_of_applicability_id,
    control_id,
    organization_id,
    snapshot_id,
    applicability,
    justification,
    created_at,
    updated_at,
    section_title
FROM
    stmt
WHERE
    %[2]s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{
		"state_of_applicability_id": stateOfApplicabilityID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query applicability_statements: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ApplicabilityStatement])
	if err != nil {
		return fmt.Errorf("cannot collect applicability_statements: %w", err)
	}

	*sacs = controls
	return nil
}

func (sacs *ApplicabilityStatements) CountByStateOfApplicabilityID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	stateOfApplicabilityID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    applicability_statements
WHERE
    %s
    AND state_of_applicability_id = @state_of_applicability_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"state_of_applicability_id": stateOfApplicabilityID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count applicability_statements: %w", err)
	}

	return count, nil
}

func (sacs *ApplicabilityStatements) LoadByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor[ApplicabilityStatementOrderField],
) error {
	q := `
WITH soac_ctrl AS (
    SELECT
        soac.id,
        soac.state_of_applicability_id,
        soac.control_id,
        soac.organization_id,
        soac.snapshot_id,
        soac.applicability,
        soac.justification,
        soac.created_at,
        soac.updated_at,
        soac.tenant_id,
        f.name || ' - ' || c.section_title AS section_title
    FROM
        applicability_statements soac
    INNER JOIN
        states_of_applicability soa ON soac.state_of_applicability_id = soa.id
    INNER JOIN
        controls c ON c.id = soac.control_id
    INNER JOIN
        frameworks f ON f.id = c.framework_id
    WHERE
        soac.%[1]s
        AND soac.control_id = @control_id
        AND soa.snapshot_id IS NULL
)
SELECT
    id,
    state_of_applicability_id,
    control_id,
    organization_id,
    snapshot_id,
    applicability,
    justification,
    created_at,
    updated_at,
    section_title
FROM
    soac_ctrl
WHERE
    %[1]s
    AND %[2]s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query applicability_statements: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ApplicabilityStatement])
	if err != nil {
		return fmt.Errorf("cannot collect applicability_statements: %w", err)
	}

	*sacs = controls
	return nil
}

func (sacs *ApplicabilityStatements) CountByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
) (int, error) {
	q := `
WITH soac_ctrl AS (
    SELECT
        soac.id,
        soac.organization_id,
        soac.tenant_id
    FROM
        applicability_statements soac
    INNER JOIN
        states_of_applicability soa ON soac.state_of_applicability_id = soa.id
    WHERE
        soac.%[1]s
        AND soac.control_id = @control_id
        AND soa.snapshot_id IS NULL
)
SELECT
    COUNT(id)
FROM
    soac_ctrl
WHERE
    %[1]s;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}
