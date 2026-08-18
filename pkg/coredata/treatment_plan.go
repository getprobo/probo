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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	TreatmentPlan struct {
		ID                 gid.GID       `db:"id"`
		OrganizationID     gid.GID       `db:"organization_id"`
		RiskID             gid.GID       `db:"risk_id"`
		RiskAnalysisID     gid.GID       `db:"risk_analysis_id"`
		Treatment          RiskTreatment `db:"treatment"`
		OwnerID            gid.GID       `db:"owner_profile_id"`
		InherentLikelihood int           `db:"inherent_likelihood"`
		InherentImpact     int           `db:"inherent_impact"`
		InherentRiskScore  int           `db:"inherent_risk_score"`
		ResidualLikelihood int           `db:"residual_likelihood"`
		ResidualImpact     int           `db:"residual_impact"`
		ResidualRiskScore  int           `db:"residual_risk_score"`
		Category           string        `db:"category"`
		CreatedAt          time.Time     `db:"created_at"`
		UpdatedAt          time.Time     `db:"updated_at"`
	}

	TreatmentPlans []*TreatmentPlan
)

func (tp *TreatmentPlan) CursorKey(orderBy TreatmentPlanOrderField) page.CursorKey {
	switch orderBy {
	case TreatmentPlanOrderFieldCreatedAt:
		return page.NewCursorKey(tp.ID, tp.CreatedAt)
	case TreatmentPlanOrderFieldTreatment:
		return page.NewCursorKey(tp.ID, tp.Treatment)
	case TreatmentPlanOrderFieldCategory:
		return page.NewCursorKey(tp.ID, tp.Category)
	case TreatmentPlanOrderFieldInherentRiskScore:
		return page.NewCursorKey(tp.ID, tp.InherentRiskScore)
	case TreatmentPlanOrderFieldResidualRiskScore:
		return page.NewCursorKey(tp.ID, tp.ResidualRiskScore)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (tp *TreatmentPlan) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM treatment_plans WHERE id = ANY(@resource_ids)`

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

func (tp *TreatmentPlan) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
WITH tps AS (
	SELECT
		tp.id,
		tp.tenant_id,
		tp.organization_id,
		tp.risk_id,
		tp.risk_analysis_id,
		tp.treatment,
		tp.owner_profile_id,
		tp.inherent_likelihood,
		tp.inherent_impact,
		tp.inherent_risk_score,
		tp.residual_likelihood,
		tp.residual_impact,
		tp.residual_risk_score,
		tp.created_at,
		tp.updated_at,
		r.category
	FROM
		treatment_plans tp
	INNER JOIN
		risks r ON r.id = tp.risk_id
)
SELECT
	id,
	organization_id,
	risk_id,
	risk_analysis_id,
	treatment,
	owner_profile_id,
	inherent_likelihood,
	inherent_impact,
	inherent_risk_score,
	residual_likelihood,
	residual_impact,
	residual_risk_score,
	created_at,
	updated_at,
	category
FROM tps
WHERE %s
	AND id = @id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query treatment plan: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TreatmentPlan])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect treatment plan: %w", err)
	}

	*tp = result

	return nil
}

func (tps *TreatmentPlans) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *TreatmentPlanFilter,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM treatment_plans
WHERE %s
	AND organization_id = @organization_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count treatment plans: %w", err)
	}

	return count, nil
}

func (tps *TreatmentPlans) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[TreatmentPlanOrderField],
	filter *TreatmentPlanFilter,
) error {
	q := `
WITH tps AS (
	SELECT
		tp.id,
		tp.tenant_id,
		tp.organization_id,
		tp.risk_id,
		tp.risk_analysis_id,
		tp.treatment,
		tp.owner_profile_id,
		tp.inherent_likelihood,
		tp.inherent_impact,
		tp.inherent_risk_score,
		tp.residual_likelihood,
		tp.residual_impact,
		tp.residual_risk_score,
		tp.created_at,
		tp.updated_at,
		r.category
	FROM
		treatment_plans tp
	INNER JOIN
		risks r ON r.id = tp.risk_id
)
SELECT
	id,
	organization_id,
	risk_id,
	risk_analysis_id,
	treatment,
	owner_profile_id,
	inherent_likelihood,
	inherent_impact,
	inherent_risk_score,
	residual_likelihood,
	residual_impact,
	residual_risk_score,
	created_at,
	updated_at,
	category
FROM tps
WHERE %s
	AND organization_id = @organization_id
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query treatment plans: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TreatmentPlan])
	if err != nil {
		return fmt.Errorf("cannot collect treatment plans: %w", err)
	}

	*tps = results

	return nil
}

func (tps *TreatmentPlans) CountByRiskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
	filter *TreatmentPlanFilter,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM treatment_plans
WHERE %s
	AND risk_id = @risk_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count treatment plans: %w", err)
	}

	return count, nil
}

