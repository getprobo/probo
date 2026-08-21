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
)

func (s *Service) Fork(
	ctx context.Context,
	scope coredata.Scoper,
	req ForkRiskAnalysisRequest,
) (*coredata.RiskAnalysis, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	forked := &coredata.RiskAnalysis{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			source := &coredata.RiskAnalysis{}
			if err := source.LoadByID(ctx, tx, scope, req.RiskAnalysisID); err != nil {
				return fmt.Errorf("cannot load risk analysis: %w", err)
			}

			*forked = coredata.RiskAnalysis{
				ID:             gid.New(scope.GetTenantID(), coredata.RiskAnalysisEntityType),
				OrganizationID: source.OrganizationID,
				Name:           req.Name,
				Description:    req.Description,
				MatrixRows:     source.MatrixRows,
				MatrixCols:     source.MatrixCols,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if req.Period != nil {
				forked.PeriodStart = req.Period.Start
				forked.PeriodEnd = req.Period.End
			}

			if err := forked.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert forked risk analysis: %w", err)
			}

			idMap := make(map[gid.GID]gid.GID)
			if err := copyDiagrams(ctx, tx, scope, source.ID, forked, now, idMap); err != nil {
				return fmt.Errorf("cannot copy diagrams: %w", err)
			}

			if err := copyTreatmentPlans(ctx, tx, scope, source.ID, forked, now, idMap); err != nil {
				return fmt.Errorf("cannot copy treatment plans: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return forked, nil
}

func copyDiagrams(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	sourceAnalysisID gid.GID,
	forked *coredata.RiskAnalysis,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) error {
	diagrams, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.RiskAnalysisDiagramOrderField]{
			Field:     coredata.RiskAnalysisDiagramOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisDiagramOrderField]) ([]*coredata.RiskAnalysisDiagram, error) {
			var batch coredata.RiskAnalysisDiagrams
			if err := batch.LoadByRiskAnalysisID(ctx, tx, scope, sourceAnalysisID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load diagrams: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load all diagrams: %w", err)
	}

	var sourceScenarioIDs []gid.GID
	for _, diagram := range diagrams {
		scenarioIDs, err := copyDiagram(ctx, tx, scope, diagram, forked, now, idMap)
		if err != nil {
			return fmt.Errorf("cannot copy diagram: %w", err)
		}

		sourceScenarioIDs = append(sourceScenarioIDs, scenarioIDs...)
	}

	// Scenario-threat links can span diagrams, so remap them only after
	// every diagram's threats are in idMap.
	if err := copyScenarioThreats(ctx, tx, scope, sourceScenarioIDs, now, idMap); err != nil {
		return fmt.Errorf("cannot copy scenario threats: %w", err)
	}

	return nil
}

func copyDiagram(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	source *coredata.RiskAnalysisDiagram,
	forked *coredata.RiskAnalysis,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) ([]gid.GID, error) {
	copied := &coredata.RiskAnalysisDiagram{
		ID:             gid.New(scope.GetTenantID(), coredata.RiskAnalysisDiagramEntityType),
		OrganizationID: forked.OrganizationID,
		RiskAnalysisID: forked.ID,
		Name:           source.Name,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	idMap[source.ID] = copied.ID

	if err := copied.Insert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot insert diagram: %w", err)
	}

	if err := copyBoundaries(ctx, tx, scope, source.ID, copied, now, idMap); err != nil {
		return nil, fmt.Errorf("cannot copy boundaries: %w", err)
	}

	if err := copyNodes(ctx, tx, scope, source.ID, copied, now, idMap); err != nil {
		return nil, fmt.Errorf("cannot copy nodes: %w", err)
	}

	if err := copyProcesses(ctx, tx, scope, source.ID, copied, now, idMap); err != nil {
		return nil, fmt.Errorf("cannot copy processes: %w", err)
	}

	if err := copyThreats(ctx, tx, scope, source.ID, copied, now, idMap); err != nil {
		return nil, fmt.Errorf("cannot copy threats: %w", err)
	}

	scenarioIDs, err := copyScenarios(ctx, tx, scope, source.ID, copied, now, idMap)
	if err != nil {
		return nil, fmt.Errorf("cannot copy scenarios: %w", err)
	}

	return scenarioIDs, nil
}

func copyBoundaries(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	sourceDiagramID gid.GID,
	copiedDiagram *coredata.RiskAnalysisDiagram,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) error {
	boundaries, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.RiskAnalysisBoundaryOrderField]{
			Field:     coredata.RiskAnalysisBoundaryOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisBoundaryOrderField]) ([]*coredata.RiskAnalysisBoundary, error) {
			var batch coredata.RiskAnalysisBoundaries
			if err := batch.LoadByRiskAnalysisDiagramID(ctx, tx, scope, sourceDiagramID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load boundaries: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load all boundaries: %w", err)
	}

	ordered, err := orderBoundaries(boundaries)
	if err != nil {
		return fmt.Errorf("cannot order boundaries: %w", err)
	}

	for _, source := range ordered {
		copied := &coredata.RiskAnalysisBoundary{
			ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisBoundaryEntityType),
			OrganizationID:        copiedDiagram.OrganizationID,
			RiskAnalysisDiagramID: copiedDiagram.ID,
			Name:                  source.Name,
			CreatedAt:             now,
			UpdatedAt:             now,
		}

		parentID, err := remapOptionalID(idMap, source.ParentBoundaryID)
		if err != nil {
			return fmt.Errorf("cannot remap parent boundary: %w", err)
		}

		copied.ParentBoundaryID = parentID
		idMap[source.ID] = copied.ID

		if err := copied.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert boundary: %w", err)
		}
	}

	return nil
}

