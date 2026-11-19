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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	CorrectiveActionPlanInternalControl struct {
		CorrectiveActionPlanID gid.GID      `db:"corrective_action_plan_id"`
		InternalControlID      gid.GID      `db:"internal_control_id"`
		OrganizationID         gid.GID      `db:"organization_id"`
		TenantID               gid.TenantID `db:"tenant_id"`
		CreatedAt              time.Time    `db:"created_at"`
	}
)

func (m CorrectiveActionPlanInternalControl) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    corrective_action_plans_internal_controls (
        corrective_action_plan_id,
        internal_control_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @corrective_action_plan_id,
    @internal_control_id,
    @organization_id,
    @tenant_id,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"corrective_action_plan_id": m.CorrectiveActionPlanID,
		"internal_control_id":       m.InternalControlID,
		"organization_id":           m.OrganizationID,
		"tenant_id":                 scope.GetTenantID(),
		"created_at":                m.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "corrective_action_plans_internal_controls_pkey" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert corrective action plan internal control: %w", err)
	}

	return nil
}

func (m CorrectiveActionPlanInternalControl) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	correctiveActionPlanID gid.GID,
	internalControlID gid.GID,
) error {
	q := `
DELETE
FROM
    corrective_action_plans_internal_controls
WHERE
    %s
    AND corrective_action_plan_id = @corrective_action_plan_id
    AND internal_control_id = @internal_control_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"corrective_action_plan_id": correctiveActionPlanID,
		"internal_control_id":       internalControlID,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete corrective action plan internal control: %w", err)
	}

	return nil
}
