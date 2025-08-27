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
	Control struct {
		ID                     gid.GID       `db:"id"`
		SectionTitle           string        `db:"section_title"`
		FrameworkID            gid.GID       `db:"framework_id"`
		Name                   string        `db:"name"`
		Description            string        `db:"description"`
		Status                 ControlStatus `db:"status"`
		ExclusionJustification *string       `db:"exclusion_justification"`
		CreatedAt              time.Time     `db:"created_at"`
		UpdatedAt              time.Time     `db:"updated_at"`
	}

	Controls []*Control

	UpdateControlParams struct {
		Name                   *string
		Description            *string
		SectionTitle           *string
		Status                 *ControlStatus
		ExclusionJustification *string
	}
)

func (c Control) CursorKey(orderBy ControlOrderField) page.CursorKey {
	switch orderBy {
	case ControlOrderFieldCreatedAt:
		return page.CursorKey{ID: c.ID, Value: c.CreatedAt}
	case ControlOrderFieldSectionTitle:
		return page.CursorKey{ID: c.ID, Value: c.SectionTitle}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *Controls) CountByDocumentID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
	filter *ControlFilter,
) (int, error) {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.tenant_id
	FROM
		controls c
	INNER JOIN
		controls_documents cp ON c.id = cp.control_id
	WHERE
		cp.document_id = @document_id
)
SELECT
	COUNT(id)
FROM
	ctrl
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (c *Controls) LoadByDocumentID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
	cursor *page.Cursor[ControlOrderField],
	filter *ControlFilter,
) error {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.section_title,
		c.framework_id,
		c.tenant_id,
		c.name,
		c.description,
		c.status,
		c.exclusion_justification,
		c.created_at,
		c.updated_at,
		c.search_vector
	FROM
		controls c
	INNER JOIN
		controls_documents cp ON c.id = cp.control_id
	WHERE
		cp.document_id = @document_id
)
SELECT
	id,
	section_title,
	framework_id,
	name,
	description,
	status,
	exclusion_justification,
	created_at,
	updated_at
FROM
	ctrl
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect controls: %w", err)
	}

	*c = controls

	return nil
}

func (c *Controls) CountByMeasureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	measureID gid.GID,
	filter *ControlFilter,
) (int, error) {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.tenant_id
	FROM
		controls c
	INNER JOIN
		controls_measures cm ON c.id = cm.control_id
	WHERE
		cm.measure_id = @measure_id
)
SELECT
	COUNT(id)
FROM
	ctrl
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (c *Controls) LoadByMeasureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	measureID gid.GID,
	cursor *page.Cursor[ControlOrderField],
	filter *ControlFilter,
) error {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.section_title,
		c.framework_id,
		c.tenant_id,
		c.name,
		c.description,
		c.status,
		c.exclusion_justification,
		c.created_at,
		c.updated_at,
		c.search_vector
	FROM
		controls c
	INNER JOIN
		controls_measures cm ON c.id = cm.control_id
	WHERE
		cm.measure_id = @measure_id
)
SELECT
	id,
	section_title,
	framework_id,
	name,
	description,
	status,
	exclusion_justification,
	created_at,
	updated_at
FROM
	ctrl
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect controls: %w", err)
	}

	*c = controls

	return nil
}