func copyNodes(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	sourceDiagramID gid.GID,
	copiedDiagram *coredata.RiskAnalysisDiagram,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) error {
	nodes, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.RiskAnalysisNodeOrderField]{
			Field:     coredata.RiskAnalysisNodeOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisNodeOrderField]) ([]*coredata.RiskAnalysisNode, error) {
			var batch coredata.RiskAnalysisNodes
			if err := batch.LoadByRiskAnalysisDiagramID(ctx, tx, scope, sourceDiagramID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load nodes: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load all nodes: %w", err)
	}

	for _, source := range nodes {
		copied := &coredata.RiskAnalysisNode{
			ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisNodeEntityType),
			OrganizationID:        copiedDiagram.OrganizationID,
			RiskAnalysisDiagramID: copiedDiagram.ID,
			NodeType:              source.NodeType,
			Name:                  source.Name,
			CreatedAt:             now,
			UpdatedAt:             now,
		}

		boundaryID, err := remapOptionalID(idMap, source.BoundaryID)
		if err != nil {
			return fmt.Errorf("cannot remap node boundary: %w", err)
		}

		copied.BoundaryID = boundaryID
		idMap[source.ID] = copied.ID

		if err := copied.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert node: %w", err)
		}
	}

	return nil
}

func copyProcesses(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	sourceDiagramID gid.GID,
	copiedDiagram *coredata.RiskAnalysisDiagram,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) error {
	processes, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.RiskAnalysisProcessOrderField]{
			Field:     coredata.RiskAnalysisProcessOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisProcessOrderField]) ([]*coredata.RiskAnalysisProcess, error) {
			var batch coredata.RiskAnalysisProcesses
			if err := batch.LoadByRiskAnalysisDiagramID(ctx, tx, scope, sourceDiagramID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load processes: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load all processes: %w", err)
	}

	for _, source := range processes {
		sourceNodeID, err := remapID(idMap, source.SourceNodeID)
		if err != nil {
			return fmt.Errorf("cannot remap process source node: %w", err)
		}

		targetNodeID, err := remapID(idMap, source.TargetNodeID)
		if err != nil {
			return fmt.Errorf("cannot remap process target node: %w", err)
		}

		copied := &coredata.RiskAnalysisProcess{
			ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisProcessEntityType),
			OrganizationID:        copiedDiagram.OrganizationID,
			RiskAnalysisDiagramID: copiedDiagram.ID,
			SourceNodeID:          sourceNodeID,
			TargetNodeID:          targetNodeID,
			Name:                  source.Name,
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		idMap[source.ID] = copied.ID

		if err := copied.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert process: %w", err)
		}
	}

	return nil
}

