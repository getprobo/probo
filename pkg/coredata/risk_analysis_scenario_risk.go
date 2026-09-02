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
	"go.probo.inc/probo/pkg/page"
)

type (
	RiskAnalysisScenarioRisk struct {
		RiskAnalysisScenarioID gid.GID   `db:"risk_analysis_scenario_id"`
		RiskID                 gid.GID   `db:"risk_id"`
		CreatedAt              time.Time `db:"created_at"`
	}

	RiskAnalysisScenarioRisks []*RiskAnalysisScenarioRisk
)

func (sr *RiskAnalysisScenarioRisk) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_analysis_scenario_risks (
	tenant_id,
	risk_analysis_scenario_id,
	risk_id,
	created_at
) VALUES (
	@tenant_id,
	@risk_analysis_scenario_id,
	@risk_id,
	@created_at
)
`
	args := pgx.StrictNamedArgs{
		"tenant_id":                 scope.GetTenantID(),
		"risk_analysis_scenario_id": sr.RiskAnalysisScenarioID,
		"risk_id":                   sr.RiskID,
		"created_at":                sr.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "risk_analysis_scenario_risks_pkey" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert risk scenario risk: %w", err)
	}

	return nil
}

func (sr *RiskAnalysisScenarioRisk) Delete(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
DELETE FROM risk_analysis_scenario_risks
WHERE
	%s
	AND risk_analysis_scenario_id = @risk_analysis_scenario_id
	AND risk_id = @risk_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"risk_analysis_scenario_id": sr.RiskAnalysisScenarioID,
		"risk_id":                   sr.RiskID,
	}
	maps.Copy(args, scope.SQLArguments())
	_, err := conn.Exec(ctx, q, args)

	return err
}

func (srs *RiskAnalysisScenarioRisks) LoadByRiskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
) error {
	q := `
SELECT
	risk_analysis_scenario_id,
	risk_id,
	created_at
FROM
	risk_analysis_scenario_risks
WHERE
	%s
	AND risk_id = @risk_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query scenario risks: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAnalysisScenarioRisk])
	if err != nil {
		return fmt.Errorf("cannot collect scenario risks: %w", err)
	}

	*srs = results

	return nil
}

func (srs *RiskAnalysisScenarioRisks) LoadByScenarioIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	scenarioIDs []gid.GID,
) error {
	if len(scenarioIDs) == 0 {
		*srs = nil
		return nil
	}

	q := `
SELECT
	risk_analysis_scenario_id,
	risk_id,
	created_at
FROM
	risk_analysis_scenario_risks
WHERE
	%s
	AND risk_analysis_scenario_id = ANY(@scenario_ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"scenario_ids": scenarioIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query scenario risks: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAnalysisScenarioRisk])
	if err != nil {
		return fmt.Errorf("cannot collect scenario risks: %w", err)
	}

	*srs = results

	return nil
}

func (rs *Risks) LoadByScenarioID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	scenarioID gid.GID,
	cursor *page.Cursor[RiskOrderField],
) error {
	q := `
WITH linked_risks AS (
	SELECT
		risk_id
	FROM
		risk_analysis_scenario_risks
	WHERE
		%s
		AND risk_analysis_scenario_id = @scenario_id
)
SELECT
	id,
	organization_id,
	name,
	description,
	category,
	treatment,
	inherent_likelihood,
	inherent_impact,
	inherent_risk_score,
	residual_likelihood,
	residual_impact,
	residual_risk_score,
	owner_profile_id,
	NULL AS owner_full_name,
	note,
	created_at,
	updated_at
FROM
	risks
WHERE
	%s
	AND id IN (SELECT risk_id FROM linked_risks)
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment(), cursor.SQLFragment())
	args := pgx.NamedArgs{"scenario_id": scenarioID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk scenario risks: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Risk])
	if err != nil {
		return fmt.Errorf("cannot collect risk scenario risks: %w", err)
	}

	*rs = results

	return nil
}

func (rs *Risks) CountByScenarioID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	scenarioID gid.GID,
) (int, error) {
	q := `
WITH linked_risks AS (
	SELECT
		risk_id
	FROM
		risk_analysis_scenario_risks
	WHERE
		%s
		AND risk_analysis_scenario_id = @scenario_id
)
SELECT
	COUNT(id)
FROM
	risks
WHERE
	%s
	AND id IN (SELECT risk_id FROM linked_risks)
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())
	args := pgx.NamedArgs{"scenario_id": scenarioID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk scenario risks: %w", err)
	}

	return count, nil
}

func (srs *RiskAnalysisScenarioRisks) DeleteByOrganizationID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
DELETE FROM risk_analysis_scenario_risks
WHERE
	%s
	AND risk_id IN (
		SELECT id
		FROM risks
		WHERE
			%s
			AND organization_id = @organization_id
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete risk analysis scenario risks: %w", err)
	}

	return nil
}