func (c *Controls) CountByRiskID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
	filter *ControlFilter,
) (int, error) {
	q := `
WITH ctrl AS (
	SELECT DISTINCT
		c.id,
		c.tenant_id
	FROM
		controls c
	LEFT JOIN
		controls_documents cp ON c.id = cp.control_id
	LEFT JOIN
		risks_documents rp ON cp.document_id = rp.document_id
	LEFT JOIN
		controls_measures cm ON c.id = cm.control_id
	LEFT JOIN
		risks_measures rm ON (rm.measure_id = cm.measure_id)
	WHERE
		rp.risk_id = @risk_id OR rm.risk_id = @risk_id
)
SELECT
	COUNT(id)
FROM
	ctrl
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (c *Controls) LoadByRiskID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
	cursor *page.Cursor[ControlOrderField],
	filter *ControlFilter,
) error {
	q := `
WITH ctrl AS (
	SELECT DISTINCT
		c.id,
		c.section_title,
		c.framework_id,
		c.tenant_id,
		c.name,
		c.description,
		c.status,
		c.exclusion_justification,
		c.created_at,
		c.updated_at,
		c.search_vector
	FROM
		controls c
	LEFT JOIN
		controls_documents cp ON c.id = cp.control_id
	LEFT JOIN
		risks_documents rp ON cp.document_id = rp.document_id
	LEFT JOIN
		controls_measures cm ON c.id = cm.control_id
	LEFT JOIN
		risks_measures rm ON (rm.measure_id = cm.measure_id)
	WHERE
		rp.risk_id = @risk_id OR rm.risk_id = @risk_id
)
SELECT
	id,
	section_title,
	framework_id,
	name,
	description,
	status,
	exclusion_justification,
	created_at,
	updated_at
FROM
	ctrl
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect controls: %w", err)
	}

	*c = controls

	return nil
}

func (c *Controls) CountByFrameworkID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	frameworkID gid.GID,
	filter *ControlFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	controls
WHERE %s
	AND framework_id = @framework_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"framework_id": frameworkID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (c *Controls) LoadByFrameworkID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	frameworkID gid.GID,
	cursor *page.Cursor[ControlOrderField],
	filter *ControlFilter,
) error {
	q := `
SELECT
    id,
    section_title,
    framework_id,
    name,
    description,
    status,
	exclusion_justification,
    created_at,
    updated_at
FROM
    controls
WHERE
    %s
    AND framework_id = @framework_id
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"framework_id": frameworkID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect controls: %w", err)
	}

	*c = controls

	return nil
}

func (c *Controls) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *ControlFilter,
) (int, error) {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.tenant_id
	FROM
		controls c
	INNER JOIN
		frameworks f ON c.framework_id = f.id
	WHERE
		f.organization_id = @organization_id
)
SELECT
	COUNT(id)
FROM
	ctrl
WHERE %s
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (c *Controls) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ControlOrderField],
	filter *ControlFilter,
) error {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.section_title,
		c.framework_id,
		c.tenant_id,
		c.name,
		c.description,
		c.status,
		c.exclusion_justification,
		c.created_at,
		c.updated_at,
		c.search_vector
	FROM
		controls c
	INNER JOIN
		frameworks f ON c.framework_id = f.id
	WHERE
		f.organization_id = @organization_id
)
SELECT
	id,
	section_title,
	framework_id,
	name,
	description,
	status,
	exclusion_justification,
	created_at,
	updated_at
FROM
	ctrl
WHERE %s
    AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect controls: %w", err)
	}

	*c = controls

	return nil
}

func (c *Control) LoadByFrameworkIDAndSectionTitle(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	frameworkID gid.GID,
	sectionTitle string,
) error {
	q := `
SELECT
    id,
    section_title,
    framework_id,
    name,
    description,
    status,
    exclusion_justification,
    created_at,
    updated_at
FROM
    controls
WHERE
    %s
    AND framework_id = @framework_id
    AND section_title = @section_title
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"framework_id": frameworkID, "section_title": sectionTitle}
	maps.Copy(args, scope.SQLArguments())
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	control, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect control: %w", err)
	}

	*c = control

	return nil
}

