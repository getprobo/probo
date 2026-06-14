// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	RiskAssessmentScenarioRisk struct {
		RiskAssessmentScenarioID gid.GID   `db:"risk_assessment_scenario_id"`
		RiskID                   gid.GID   `db:"risk_id"`
		CreatedAt                time.Time `db:"created_at"`
	}

	RiskAssessmentScenarioRisks []*RiskAssessmentScenarioRisk
)

func (sr *RiskAssessmentScenarioRisk) Insert(ctx context.Context, conn pg.Tx, predicate Predicater) error {
	q := `
INSERT INTO risk_assessment_scenario_risks (
	tenant_id,
	risk_assessment_scenario_id,
	risk_id,
	created_at
) VALUES (
	@tenant_id,
	@risk_assessment_scenario_id,
	@risk_id,
	@created_at
)
`
	args := pgx.StrictNamedArgs{
		"tenant_id":                   predicate.GetTenantID(),
		"risk_assessment_scenario_id": sr.RiskAssessmentScenarioID,
		"risk_id":                     sr.RiskID,
		"created_at":                  sr.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "risk_assessment_scenario_risks_pkey" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert risk scenario risk: %w", err)
	}

	return nil
}

func (sr *RiskAssessmentScenarioRisk) Delete(ctx context.Context, conn pg.Tx, predicate Predicater) error {
	q := `
DELETE FROM risk_assessment_scenario_risks
WHERE
	%s
	AND risk_assessment_scenario_id = @risk_assessment_scenario_id
	AND risk_id = @risk_id
`
	q = fmt.Sprintf(q, predicate.SQLFragment())
	args := pgx.StrictNamedArgs{
		"risk_assessment_scenario_id": sr.RiskAssessmentScenarioID,
		"risk_id":                     sr.RiskID,
	}
	maps.Copy(args, predicate.SQLArguments())
	_, err := conn.Exec(ctx, q, args)

	return err
}

func (rs *Risks) LoadByScenarioID(
	ctx context.Context,
	conn pg.Querier,
	predicate Predicater,
	scenarioID gid.GID,
	cursor *page.Cursor[RiskOrderField],
) error {
	q := `
WITH linked_risks AS (
	SELECT
		risk_id
	FROM
		risk_assessment_scenario_risks
	WHERE
		%s
		AND risk_assessment_scenario_id = @scenario_id
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
	q = fmt.Sprintf(q, predicate.SQLFragment(), predicate.SQLFragment(), cursor.SQLFragment())
	args := pgx.NamedArgs{"scenario_id": scenarioID}
	maps.Copy(args, predicate.SQLArguments())
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
	predicate Predicater,
	scenarioID gid.GID,
) (int, error) {
	q := `
WITH linked_risks AS (
	SELECT
		risk_id
	FROM
		risk_assessment_scenario_risks
	WHERE
		%s
		AND risk_assessment_scenario_id = @scenario_id
)
SELECT
	COUNT(id)
FROM
	risks
WHERE
	%s
	AND id IN (SELECT risk_id FROM linked_risks)
`
	q = fmt.Sprintf(q, predicate.SQLFragment(), predicate.SQLFragment())
	args := pgx.NamedArgs{"scenario_id": scenarioID}
	maps.Copy(args, predicate.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count risk scenario risks: %w", err)
	}

	return count, nil
}
