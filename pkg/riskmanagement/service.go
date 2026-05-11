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
	CreateRiskAssessmentRequest struct {
		OrganizationID gid.GID
		Name           string
		Description    *string
	}

	UpdateRiskAssessmentRequest struct {
		ID          gid.GID
		Name        *string
		Description **string
	}

	CreateRiskAssessmentScopeRequest struct {
		RiskAssessmentID gid.GID
		Name             string
	}

	UpdateRiskAssessmentScopeRequest struct {
		ID   gid.GID
		Name *string
	}

	CreateRiskAssessmentNodeRequest struct {
		RiskAssessmentScopeID gid.GID
		NodeType              coredata.RiskAssessmentNodeType
		Name                  string
	}

	UpdateRiskAssessmentNodeRequest struct {
		ID       gid.GID
		NodeType *coredata.RiskAssessmentNodeType
		Name     *string
	}

	CreateRiskAssessmentProcessRequest struct {
		RiskAssessmentScopeID gid.GID
		SourceNodeID          gid.GID
		TargetNodeID          gid.GID
		Name                  string
	}

	UpdateRiskAssessmentProcessRequest struct {
		ID           gid.GID
		SourceNodeID *gid.GID
		TargetNodeID *gid.GID
		Name         *string
	}

	CreateRiskAssessmentThreatRequest struct {
		RiskAssessmentScopeID gid.GID
		ProcessID             gid.GID
		Name                  string
		Category              string
	}

	UpdateRiskAssessmentThreatRequest struct {
		ID        gid.GID
		ProcessID *gid.GID
		Name      *string
		Category  *string
	}

	CreateRiskAssessmentScenarioRequest struct {
		RiskAssessmentScopeID gid.GID
		Name                  string
		Description           *string
	}

	UpdateRiskAssessmentScenarioRequest struct {
		ID          gid.GID
		Name        *string
		Description **string
	}

	LinkRiskAssessmentScenarioThreatRequest struct {
		RiskAssessmentScenarioID gid.GID
		ThreatID                 gid.GID
	}

	UnlinkRiskAssessmentScenarioThreatRequest struct {
		RiskAssessmentScenarioID gid.GID
		ThreatID                 gid.GID
	}

	LinkRiskAssessmentScenarioRiskRequest struct {
		RiskAssessmentScenarioID gid.GID
		RiskID                   gid.GID
	}

	UnlinkRiskAssessmentScenarioRiskRequest struct {
		RiskAssessmentScenarioID gid.GID
		RiskID                   gid.GID
	}
)

func (r *CreateRiskAssessmentRequest) Validate() error {
	v := validator.New()
	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	return v.Error()
}

func (r *UpdateRiskAssessmentRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAssessmentEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	return v.Error()
}

func (r *CreateRiskAssessmentScopeRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAssessmentID, "risk_assessment_id", validator.Required(), validator.GID(coredata.RiskAssessmentEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	return v.Error()
}

func (r *UpdateRiskAssessmentScopeRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAssessmentScopeEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	return v.Error()
}

func (r *CreateRiskAssessmentNodeRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAssessmentScopeID, "risk_assessment_scope_id", validator.Required(), validator.GID(coredata.RiskAssessmentScopeEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.NodeType, "node_type", validator.Required(), validator.OneOfSlice(coredata.RiskAssessmentNodeTypes()))
	return v.Error()
}

func (r *UpdateRiskAssessmentNodeRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAssessmentNodeEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.NodeType, "node_type", validator.OneOfSlice(coredata.RiskAssessmentNodeTypes()))
	return v.Error()
}

func (r *CreateRiskAssessmentProcessRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAssessmentScopeID, "risk_assessment_scope_id", validator.Required(), validator.GID(coredata.RiskAssessmentScopeEntityType))
	v.Check(r.SourceNodeID, "source_node_id", validator.Required(), validator.GID(coredata.RiskAssessmentNodeEntityType))
	v.Check(r.TargetNodeID, "target_node_id", validator.Required(), validator.GID(coredata.RiskAssessmentNodeEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	return v.Error()
}

func (r *UpdateRiskAssessmentProcessRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAssessmentProcessEntityType))
	v.Check(r.SourceNodeID, "source_node_id", validator.GID(coredata.RiskAssessmentNodeEntityType))
	v.Check(r.TargetNodeID, "target_node_id", validator.GID(coredata.RiskAssessmentNodeEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	return v.Error()
}

func (r *CreateRiskAssessmentThreatRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAssessmentScopeID, "risk_assessment_scope_id", validator.Required(), validator.GID(coredata.RiskAssessmentScopeEntityType))
	v.Check(r.ProcessID, "process_id", validator.Required(), validator.GID(coredata.RiskAssessmentProcessEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Category, "category", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	return v.Error()
}

func (r *UpdateRiskAssessmentThreatRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAssessmentThreatEntityType))
	v.Check(r.ProcessID, "process_id", validator.GID(coredata.RiskAssessmentProcessEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Category, "category", validator.SafeTextNoNewLine(TitleMaxLength))
	return v.Error()
}

func (r *CreateRiskAssessmentScenarioRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAssessmentScopeID, "risk_assessment_scope_id", validator.Required(), validator.GID(coredata.RiskAssessmentScopeEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	return v.Error()
}

func (r *LinkRiskAssessmentScenarioThreatRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAssessmentScenarioID, "risk_scenario_id", validator.Required(), validator.GID(coredata.RiskAssessmentScenarioEntityType))
	v.Check(r.ThreatID, "threat_id", validator.Required(), validator.GID(coredata.RiskAssessmentThreatEntityType))
	return v.Error()
}

func (r *UnlinkRiskAssessmentScenarioThreatRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAssessmentScenarioID, "risk_scenario_id", validator.Required(), validator.GID(coredata.RiskAssessmentScenarioEntityType))
	v.Check(r.ThreatID, "threat_id", validator.Required(), validator.GID(coredata.RiskAssessmentThreatEntityType))
	return v.Error()
}

func (r *LinkRiskAssessmentScenarioRiskRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAssessmentScenarioID, "risk_scenario_id", validator.Required(), validator.GID(coredata.RiskAssessmentScenarioEntityType))
	v.Check(r.RiskID, "risk_id", validator.Required(), validator.GID(coredata.RiskEntityType))
	return v.Error()
}

func (r *UnlinkRiskAssessmentScenarioRiskRequest) Validate() error {
	v := validator.New()
	v.Check(r.RiskAssessmentScenarioID, "risk_scenario_id", validator.Required(), validator.GID(coredata.RiskAssessmentScenarioEntityType))
	v.Check(r.RiskID, "risk_id", validator.Required(), validator.GID(coredata.RiskEntityType))
	return v.Error()
}

func (r *UpdateRiskAssessmentScenarioRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.RiskAssessmentScenarioEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	return v.Error()
}

