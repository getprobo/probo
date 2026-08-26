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

package riskmanagement

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	CreateTreatmentPlanRequest struct {
		RiskID             gid.GID
		RiskAnalysisID     gid.GID
		Treatment          coredata.RiskTreatment
		OwnerID            gid.GID
		InherentLikelihood int
		InherentImpact     int
		ResidualLikelihood *int
		ResidualImpact     *int
	}

	UpdateTreatmentPlanRequest struct {
		ID                 gid.GID
		Treatment          *coredata.RiskTreatment
		OwnerID            **gid.GID
		InherentLikelihood *int
		InherentImpact     *int
		ResidualLikelihood *int
		ResidualImpact     *int
	}
)

func (r *CreateTreatmentPlanRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskID, "risk_id", validator.Required(), validator.GID(coredata.RiskEntityType))
	v.Check(r.RiskAnalysisID, "risk_analysis_id", validator.Required(), validator.GID(coredata.RiskAnalysisEntityType))
	v.Check(r.Treatment, "treatment", validator.Required(), validator.OneOfSlice(coredata.RiskTreatments()))
	v.Check(r.OwnerID, "owner_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	v.Check(r.InherentLikelihood, "inherent_likelihood", validator.Required(), validator.Min(1))
	v.Check(r.InherentImpact, "inherent_impact", validator.Required(), validator.Min(1))
	v.Check(r.ResidualLikelihood, "residual_likelihood", validator.Min(1))
	v.Check(r.ResidualImpact, "residual_impact", validator.Min(1))
	requireResidualPair(v, r.ResidualLikelihood, r.ResidualImpact)
	requireAcceptedScoresMatchInherent(
		v,
		r.Treatment,
		r.InherentLikelihood,
		r.InherentImpact,
		residualOrInherent(r.ResidualLikelihood, r.InherentLikelihood),
		residualOrInherent(r.ResidualImpact, r.InherentImpact),
	)

	return v.Error()
}

func (r *UpdateTreatmentPlanRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.TreatmentPlanEntityType))
	v.Check(r.Treatment, "treatment", validator.OneOfSlice(coredata.RiskTreatments()))

	if r.OwnerID != nil {
		v.Check(r.OwnerID, "owner_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	}

	v.Check(r.InherentLikelihood, "inherent_likelihood", validator.Min(1))
	v.Check(r.InherentImpact, "inherent_impact", validator.Min(1))
	v.Check(r.ResidualLikelihood, "residual_likelihood", validator.Min(1))
	v.Check(r.ResidualImpact, "residual_impact", validator.Min(1))

	return v.Error()
}

func requireResidualPair(v *validator.Validator, likelihood, impact *int) {
	if (likelihood == nil) == (impact == nil) {
		return
	}

	missing := "residual_likelihood"
	other := "residual_impact"

	if likelihood != nil {
		missing = "residual_impact"
		other = "residual_likelihood"
	}

	v.Check(
		missing,
		missing,
		func(any) *validator.ValidationError {
			return &validator.ValidationError{
				Code:    validator.ErrorCodeCustom,
				Message: fmt.Sprintf("must be set together with %s", other),
			}
		},
	)
}

func residualOrInherent(residual *int, inherent int) int {
	if residual != nil {
		return *residual
	}

	return inherent
}

func requireAcceptedScoresMatchInherent(
	v *validator.Validator,
	treatment coredata.RiskTreatment,
	inherentLikelihood int,
	inherentImpact int,
	residualLikelihood int,
	residualImpact int,
) {
	if treatment != coredata.RiskTreatmentAccepted {
		return
	}

	if residualLikelihood != inherentLikelihood {
		v.Check(
			residualLikelihood,
			"residual_likelihood",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeInvalidFormat,
					Message: "must equal inherent_likelihood when treatment is ACCEPTED",
				}
			},
		)
	}

	if residualImpact != inherentImpact {
		v.Check(
			residualImpact,
			"residual_impact",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeInvalidFormat,
					Message: "must equal inherent_impact when treatment is ACCEPTED",
				}
			},
		)
	}
}