func copyThreats(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	sourceDiagramID gid.GID,
	copiedDiagram *coredata.RiskAnalysisDiagram,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) error {
	threats, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.RiskAnalysisThreatOrderField]{
			Field:     coredata.RiskAnalysisThreatOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisThreatOrderField]) ([]*coredata.RiskAnalysisThreat, error) {
			var batch coredata.RiskAnalysisThreats
			if err := batch.LoadByRiskAnalysisDiagramID(ctx, tx, scope, sourceDiagramID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load threats: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load all threats: %w", err)
	}

	for _, source := range threats {
		processID, err := remapID(idMap, source.ProcessID)
		if err != nil {
			return fmt.Errorf("cannot remap threat process: %w", err)
		}

		copied := &coredata.RiskAnalysisThreat{
			ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisThreatEntityType),
			OrganizationID:        copiedDiagram.OrganizationID,
			RiskAnalysisDiagramID: copiedDiagram.ID,
			ProcessID:             processID,
			Name:                  source.Name,
			Category:              source.Category,
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		idMap[source.ID] = copied.ID

		if err := copied.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert threat: %w", err)
		}
	}

	return nil
}

func copyScenarios(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	sourceDiagramID gid.GID,
	copiedDiagram *coredata.RiskAnalysisDiagram,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) ([]gid.GID, error) {
	scenarios, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.RiskAnalysisScenarioOrderField]{
			Field:     coredata.RiskAnalysisScenarioOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.RiskAnalysisScenarioOrderField]) ([]*coredata.RiskAnalysisScenario, error) {
			var batch coredata.RiskAnalysisScenarios
			if err := batch.LoadByRiskAnalysisDiagramID(ctx, tx, scope, sourceDiagramID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load scenarios: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load all scenarios: %w", err)
	}

	sourceScenarioIDs := make([]gid.GID, 0, len(scenarios))
	for _, source := range scenarios {
		copied := &coredata.RiskAnalysisScenario{
			ID:                    gid.New(scope.GetTenantID(), coredata.RiskAnalysisScenarioEntityType),
			OrganizationID:        copiedDiagram.OrganizationID,
			RiskAnalysisDiagramID: copiedDiagram.ID,
			Name:                  source.Name,
			Description:           source.Description,
			CreatedAt:             now,
			UpdatedAt:             now,
		}
		idMap[source.ID] = copied.ID
		sourceScenarioIDs = append(sourceScenarioIDs, source.ID)

		if err := copied.Insert(ctx, tx, scope); err != nil {
			return nil, fmt.Errorf("cannot insert scenario: %w", err)
		}
	}

	if err := copyScenarioRisks(ctx, tx, scope, sourceScenarioIDs, now, idMap); err != nil {
		return nil, fmt.Errorf("cannot copy scenario risks: %w", err)
	}

	return sourceScenarioIDs, nil
}

func copyScenarioThreats(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	sourceScenarioIDs []gid.GID,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) error {
	var links coredata.RiskAnalysisScenarioThreats
	if err := links.LoadByScenarioIDs(ctx, tx, scope, sourceScenarioIDs); err != nil {
		return fmt.Errorf("cannot load scenario threats: %w", err)
	}

	for _, link := range links {
		scenarioID, err := remapID(idMap, link.RiskAnalysisScenarioID)
		if err != nil {
			return fmt.Errorf("cannot remap scenario threat scenario: %w", err)
		}

		threatID, err := remapID(idMap, link.RiskAnalysisThreatID)
		if err != nil {
			return fmt.Errorf("cannot remap scenario threat: %w", err)
		}

		copied := &coredata.RiskAnalysisScenarioThreat{
			RiskAnalysisScenarioID: scenarioID,
			RiskAnalysisThreatID:   threatID,
			CreatedAt:              now,
		}
		if err := copied.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert scenario threat: %w", err)
		}
	}

	return nil
}

func copyScenarioRisks(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	sourceScenarioIDs []gid.GID,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) error {
	var links coredata.RiskAnalysisScenarioRisks
	if err := links.LoadByScenarioIDs(ctx, tx, scope, sourceScenarioIDs); err != nil {
		return fmt.Errorf("cannot load scenario risks: %w", err)
	}

	for _, link := range links {
		scenarioID, err := remapID(idMap, link.RiskAnalysisScenarioID)
		if err != nil {
			return fmt.Errorf("cannot remap scenario risk scenario: %w", err)
		}

		copied := &coredata.RiskAnalysisScenarioRisk{
			RiskAnalysisScenarioID: scenarioID,
			RiskID:                 link.RiskID,
			CreatedAt:              now,
		}
		if err := copied.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert scenario risk: %w", err)
		}
	}

	return nil
}