func (c *Control) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
) error {
	q := `
SELECT
    id,
    section_title,
    framework_id,
    name,
    description,
    status,
    exclusion_justification,
    created_at,
    updated_at
FROM
    controls
WHERE
    %s
    AND id = @control_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	control, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect control: %w", err)
	}

	*c = control

	return nil
}

func (c Control) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    controls (
        tenant_id,
        id,
        framework_id,
        section_title,
        name,
        description,
        status,
        exclusion_justification,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @control_id,
    @framework_id,
    @section_title,
    @name,
    @description,
	@status,
	@exclusion_justification,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":               scope.GetTenantID(),
		"control_id":              c.ID,
		"framework_id":            c.FrameworkID,
		"section_title":           c.SectionTitle,
		"name":                    c.Name,
		"description":             c.Description,
		"status":                  c.Status,
		"exclusion_justification": c.ExclusionJustification,
		"created_at":              c.CreatedAt,
		"updated_at":              c.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (c Control) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE
FROM
    controls
WHERE
    %s
    AND id = @control_id;
`

	args := pgx.StrictNamedArgs{"control_id": c.ID}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (c *Control) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	params UpdateControlParams,
) error {
	q := `
UPDATE controls SET
    name = COALESCE(@name, name),
    description = COALESCE(@description, description),
    section_title = COALESCE(@section_title, section_title),
	status = COALESCE(@status, status),
	exclusion_justification = COALESCE(@exclusion_justification, exclusion_justification),
    updated_at = @updated_at
WHERE %s
    AND id = @control_id
RETURNING
    id,
    framework_id,
    name,
    description,
	section_title,
	status,
	exclusion_justification,
    created_at,
    updated_at
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"control_id":              c.ID,
		"section_title":           params.SectionTitle,
		"status":                  params.Status,
		"exclusion_justification": params.ExclusionJustification,
		"updated_at":              time.Now(),
	}

	if params.Name != nil {
		args["name"] = *params.Name
	}
	if params.Description != nil {
		args["description"] = *params.Description
	}

	if params.Status != nil {
		args["status"] = *params.Status
	}
	if params.ExclusionJustification != nil {
		args["exclusion_justification"] = *params.ExclusionJustification
	}

	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	control, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect control: %w", err)
	}

	*c = control

	return nil
}

func (c *Controls) CountByAuditID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	auditID gid.GID,
	filter *ControlFilter,
) (int, error) {
	q := `
WITH ctrl AS (
		SELECT
			c.id,
			c.tenant_id,
			c.search_vector
		FROM
			controls c
		INNER JOIN
			controls_audits ca ON c.id = ca.control_id
		WHERE
			ca.audit_id = @audit_id
	)
	SELECT
		COUNT(id)
	FROM
		ctrl
	WHERE %s
		AND %s
	`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"audit_id": auditID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (c *Controls) LoadByAuditID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	auditID gid.GID,
	cursor *page.Cursor[ControlOrderField],
	filter *ControlFilter,
) error {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.section_title,
		c.framework_id,
		c.tenant_id,
		c.name,
		c.description,
		c.status,
		c.exclusion_justification,
		c.created_at,
		c.updated_at,
		c.search_vector
	FROM
		controls c
	INNER JOIN
		controls_audits ca ON c.id = ca.control_id
	WHERE
		ca.audit_id = @audit_id
)
SELECT
	id,
	section_title,
	framework_id,
	name,
	description,
	status,
	exclusion_justification,
	created_at,
	updated_at
FROM
	ctrl
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"audit_id": auditID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect controls: %w", err)
	}

	*c = controls

	return nil
}

func (c *Controls) CountBySnapshotID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	snapshotID gid.GID,
	filter *ControlFilter,
) (int, error) {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.tenant_id,
		c.search_vector
	FROM
		controls c
	INNER JOIN
		controls_snapshots cs ON c.id = cs.control_id
	WHERE
		cs.snapshot_id = @snapshot_id
)
SELECT
	COUNT(id)
FROM
	ctrl
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"snapshot_id": snapshotID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (c *Controls) LoadBySnapshotID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	snapshotID gid.GID,
	cursor *page.Cursor[ControlOrderField],
	filter *ControlFilter,
) error {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.section_title,
		c.framework_id,
		c.tenant_id,
		c.name,
		c.description,
		c.status,
		c.exclusion_justification,
		c.created_at,
		c.updated_at,
		c.search_vector
	FROM
		controls c
	INNER JOIN
		controls_snapshots cs ON c.id = cs.control_id
	WHERE
		cs.snapshot_id = @snapshot_id
)
SELECT
	id,
	section_title,
	framework_id,
	name,
	description,
	status,
	exclusion_justification,
	created_at,
	updated_at
FROM
	ctrl
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"snapshot_id": snapshotID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query controls: %w", err)
	}

	controls, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Control])
	if err != nil {
		return fmt.Errorf("cannot collect controls: %w", err)
	}

	*c = controls

	return nil
}