func validateScoresAgainstMatrix(
	v *validator.Validator,
	analysis *coredata.RiskAnalysis,
	likelihood int,
	impact int,
	likelihoodField string,
	impactField string,
) {
	v.Check(likelihood, likelihoodField, validator.Max(analysis.MatrixRows))
	v.Check(impact, impactField, validator.Max(analysis.MatrixCols))
}

func riskNotOnAnalysisScenarioError() error {
	return validator.ValidationErrors{
		{
			Field:   "risk_id",
			Code:    validator.ErrorCodeCustom,
			Message: "must be linked to a scenario of this analysis",
		},
	}
}

func (s *Service) CreateTreatmentPlan(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateTreatmentPlanRequest,
) (*coredata.TreatmentPlan, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	tp := &coredata.TreatmentPlan{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			risk := &coredata.Risk{}
			if err := risk.LoadByID(ctx, tx, scope, req.RiskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			analysis := &coredata.RiskAnalysis{}
			if err := analysis.LoadByID(ctx, tx, scope, req.RiskAnalysisID); err != nil {
				return fmt.Errorf("cannot load risk analysis: %w", err)
			}

			if risk.OrganizationID != analysis.OrganizationID {
				return fmt.Errorf("cannot verify risk organization: %w", coredata.ErrResourceNotFound)
			}

			linked, err := s.riskLinkedToAnalysis(ctx, tx, scope, req.RiskID, req.RiskAnalysisID)
			if err != nil {
				return fmt.Errorf("cannot check scenario link: %w", err)
			}

			if !linked {
				return fmt.Errorf("cannot create treatment plan without scenario link: %w", riskNotOnAnalysisScenarioError())
			}

			owner := &coredata.MembershipProfile{}
			if err := owner.LoadByID(ctx, tx, scope, req.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner profile: %w", err)
			}

			if owner.OrganizationID != risk.OrganizationID {
				return fmt.Errorf("cannot verify owner organization: %w", coredata.ErrResourceNotFound)
			}

			residualLikelihood := req.InherentLikelihood
			residualImpact := req.InherentImpact

			if req.ResidualLikelihood != nil {
				residualLikelihood = *req.ResidualLikelihood
			}

			if req.ResidualImpact != nil {
				residualImpact = *req.ResidualImpact
			}

			v := validator.New()
			validateScoresAgainstMatrix(
				v,
				analysis,
				req.InherentLikelihood,
				req.InherentImpact,
				"inherent_likelihood",
				"inherent_impact",
			)
			validateScoresAgainstMatrix(
				v,
				analysis,
				residualLikelihood,
				residualImpact,
				"residual_likelihood",
				"residual_impact",
			)

			if err := v.Error(); err != nil {
				return fmt.Errorf("cannot validate scores: %w", err)
			}

			*tp = coredata.TreatmentPlan{
				ID:                 gid.New(scope.GetTenantID(), coredata.TreatmentPlanEntityType),
				OrganizationID:     risk.OrganizationID,
				RiskID:             risk.ID,
				RiskAnalysisID:     analysis.ID,
				Treatment:          req.Treatment,
				OwnerID:            req.OwnerID,
				InherentLikelihood: req.InherentLikelihood,
				InherentImpact:     req.InherentImpact,
				ResidualLikelihood: residualLikelihood,
				ResidualImpact:     residualImpact,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if err := tp.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert treatment plan: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create treatment plan: %w", err)
	}

	return tp, nil
}

func (s *Service) GetTreatmentPlan(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.TreatmentPlan, error) {
	tp := &coredata.TreatmentPlan{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tp.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load treatment plan: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get treatment plan: %w", err)
	}

	return tp, nil
}

func (s *Service) UpdateTreatmentPlan(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateTreatmentPlanRequest,
) (*coredata.TreatmentPlan, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	tp := &coredata.TreatmentPlan{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := tp.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load treatment plan: %w", err)
			}

			previousTreatment := tp.Treatment

			if req.Treatment != nil {
				tp.Treatment = *req.Treatment
			}

			if req.OwnerID != nil {
				tp.OwnerID = **req.OwnerID
			}

			if req.InherentLikelihood != nil {
				tp.InherentLikelihood = *req.InherentLikelihood
			}

			if req.InherentImpact != nil {
				tp.InherentImpact = *req.InherentImpact
			}

			if req.ResidualLikelihood != nil {
				tp.ResidualLikelihood = *req.ResidualLikelihood
			}

			if req.ResidualImpact != nil {
				tp.ResidualImpact = *req.ResidualImpact
			}

			owner := &coredata.MembershipProfile{}
			if err := owner.LoadByID(ctx, tx, scope, tp.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner profile: %w", err)
			}

			if owner.OrganizationID != tp.OrganizationID {
				return fmt.Errorf("cannot verify owner organization: %w", coredata.ErrResourceNotFound)
			}

			analysis := &coredata.RiskAnalysis{}
			if err := analysis.LoadByID(ctx, tx, scope, tp.RiskAnalysisID); err != nil {
				return fmt.Errorf("cannot load risk analysis: %w", err)
			}

			v := validator.New()
			validateScoresAgainstMatrix(
				v,
				analysis,
				tp.InherentLikelihood,
				tp.InherentImpact,
				"inherent_likelihood",
				"inherent_impact",
			)
			validateScoresAgainstMatrix(
				v,
				analysis,
				tp.ResidualLikelihood,
				tp.ResidualImpact,
				"residual_likelihood",
				"residual_impact",
			)
			requireAcceptedScoresMatchInherent(
				v,
				tp.Treatment,
				tp.InherentLikelihood,
				tp.InherentImpact,
				tp.ResidualLikelihood,
				tp.ResidualImpact,
			)

			if err := v.Error(); err != nil {
				return fmt.Errorf("cannot validate scores: %w", err)
			}

			if tp.Treatment == coredata.RiskTreatmentAccepted && previousTreatment != coredata.RiskTreatmentAccepted {
				var mappings coredata.TreatmentPlanMeasures
				if err := mappings.LoadByTreatmentPlanIDs(ctx, tx, scope, []gid.GID{tp.ID}); err != nil {
					return fmt.Errorf("cannot load treatment plan measures: %w", err)
				}

				if len(mappings) > 0 {
					return fmt.Errorf("cannot accept treatment plan with measures: %w", errMeasuresNotAllowedForAccepted())
				}
			}

			tp.UpdatedAt = time.Now()
			if err := tp.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot persist treatment plan: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot update treatment plan: %w", err)
	}

	return tp, nil
}

func (s *Service) DeleteTreatmentPlan(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			tp := &coredata.TreatmentPlan{}
			if err := tp.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete treatment plan row: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot delete treatment plan: %w", err)
	}

	return nil
}

