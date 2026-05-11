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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	RiskAssessmentScenarioThreat struct {
		RiskAssessmentScenarioID gid.GID   `db:"risk_assessment_scenario_id"`
		RiskAssessmentThreatID   gid.GID   `db:"risk_assessment_threat_id"`
		CreatedAt                time.Time `db:"created_at"`
	}

	RiskAssessmentScenarioThreats []*RiskAssessmentScenarioThreat
)

func (st *RiskAssessmentScenarioThreat) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO risk_assessment_scenario_threats (
	tenant_id,
	risk_assessment_scenario_id,
	risk_assessment_threat_id,
	created_at
) VALUES (
	@tenant_id,
	@risk_assessment_scenario_id,
	@risk_assessment_threat_id,
	@created_at
)
`
	args := pgx.StrictNamedArgs{
		"tenant_id":                   scope.GetTenantID(),
		"risk_assessment_scenario_id": st.RiskAssessmentScenarioID,
		"risk_assessment_threat_id":   st.RiskAssessmentThreatID,
		"created_at":                  st.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return ErrResourceAlreadyExists
		}
		return fmt.Errorf("cannot insert risk scenario threat: %w", err)
	}
	return nil
}

func (st *RiskAssessmentScenarioThreat) Delete(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
DELETE FROM risk_assessment_scenario_threats
WHERE
	%s
	AND risk_assessment_scenario_id = @risk_assessment_scenario_id
	AND risk_assessment_threat_id = @risk_assessment_threat_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"risk_assessment_scenario_id": st.RiskAssessmentScenarioID,
		"risk_assessment_threat_id":   st.RiskAssessmentThreatID,
	}
	maps.Copy(args, scope.SQLArguments())
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (ts *RiskAssessmentThreats) LoadByScenarioID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	scenarioID gid.GID,
	cursor *page.Cursor[RiskAssessmentThreatOrderField],
) error {
	q := `
WITH linked_threats AS (
	SELECT
		risk_assessment_threat_id
	FROM
		risk_assessment_scenario_threats
	WHERE
		%s
		AND risk_assessment_scenario_id = @scenario_id
)
SELECT
	id,
	organization_id,
	risk_assessment_scope_id,
	process_id,
	name,
	category,
	created_at,
	updated_at
FROM
	risk_assessment_threats
WHERE
	%s
	AND id IN (SELECT risk_assessment_threat_id FROM linked_threats)
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
	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAssessmentThreat])
	if err != nil {
		return fmt.Errorf("cannot collect risk scenario threats: %w", err)
	}
	*ts = results
	return nil
}

func (ts *RiskAssessmentThreats) CountByScenarioID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	scenarioID gid.GID,
) (int, error) {
	q := `
WITH linked_threats AS (
	SELECT
		risk_assessment_threat_id
	FROM
		risk_assessment_scenario_threats
	WHERE
		%s
		AND risk_assessment_scenario_id = @scenario_id
)
SELECT
	COUNT(id)
FROM
	risk_assessment_threats
WHERE
	%s
	AND id IN (SELECT risk_assessment_threat_id FROM linked_threats)
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