func copyTreatmentPlans(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	sourceAnalysisID gid.GID,
	forked *coredata.RiskAnalysis,
	now time.Time,
	idMap map[gid.GID]gid.GID,
) error {
	plans, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.TreatmentPlanOrderField]{
			Field:     coredata.TreatmentPlanOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.TreatmentPlanOrderField]) ([]*coredata.TreatmentPlan, error) {
			var batch coredata.TreatmentPlans
			if err := batch.LoadByRiskAnalysisID(ctx, tx, scope, sourceAnalysisID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load treatment plans: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load all treatment plans: %w", err)
	}

	sourcePlanIDs := make([]gid.GID, 0, len(plans))
	for _, source := range plans {
		copied := &coredata.TreatmentPlan{
			ID:                 gid.New(scope.GetTenantID(), coredata.TreatmentPlanEntityType),
			OrganizationID:     forked.OrganizationID,
			RiskID:             source.RiskID,
			RiskAnalysisID:     forked.ID,
			Treatment:          source.Treatment,
			OwnerID:            source.OwnerID,
			InherentLikelihood: source.InherentLikelihood,
			InherentImpact:     source.InherentImpact,
			ResidualLikelihood: source.ResidualLikelihood,
			ResidualImpact:     source.ResidualImpact,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
		idMap[source.ID] = copied.ID
		sourcePlanIDs = append(sourcePlanIDs, source.ID)

		if err := copied.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert treatment plan: %w", err)
		}
	}

	var mappings coredata.TreatmentPlanMeasures
	if err := mappings.LoadByTreatmentPlanIDs(ctx, tx, scope, sourcePlanIDs); err != nil {
		return fmt.Errorf("cannot load treatment plan measures: %w", err)
	}

	for _, mapping := range mappings {
		planID, err := remapID(idMap, mapping.TreatmentPlanID)
		if err != nil {
			return fmt.Errorf("cannot remap treatment plan measure: %w", err)
		}

		copied := coredata.TreatmentPlanMeasure{
			TreatmentPlanID: planID,
			MeasureID:       mapping.MeasureID,
			OrganizationID:  forked.OrganizationID,
			CreatedAt:       now,
		}
		if err := copied.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert treatment plan measure: %w", err)
		}
	}

	return nil
}

func orderBoundaries(boundaries []*coredata.RiskAnalysisBoundary) ([]*coredata.RiskAnalysisBoundary, error) {
	remaining := make(map[gid.GID]*coredata.RiskAnalysisBoundary, len(boundaries))
	for _, boundary := range boundaries {
		remaining[boundary.ID] = boundary
	}

	ordered := make([]*coredata.RiskAnalysisBoundary, 0, len(boundaries))
	emitted := make(map[gid.GID]struct{}, len(boundaries))

	for len(remaining) > 0 {
		progress := false

		for id, boundary := range remaining {
			if boundary.ParentBoundaryID != nil {
				if _, ok := emitted[*boundary.ParentBoundaryID]; !ok {
					continue
				}
			}

			ordered = append(ordered, boundary)
			emitted[id] = struct{}{}
			delete(remaining, id)
			progress = true
		}

		if !progress {
			return nil, fmt.Errorf("cannot order boundaries: cycle detected")
		}
	}

	return ordered, nil
}

func remapID(idMap map[gid.GID]gid.GID, old gid.GID) (gid.GID, error) {
	mapped, ok := idMap[old]
	if !ok {
		return gid.GID{}, fmt.Errorf("cannot remap id %s: mapping not found", old)
	}

	return mapped, nil
}

func remapOptionalID(idMap map[gid.GID]gid.GID, old *gid.GID) (*gid.GID, error) {
	if old == nil {
		return nil, nil
	}

	mapped, err := remapID(idMap, *old)
	if err != nil {
		return nil, err
	}

	return &mapped, nil
}