func (s *Service) ListTreatmentPlansForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.TreatmentPlanOrderField],
	filter *coredata.TreatmentPlanFilter,
) (*page.Page[*coredata.TreatmentPlan, coredata.TreatmentPlanOrderField], error) {
	var results coredata.TreatmentPlans

	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return nil, err
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter); err != nil {
				return fmt.Errorf("cannot list treatment plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list organization treatment plans: %w", err)
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountTreatmentPlansForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.TreatmentPlanFilter,
) (int, error) {
	var count int

	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return 0, err
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			tps := &coredata.TreatmentPlans{}

			count, err = tps.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count treatment plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count organization treatment plans: %w", err)
	}

	return count, nil
}

func (s *Service) ListTreatmentPlansForRiskID(
	ctx context.Context,
	scope coredata.Scoper,
	riskID gid.GID,
	cursor *page.Cursor[coredata.TreatmentPlanOrderField],
	filter *coredata.TreatmentPlanFilter,
) (*page.Page[*coredata.TreatmentPlan, coredata.TreatmentPlanOrderField], error) {
	var results coredata.TreatmentPlans

	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return nil, err
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskID(ctx, conn, scope, riskID, cursor, filter); err != nil {
				return fmt.Errorf("cannot list treatment plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list risk treatment plans: %w", err)
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountTreatmentPlansForRiskID(
	ctx context.Context,
	scope coredata.Scoper,
	riskID gid.GID,
	filter *coredata.TreatmentPlanFilter,
) (int, error) {
	var count int

	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return 0, err
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			tps := &coredata.TreatmentPlans{}

			count, err = tps.CountByRiskID(ctx, conn, scope, riskID, filter)
			if err != nil {
				return fmt.Errorf("cannot count treatment plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count risk treatment plans: %w", err)
	}

	return count, nil
}

func (s *Service) ListTreatmentPlansForRiskAnalysisID(
	ctx context.Context,
	scope coredata.Scoper,
	riskAnalysisID gid.GID,
	cursor *page.Cursor[coredata.TreatmentPlanOrderField],
	filter *coredata.TreatmentPlanFilter,
) (*page.Page[*coredata.TreatmentPlan, coredata.TreatmentPlanOrderField], error) {
	var results coredata.TreatmentPlans

	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return nil, err
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAnalysisID(ctx, conn, scope, riskAnalysisID, cursor, filter); err != nil {
				return fmt.Errorf("cannot list treatment plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list analysis treatment plans: %w", err)
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountTreatmentPlansForRiskAnalysisID(
	ctx context.Context,
	scope coredata.Scoper,
	riskAnalysisID gid.GID,
	filter *coredata.TreatmentPlanFilter,
) (int, error) {
	var count int

	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return 0, err
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			tps := &coredata.TreatmentPlans{}

			count, err = tps.CountByRiskAnalysisID(ctx, conn, scope, riskAnalysisID, filter)
			if err != nil {
				return fmt.Errorf("cannot count treatment plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count analysis treatment plans: %w", err)
	}

	return count, nil
}

func errMeasuresNotAllowedForAccepted() error {
	return validator.ValidationErrors{
		{
			Field:   "measure_id",
			Code:    validator.ErrorCodeCustom,
			Message: "cannot be linked to an accepted treatment plan",
		},
	}
}

func (s *Service) CreateMeasureMapping(
	ctx context.Context,
	scope coredata.Scoper,
	treatmentPlanID gid.GID,
	measureID gid.GID,
) (*coredata.TreatmentPlan, *coredata.Measure, error) {
	tp := &coredata.TreatmentPlan{}
	measure := &coredata.Measure{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := tp.LoadByID(ctx, tx, scope, treatmentPlanID); err != nil {
				return fmt.Errorf("cannot load treatment plan: %w", err)
			}

			if tp.Treatment == coredata.RiskTreatmentAccepted {
				return fmt.Errorf("cannot link measures to accepted treatment plan: %w", errMeasuresNotAllowedForAccepted())
			}

			if err := measure.LoadByID(ctx, tx, scope, measureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			if measure.OrganizationID != tp.OrganizationID {
				return fmt.Errorf("cannot verify measure organization: %w", coredata.ErrResourceNotFound)
			}

			mapping := &coredata.TreatmentPlanMeasure{
				TreatmentPlanID: tp.ID,
				MeasureID:       measure.ID,
				OrganizationID:  tp.OrganizationID,
				CreatedAt:       time.Now(),
			}

			if err := mapping.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert treatment plan measure: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create treatment plan measure: %w", err)
	}

	return tp, measure, nil
}

func (s *Service) DeleteMeasureMapping(
	ctx context.Context,
	scope coredata.Scoper,
	treatmentPlanID gid.GID,
	measureID gid.GID,
) (*coredata.TreatmentPlan, *coredata.Measure, error) {
	tp := &coredata.TreatmentPlan{}
	measure := &coredata.Measure{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := tp.LoadByID(ctx, tx, scope, treatmentPlanID); err != nil {
				return fmt.Errorf("cannot load treatment plan: %w", err)
			}

			if err := measure.LoadByID(ctx, tx, scope, measureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			mapping := coredata.TreatmentPlanMeasure{}
			if err := mapping.Delete(ctx, tx, scope, tp.ID, measure.ID); err != nil {
				return fmt.Errorf("cannot delete treatment plan measure: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot remove treatment plan measure: %w", err)
	}

	return tp, measure, nil
}

func (s *Service) ListTreatmentPlansForMeasureID(
	ctx context.Context,
	scope coredata.Scoper,
	measureID gid.GID,
	cursor *page.Cursor[coredata.TreatmentPlanOrderField],
	filter *coredata.TreatmentPlanFilter,
) (*page.Page[*coredata.TreatmentPlan, coredata.TreatmentPlanOrderField], error) {
	var results coredata.TreatmentPlans

	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return nil, err
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByMeasureID(ctx, conn, scope, measureID, cursor, filter); err != nil {
				return fmt.Errorf("cannot list treatment plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list measure treatment plans: %w", err)
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountTreatmentPlansForMeasureID(
	ctx context.Context,
	scope coredata.Scoper,
	measureID gid.GID,
	filter *coredata.TreatmentPlanFilter,
) (int, error) {
	var count int

	filter, err := prepareTreatmentPlanFilter(filter)
	if err != nil {
		return 0, err
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			tps := &coredata.TreatmentPlans{}

			count, err = tps.CountByMeasureID(ctx, conn, scope, measureID, filter)
			if err != nil {
				return fmt.Errorf("cannot count treatment plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count measure treatment plans: %w", err)
	}

	return count, nil
}

func (s *Service) GetRiskAnalysisMatrixCells(
	ctx context.Context,
	scope coredata.Scoper,
	riskAnalysisID gid.GID,
) ([]*coredata.RiskAnalysisMatrixCell, error) {
	var counts []*coredata.RiskAnalysisMatrixCell

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			tps := coredata.TreatmentPlans{}

			var err error

			counts, err = tps.CountMatrixCellsByRiskAnalysisID(ctx, conn, scope, riskAnalysisID)
			if err != nil {
				return fmt.Errorf("cannot count risk analysis matrix cells: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get risk analysis matrix cells: %w", err)
	}

	if counts == nil {
		return []*coredata.RiskAnalysisMatrixCell{}, nil
	}

	return counts, nil
}

func treatmentPlanFilterOrEmpty(filter *coredata.TreatmentPlanFilter) *coredata.TreatmentPlanFilter {
	if filter != nil {
		return filter
	}

	return coredata.NewTreatmentPlanFilter(nil, nil, nil)
}

func prepareTreatmentPlanFilter(filter *coredata.TreatmentPlanFilter) (*coredata.TreatmentPlanFilter, error) {
	filter = treatmentPlanFilterOrEmpty(filter)
	if err := validateTreatmentPlanFilter(filter); err != nil {
		return nil, err
	}

	return filter, nil
}

func validateTreatmentPlanFilter(filter *coredata.TreatmentPlanFilter) error {
	if filter.IsEmpty() {
		return nil
	}

	v := validator.New()
	v.Check(
		filter.ScoreType(),
		"score_type",
		validator.Required(),
		validator.OneOfSlice(coredata.TreatmentPlanScoreTypes()),
	)
	v.Check(filter.Likelihood(), "likelihood", validator.Required(), validator.Min(1))
	v.Check(filter.Impact(), "impact", validator.Required(), validator.Min(1))

	return v.Error()
}

func (s *Service) riskLinkedToAnalysis(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	riskID gid.GID,
	analysisID gid.GID,
) (bool, error) {
	var links coredata.RiskAnalysisScenarioRisks
	if err := links.LoadByRiskID(ctx, conn, scope, riskID); err != nil {
		return false, fmt.Errorf("cannot load scenario risks: %w", err)
	}

	if len(links) == 0 {
		return false, nil
	}

	scenarioIDs := make([]gid.GID, 0, len(links))
	seenScenarios := make(map[gid.GID]struct{}, len(links))

	for _, link := range links {
		if _, seen := seenScenarios[link.RiskAnalysisScenarioID]; seen {
			continue
		}

		seenScenarios[link.RiskAnalysisScenarioID] = struct{}{}
		scenarioIDs = append(scenarioIDs, link.RiskAnalysisScenarioID)
	}

	var scenarios coredata.RiskAnalysisScenarios
	if err := scenarios.LoadByIDs(ctx, conn, scope, scenarioIDs); err != nil {
		return false, fmt.Errorf("cannot load risk analysis scenarios: %w", err)
	}

	diagramIDs := make([]gid.GID, 0, len(scenarios))
	seenDiagrams := make(map[gid.GID]struct{}, len(scenarios))

	for _, scenario := range scenarios {
		if _, seen := seenDiagrams[scenario.RiskAnalysisDiagramID]; seen {
			continue
		}

		seenDiagrams[scenario.RiskAnalysisDiagramID] = struct{}{}
		diagramIDs = append(diagramIDs, scenario.RiskAnalysisDiagramID)
	}

	var diagrams coredata.RiskAnalysisDiagrams
	if err := diagrams.LoadByIDs(ctx, conn, scope, diagramIDs); err != nil {
		return false, fmt.Errorf("cannot load risk analysis diagrams: %w", err)
	}

	for _, diagram := range diagrams {
		if diagram.RiskAnalysisID == analysisID {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) GetTreatmentProgress(
	ctx context.Context,
	scope coredata.Scoper,
	treatmentPlanID gid.GID,
) (TreatmentProgress, error) {
	progressByID, err := s.GetTreatmentProgressByIDs(ctx, scope, []gid.GID{treatmentPlanID})
	if err != nil {
		return TreatmentProgress{}, err
	}

	return progressByID[treatmentPlanID], nil
}

func (s *Service) GetTreatmentProgressByIDs(
	ctx context.Context,
	scope coredata.Scoper,
	treatmentPlanIDs []gid.GID,
) (map[gid.GID]TreatmentProgress, error) {
	var progress map[gid.GID]TreatmentProgress

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			progress, err = s.loadTreatmentProgress(ctx, conn, scope, treatmentPlanIDs)
			if err != nil {
				return fmt.Errorf("cannot load treatment progress: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get treatment progress: %w", err)
	}

	return progress, nil
}

func (s *Service) loadTreatmentProgress(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	treatmentPlanIDs []gid.GID,
) (map[gid.GID]TreatmentProgress, error) {
	progress := make(map[gid.GID]TreatmentProgress, len(treatmentPlanIDs))
	for _, id := range treatmentPlanIDs {
		progress[id] = TreatmentProgress{}
	}

	if len(treatmentPlanIDs) == 0 {
		return progress, nil
	}

	mappings := coredata.TreatmentPlanMeasures{}
	if err := mappings.LoadByTreatmentPlanIDs(ctx, conn, scope, treatmentPlanIDs); err != nil {
		return nil, fmt.Errorf("cannot load treatment plan measures: %w", err)
	}

	measureIDs := make([]gid.GID, 0, len(mappings))
	seenMeasures := make(map[gid.GID]struct{}, len(mappings))

	for _, mapping := range mappings {
		if _, seen := seenMeasures[mapping.MeasureID]; seen {
			continue
		}

		seenMeasures[mapping.MeasureID] = struct{}{}
		measureIDs = append(measureIDs, mapping.MeasureID)
	}

	measures := coredata.Measures{}
	if len(measureIDs) > 0 {
		if err := measures.LoadByIDs(ctx, conn, scope, measureIDs); err != nil {
			return nil, fmt.Errorf("cannot load measures: %w", err)
		}
	}

	byID := make(map[gid.GID]*coredata.Measure, len(measures))
	for _, measure := range measures {
		byID[measure.ID] = measure
	}

	for _, mapping := range mappings {
		planProgress := progress[mapping.TreatmentPlanID]
		planProgress.Total++

		measure := byID[mapping.MeasureID]
		if measure == nil {
			return nil, fmt.Errorf("cannot load measure %q for treatment plan", mapping.MeasureID)
		}

		switch measure.State {
		case coredata.MeasureStateImplemented:
			planProgress.Done++
		case coredata.MeasureStateInProgress:
			planProgress.InProgress++
		case coredata.MeasureStateNotImplemented:
			planProgress.NotImplemented++
		}

		progress[mapping.TreatmentPlanID] = planProgress
	}

	return progress, nil
}
