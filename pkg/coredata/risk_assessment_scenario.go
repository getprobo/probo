// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	RiskAssessmentScenario struct {
		ID                    gid.GID   `db:"id"`
		OrganizationID        gid.GID   `db:"organization_id"`
		RiskAssessmentScopeID gid.GID   `db:"risk_assessment_scope_id"`
		Name                  string    `db:"name"`
		Description           *string   `db:"description"`
		CreatedAt             time.Time `db:"created_at"`
		UpdatedAt             time.Time `db:"updated_at"`
	}

	RiskAssessmentScenarios []*RiskAssessmentScenario
)

func (s *RiskAssessmentScenario) CursorKey(orderBy RiskAssessmentScenarioOrderField) page.CursorKey {
	switch orderBy {
	case RiskAssessmentScenarioOrderFieldCreatedAt:
		return page.CursorKey{ID: s.ID, Value: s.CreatedAt}
	case RiskAssessmentScenarioOrderFieldName:
		return page.CursorKey{ID: s.ID, Value: s.Name}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *RiskAssessmentScenario) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM risk_assessment_scenarios WHERE id = ANY(@resource_ids::text[])`

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

func (ss *RiskAssessmentScenarios) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[RiskAssessmentScenarioOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	name,
	description,
	created_at,
	updated_at
FROM
	risk_assessment_scenarios
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())
	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk scenarios: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentScenario])
	if err != nil {
		return fmt.Errorf("cannot collect risk scenarios: %w", err)
	}

	*ss = results

	return nil
}

func (ss *RiskAssessmentScenarios) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	risk_assessment_scenarios
WHERE
	%s
	AND organization_id = @organization_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk scenarios: %w", err)
	}

	return count, nil
}

func (ss *RiskAssessmentScenarios) LoadByRiskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
	cursor *page.Cursor[RiskAssessmentScenarioOrderField],
) error {
	q := `
WITH linked_scenarios AS (
	SELECT
		risk_assessment_scenario_id
	FROM
		risk_assessment_scenario_risks
	WHERE
		%s
		AND risk_id = @risk_id
)
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	name,
	description,
	created_at,
	updated_at
FROM
	risk_assessment_scenarios
WHERE
	%s
	AND id IN (SELECT risk_assessment_scenario_id FROM linked_scenarios)
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment(), cursor.SQLFragment())
	args := pgx.NamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk scenarios: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentScenario])
	if err != nil {
		return fmt.Errorf("cannot collect risk scenarios: %w", err)
	}

	*ss = results

	return nil
}

func (ss *RiskAssessmentScenarios) CountByRiskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
) (int, error) {
	q := `
WITH linked_scenarios AS (
	SELECT
		risk_assessment_scenario_id
	FROM
		risk_assessment_scenario_risks
	WHERE
		%s
		AND risk_id = @risk_id
)
SELECT
	COUNT(id)
FROM
	risk_assessment_scenarios
WHERE
	%s
	AND id IN (SELECT risk_assessment_scenario_id FROM linked_scenarios)
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())
	args := pgx.NamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk scenarios: %w", err)
	}

	return count, nil
}

func (ss *RiskAssessmentScenarios) LoadByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
	cursor *page.Cursor[RiskAssessmentScenarioOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	name,
	description,
	created_at,
	updated_at
FROM
	risk_assessment_scenarios
WHERE
	%s
	AND risk_assessment_scope_id = @risk_assessment_scope_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_scope_id": riskAssessmentScopeID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk scenarios: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentScenario])
	if err != nil {
		return fmt.Errorf("cannot collect risk scenarios: %w", err)
	}

	*ss = results

	return nil
}

func (ss *RiskAssessmentScenarios) CountByRiskAssessmentScopeID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAssessmentScopeID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	risk_assessment_scenarios
WHERE
	%s
	AND risk_assessment_scope_id = @risk_assessment_scope_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.NamedArgs{"risk_assessment_scope_id": riskAssessmentScopeID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk scenarios: %w", err)
	}

	return count, nil
}

func (s *RiskAssessmentScenario) LoadByID(ctx context.Context, conn pg.Querier, scope Scoper, id gid.GID) error {
	q := `
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	name,
	description,
	created_at,
	updated_at
FROM
	risk_assessment_scenarios
WHERE
	%s
	AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk scenario: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[RiskAssessmentScenario])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect risk scenario: %w", err)
	}

	*s = result

	return nil
}

func (s *RiskAssessmentScenario) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_assessment_scenarios (
	id,
	tenant_id,
	organization_id,
	risk_assessment_scope_id,
	name,
	description,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@risk_assessment_scope_id,
	@name,
	@description,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                       s.ID,
		"tenant_id":                scope.GetTenantID(),
		"organization_id":          s.OrganizationID,
		"risk_assessment_scope_id": s.RiskAssessmentScopeID,
		"name":                     s.Name,
		"description":              s.Description,
		"created_at":               s.CreatedAt,
		"updated_at":               s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert risk scenario: %w", err)
	}

	return nil
}

func (s *RiskAssessmentScenario) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE risk_assessment_scenarios
SET
	name = @name,
	description = @description,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": s.ID, "name": s.Name, "description": s.Description, "updated_at": s.UpdatedAt}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update risk scenario: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *RiskAssessmentScenario) Delete(ctx context.Context, conn pg.Tx, scope Scoper, id gid.GID) error {
	q := `
DELETE FROM risk_assessment_scenarios
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())
	_, err := conn.Exec(ctx, q, args)

	return err
}