func (tps *TreatmentPlans) LoadByRiskID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskID gid.GID,
	cursor *page.Cursor[TreatmentPlanOrderField],
	filter *TreatmentPlanFilter,
) error {
	q := `
WITH tps AS (
	SELECT
		tp.id,
		tp.tenant_id,
		tp.organization_id,
		tp.risk_id,
		tp.risk_analysis_id,
		tp.treatment,
		tp.owner_profile_id,
		tp.inherent_likelihood,
		tp.inherent_impact,
		tp.inherent_risk_score,
		tp.residual_likelihood,
		tp.residual_impact,
		tp.residual_risk_score,
		tp.created_at,
		tp.updated_at,
		r.category
	FROM
		treatment_plans tp
	INNER JOIN
		risks r ON r.id = tp.risk_id
)
SELECT
	id,
	organization_id,
	risk_id,
	risk_analysis_id,
	treatment,
	owner_profile_id,
	inherent_likelihood,
	inherent_impact,
	inherent_risk_score,
	residual_likelihood,
	residual_impact,
	residual_risk_score,
	created_at,
	updated_at,
	category
FROM tps
WHERE %s
	AND risk_id = @risk_id
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"risk_id": riskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query treatment plans: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TreatmentPlan])
	if err != nil {
		return fmt.Errorf("cannot collect treatment plans: %w", err)
	}

	*tps = results

	return nil
}

func (tps *TreatmentPlans) CountByMeasureID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	measureID gid.GID,
	filter *TreatmentPlanFilter,
) (int, error) {
	q := `
WITH tps AS (
	SELECT
		tp.id,
		tp.tenant_id,
		tp.inherent_likelihood,
		tp.inherent_impact,
		tp.residual_likelihood,
		tp.residual_impact
	FROM
		treatment_plans tp
	INNER JOIN
		treatment_plans_measures tpm ON tpm.treatment_plan_id = tp.id
	WHERE
		tpm.measure_id = @measure_id
)
SELECT COUNT(id)
FROM tps
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count treatment plans: %w", err)
	}

	return count, nil
}

func (tps *TreatmentPlans) LoadByMeasureID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	measureID gid.GID,
	cursor *page.Cursor[TreatmentPlanOrderField],
	filter *TreatmentPlanFilter,
) error {
	q := `
WITH tps AS (
	SELECT
		tp.id,
		tp.tenant_id,
		tp.organization_id,
		tp.risk_id,
		tp.risk_analysis_id,
		tp.treatment,
		tp.owner_profile_id,
		tp.inherent_likelihood,
		tp.inherent_impact,
		tp.inherent_risk_score,
		tp.residual_likelihood,
		tp.residual_impact,
		tp.residual_risk_score,
		tp.created_at,
		tp.updated_at,
		r.category
	FROM
		treatment_plans tp
	INNER JOIN
		risks r ON r.id = tp.risk_id
	INNER JOIN
		treatment_plans_measures tpm ON tpm.treatment_plan_id = tp.id
	WHERE
		tpm.measure_id = @measure_id
)
SELECT
	id,
	organization_id,
	risk_id,
	risk_analysis_id,
	treatment,
	owner_profile_id,
	inherent_likelihood,
	inherent_impact,
	inherent_risk_score,
	residual_likelihood,
	residual_impact,
	residual_risk_score,
	created_at,
	updated_at,
	category
FROM tps
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query treatment plans: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TreatmentPlan])
	if err != nil {
		return fmt.Errorf("cannot collect treatment plans: %w", err)
	}

	*tps = results

	return nil
}

func (tps *TreatmentPlans) CountByRiskAnalysisID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAnalysisID gid.GID,
	filter *TreatmentPlanFilter,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM treatment_plans
WHERE %s
	AND risk_analysis_id = @risk_analysis_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"risk_analysis_id": riskAnalysisID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count treatment plans: %w", err)
	}

	return count, nil
}

func (tps *TreatmentPlans) LoadByRiskAnalysisID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAnalysisID gid.GID,
	cursor *page.Cursor[TreatmentPlanOrderField],
	filter *TreatmentPlanFilter,
) error {
	q := `
WITH tps AS (
	SELECT
		tp.id,
		tp.tenant_id,
		tp.organization_id,
		tp.risk_id,
		tp.risk_analysis_id,
		tp.treatment,
		tp.owner_profile_id,
		tp.inherent_likelihood,
		tp.inherent_impact,
		tp.inherent_risk_score,
		tp.residual_likelihood,
		tp.residual_impact,
		tp.residual_risk_score,
		tp.created_at,
		tp.updated_at,
		r.category
	FROM
		treatment_plans tp
	INNER JOIN
		risks r ON r.id = tp.risk_id
)
SELECT
	id,
	organization_id,
	risk_id,
	risk_analysis_id,
	treatment,
	owner_profile_id,
	inherent_likelihood,
	inherent_impact,
	inherent_risk_score,
	residual_likelihood,
	residual_impact,
	residual_risk_score,
	created_at,
	updated_at,
	category
FROM tps
WHERE %s
	AND risk_analysis_id = @risk_analysis_id
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"risk_analysis_id": riskAnalysisID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query treatment plans: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TreatmentPlan])
	if err != nil {
		return fmt.Errorf("cannot collect treatment plans: %w", err)
	}

	*tps = results

	return nil
}

