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
	TreatmentPlanInternalControl struct {
		TreatmentPlanID   gid.GID      `db:"treatment_plan_id"`
		InternalControlID gid.GID      `db:"internal_control_id"`
		OrganizationID    gid.GID      `db:"organization_id"`
		TenantID          gid.TenantID `db:"tenant_id"`
		CreatedAt         time.Time    `db:"created_at"`
	}

	TreatmentPlanInternalControls []*TreatmentPlanInternalControl
)

func (tpm TreatmentPlanInternalControl) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    treatment_plans_internal_controls (
        treatment_plan_id,
        internal_control_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @treatment_plan_id,
    @internal_control_id,
    @organization_id,
    @tenant_id,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"treatment_plan_id":   tpm.TreatmentPlanID,
		"internal_control_id": tpm.InternalControlID,
		"organization_id":     tpm.OrganizationID,
		"tenant_id":           scope.GetTenantID(),
		"created_at":          tpm.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "treatment_plans_internal_controls_pkey" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert treatment plan internalControl: %w", err)
	}

	return nil
}

func (tpm TreatmentPlanInternalControl) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	treatmentPlanID gid.GID,
	internalControlID gid.GID,
) (bool, error) {
	q := `
DELETE
FROM
    treatment_plans_internal_controls
WHERE
    %s
    AND treatment_plan_id = @treatment_plan_id
    AND internal_control_id = @internal_control_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"treatment_plan_id":   treatmentPlanID,
		"internal_control_id": internalControlID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot delete treatment plan internalControl: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (tpms *TreatmentPlanInternalControls) LoadByTreatmentPlanIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	treatmentPlanIDs []gid.GID,
) error {
	if len(treatmentPlanIDs) == 0 {
		*tpms = nil
		return nil
	}

	q := `
SELECT
    treatment_plan_id,
    internal_control_id,
    organization_id,
    tenant_id,
    created_at
FROM
    treatment_plans_internal_controls
WHERE
    %s
    AND treatment_plan_id = ANY(@treatment_plan_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"treatment_plan_ids": treatmentPlanIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query treatment plan internalControls: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TreatmentPlanInternalControl])
	if err != nil {
		return fmt.Errorf("cannot collect treatment plan internalControls: %w", err)
	}

	*tpms = results

	return nil
}