func (s *Service) Create(ctx context.Context, scope coredata.Scoper, req CreateRiskAssessmentRequest) (*coredata.RiskAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	ra := &coredata.RiskAssessment{
		ID:             gid.New(scope.GetTenantID(), coredata.RiskAssessmentEntityType),
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		CreatedAt:      now,
		UpdatedAt:      now,
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

func (s *Service) Get(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAssessment, error) {
	ra := &coredata.RiskAssessment{}
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

func (s *Service) Update(ctx context.Context, scope coredata.Scoper, req UpdateRiskAssessmentRequest) (*coredata.RiskAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	ra := &coredata.RiskAssessment{}
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
			ra := &coredata.RiskAssessment{}
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
	cursor *page.Cursor[coredata.RiskAssessmentOrderField],
) (*page.Page[*coredata.RiskAssessment, coredata.RiskAssessmentOrderField], error) {
	var results coredata.RiskAssessments
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
			ras := &coredata.RiskAssessments{}
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

func (s *Service) CreateScope(ctx context.Context, scope coredata.Scoper, req CreateRiskAssessmentScopeRequest) (*coredata.RiskAssessmentScope, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	raScope := &coredata.RiskAssessmentScope{
		ID:               gid.New(scope.GetTenantID(), coredata.RiskAssessmentScopeEntityType),
		RiskAssessmentID: req.RiskAssessmentID,
		Name:             req.Name,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			ra := coredata.RiskAssessment{}
			if err := ra.LoadByID(ctx, tx, scope, req.RiskAssessmentID); err != nil {
				return fmt.Errorf("cannot load risk assessment: %w", err)
			}
			raScope.OrganizationID = ra.OrganizationID
			if err := raScope.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert risk assessment scope: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return raScope, nil
}

func (s *Service) GetScope(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAssessmentScope, error) {
	raScope := &coredata.RiskAssessmentScope{}
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := raScope.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load risk assessment scope: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return raScope, nil
}

func (s *Service) UpdateScope(ctx context.Context, scope coredata.Scoper, req UpdateRiskAssessmentScopeRequest) (*coredata.RiskAssessmentScope, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	raScope := &coredata.RiskAssessmentScope{}
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := raScope.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk assessment scope: %w", err)
			}
			if req.Name != nil {
				raScope.Name = *req.Name
			}
			raScope.UpdatedAt = time.Now()
			if err := raScope.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update risk assessment scope: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return raScope, nil
}

func (s *Service) DeleteScope(ctx context.Context, scope coredata.Scoper, id gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raScope := &coredata.RiskAssessmentScope{}
			if err := raScope.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk assessment scope: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) ListScopesForRiskAssessmentID(
	ctx context.Context,
	scope coredata.Scoper,
	riskAssessmentID gid.GID,
	cursor *page.Cursor[coredata.RiskAssessmentScopeOrderField],
) (*page.Page[*coredata.RiskAssessmentScope, coredata.RiskAssessmentScopeOrderField], error) {
	var results coredata.RiskAssessmentScopes
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAssessmentID(ctx, conn, scope, riskAssessmentID, cursor); err != nil {
				return fmt.Errorf("cannot list risk assessment scopes: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return page.NewPage(results, cursor), nil
}

func (s *Service) CountScopesForRiskAssessmentID(ctx context.Context, scope coredata.Scoper, riskAssessmentID gid.GID) (int, error) {
	var count int
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ss := &coredata.RiskAssessmentScopes{}
			count, err = ss.CountByRiskAssessmentID(ctx, conn, scope, riskAssessmentID)
			if err != nil {
				return fmt.Errorf("cannot count risk assessment scopes: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Service) CreateNode(ctx context.Context, scope coredata.Scoper, req CreateRiskAssessmentNodeRequest) (*coredata.RiskAssessmentNode, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	node := &coredata.RiskAssessmentNode{
		ID:                    gid.New(scope.GetTenantID(), coredata.RiskAssessmentNodeEntityType),
		RiskAssessmentScopeID: req.RiskAssessmentScopeID,
		NodeType:              req.NodeType,
		Name:                  req.Name,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raScope := coredata.RiskAssessmentScope{}
			if err := raScope.LoadByID(ctx, tx, scope, req.RiskAssessmentScopeID); err != nil {
				return fmt.Errorf("cannot load risk assessment scope: %w", err)
			}
			node.OrganizationID = raScope.OrganizationID
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

func (s *Service) GetNode(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAssessmentNode, error) {
	node := &coredata.RiskAssessmentNode{}
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

func (s *Service) UpdateNode(ctx context.Context, scope coredata.Scoper, req UpdateRiskAssessmentNodeRequest) (*coredata.RiskAssessmentNode, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	node := &coredata.RiskAssessmentNode{}
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
			node := &coredata.RiskAssessmentNode{}
			if err := node.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk assessment node: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) ListNodesForScopeID(
	ctx context.Context,
	scope coredata.Scoper,
	scopeID gid.GID,
	cursor *page.Cursor[coredata.RiskAssessmentNodeOrderField],
) (*page.Page[*coredata.RiskAssessmentNode, coredata.RiskAssessmentNodeOrderField], error) {
	var results coredata.RiskAssessmentNodes
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAssessmentScopeID(ctx, conn, scope, scopeID, cursor); err != nil {
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

func (s *Service) CountNodesForScopeID(ctx context.Context, scope coredata.Scoper, scopeID gid.GID) (int, error) {
	var count int
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ns := &coredata.RiskAssessmentNodes{}
			count, err = ns.CountByRiskAssessmentScopeID(ctx, conn, scope, scopeID)
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

func (s *Service) CreateProcess(ctx context.Context, scope coredata.Scoper, req CreateRiskAssessmentProcessRequest) (*coredata.RiskAssessmentProcess, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	process := &coredata.RiskAssessmentProcess{
		ID:                    gid.New(scope.GetTenantID(), coredata.RiskAssessmentProcessEntityType),
		RiskAssessmentScopeID: req.RiskAssessmentScopeID,
		SourceNodeID:          req.SourceNodeID,
		TargetNodeID:          req.TargetNodeID,
		Name:                  req.Name,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raScope := coredata.RiskAssessmentScope{}
			if err := raScope.LoadByID(ctx, tx, scope, req.RiskAssessmentScopeID); err != nil {
				return fmt.Errorf("cannot load risk assessment scope: %w", err)
			}
			process.OrganizationID = raScope.OrganizationID
			if err := s.assertNodeInScope(ctx, tx, scope, req.SourceNodeID, req.RiskAssessmentScopeID, "source_node_id"); err != nil {
				return err
			}
			if err := s.assertNodeInScope(ctx, tx, scope, req.TargetNodeID, req.RiskAssessmentScopeID, "target_node_id"); err != nil {
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

func (s *Service) GetProcess(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAssessmentProcess, error) {
	process := &coredata.RiskAssessmentProcess{}
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

func (s *Service) UpdateProcess(ctx context.Context, scope coredata.Scoper, req UpdateRiskAssessmentProcessRequest) (*coredata.RiskAssessmentProcess, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	process := &coredata.RiskAssessmentProcess{}
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := process.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk assessment process: %w", err)
			}
			if req.SourceNodeID != nil {
				if err := s.assertNodeInScope(ctx, tx, scope, *req.SourceNodeID, process.RiskAssessmentScopeID, "source_node_id"); err != nil {
					return err
				}
				process.SourceNodeID = *req.SourceNodeID
			}
			if req.TargetNodeID != nil {
				if err := s.assertNodeInScope(ctx, tx, scope, *req.TargetNodeID, process.RiskAssessmentScopeID, "target_node_id"); err != nil {
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
			process := &coredata.RiskAssessmentProcess{}
			if err := process.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk assessment process: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) ListProcessesForScopeID(
	ctx context.Context,
	scope coredata.Scoper,
	scopeID gid.GID,
	cursor *page.Cursor[coredata.RiskAssessmentProcessOrderField],
) (*page.Page[*coredata.RiskAssessmentProcess, coredata.RiskAssessmentProcessOrderField], error) {
	var results coredata.RiskAssessmentProcesses
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAssessmentScopeID(ctx, conn, scope, scopeID, cursor); err != nil {
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

func (s *Service) CountProcessesForScopeID(ctx context.Context, scope coredata.Scoper, scopeID gid.GID) (int, error) {
	var count int
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ps := &coredata.RiskAssessmentProcesses{}
			count, err = ps.CountByRiskAssessmentScopeID(ctx, conn, scope, scopeID)
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

func (s *Service) CreateThreat(ctx context.Context, scope coredata.Scoper, req CreateRiskAssessmentThreatRequest) (*coredata.RiskAssessmentThreat, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	threat := &coredata.RiskAssessmentThreat{
		ID:                    gid.New(scope.GetTenantID(), coredata.RiskAssessmentThreatEntityType),
		RiskAssessmentScopeID: req.RiskAssessmentScopeID,
		ProcessID:             req.ProcessID,
		Name:                  req.Name,
		Category:              req.Category,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raScope := coredata.RiskAssessmentScope{}
			if err := raScope.LoadByID(ctx, tx, scope, req.RiskAssessmentScopeID); err != nil {
				return fmt.Errorf("cannot load risk assessment scope: %w", err)
			}
			threat.OrganizationID = raScope.OrganizationID
			if err := s.assertProcessInScope(ctx, tx, scope, req.ProcessID, req.RiskAssessmentScopeID, "process_id"); err != nil {
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

func (s *Service) GetThreat(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAssessmentThreat, error) {
	threat := &coredata.RiskAssessmentThreat{}
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

func (s *Service) UpdateThreat(ctx context.Context, scope coredata.Scoper, req UpdateRiskAssessmentThreatRequest) (*coredata.RiskAssessmentThreat, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	threat := &coredata.RiskAssessmentThreat{}
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := threat.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk threat: %w", err)
			}
			if req.ProcessID != nil {
				if err := s.assertProcessInScope(ctx, tx, scope, *req.ProcessID, threat.RiskAssessmentScopeID, "process_id"); err != nil {
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
			threat := &coredata.RiskAssessmentThreat{}
			if err := threat.Delete(ctx, tx, scope, id); err != nil {
				return fmt.Errorf("cannot delete risk threat: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) ListThreatsForScopeID(
	ctx context.Context,
	scope coredata.Scoper,
	scopeID gid.GID,
	cursor *page.Cursor[coredata.RiskAssessmentThreatOrderField],
) (*page.Page[*coredata.RiskAssessmentThreat, coredata.RiskAssessmentThreatOrderField], error) {
	var results coredata.RiskAssessmentThreats
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAssessmentScopeID(ctx, conn, scope, scopeID, cursor); err != nil {
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

func (s *Service) CountThreatsForScopeID(ctx context.Context, scope coredata.Scoper, scopeID gid.GID) (int, error) {
	var count int
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ts := &coredata.RiskAssessmentThreats{}
			count, err = ts.CountByRiskAssessmentScopeID(ctx, conn, scope, scopeID)
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

func (s *Service) CreateScenario(ctx context.Context, scope coredata.Scoper, req CreateRiskAssessmentScenarioRequest) (*coredata.RiskAssessmentScenario, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	scenario := &coredata.RiskAssessmentScenario{
		ID:                    gid.New(scope.GetTenantID(), coredata.RiskAssessmentScenarioEntityType),
		RiskAssessmentScopeID: req.RiskAssessmentScopeID,
		Name:                  req.Name,
		Description:           req.Description,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			raScope := coredata.RiskAssessmentScope{}
			if err := raScope.LoadByID(ctx, tx, scope, req.RiskAssessmentScopeID); err != nil {
				return fmt.Errorf("cannot load risk assessment scope: %w", err)
			}
			scenario.OrganizationID = raScope.OrganizationID
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

func (s *Service) GetScenario(ctx context.Context, scope coredata.Scoper, id gid.GID) (*coredata.RiskAssessmentScenario, error) {
	scenario := &coredata.RiskAssessmentScenario{}
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

func (s *Service) UpdateScenario(ctx context.Context, scope coredata.Scoper, req UpdateRiskAssessmentScenarioRequest) (*coredata.RiskAssessmentScenario, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	scenario := &coredata.RiskAssessmentScenario{}
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
			scenario := &coredata.RiskAssessmentScenario{}
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
	cursor *page.Cursor[coredata.RiskAssessmentScenarioOrderField],
) (*page.Page[*coredata.RiskAssessmentScenario, coredata.RiskAssessmentScenarioOrderField], error) {
	var results coredata.RiskAssessmentScenarios
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
			ss := &coredata.RiskAssessmentScenarios{}
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
	cursor *page.Cursor[coredata.RiskAssessmentScenarioOrderField],
) (*page.Page[*coredata.RiskAssessmentScenario, coredata.RiskAssessmentScenarioOrderField], error) {
	var results coredata.RiskAssessmentScenarios
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
			ss := &coredata.RiskAssessmentScenarios{}
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

func (s *Service) ListScenariosForScopeID(
	ctx context.Context,
	scope coredata.Scoper,
	scopeID gid.GID,
	cursor *page.Cursor[coredata.RiskAssessmentScenarioOrderField],
) (*page.Page[*coredata.RiskAssessmentScenario, coredata.RiskAssessmentScenarioOrderField], error) {
	var results coredata.RiskAssessmentScenarios
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := results.LoadByRiskAssessmentScopeID(ctx, conn, scope, scopeID, cursor); err != nil {
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

func (s *Service) CountScenariosForScopeID(ctx context.Context, scope coredata.Scoper, scopeID gid.GID) (int, error) {
	var count int
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			ss := &coredata.RiskAssessmentScenarios{}
			count, err = ss.CountByRiskAssessmentScopeID(ctx, conn, scope, scopeID)
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

func (s *Service) LinkScenarioThreat(ctx context.Context, scope coredata.Scoper, req LinkRiskAssessmentScenarioThreatRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scenario := coredata.RiskAssessmentScenario{}
			if err := scenario.LoadByID(ctx, tx, scope, req.RiskAssessmentScenarioID); err != nil {
				return fmt.Errorf("cannot load risk scenario: %w", err)
			}
			threat := coredata.RiskAssessmentThreat{}
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
			link := &coredata.RiskAssessmentScenarioThreat{
				RiskAssessmentScenarioID: req.RiskAssessmentScenarioID,
				RiskAssessmentThreatID:   req.ThreatID,
				CreatedAt:                time.Now(),
			}
			if err := link.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot link scenario threat: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) UnlinkScenarioThreat(ctx context.Context, scope coredata.Scoper, req UnlinkRiskAssessmentScenarioThreatRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.RiskAssessmentScenarioThreat{
				RiskAssessmentScenarioID: req.RiskAssessmentScenarioID,
				RiskAssessmentThreatID:   req.ThreatID,
			}
			if err := link.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot unlink scenario threat: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) LinkScenarioRisk(ctx context.Context, scope coredata.Scoper, req LinkRiskAssessmentScenarioRiskRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scenario := coredata.RiskAssessmentScenario{}
			if err := scenario.LoadByID(ctx, tx, scope, req.RiskAssessmentScenarioID); err != nil {
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
			link := &coredata.RiskAssessmentScenarioRisk{
				RiskAssessmentScenarioID: req.RiskAssessmentScenarioID,
				RiskID:                   req.RiskID,
				CreatedAt:                time.Now(),
			}
			if err := link.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot link scenario risk: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) UnlinkScenarioRisk(ctx context.Context, scope coredata.Scoper, req UnlinkRiskAssessmentScenarioRiskRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			link := &coredata.RiskAssessmentScenarioRisk{
				RiskAssessmentScenarioID: req.RiskAssessmentScenarioID,
				RiskID:                   req.RiskID,
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
	cursor *page.Cursor[coredata.RiskAssessmentThreatOrderField],
) (*page.Page[*coredata.RiskAssessmentThreat, coredata.RiskAssessmentThreatOrderField], error) {
	var results coredata.RiskAssessmentThreats
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
			ts := &coredata.RiskAssessmentThreats{}
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

func (s *Service) assertNodeInScope(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	nodeID gid.GID,
	scopeID gid.GID,
	field string,
) error {
	node := &coredata.RiskAssessmentNode{}
	if err := node.LoadByID(ctx, tx, scope, nodeID); err != nil {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "node not found",
		}}
	}
	if node.RiskAssessmentScopeID != scopeID {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "node does not belong to this scope",
		}}
	}
	return nil
}

func (s *Service) assertProcessInScope(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	processID gid.GID,
	scopeID gid.GID,
	field string,
) error {
	process := &coredata.RiskAssessmentProcess{}
	if err := process.LoadByID(ctx, tx, scope, processID); err != nil {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "process not found",
		}}
	}
	if process.RiskAssessmentScopeID != scopeID {
		return validator.ValidationErrors{{
			Field:   field,
			Code:    validator.ErrorCodeCustom,
			Message: "process does not belong to this scope",
		}}
	}
	return nil
}
