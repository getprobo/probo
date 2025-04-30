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
		ID          gid.GID      `db:"id"`
		ReferenceID string       `db:"reference_id"`
		TenantID    gid.TenantID `db:"tenant_id"`
		FrameworkID gid.GID      `db:"framework_id"`
		Name        string       `db:"name"`
		Description string       `db:"description"`
		CreatedAt   time.Time    `db:"created_at"`
		UpdatedAt   time.Time    `db:"updated_at"`
	}

	Controls []*Control

	UpdateControlParams struct {
		ExpectedVersion int
		Name            *string
		Description     *string
	}
)

func (c Control) CursorKey(orderBy ControlOrderField) page.CursorKey {
	switch orderBy {
	case ControlOrderFieldCreatedAt:
		return page.CursorKey{ID: c.ID, Value: c.CreatedAt}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *Controls) LoadByPolicyID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	policyID gid.GID,
	cursor *page.Cursor[ControlOrderField],
) error {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.reference_id,
		c.framework_id,
		c.tenant_id,
		c.name,
		c.description,
		c.created_at,
		c.updated_at
	FROM
		controls c
	INNER JOIN
		controls_policies cp ON c.id = cp.control_id
	WHERE
		cp.policy_id = @policy_id
)
SELECT
	id,
	reference_id,
	framework_id,
	tenant_id,
	name,
	description,
	created_at,
	updated_at
FROM
	ctrl
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"policy_id": policyID}
	maps.Copy(args, scope.SQLArguments())
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

func (c *Controls) LoadByMeasureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	measureID gid.GID,
	cursor *page.Cursor[ControlOrderField],
) error {
	q := `
WITH ctrl AS (
	SELECT
		c.id,
		c.reference_id,
		c.framework_id,
		c.tenant_id,
		c.name,
		c.description,
		c.created_at,
		c.updated_at
	FROM
		controls c
	INNER JOIN
		controls_measures cm ON c.id = cm.control_id
	WHERE
		cm.measure_id = @measure_id
)
SELECT
	id,
	reference_id,
	framework_id,
    tenant_id,
	name,
	description,
	created_at,
	updated_at
FROM
	ctrl
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())

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

func (c *Controls) LoadByRiskID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
	cursor *page.Cursor[ControlOrderField],
) error {
	q := `
WITH ctrl AS (
	SELECT DISTINCT
		c.id,
		c.reference_id,
		c.framework_id,
		c.tenant_id,
		c.name,
		c.description,
		c.created_at,
		c.updated_at
	FROM
		controls c
	LEFT JOIN
		controls_policies cp ON c.id = cp.control_id
	LEFT JOIN
		risks_policies rp ON cp.policy_id = rp.policy_id
	LEFT JOIN
		controls_measures cm ON c.id = cm.control_id
	LEFT JOIN
		risks_measures rm ON (rm.measure_id = cm.measure_id)
	WHERE
		rp.risk_id = @risk_id OR rm.risk_id = @risk_id
)
SELECT
	id,
	reference_id,
	framework_id,
    tenant_id,
	name,
	description,
	created_at,
	updated_at
FROM
	ctrl
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
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

func (c *Controls) LoadByFrameworkID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	frameworkID gid.GID,
	cursor *page.Cursor[ControlOrderField],
) error {
	q := `
SELECT
    id,
    reference_id,
    framework_id,
    tenant_id,
    name,
    description,
    created_at,
    updated_at
FROM
    controls
WHERE
    %s
    AND framework_id = @framework_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"framework_id": frameworkID}
	maps.Copy(args, scope.SQLArguments())

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

func (c *Control) LoadByFrameworkIDAndReferenceID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	frameworkID gid.GID,
	referenceID string,
) error {
	q := `
SELECT
    id,
    reference_id,
    framework_id,
    tenant_id,
    name,
    description,
    created_at,
    updated_at
FROM
    controls
WHERE
    %s
    AND framework_id = @framework_id
    AND reference_id = @reference_id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"framework_id": frameworkID, "reference_id": referenceID}
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
    reference_id,
    framework_id,
    tenant_id,
    name,
    description,
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
        reference_id,
        name,
        description,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @control_id,
    @framework_id,
    @reference_id,
    @name,
    @description,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":    scope.GetTenantID(),
		"control_id":   c.ID,
		"framework_id": c.FrameworkID,
		"reference_id": c.ReferenceID,
		"name":         c.Name,
		"description":  c.Description,
		"created_at":   c.CreatedAt,
		"updated_at":   c.UpdatedAt,
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
    updated_at = @updated_at
WHERE %s
    AND id = @control_id
RETURNING 
    id,
    framework_id,
    tenant_id,
    name,
    description,
    created_at,
    updated_at
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"control_id":       c.ID,
		"expected_version": params.ExpectedVersion,
		"updated_at":       time.Now(),
	}

	if params.Name != nil {
		args["name"] = *params.Name
	}
	if params.Description != nil {
		args["description"] = *params.Description
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
