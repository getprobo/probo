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
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	TreatmentPlanEvent struct {
		OrganizationID         gid.GID                `db:"organization_id"`
		RiskAnalysisID         gid.GID                `db:"risk_analysis_id"`
		TreatmentPlanID        gid.GID                `db:"treatment_plan_id"`
		EventType              TreatmentPlanEventType `db:"event_type"`
		RiskID                 gid.GID                `db:"risk_id"`
		OwnerID                gid.GID                `db:"owner_profile_id"`
		Treatment              RiskTreatment          `db:"treatment"`
		InherentLikelihood     int                    `db:"inherent_likelihood"`
		InherentImpact         int                    `db:"inherent_impact"`
		ResidualLikelihood     int                    `db:"residual_likelihood"`
		ResidualImpact         int                    `db:"residual_impact"`
		MeasureIDs             []string               `db:"measure_ids"`
		Category               string                 `db:"category"`
		TreatmentPlanCreatedAt time.Time              `db:"treatment_plan_created_at"`
		TreatmentPlanUpdatedAt time.Time              `db:"treatment_plan_updated_at"`
		CreatedAt              time.Time              `db:"created_at"`
	}

	TreatmentPlanEvents []*TreatmentPlanEvent
)

func MeasureIDStrings(ids []gid.GID) []string {
	if len(ids) == 0 {
		return []string{}
	}

	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}

	slices.Sort(out)

	return out
}

func NewTreatmentPlanEvent(
	tp *TreatmentPlan,
	eventType TreatmentPlanEventType,
	measureIDs []gid.GID,
	now time.Time,
) *TreatmentPlanEvent {
	return &TreatmentPlanEvent{
		OrganizationID:         tp.OrganizationID,
		RiskAnalysisID:         tp.RiskAnalysisID,
		TreatmentPlanID:        tp.ID,
		EventType:              eventType,
		RiskID:                 tp.RiskID,
		OwnerID:                tp.OwnerID,
		Treatment:              tp.Treatment,
		InherentLikelihood:     tp.InherentLikelihood,
		InherentImpact:         tp.InherentImpact,
		ResidualLikelihood:     tp.ResidualLikelihood,
		ResidualImpact:         tp.ResidualImpact,
		MeasureIDs:             MeasureIDStrings(measureIDs),
		Category:               tp.Category,
		TreatmentPlanCreatedAt: tp.CreatedAt,
		TreatmentPlanUpdatedAt: tp.UpdatedAt,
		CreatedAt:              now,
	}
}

