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
	RiskAnalysisScenarioThreat struct {
		RiskAnalysisScenarioID gid.GID   `db:"risk_analysis_scenario_id"`
		RiskAnalysisThreatID   gid.GID   `db:"risk_analysis_threat_id"`
		CreatedAt              time.Time `db:"created_at"`
	}

	RiskAnalysisScenarioThreats []*RiskAnalysisScenarioThreat
)

func (st *RiskAnalysisScenarioThreat) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_analysis_scenario_threats (
	tenant_id,
	risk_analysis_scenario_id,
	risk_analysis_threat_id,
	created_at
) VALUES (
	@tenant_id,
	@risk_analysis_scenario_id,
	@risk_analysis_threat_id,
	@created_at
)
`
	args := pgx.StrictNamedArgs{
		"tenant_id":                 scope.GetTenantID(),
		"risk_analysis_scenario_id": st.RiskAnalysisScenarioID,
		"risk_analysis_threat_id":   st.RiskAnalysisThreatID,
		"created_at":                st.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "risk_analysis_scenario_threats_pkey" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert risk scenario threat: %w", err)
	}

	return nil
}

func (st *RiskAnalysisScenarioThreat) Delete(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
DELETE FROM risk_analysis_scenario_threats
WHERE
	%s
	AND risk_analysis_scenario_id = @risk_analysis_scenario_id
	AND risk_analysis_threat_id = @risk_analysis_threat_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"risk_analysis_scenario_id": st.RiskAnalysisScenarioID,
		"risk_analysis_threat_id":   st.RiskAnalysisThreatID,
	}
	maps.Copy(args, scope.SQLArguments())
	_, err := conn.Exec(ctx, q, args)

	return err
}

func (ts *RiskAnalysisThreats) LoadByScenarioID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	scenarioID gid.GID,
	cursor *page.Cursor[RiskAnalysisThreatOrderField],
) error {
	q := `
WITH linked_threats AS (
	SELECT
		risk_analysis_threat_id
	FROM
		risk_analysis_scenario_threats
	WHERE
		%s
		AND risk_analysis_scenario_id = @scenario_id
)
SELECT
	id,
	organization_id,
	risk_analysis_diagram_id,
	process_id,
	name,
	category,
	created_at,
	updated_at
FROM
	risk_analysis_threats
WHERE
	%s
	AND id IN (SELECT risk_analysis_threat_id FROM linked_threats)
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment(), cursor.SQLFragment())
	args := pgx.NamedArgs{"scenario_id": scenarioID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk scenario threats: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAnalysisThreat])
	if err != nil {
		return fmt.Errorf("cannot collect risk scenario threats: %w", err)
	}

	*ts = results

	return nil
}

func (ts *RiskAnalysisThreats) CountByScenarioID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	scenarioID gid.GID,
) (int, error) {
	q := `
WITH linked_threats AS (
	SELECT
		risk_analysis_threat_id
	FROM
		risk_analysis_scenario_threats
	WHERE
		%s
		AND risk_analysis_scenario_id = @scenario_id
)
SELECT
	COUNT(id)
FROM
	risk_analysis_threats
WHERE
	%s
	AND id IN (SELECT risk_analysis_threat_id FROM linked_threats)
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())
	args := pgx.NamedArgs{"scenario_id": scenarioID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk scenario threats: %w", err)
	}

	return count, nil
}
