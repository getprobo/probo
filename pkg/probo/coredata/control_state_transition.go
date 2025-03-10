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

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	ControlStateTransition struct {
		StateTransition[ControlState]

		ControlID gid.GID `db:"control_id"`
	}

	ControlStateTransitions []*ControlStateTransition
)

func (cst ControlStateTransition) CursorKey() page.CursorKey {
	return page.NewCursorKey(cst.ID, cst.CreatedAt)
}

func (cst ControlStateTransition) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    control_state_transitions (
		tenant_id,
        id,
        control_id,
        from_state,
        to_state,
        reason,
        created_at,
        updated_at
    )
VALUES (
	@tenant_id,
    @control_state_transition_id,
    @control_id,
    @from_state,
    @to_state,
    @reason,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                   scope.GetTenantID(),
		"control_state_transition_id": cst.ID,
		"control_id":                  cst.ControlID,
		"from_state":                  cst.FromState,
		"to_state":                    cst.ToState,
		"reason":                      cst.Reason,
		"created_at":                  cst.CreatedAt,
		"updated_at":                  cst.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (cst *ControlStateTransitions) LoadByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	cursor *page.Cursor,
) error {
	q := `
SELECT
    id,
	tenant_id,
    control_id,
    from_state,
    to_state,
    reason,
    created_at,
    updated_at
FROM
    control_state_transitions
WHERE
    %s
    AND control_id = @control_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query control state transitions: %w", err)
	}

	controlStateTransitions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ControlStateTransition])
	if err != nil {
		return fmt.Errorf("cannot collect control state transitions: %w", err)
	}

	*cst = controlStateTransitions

	return nil
}