func (e *TreatmentPlanEvent) LinkedMeasureIDs() ([]gid.GID, error) {
	ids := make([]gid.GID, 0, len(e.MeasureIDs))
	for _, raw := range e.MeasureIDs {
		id, err := gid.ParseGID(raw)
		if err != nil {
			return nil, fmt.Errorf("cannot parse measure id %q: %w", raw, err)
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (e *TreatmentPlanEvent) TreatmentPlan() *TreatmentPlan {
	return &TreatmentPlan{
		ID:                 e.TreatmentPlanID,
		OrganizationID:     e.OrganizationID,
		RiskID:             e.RiskID,
		RiskAnalysisID:     e.RiskAnalysisID,
		Treatment:          e.Treatment,
		OwnerID:            e.OwnerID,
		InherentLikelihood: e.InherentLikelihood,
		InherentImpact:     e.InherentImpact,
		InherentRiskScore:  e.InherentLikelihood * e.InherentImpact,
		ResidualLikelihood: e.ResidualLikelihood,
		ResidualImpact:     e.ResidualImpact,
		ResidualRiskScore:  e.ResidualLikelihood * e.ResidualImpact,
		Category:           e.Category,
		CreatedAt:          e.TreatmentPlanCreatedAt,
		UpdatedAt:          e.TreatmentPlanUpdatedAt,
	}
}

func (e *TreatmentPlanEvent) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    treatment_plan_events (
        tenant_id,
        organization_id,
        risk_analysis_id,
        treatment_plan_id,
        event_type,
        risk_id,
        owner_profile_id,
        treatment,
        inherent_likelihood,
        inherent_impact,
        residual_likelihood,
        residual_impact,
        measure_ids,
        category,
        treatment_plan_created_at,
        treatment_plan_updated_at,
        created_at
    )
VALUES (
    @tenant_id,
    @organization_id,
    @risk_analysis_id,
    @treatment_plan_id,
    @event_type,
    @risk_id,
    @owner_profile_id,
    @treatment,
    @inherent_likelihood,
    @inherent_impact,
    @residual_likelihood,
    @residual_impact,
    @measure_ids,
    @category,
    @treatment_plan_created_at,
    @treatment_plan_updated_at,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           e.OrganizationID,
		"risk_analysis_id":          e.RiskAnalysisID,
		"treatment_plan_id":         e.TreatmentPlanID,
		"event_type":                e.EventType,
		"risk_id":                   e.RiskID,
		"owner_profile_id":          e.OwnerID,
		"treatment":                 e.Treatment,
		"inherent_likelihood":       e.InherentLikelihood,
		"inherent_impact":           e.InherentImpact,
		"residual_likelihood":       e.ResidualLikelihood,
		"residual_impact":           e.ResidualImpact,
		"measure_ids":               e.MeasureIDs,
		"category":                  e.Category,
		"treatment_plan_created_at": e.TreatmentPlanCreatedAt,
		"treatment_plan_updated_at": e.TreatmentPlanUpdatedAt,
		"created_at":                e.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert treatment plan event: %w", err)
	}

	return nil
}

func (es *TreatmentPlanEvents) LoadLatestByRiskAnalysisIDAsOf(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAnalysisID gid.GID,
	asOf time.Time,
) error {
	q := `
SELECT
    organization_id,
    risk_analysis_id,
    treatment_plan_id,
    event_type,
    risk_id,
    owner_profile_id,
    treatment,
    inherent_likelihood,
    inherent_impact,
    residual_likelihood,
    residual_impact,
    measure_ids,
    category,
    treatment_plan_created_at,
    treatment_plan_updated_at,
    created_at
FROM (
    SELECT DISTINCT ON (treatment_plan_id)
        organization_id,
        risk_analysis_id,
        treatment_plan_id,
        event_type,
        risk_id,
        owner_profile_id,
        treatment,
        inherent_likelihood,
        inherent_impact,
        residual_likelihood,
        residual_impact,
        measure_ids,
        category,
        treatment_plan_created_at,
        treatment_plan_updated_at,
        created_at
    FROM
        treatment_plan_events
    WHERE
        %s
        AND risk_analysis_id = @risk_analysis_id
        AND created_at < @as_of
    ORDER BY
        treatment_plan_id,
        created_at DESC
) latest
WHERE
    event_type <> @deleted
ORDER BY
    treatment_plan_created_at ASC,
    treatment_plan_id ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"risk_analysis_id": riskAnalysisID,
		"as_of":            asOf,
		"deleted":          TreatmentPlanEventTypeDeleted,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query treatment plan events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TreatmentPlanEvent])
	if err != nil {
		return fmt.Errorf("cannot collect treatment plan events: %w", err)
	}

	*es = events

	return nil
}

func (es *TreatmentPlanEvents) LoadLatestByTreatmentPlanIDsAsOf(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	riskAnalysisID gid.GID,
	treatmentPlanIDs []gid.GID,
	asOf time.Time,
) error {
	if len(treatmentPlanIDs) == 0 {
		*es = nil
		return nil
	}

	q := `
SELECT
    organization_id,
    risk_analysis_id,
    treatment_plan_id,
    event_type,
    risk_id,
    owner_profile_id,
    treatment,
    inherent_likelihood,
    inherent_impact,
    residual_likelihood,
    residual_impact,
    measure_ids,
    category,
    treatment_plan_created_at,
    treatment_plan_updated_at,
    created_at
FROM (
    SELECT DISTINCT ON (treatment_plan_id)
        organization_id,
        risk_analysis_id,
        treatment_plan_id,
        event_type,
        risk_id,
        owner_profile_id,
        treatment,
        inherent_likelihood,
        inherent_impact,
        residual_likelihood,
        residual_impact,
        measure_ids,
        category,
        treatment_plan_created_at,
        treatment_plan_updated_at,
        created_at
    FROM
        treatment_plan_events
    WHERE
        %s
        AND risk_analysis_id = @risk_analysis_id
        AND treatment_plan_id = ANY(@treatment_plan_ids)
        AND created_at < @as_of
    ORDER BY
        treatment_plan_id,
        created_at DESC
) latest
WHERE
    event_type <> @deleted
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"risk_analysis_id":   riskAnalysisID,
		"treatment_plan_ids": treatmentPlanIDs,
		"as_of":              asOf,
		"deleted":            TreatmentPlanEventTypeDeleted,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query treatment plan events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TreatmentPlanEvent])
	if err != nil {
		return fmt.Errorf("cannot collect treatment plan events: %w", err)
	}

	*es = events

	return nil
}