func (tps *TreatmentPlans) CountMatrixCellsByRiskAnalysisID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAnalysisID gid.GID,
) ([]*RiskAnalysisMatrixCell, error) {
	implementedSQL := treatmentPlanAllMeasuresImplementedSQL()
	q := `
SELECT
	@score_type_inherent::text AS score_type,
	inherent_likelihood AS likelihood,
	inherent_impact AS impact,
	COUNT(*)::int AS count
FROM treatment_plans
WHERE %s
	AND risk_analysis_id = @risk_analysis_id
GROUP BY inherent_likelihood, inherent_impact

UNION ALL

SELECT
	@score_type_residual::text,
	residual_likelihood,
	residual_impact,
	COUNT(*)::int
FROM treatment_plans
WHERE %s
	AND risk_analysis_id = @risk_analysis_id
GROUP BY residual_likelihood, residual_impact

UNION ALL

SELECT
	@score_type_net::text,
	CASE
		WHEN %s THEN residual_likelihood
		ELSE inherent_likelihood
	END,
	CASE
		WHEN %s THEN residual_impact
		ELSE inherent_impact
	END,
	COUNT(*)::int
FROM treatment_plans
WHERE %s
	AND risk_analysis_id = @risk_analysis_id
GROUP BY 2, 3
`
	q = fmt.Sprintf(
		q,
		scope.SQLFragment(),
		scope.SQLFragment(),
		implementedSQL,
		implementedSQL,
		scope.SQLFragment(),
	)

	args := pgx.StrictNamedArgs{
		"risk_analysis_id":       riskAnalysisID,
		"score_type_inherent":    TreatmentPlanScoreTypeInherent,
		"score_type_residual":    TreatmentPlanScoreTypeResidual,
		"score_type_net":         TreatmentPlanScoreTypeNet,
		"filter_net_implemented": MeasureStateImplemented,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query risk analysis matrix cells: %w", err)
	}

	counts, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[RiskAnalysisMatrixCell])
	if err != nil {
		return nil, fmt.Errorf("cannot collect treatment plan matrix counts: %w", err)
	}

	return counts, nil
}

func (tp *TreatmentPlan) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO treatment_plans (
	id,
	tenant_id,
	organization_id,
	risk_id,
	risk_analysis_id,
	treatment,
	owner_profile_id,
	inherent_likelihood,
	inherent_impact,
	residual_likelihood,
	residual_impact,
	created_at,
	updated_at
)
VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@risk_id,
	@risk_analysis_id,
	@treatment,
	@owner_profile_id,
	@inherent_likelihood,
	@inherent_impact,
	@residual_likelihood,
	@residual_impact,
	@created_at,
	@updated_at
)
RETURNING inherent_risk_score, residual_risk_score
`

	args := pgx.StrictNamedArgs{
		"id":                  tp.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     tp.OrganizationID,
		"risk_id":             tp.RiskID,
		"risk_analysis_id":    tp.RiskAnalysisID,
		"treatment":           tp.Treatment,
		"owner_profile_id":    tp.OwnerID,
		"inherent_likelihood": tp.InherentLikelihood,
		"inherent_impact":     tp.InherentImpact,
		"residual_likelihood": tp.ResidualLikelihood,
		"residual_impact":     tp.ResidualImpact,
		"created_at":          tp.CreatedAt,
		"updated_at":          tp.UpdatedAt,
	}

	err := conn.QueryRow(ctx, q, args).Scan(&tp.InherentRiskScore, &tp.ResidualRiskScore)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "treatment_plans_risk_analysis_unique" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert treatment plan: %w", err)
	}

	return nil
}

func (tp *TreatmentPlan) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE treatment_plans
SET
	treatment = @treatment,
	owner_profile_id = @owner_profile_id,
	inherent_likelihood = @inherent_likelihood,
	inherent_impact = @inherent_impact,
	residual_likelihood = @residual_likelihood,
	residual_impact = @residual_impact,
	updated_at = @updated_at
WHERE %s
	AND id = @id
RETURNING inherent_risk_score, residual_risk_score
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  tp.ID,
		"treatment":           tp.Treatment,
		"owner_profile_id":    tp.OwnerID,
		"inherent_likelihood": tp.InherentLikelihood,
		"inherent_impact":     tp.InherentImpact,
		"residual_likelihood": tp.ResidualLikelihood,
		"residual_impact":     tp.ResidualImpact,
		"updated_at":          tp.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	err := conn.QueryRow(ctx, q, args).Scan(&tp.InherentRiskScore, &tp.ResidualRiskScore)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot update treatment plan: %w", err)
	}

	return nil
}

func (tp *TreatmentPlan) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	id gid.GID,
) error {
	q := `
DELETE FROM treatment_plans WHERE %s AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete treatment plan: %w", err)
	}

	return nil
}
