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

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	ControlObligation struct {
		ControlID    gid.GID   `db:"control_id"`
		ObligationID gid.GID   `db:"obligation_id"`
		CreatedAt    time.Time `db:"created_at"`
	}

	ControlObligations []*ControlObligation
)

func (co ControlObligation) Upsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    controls_obligations (
        control_id,
        obligation_id,
        tenant_id,
        created_at
    )
VALUES (
    @control_id,
    @obligation_id,
    @tenant_id,
    @created_at
)
ON CONFLICT (control_id, obligation_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"control_id":    co.ControlID,
		"obligation_id": co.ObligationID,
		"tenant_id":     scope.GetTenantID(),
		"created_at":    co.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (co ControlObligation) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	obligationID gid.GID,
) error {
	q := `
DELETE
FROM
    controls_obligations
WHERE
    %s
    AND control_id = @control_id
    AND obligation_id = @obligation_id;
`

	args := pgx.StrictNamedArgs{
		"control_id":    controlID,
		"obligation_id": obligationID,
	}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (cos *ControlObligations) LoadByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
) error {
	q := `
SELECT
    control_id,
    obligation_id,
    created_at
FROM
    controls_obligations
WHERE
    %s
    AND control_id = @control_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query control_obligations: %w", err)
	}

	controlObligations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ControlObligation])
	if err != nil {
		return fmt.Errorf("cannot collect control_obligations: %w", err)
	}

	*cos = controlObligations
	return nil
}

func (cos *ControlObligations) LoadByObligationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	obligationID gid.GID,
) error {
	q := `
SELECT
    control_id,
    obligation_id,
    created_at
FROM
    controls_obligations
WHERE
    %s
    AND obligation_id = @obligation_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"obligation_id": obligationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query control_obligations: %w", err)
	}

	controlObligations, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ControlObligation])
	if err != nil {
		return fmt.Errorf("cannot collect control_obligations: %w", err)
	}

	*cos = controlObligations
	return nil
}

func (cos *ControlObligations) CountByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	filter *ControlObligationFilter,
) (int, error) {
	q := `
WITH control_obls AS (
	SELECT
		co.control_id,
		o.type,
		o.tenant_id
	FROM
		controls_obligations co
	INNER JOIN
		obligations o ON co.obligation_id = o.id
	WHERE
		co.control_id = @control_id
)
SELECT
	COUNT(*)
FROM
	control_obls
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count control obligations: %w", err)
	}

	return count, nil
}
