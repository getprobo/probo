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
	CorrectiveActionPlan struct {
		ID             gid.GID    `db:"id"`
		OrganizationID gid.GID    `db:"organization_id"`
		FindingID      gid.GID    `db:"finding_id"`
		Title          string     `db:"title"`
		Description    *string    `db:"description"`
		OwnerID        *gid.GID   `db:"owner_profile_id"`
		DueDate        *time.Time `db:"due_date"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}

	CorrectiveActionPlans []*CorrectiveActionPlan
)

func (p *CorrectiveActionPlan) CursorKey(field CorrectiveActionPlanOrderField) page.CursorKey {
	switch field {
	case CorrectiveActionPlanOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	case CorrectiveActionPlanOrderFieldDueDate:
		return page.NewCursorKey(p.ID, p.DueDate)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (p *CorrectiveActionPlan) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM corrective_action_plans WHERE id = ANY(@resource_ids::text[])`

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

func (p *CorrectiveActionPlan) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	correctiveActionPlanID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	finding_id,
	title,
	description,
	owner_profile_id,
	due_date,
	created_at,
	updated_at
FROM
	corrective_action_plans
WHERE
	%s
	AND id = @corrective_action_plan_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"corrective_action_plan_id": correctiveActionPlanID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query corrective action plan: %w", err)
	}

	plan, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CorrectiveActionPlan])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect corrective action plan: %w", err)
	}

	*p = plan

	return nil
}

func (p *CorrectiveActionPlan) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO corrective_action_plans (
	id,
	tenant_id,
	organization_id,
	finding_id,
	title,
	description,
	owner_profile_id,
	due_date,
	created_at,
	updated_at
)
VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@finding_id,
	@title,
	@description,
	@owner_profile_id,
	@due_date,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               p.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  p.OrganizationID,
		"finding_id":       p.FindingID,
		"title":            p.Title,
		"description":      p.Description,
		"owner_profile_id": p.OwnerID,
		"due_date":         p.DueDate,
		"created_at":       p.CreatedAt,
		"updated_at":       p.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert corrective action plan: %w", err)
	}

	return nil
}

func (p *CorrectiveActionPlan) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE corrective_action_plans
SET
	title = @title,
	description = @description,
	owner_profile_id = @owner_profile_id,
	due_date = @due_date,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":               p.ID,
		"title":            p.Title,
		"description":      p.Description,
		"owner_profile_id": p.OwnerID,
		"due_date":         p.DueDate,
		"updated_at":       p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update corrective action plan: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (p *CorrectiveActionPlan) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM corrective_action_plans
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": p.ID}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete corrective action plan: %w", err)
	}

	return nil
}

func (ps *CorrectiveActionPlans) LoadByFindingID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	findingID gid.GID,
	cursor *page.Cursor[CorrectiveActionPlanOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	finding_id,
	title,
	description,
	owner_profile_id,
	due_date,
	created_at,
	updated_at
FROM
	corrective_action_plans
WHERE
	%s
	AND finding_id = @finding_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"finding_id": findingID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query corrective action plans: %w", err)
	}

	plans, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CorrectiveActionPlan])
	if err != nil {
		return fmt.Errorf("cannot collect corrective action plans: %w", err)
	}

	*ps = plans

	return nil
}

func (ps *CorrectiveActionPlans) CountByFindingID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	findingID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	corrective_action_plans
WHERE
	%s
	AND finding_id = @finding_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"finding_id": findingID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count corrective action plans: %w", err)
	}

	return count, nil
}

func (ps *CorrectiveActionPlans) LoadByInternalControlID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	internalControlID gid.GID,
	cursor *page.Cursor[CorrectiveActionPlanOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	finding_id,
	title,
	description,
	owner_profile_id,
	due_date,
	created_at,
	updated_at
FROM
	corrective_action_plans
WHERE
	%s
	AND id IN (
		SELECT
			corrective_action_plan_id
		FROM
			corrective_action_plans_internal_controls
		WHERE
			internal_control_id = @internal_control_id
	)
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"internal_control_id": internalControlID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query corrective action plans: %w", err)
	}

	plans, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CorrectiveActionPlan])
	if err != nil {
		return fmt.Errorf("cannot collect corrective action plans: %w", err)
	}

	*ps = plans

	return nil
}

func (ps *CorrectiveActionPlans) CountByInternalControlID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	internalControlID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	corrective_action_plans
WHERE
	%s
	AND id IN (
		SELECT
			corrective_action_plan_id
		FROM
			corrective_action_plans_internal_controls
		WHERE
			internal_control_id = @internal_control_id
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"internal_control_id": internalControlID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count corrective action plans: %w", err)
	}

	return count, nil
}
