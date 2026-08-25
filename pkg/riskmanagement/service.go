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

const (
	TitleMaxLength   = 1000
	ContentMaxLength = 5000
)

type Service struct {
	pg *pg.Client
}

func NewService(pgClient *pg.Client) *Service {
	return &Service{pg: pgClient}
}

type (
	Period struct {
		Start *time.Time
		End   *time.Time
	}

	MatrixSize struct {
		Rows int
		Cols int
	}

	CreateRiskAnalysisRequest struct {
		OrganizationID gid.GID
		Name           string
		Description    *string
		Period         *Period
		MatrixSize     *MatrixSize
	}

	UpdateRiskAnalysisRequest struct {
		ID          gid.GID
		Name        *string
		Description **string
		Period      *Period
	}

	CreateRiskAnalysisDiagramRequest struct {
		RiskAnalysisID gid.GID
		Name           string
	}

	UpdateRiskAnalysisDiagramRequest struct {
		ID   gid.GID
		Name *string
	}

	CreateRiskAnalysisBoundaryRequest struct {
		RiskAnalysisDiagramID gid.GID
		ParentBoundaryID      *gid.GID
		Name                  string
	}

	UpdateRiskAnalysisBoundaryRequest struct {
		ID               gid.GID
		ParentBoundaryID **gid.GID
		Name             *string
	}

	CreateRiskAnalysisNodeRequest struct {
		RiskAnalysisDiagramID gid.GID
		BoundaryID            *gid.GID
		NodeType              coredata.RiskAnalysisNodeType
		Name                  string
	}

	UpdateRiskAnalysisNodeRequest struct {
		ID         gid.GID
		BoundaryID **gid.GID
		NodeType   *coredata.RiskAnalysisNodeType
		Name       *string
	}

	CreateRiskAnalysisProcessRequest struct {
		RiskAnalysisDiagramID gid.GID
		SourceNodeID          gid.GID
		TargetNodeID          gid.GID
		Name                  string
	}

	UpdateRiskAnalysisProcessRequest struct {
		ID           gid.GID
		SourceNodeID *gid.GID
		TargetNodeID *gid.GID
		Name         *string
	}

	CreateRiskAnalysisThreatRequest struct {
		RiskAnalysisDiagramID gid.GID
		ProcessID             gid.GID
		Name                  string
		Category              string
	}

	UpdateRiskAnalysisThreatRequest struct {
		ID        gid.GID
		ProcessID *gid.GID
		Name      *string
		Category  *string
	}

	CreateRiskAnalysisScenarioRequest struct {
		RiskAnalysisDiagramID gid.GID
		Name                  string
		Description           *string
	}

	UpdateRiskAnalysisScenarioRequest struct {
		ID          gid.GID
		Name        *string
		Description **string
	}

	LinkRiskAnalysisScenarioThreatRequest struct {
		RiskAnalysisScenarioID gid.GID
		ThreatID               gid.GID
	}

	UnlinkRiskAnalysisScenarioThreatRequest struct {
		RiskAnalysisScenarioID gid.GID
		ThreatID               gid.GID
	}

	LinkRiskAnalysisScenarioRiskRequest struct {
		RiskAnalysisScenarioID gid.GID
		RiskID                 gid.GID
	}

	UnlinkRiskAnalysisScenarioRiskRequest struct {
		RiskAnalysisScenarioID gid.GID
		RiskID                 gid.GID
	}

	ForkRiskAnalysisRequest struct {
		RiskAnalysisID gid.GID
		Name           string
		Description    *string
		Period         *Period
	}
)

func (r *CreateRiskAnalysisRequest) Validate() error {
	v := validator.New()
	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(r.MatrixSize, "matrix_size", validator.Required())
	validateMatrixSize(v, r.MatrixSize)

	if r.Period != nil {
		validatePeriodRange(v, r.Period.Start, r.Period.End)
	}

	return v.Error()
}

func (r *UpdateRiskAnalysisRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAnalysisEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	if r.Period != nil {
		validatePeriodRange(v, r.Period.Start, r.Period.End)
	}

	return v.Error()
}

func (r *ForkRiskAnalysisRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisID, "risk_analysis_id", validator.Required(), validator.GID(coredata.RiskAnalysisEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	if r.Period != nil {
		validatePeriodRange(v, r.Period.Start, r.Period.End)
	}

	return v.Error()
}

func validatePeriodRange(v *validator.Validator, periodStart, periodEnd *time.Time) {
	if periodStart == nil || periodEnd == nil {
		return
	}

	if periodEnd.Before(*periodStart) {
		v.Check(
			periodEnd,
			"period_end",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeOutOfRange,
					Message: fmt.Sprintf("must be on or after %s", periodStart.Format(time.DateOnly)),
				}
			},
		)
	}
}

func validateMatrixSize(v *validator.Validator, size *MatrixSize) {
	if size == nil {
		return
	}

	v.Check(size.Rows, "matrix_size.rows", validator.OneOfSlice([]int{3, 4, 5}))
	v.Check(size.Cols, "matrix_size.cols", validator.OneOfSlice([]int{3, 4, 5}))

	if size.Rows != size.Cols {
		v.Check(
			size.Cols,
			"matrix_size.cols",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeInvalidFormat,
					Message: "must equal matrix_size.rows",
				}
			},
		)
	}
}

func (r *CreateRiskAnalysisDiagramRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisID, "risk_analysis_id", validator.Required(), validator.GID(coredata.RiskAnalysisEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (r *UpdateRiskAnalysisDiagramRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAnalysisDiagramEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (r *CreateRiskAnalysisBoundaryRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisDiagramID, "risk_analysis_diagram_id", validator.Required(), validator.GID(coredata.RiskAnalysisDiagramEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))

	if r.ParentBoundaryID != nil {
		v.Check(*r.ParentBoundaryID, "parent_boundary_id", validator.Required(), validator.GID(coredata.RiskAnalysisBoundaryEntityType))
	}

	return v.Error()
}

func (r *UpdateRiskAnalysisBoundaryRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAnalysisBoundaryEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))

	if r.ParentBoundaryID != nil && *r.ParentBoundaryID != nil {
		v.Check(**r.ParentBoundaryID, "parent_boundary_id", validator.Required(), validator.GID(coredata.RiskAnalysisBoundaryEntityType))
	}

	return v.Error()
}

func (r *CreateRiskAnalysisNodeRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisDiagramID, "risk_analysis_diagram_id", validator.Required(), validator.GID(coredata.RiskAnalysisDiagramEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.NodeType, "node_type", validator.Required(), validator.OneOfSlice(coredata.RiskAnalysisNodeTypes()))

	if r.BoundaryID != nil {
		v.Check(*r.BoundaryID, "boundary_id", validator.Required(), validator.GID(coredata.RiskAnalysisBoundaryEntityType))
	}

	return v.Error()
}

func (r *UpdateRiskAnalysisNodeRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAnalysisNodeEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.NodeType, "node_type", validator.OneOfSlice(coredata.RiskAnalysisNodeTypes()))

	if r.BoundaryID != nil && *r.BoundaryID != nil {
		v.Check(**r.BoundaryID, "boundary_id", validator.Required(), validator.GID(coredata.RiskAnalysisBoundaryEntityType))
	}

	return v.Error()
}

func (r *CreateRiskAnalysisProcessRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisDiagramID, "risk_analysis_diagram_id", validator.Required(), validator.GID(coredata.RiskAnalysisDiagramEntityType))
	v.Check(r.SourceNodeID, "source_node_id", validator.Required(), validator.GID(coredata.RiskAnalysisNodeEntityType))
	v.Check(r.TargetNodeID, "target_node_id", validator.Required(), validator.GID(coredata.RiskAnalysisNodeEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (r *UpdateRiskAnalysisProcessRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAnalysisProcessEntityType))
	v.Check(r.SourceNodeID, "source_node_id", validator.GID(coredata.RiskAnalysisNodeEntityType))
	v.Check(r.TargetNodeID, "target_node_id", validator.GID(coredata.RiskAnalysisNodeEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (r *CreateRiskAnalysisThreatRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisDiagramID, "risk_analysis_diagram_id", validator.Required(), validator.GID(coredata.RiskAnalysisDiagramEntityType))
	v.Check(r.ProcessID, "process_id", validator.Required(), validator.GID(coredata.RiskAnalysisProcessEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Category, "category", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (r *UpdateRiskAnalysisThreatRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAnalysisThreatEntityType))
	v.Check(r.ProcessID, "process_id", validator.GID(coredata.RiskAnalysisProcessEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Category, "category", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (r *CreateRiskAnalysisScenarioRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisDiagramID, "risk_analysis_diagram_id", validator.Required(), validator.GID(coredata.RiskAnalysisDiagramEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (r *LinkRiskAnalysisScenarioThreatRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisScenarioID, "risk_analysis_scenario_id", validator.Required(), validator.GID(coredata.RiskAnalysisScenarioEntityType))
	v.Check(r.ThreatID, "threat_id", validator.Required(), validator.GID(coredata.RiskAnalysisThreatEntityType))

	return v.Error()
}

func (r *UnlinkRiskAnalysisScenarioThreatRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisScenarioID, "risk_analysis_scenario_id", validator.Required(), validator.GID(coredata.RiskAnalysisScenarioEntityType))
	v.Check(r.ThreatID, "threat_id", validator.Required(), validator.GID(coredata.RiskAnalysisThreatEntityType))

	return v.Error()
}

func (r *LinkRiskAnalysisScenarioRiskRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisScenarioID, "risk_analysis_scenario_id", validator.Required(), validator.GID(coredata.RiskAnalysisScenarioEntityType))
	v.Check(r.RiskID, "risk_id", validator.Required(), validator.GID(coredata.RiskEntityType))

	return v.Error()
}

func (r *UnlinkRiskAnalysisScenarioRiskRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAnalysisScenarioID, "risk_analysis_scenario_id", validator.Required(), validator.GID(coredata.RiskAnalysisScenarioEntityType))
	v.Check(r.RiskID, "risk_id", validator.Required(), validator.GID(coredata.RiskEntityType))

	return v.Error()
}

func (r *UpdateRiskAnalysisScenarioRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAnalysisScenarioEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s *Service) Create(ctx context.Context, scope coredata.Scoper, req CreateRiskAnalysisRequest) (*coredata.RiskAnalysis, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	ra := &coredata.RiskAnalysis{
		ID:             gid.New(scope.GetTenantID(), coredata.RiskAnalysisEntityType),
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		MatrixRows:     req.MatrixSize.Rows,
		MatrixCols:     req.MatrixSize.Cols,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if req.Period != nil {
		ra.PeriodStart = req.Period.Start
		ra.PeriodEnd = req.Period.End
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := ra.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert risk assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return ra, nil
}

func (s *Service) Get(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAnalysis, error) {
	ra := &coredata.RiskAnalysis{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := ra.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load risk assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return ra, nil
}

func (s *Service) Update(ctx context.Context, scope coredata.Scoper, req UpdateRiskAnalysisRequest) (*coredata.RiskAnalysis, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	ra := &coredata.RiskAnalysis{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := ra.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk assessment: %w", err)
			}

			if req.Name != nil {
				ra.Name = *req.Name
			}

			if req.Description != nil {
				ra.Description = *req.Description
			}

			if req.Period != nil {
				ra.PeriodStart = req.Period.Start
				ra.PeriodEnd = req.Period.End
			}

			v := validator.New()
			validatePeriodRange(v, ra.PeriodStart, ra.PeriodEnd)

			if err := v.Error(); err != nil {
				return fmt.Errorf("invalid request: %w", err)
			}

			ra.UpdatedAt = time.Now()
			if err := ra.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update risk assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return ra, nil
}

func (s *Service) Delete(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			ra := &coredata.RiskAnalysis{}
			if err := ra.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk assessment: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisOrderField],
) (*page.Page[*coredata.RiskAnalysis, coredata.RiskAnalysisOrderField], error) {
	var results coredata.RiskAnalyses

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
				return fmt.Errorf("cannot list risk assessments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ras := &coredata.RiskAnalyses{}

			count, err = ras.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count risk assessments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CreateDiagram(ctx context.Context, scope coredata.Scoper, req CreateRiskAnalysisDiagramRequest) (*coredata.RiskAnalysisDiagram, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	raDiagram := &coredata.RiskAnalysisDiagram{
		ID:             gid.New(scope.GetTenantID(), coredata.RiskAnalysisDiagramEntityType),
		RiskAnalysisID: req.RiskAnalysisID,
		Name:           req.Name,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			ra := coredata.RiskAnalysis{}
			if err := ra.LoadByID(ctx, tx, scope, req.RiskAnalysisID); err != nil {
				return fmt.Errorf("cannot load risk assessment: %w", err)
			}

			raDiagram.OrganizationID = ra.OrganizationID
			if err := raDiagram.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert risk analysis diagram: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return raDiagram, nil
}

func (s *Service) GetDiagram(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAnalysisDiagram, error) {
	raDiagram := &coredata.RiskAnalysisDiagram{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := raDiagram.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load risk analysis diagram: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return raDiagram, nil
}

func (s *Service) UpdateDiagram(ctx context.Context, scope coredata.Scoper, req UpdateRiskAnalysisDiagramRequest) (*coredata.RiskAnalysisDiagram, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	raDiagram := &coredata.RiskAnalysisDiagram{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := raDiagram.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk analysis diagram: %w", err)
			}

			if req.Name != nil {
				raDiagram.Name = *req.Name
			}

			raDiagram.UpdatedAt = time.Now()
			if err := raDiagram.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update risk analysis diagram: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return raDiagram, nil
}

func (s *Service) DeleteDiagram(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raDiagram := &coredata.RiskAnalysisDiagram{}
			if err := raDiagram.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk analysis diagram: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListDiagramsForRiskAnalysisID(
	ctx context.Context,
	scope coredata.Scoper,
	riskAnalysisID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisDiagramOrderField],
) (*page.Page[*coredata.RiskAnalysisDiagram, coredata.RiskAnalysisDiagramOrderField], error) {
	var results coredata.RiskAnalysisDiagrams

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAnalysisID(ctx, conn, scope, riskAnalysisID, cursor); err != nil {
				return fmt.Errorf("cannot list risk analysis diagrams: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountDiagramsForRiskAnalysisID(ctx context.Context, scope coredata.Scoper, riskAnalysisID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ss := &coredata.RiskAnalysisDiagrams{}

			count, err = ss.CountByRiskAnalysisID(ctx, conn, scope, riskAnalysisID)
			if err != nil {
				return fmt.Errorf("cannot count risk analysis diagrams: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CreateNode(ctx context.Context, scope coredata.Scoper, req CreateRiskAnalysisNodeRequest) (*coredata.RiskAnalysisNode, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	node := &coredata.RiskAnalysisNode{
		ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisNodeEntityType),
		RiskAnalysisDiagramID: req.RiskAnalysisDiagramID,
		BoundaryID:            req.BoundaryID,
		NodeType:              req.NodeType,
		Name:                  req.Name,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raDiagram := coredata.RiskAnalysisDiagram{}
			if err := raDiagram.LoadByID(ctx, tx, scope, req.RiskAnalysisDiagramID); err != nil {
				return fmt.Errorf("cannot load risk analysis diagram: %w", err)
			}

			if req.BoundaryID != nil {
				if err := s.assertBoundaryInDiagram(ctx, tx, scope, *req.BoundaryID, req.RiskAnalysisDiagramID, "boundary_id"); err != nil {
					return err
				}
			}

			node.OrganizationID = raDiagram.OrganizationID
			if err := node.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert risk assessment node: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *Service) GetNode(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAnalysisNode, error) {
	node := &coredata.RiskAnalysisNode{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := node.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load risk assessment node: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *Service) UpdateNode(ctx context.Context, scope coredata.Scoper, req UpdateRiskAnalysisNodeRequest) (*coredata.RiskAnalysisNode, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	node := &coredata.RiskAnalysisNode{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := node.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk assessment node: %w", err)
			}

			if req.Name != nil {
				node.Name = *req.Name
			}

			if req.NodeType != nil {
				node.NodeType = *req.NodeType
			}

			if req.BoundaryID != nil {
				if *req.BoundaryID != nil {
					if err := s.assertBoundaryInDiagram(ctx, tx, scope, **req.BoundaryID, node.RiskAnalysisDiagramID, "boundary_id"); err != nil {
						return err
					}
				}

				node.BoundaryID = *req.BoundaryID
			}

			node.UpdatedAt = time.Now()
			if err := node.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update risk assessment node: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (s *Service) DeleteNode(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			node := &coredata.RiskAnalysisNode{}
			if err := node.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk assessment node: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListNodesForDiagramID(
	ctx context.Context,
	scope coredata.Scoper,
	diagramID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisNodeOrderField],
) (*page.Page[*coredata.RiskAnalysisNode, coredata.RiskAnalysisNodeOrderField], error) {
	var results coredata.RiskAnalysisNodes

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagramID, cursor); err != nil {
				return fmt.Errorf("cannot list risk assessment nodes: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountNodesForDiagramID(ctx context.Context, scope coredata.Scoper, diagramID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ns := &coredata.RiskAnalysisNodes{}

			count, err = ns.CountByRiskAnalysisDiagramID(ctx, conn, scope, diagramID)
			if err != nil {
				return fmt.Errorf("cannot count risk assessment nodes: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CreateBoundary(ctx context.Context, scope coredata.Scoper, req CreateRiskAnalysisBoundaryRequest) (*coredata.RiskAnalysisBoundary, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	boundary := &coredata.RiskAnalysisBoundary{
		ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisBoundaryEntityType),
		RiskAnalysisDiagramID: req.RiskAnalysisDiagramID,
		ParentBoundaryID:      req.ParentBoundaryID,
		Name:                  req.Name,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raDiagram := coredata.RiskAnalysisDiagram{}
			if err := raDiagram.LoadByID(ctx, tx, scope, req.RiskAnalysisDiagramID); err != nil {
				return fmt.Errorf("cannot load risk analysis diagram: %w", err)
			}

			if req.ParentBoundaryID != nil {
				if err := s.assertBoundaryInDiagram(ctx, tx, scope, *req.ParentBoundaryID, req.RiskAnalysisDiagramID, "parent_boundary_id"); err != nil {
					return err
				}
			}

			boundary.OrganizationID = raDiagram.OrganizationID
			if err := boundary.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert risk assessment boundary: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return boundary, nil
}

func (s *Service) GetBoundary(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAnalysisBoundary, error) {
	boundary := &coredata.RiskAnalysisBoundary{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := boundary.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load risk assessment boundary: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return boundary, nil
}

func (s *Service) UpdateBoundary(ctx context.Context, scope coredata.Scoper, req UpdateRiskAnalysisBoundaryRequest) (*coredata.RiskAnalysisBoundary, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	boundary := &coredata.RiskAnalysisBoundary{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := boundary.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk assessment boundary: %w", err)
			}

			if req.Name != nil {
				boundary.Name = *req.Name
			}

			if req.ParentBoundaryID != nil {
				if *req.ParentBoundaryID != nil {
					if err := s.assertBoundaryInDiagram(ctx, tx, scope, **req.ParentBoundaryID, boundary.RiskAnalysisDiagramID, "parent_boundary_id"); err != nil {
						return err
					}

					if err := s.assertNoBoundaryCycle(ctx, tx, scope, boundary.ID, **req.ParentBoundaryID, "parent_boundary_id"); err != nil {
						return err
					}
				}

				boundary.ParentBoundaryID = *req.ParentBoundaryID
			}

			boundary.UpdatedAt = time.Now()
			if err := boundary.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update risk assessment boundary: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return boundary, nil
}

func (s *Service) DeleteBoundary(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			boundary := &coredata.RiskAnalysisBoundary{}
			if err := boundary.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk assessment boundary: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListBoundariesForDiagramID(
	ctx context.Context,
	scope coredata.Scoper,
	diagramID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisBoundaryOrderField],
) (*page.Page[*coredata.RiskAnalysisBoundary, coredata.RiskAnalysisBoundaryOrderField], error) {
	var results coredata.RiskAnalysisBoundaries

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagramID, cursor); err != nil {
				return fmt.Errorf("cannot list risk assessment boundaries: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountBoundariesForDiagramID(ctx context.Context, scope coredata.Scoper, diagramID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			bs := &coredata.RiskAnalysisBoundaries{}

			count, err = bs.CountByRiskAnalysisDiagramID(ctx, conn, scope, diagramID)
			if err != nil {
				return fmt.Errorf("cannot count risk assessment boundaries: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CreateProcess(ctx context.Context, scope coredata.Scoper, req CreateRiskAnalysisProcessRequest) (*coredata.RiskAnalysisProcess, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	process := &coredata.RiskAnalysisProcess{
		ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisProcessEntityType),
		RiskAnalysisDiagramID: req.RiskAnalysisDiagramID,
		SourceNodeID:          req.SourceNodeID,
		TargetNodeID:          req.TargetNodeID,
		Name:                  req.Name,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raDiagram := coredata.RiskAnalysisDiagram{}
			if err := raDiagram.LoadByID(ctx, tx, scope, req.RiskAnalysisDiagramID); err != nil {
				return fmt.Errorf("cannot load risk analysis diagram: %w", err)
			}

			process.OrganizationID = raDiagram.OrganizationID

			if err := s.assertNodeInDiagram(ctx, tx, scope, req.SourceNodeID, req.RiskAnalysisDiagramID, "source_node_id"); err != nil {
				return err
			}

			if err := s.assertNodeInDiagram(ctx, tx, scope, req.TargetNodeID, req.RiskAnalysisDiagramID, "target_node_id"); err != nil {
				return err
			}

			if err := process.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert risk assessment process: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return process, nil
}

func (s *Service) GetProcess(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAnalysisProcess, error) {
	process := &coredata.RiskAnalysisProcess{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := process.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load risk assessment process: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return process, nil
}

func (s *Service) UpdateProcess(ctx context.Context, scope coredata.Scoper, req UpdateRiskAnalysisProcessRequest) (*coredata.RiskAnalysisProcess, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	process := &coredata.RiskAnalysisProcess{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := process.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk assessment process: %w", err)
			}

			if req.SourceNodeID != nil {
				if err := s.assertNodeInDiagram(ctx, tx, scope, *req.SourceNodeID, process.RiskAnalysisDiagramID, "source_node_id"); err != nil {
					return err
				}

				process.SourceNodeID = *req.SourceNodeID
			}

			if req.TargetNodeID != nil {
				if err := s.assertNodeInDiagram(ctx, tx, scope, *req.TargetNodeID, process.RiskAnalysisDiagramID, "target_node_id"); err != nil {
					return err
				}

				process.TargetNodeID = *req.TargetNodeID
			}

			if req.Name != nil {
				process.Name = *req.Name
			}

			process.UpdatedAt = time.Now()
			if err := process.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update risk assessment process: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return process, nil
}

func (s *Service) DeleteProcess(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			process := &coredata.RiskAnalysisProcess{}
			if err := process.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk assessment process: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListProcessesForDiagramID(
	ctx context.Context,
	scope coredata.Scoper,
	diagramID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisProcessOrderField],
) (*page.Page[*coredata.RiskAnalysisProcess, coredata.RiskAnalysisProcessOrderField], error) {
	var results coredata.RiskAnalysisProcesses

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagramID, cursor); err != nil {
				return fmt.Errorf("cannot list risk assessment processes: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountProcessesForDiagramID(ctx context.Context, scope coredata.Scoper, diagramID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ps := &coredata.RiskAnalysisProcesses{}

			count, err = ps.CountByRiskAnalysisDiagramID(ctx, conn, scope, diagramID)
			if err != nil {
				return fmt.Errorf("cannot count risk assessment processes: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CreateThreat(ctx context.Context, scope coredata.Scoper, req CreateRiskAnalysisThreatRequest) (*coredata.RiskAnalysisThreat, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	threat := &coredata.RiskAnalysisThreat{
		ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisThreatEntityType),
		RiskAnalysisDiagramID: req.RiskAnalysisDiagramID,
		ProcessID:             req.ProcessID,
		Name:                  req.Name,
		Category:              req.Category,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raDiagram := coredata.RiskAnalysisDiagram{}
			if err := raDiagram.LoadByID(ctx, tx, scope, req.RiskAnalysisDiagramID); err != nil {
				return fmt.Errorf("cannot load risk analysis diagram: %w", err)
			}

			threat.OrganizationID = raDiagram.OrganizationID

			if err := s.assertProcessInDiagram(ctx, tx, scope, req.ProcessID, req.RiskAnalysisDiagramID, "process_id"); err != nil {
				return err
			}

			if err := threat.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert risk threat: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return threat, nil
}

func (s *Service) GetThreat(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAnalysisThreat, error) {
	threat := &coredata.RiskAnalysisThreat{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := threat.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load risk threat: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return threat, nil
}

func (s *Service) UpdateThreat(ctx context.Context, scope coredata.Scoper, req UpdateRiskAnalysisThreatRequest) (*coredata.RiskAnalysisThreat, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	threat := &coredata.RiskAnalysisThreat{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := threat.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk threat: %w", err)
			}

			if req.ProcessID != nil {
				if err := s.assertProcessInDiagram(ctx, tx, scope, *req.ProcessID, threat.RiskAnalysisDiagramID, "process_id"); err != nil {
					return err
				}

				threat.ProcessID = *req.ProcessID
			}

			if req.Name != nil {
				threat.Name = *req.Name
			}

			if req.Category != nil {
				threat.Category = *req.Category
			}

			threat.UpdatedAt = time.Now()
			if err := threat.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update risk threat: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return threat, nil
}

func (s *Service) DeleteThreat(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			threat := &coredata.RiskAnalysisThreat{}
			if err := threat.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk threat: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListThreatsForDiagramID(
	ctx context.Context,
	scope coredata.Scoper,
	diagramID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisThreatOrderField],
) (*page.Page[*coredata.RiskAnalysisThreat, coredata.RiskAnalysisThreatOrderField], error) {
	var results coredata.RiskAnalysisThreats

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagramID, cursor); err != nil {
				return fmt.Errorf("cannot list risk threats: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountThreatsForDiagramID(ctx context.Context, scope coredata.Scoper, diagramID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ts := &coredata.RiskAnalysisThreats{}

			count, err = ts.CountByRiskAnalysisDiagramID(ctx, conn, scope, diagramID)
			if err != nil {
				return fmt.Errorf("cannot count risk threats: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CreateScenario(ctx context.Context, scope coredata.Scoper, req CreateRiskAnalysisScenarioRequest) (*coredata.RiskAnalysisScenario, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	scenario := &coredata.RiskAnalysisScenario{
		ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisScenarioEntityType),
		RiskAnalysisDiagramID: req.RiskAnalysisDiagramID,
		Name:                  req.Name,
		Description:           req.Description,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raDiagram := coredata.RiskAnalysisDiagram{}
			if err := raDiagram.LoadByID(ctx, tx, scope, req.RiskAnalysisDiagramID); err != nil {
				return fmt.Errorf("cannot load risk analysis diagram: %w", err)
			}

			scenario.OrganizationID = raDiagram.OrganizationID
			if err := scenario.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert risk scenario: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return scenario, nil
}

func (s *Service) GetScenario(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAnalysisScenario, error) {
	scenario := &coredata.RiskAnalysisScenario{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := scenario.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load risk scenario: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return scenario, nil
}

func (s *Service) UpdateScenario(ctx context.Context, scope coredata.Scoper, req UpdateRiskAnalysisScenarioRequest) (*coredata.RiskAnalysisScenario, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	scenario := &coredata.RiskAnalysisScenario{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := scenario.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk scenario: %w", err)
			}

			if req.Name != nil {
				scenario.Name = *req.Name
			}

			if req.Description != nil {
				scenario.Description = *req.Description
			}

			scenario.UpdatedAt = time.Now()
			if err := scenario.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update risk scenario: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return scenario, nil
}

func (s *Service) DeleteScenario(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scenario := &coredata.RiskAnalysisScenario{}
			if err := scenario.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk scenario: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListScenariosForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisScenarioOrderField],
) (*page.Page[*coredata.RiskAnalysisScenario, coredata.RiskAnalysisScenarioOrderField], error) {
	var results coredata.RiskAnalysisScenarios

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
				return fmt.Errorf("cannot list risk scenarios: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountScenariosForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ss := &coredata.RiskAnalysisScenarios{}

			count, err = ss.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count risk scenarios: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListScenariosForRiskID(
	ctx context.Context,
	scope coredata.Scoper,
	riskID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisScenarioOrderField],
) (*page.Page[*coredata.RiskAnalysisScenario, coredata.RiskAnalysisScenarioOrderField], error) {
	var results coredata.RiskAnalysisScenarios

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskID(ctx, conn, scope, riskID, cursor); err != nil {
				return fmt.Errorf("cannot list risk scenarios: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountScenariosForRiskID(ctx context.Context, scope coredata.Scoper, riskID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ss := &coredata.RiskAnalysisScenarios{}

			count, err = ss.CountByRiskID(ctx, conn, scope, riskID)
			if err != nil {
				return fmt.Errorf("cannot count risk scenarios: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListScenariosForDiagramID(
	ctx context.Context,
	scope coredata.Scoper,
	diagramID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisScenarioOrderField],
) (*page.Page[*coredata.RiskAnalysisScenario, coredata.RiskAnalysisScenarioOrderField], error) {
	var results coredata.RiskAnalysisScenarios

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagramID, cursor); err != nil {
				return fmt.Errorf("cannot list risk scenarios: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountScenariosForDiagramID(ctx context.Context, scope coredata.Scoper, diagramID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ss := &coredata.RiskAnalysisScenarios{}

			count, err = ss.CountByRiskAnalysisDiagramID(ctx, conn, scope, diagramID)
			if err != nil {
				return fmt.Errorf("cannot count risk scenarios: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) LinkScenarioThreat(ctx context.Context, scope coredata.Scoper, req LinkRiskAnalysisScenarioThreatRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scenario := coredata.RiskAnalysisScenario{}
			if err := scenario.LoadByID(ctx, tx, scope, req.RiskAnalysisScenarioID); err != nil {
				return fmt.Errorf("cannot load risk scenario: %w", err)
			}

			threat := coredata.RiskAnalysisThreat{}
			if err := threat.LoadByID(ctx, tx, scope, req.ThreatID); err != nil {
				return fmt.Errorf("cannot load threat: %w", err)
			}

			if scenario.OrganizationID != threat.OrganizationID {
				return validator.ValidationErrors{{
					Field:   "threat_id",
					Code:    validator.ErrorCodeCustom,
					Message: "threat and scenario must belong to the same organization",
				}}
			}

			link := &coredata.RiskAnalysisScenarioThreat{
				RiskAnalysisScenarioID: req.RiskAnalysisScenarioID,
				RiskAnalysisThreatID:   req.ThreatID,
				CreatedAt:              time.Now(),
			}
			if err := link.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot link scenario threat: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) UnlinkScenarioThreat(ctx context.Context, scope coredata.Scoper, req UnlinkRiskAnalysisScenarioThreatRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.RiskAnalysisScenarioThreat{
				RiskAnalysisScenarioID: req.RiskAnalysisScenarioID,
				RiskAnalysisThreatID:   req.ThreatID,
			}
			if err := link.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot unlink scenario threat: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) LinkScenarioRisk(ctx context.Context, scope coredata.Scoper, req LinkRiskAnalysisScenarioRiskRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scenario := coredata.RiskAnalysisScenario{}
			if err := scenario.LoadByID(ctx, tx, scope, req.RiskAnalysisScenarioID); err != nil {
				return fmt.Errorf("cannot load risk scenario: %w", err)
			}

			risk := coredata.Risk{}
			if err := risk.LoadByID(ctx, tx, scope, req.RiskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			if scenario.OrganizationID != risk.OrganizationID {
				return validator.ValidationErrors{{
					Field:   "risk_id",
					Code:    validator.ErrorCodeCustom,
					Message: "risk and scenario must belong to the same organization",
				}}
			}

			link := &coredata.RiskAnalysisScenarioRisk{
				RiskAnalysisScenarioID: req.RiskAnalysisScenarioID,
				RiskID:                 req.RiskID,
				CreatedAt:              time.Now(),
			}
			if err := link.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot link scenario risk: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) UnlinkScenarioRisk(ctx context.Context, scope coredata.Scoper, req UnlinkRiskAnalysisScenarioRiskRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.RiskAnalysisScenarioRisk{
				RiskAnalysisScenarioID: req.RiskAnalysisScenarioID,
				RiskID:                 req.RiskID,
			}
			if err := link.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot unlink scenario risk: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListThreatsForScenarioID(
	ctx context.Context,
	scope coredata.Scoper,
	scenarioID gid.GID,
	cursor *page.Cursor[coredata.RiskAnalysisThreatOrderField],
) (*page.Page[*coredata.RiskAnalysisThreat, coredata.RiskAnalysisThreatOrderField], error) {
	var results coredata.RiskAnalysisThreats

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByScenarioID(ctx, conn, scope, scenarioID, cursor); err != nil {
				return fmt.Errorf("cannot list scenario threats: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountThreatsForScenarioID(ctx context.Context, scope coredata.Scoper, scenarioID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ts := &coredata.RiskAnalysisThreats{}

			count, err = ts.CountByScenarioID(ctx, conn, scope, scenarioID)
			if err != nil {
				return fmt.Errorf("cannot count scenario threats: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListRisksForScenarioID(
	ctx context.Context,
	scope coredata.Scoper,
	scenarioID gid.GID,
	cursor *page.Cursor[coredata.RiskOrderField],
) (*page.Page[*coredata.Risk, coredata.RiskOrderField], error) {
	var results coredata.Risks

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByScenarioID(ctx, conn, scope, scenarioID, cursor); err != nil {
				return fmt.Errorf("cannot list scenario risks: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func untreatedScenarioRiskIDs(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	analysisID gid.GID,
) ([]gid.GID, error) {
	diagrams, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.RiskAnalysisDiagramOrderField]{
			Field:     coredata.RiskAnalysisDiagramOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisDiagramOrderField]) ([]*coredata.RiskAnalysisDiagram, error) {
			var batch coredata.RiskAnalysisDiagrams
			if err := batch.LoadByRiskAnalysisID(ctx, conn, scope, analysisID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load risk analysis diagrams: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return nil, err
	}

	var scenarioIDs []gid.GID

	for _, diagram := range diagrams {
		scenarios, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.RiskAnalysisScenarioOrderField]{
				Field:     coredata.RiskAnalysisScenarioOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisScenarioOrderField]) ([]*coredata.RiskAnalysisScenario, error) {
				var batch coredata.RiskAnalysisScenarios
				if err := batch.LoadByRiskAnalysisDiagramID(ctx, conn, scope, diagram.ID, cursor); err != nil {
					return nil, fmt.Errorf("cannot load risk analysis scenarios: %w", err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return nil, err
		}

		for _, scenario := range scenarios {
			scenarioIDs = append(scenarioIDs, scenario.ID)
		}
	}

	var links coredata.RiskAnalysisScenarioRisks
	if err := links.LoadByScenarioIDs(ctx, conn, scope, scenarioIDs); err != nil {
		return nil, fmt.Errorf("cannot load scenario risks: %w", err)
	}

	plans, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.TreatmentPlanOrderField]{
			Field:     coredata.TreatmentPlanOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.TreatmentPlanOrderField]) ([]*coredata.TreatmentPlan, error) {
			var batch coredata.TreatmentPlans
			if err := batch.LoadByRiskAnalysisID(
				ctx,
				conn,
				scope,
				analysisID,
				cursor,
				coredata.NewTreatmentPlanFilter(nil, nil, nil),
			); err != nil {
				return nil, fmt.Errorf("cannot load treatment plans: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return nil, err
	}

	planRiskIDs := make([]gid.GID, 0, len(plans))
	for _, plan := range plans {
		planRiskIDs = append(planRiskIDs, plan.RiskID)
	}

	planned := gid.NewSet(planRiskIDs...)

	seen := gid.NewSet()
	untreated := make([]gid.GID, 0)

	for _, link := range links {
		if planned.Contains(link.RiskID) || seen.Contains(link.RiskID) {
			continue
		}

		seen[link.RiskID] = struct{}{}
		untreated = append(untreated, link.RiskID)
	}

	return untreated, nil
}

func (s *Service) ListRisksForRiskAnalysisID(
	ctx context.Context,
	scope coredata.Scoper,
	analysisID gid.GID,
	cursor *page.Cursor[coredata.RiskOrderField],
	filter *coredata.RiskFilter,
) (*page.Page[*coredata.Risk, coredata.RiskOrderField], error) {
	var results coredata.Risks

	if filter == nil {
		filter = coredata.NewRiskFilter(nil)
	}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			riskIDs, err := untreatedScenarioRiskIDs(ctx, conn, scope, analysisID)
			if err != nil {
				return err
			}

			if err := results.LoadByRiskIDs(ctx, conn, scope, riskIDs, cursor, filter); err != nil {
				return fmt.Errorf("cannot list scenario risks on analysis: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(results, cursor), nil
}

func (s *Service) CountRisksForRiskAnalysisID(
	ctx context.Context,
	scope coredata.Scoper,
	analysisID gid.GID,
	filter *coredata.RiskFilter,
) (int, error) {
	var count int

	if filter == nil {
		filter = coredata.NewRiskFilter(nil)
	}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			riskIDs, err := untreatedScenarioRiskIDs(ctx, conn, scope, analysisID)
			if err != nil {
				return err
			}

			rs := &coredata.Risks{}

			count, err = rs.CountByRiskIDs(ctx, conn, scope, riskIDs, filter)
			if err != nil {
				return fmt.Errorf("cannot count scenario risks on analysis: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CountRisksForScenarioID(ctx context.Context, scope coredata.Scoper, scenarioID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			rs := &coredata.Risks{}

			count, err = rs.CountByScenarioID(ctx, conn, scope, scenarioID)
			if err != nil {
				return fmt.Errorf("cannot count scenario risks: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) assertNodeInDiagram(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	nodeID gid.GID,
	diagramID gid.GID,
	field string,
) error {
	node := &coredata.RiskAnalysisNode{}
	if err := node.LoadByID(ctx, tx, scope, nodeID); err != nil {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "node not found",
		}}
	}

	if node.RiskAnalysisDiagramID != diagramID {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "node does not belong to this diagram",
		}}
	}

	return nil
}

func (s *Service) assertBoundaryInDiagram(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	boundaryID gid.GID,
	diagramID gid.GID,
	field string,
) error {
	boundary := &coredata.RiskAnalysisBoundary{}
	if err := boundary.LoadByID(ctx, tx, scope, boundaryID); err != nil {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "boundary not found",
		}}
	}

	// A boundary in a different diagram is reported identically to a missing
	// one so the error does not reveal that the resource exists elsewhere.
	if boundary.RiskAnalysisDiagramID != diagramID {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "boundary not found",
		}}
	}

	return nil
}

// assertNoBoundaryCycle walks the ancestor chain starting from the proposed
// parent. If it reaches the boundary being updated, the new parent would make
// the boundary an ancestor of itself (a cycle), which is rejected. A visited
// set guards against any pre-existing cycle in stored data.
func (s *Service) assertNoBoundaryCycle(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	boundaryID gid.GID,
	proposedParentID gid.GID,
	field string,
) error {
	visited := make(map[gid.GID]bool)
	currentID := proposedParentID

	for {
		if currentID == boundaryID {
			return validator.ValidationErrors{{
				Field:   field,
				Code:    validator.ErrorCodeCustom,
				Message: "boundary cannot be nested under itself or one of its descendants",
			}}
		}

		if visited[currentID] {
			return nil
		}

		visited[currentID] = true

		current := &coredata.RiskAnalysisBoundary{}
		if err := current.LoadByID(ctx, tx, scope, currentID); err != nil {
			return fmt.Errorf("cannot load parent boundary: %w", err)
		}

		if current.ParentBoundaryID == nil {
			return nil
		}

		currentID = *current.ParentBoundaryID
	}
}

func (s *Service) assertProcessInDiagram(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	processID gid.GID,
	diagramID gid.GID,
	field string,
) error {
	process := &coredata.RiskAnalysisProcess{}
	if err := process.LoadByID(ctx, tx, scope, processID); err != nil {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "process not found",
		}}
	}

	if process.RiskAnalysisDiagramID != diagramID {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "process does not belong to this diagram",
		}}
	}

	return nil
}
