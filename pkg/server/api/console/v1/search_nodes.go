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

package console_v1

import (
	"context"
	"fmt"
	"maps"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlx"
	"go.probo.inc/probo/pkg/task"
)

func (r *queryResolver) searchNodes(
	ctx context.Context,
	results []*coredata.SearchResult,
) ([]types.Taggable, error) {
	if searchRequestsIdentityOnly(ctx) {
		nodes, ok := searchIdentityNodes(results)
		if ok {
			return nodes, nil
		}
	}

	return r.loadSearchNodes(ctx, results)
}

func searchRequestsIdentityOnly(ctx context.Context) bool {
	fieldContext := graphql.GetFieldContext(ctx)
	if fieldContext == nil {
		return false
	}

	for _, selection := range fieldContext.Field.Selections {
		selected, ok := selection.(*ast.Field)
		if !ok || selected.Name != "edges" {
			continue
		}

		return edgeSelectionIsIdentity(selected.SelectionSet)
	}

	return true
}

func edgeSelectionIsIdentity(selections ast.SelectionSet) bool {
	for _, selection := range selections {
		switch selected := selection.(type) {
		case *ast.Field:
			if selected.Name == "node" {
				return nodeSelectionIsIdentity(selected.SelectionSet)
			}

			if selected.Name != "cursor" && selected.Name != "title" && selected.Name != "__typename" {
				return false
			}
		default:
			return false
		}
	}

	return true
}

func nodeSelectionIsIdentity(selections ast.SelectionSet) bool {
	for _, selection := range selections {
		selected, ok := selection.(*ast.Field)
		if !ok {
			return false
		}

		if selected.Name != "id" && selected.Name != "__typename" {
			return false
		}
	}

	return true
}

func searchIdentityNodes(results []*coredata.SearchResult) ([]types.Taggable, bool) {
	nodes := make([]types.Taggable, len(results))

	for i, result := range results {
		node, ok := searchIdentityNode(result.ID)
		if !ok {
			return nil, false
		}

		nodes[i] = node
	}

	return nodes, true
}

func searchIdentityNode(id gid.GID) (types.Taggable, bool) {
	switch id.EntityType() {
	case coredata.AiSystemEntityType:
		return &types.AiSystem{ID: id}, true
	case coredata.AssetEntityType:
		return &types.Asset{ID: id}, true
	case coredata.AuditEntityType:
		return &types.Audit{ID: id}, true
	case coredata.BusinessFunctionEntityType:
		return &types.BusinessFunction{ID: id}, true
	case coredata.ControlEntityType:
		return &types.Control{ID: id}, true
	case coredata.DataProtectionImpactAssessmentEntityType:
		return &types.DataProtectionImpactAssessment{ID: id}, true
	case coredata.DatumEntityType:
		return &types.Datum{ID: id}, true
	case coredata.DeviceEntityType:
		return &types.Device{ID: id}, true
	case coredata.DocumentEntityType:
		return &types.Document{ID: id}, true
	case coredata.EvidenceEntityType:
		return &types.Evidence{ID: id}, true
	case coredata.FindingEntityType:
		return &types.Finding{ID: id}, true
	case coredata.FrameworkEntityType:
		return &types.Framework{ID: id}, true
	case coredata.InternalControlEntityType:
		return &types.InternalControl{ID: id}, true
	case coredata.ObligationEntityType:
		return &types.Obligation{ID: id}, true
	case coredata.ProcessingActivityEntityType:
		return &types.ProcessingActivity{ID: id}, true
	case coredata.RightsRequestEntityType:
		return &types.RightsRequest{ID: id}, true
	case coredata.RiskEntityType:
		return &types.Risk{ID: id}, true
	case coredata.RiskAnalysisEntityType:
		return &types.RiskAnalysis{ID: id}, true
	case coredata.RiskAnalysisBoundaryEntityType:
		return &types.RiskAnalysisBoundary{ID: id}, true
	case coredata.RiskAnalysisDiagramEntityType:
		return &types.RiskAnalysisDiagram{ID: id}, true
	case coredata.RiskAnalysisProcessEntityType:
		return &types.RiskAnalysisProcess{ID: id}, true
	case coredata.RiskAnalysisScenarioEntityType:
		return &types.RiskAnalysisScenario{ID: id}, true
	case coredata.RiskAnalysisThreatEntityType:
		return &types.RiskAnalysisThreat{ID: id}, true
	case coredata.StatementOfApplicabilityEntityType:
		return &types.StatementOfApplicability{ID: id}, true
	case coredata.TaskEntityType:
		return &types.Task{ID: id}, true
	case coredata.ThirdPartyEntityType:
		return &types.ThirdParty{ID: id}, true
	case coredata.TransferImpactAssessmentEntityType:
		return &types.TransferImpactAssessment{ID: id}, true
	case coredata.TreatmentPlanEntityType:
		return &types.TreatmentPlan{ID: id}, true
	default:
		return nil, false
	}
}

func (r *queryResolver) loadSearchNodes(
	ctx context.Context,
	results []*coredata.SearchResult,
) ([]types.Taggable, error) {
	grouped := map[uint16][]gid.GID{}

	for _, result := range results {
		entityType := result.ID.EntityType()
		grouped[entityType] = append(grouped[entityType], result.ID)
	}

	loaded := map[gid.GID]types.Taggable{}

	for entityType, ids := range grouped {
		nodes, err := r.loadSearchNodeGroup(ctx, entityType, ids)
		if err != nil {
			return nil, err
		}

		maps.Copy(loaded, nodes)
	}

	nodes := make([]types.Taggable, len(results))
	for i, result := range results {
		node, ok := loaded[result.ID]
		if !ok {
			return nil, gqlx.NotFound(ctx, fmt.Errorf("node %q not found", result.ID))
		}

		nodes[i] = node
	}

	return nodes, nil
}

func (r *queryResolver) loadSearchNodeGroup(
	ctx context.Context,
	entityType uint16,
	ids []gid.GID,
) (map[gid.GID]types.Taggable, error) {
	switch entityType {
	case coredata.DocumentEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionDocumentGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Document, error) {
				return r.probo.Documents.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Document) gid.GID { return item.ID }, func(item *coredata.Document) types.Taggable {
				return types.NewDocument(item)
			})
		})
	case coredata.TaskEntityType:
		return authorizeSearchGroup(ctx, r, task.ActionTaskGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Task, error) {
				return r.task.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Task) gid.GID { return item.ID }, func(item *coredata.Task) types.Taggable {
				return types.NewTask(item)
			})
		})
	case coredata.RiskEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionRiskGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Risk, error) {
				return r.probo.Risks.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Risk) gid.GID { return item.ID }, func(item *coredata.Risk) types.Taggable {
				return types.NewRisk(item)
			})
		})
	case coredata.ThirdPartyEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionThirdPartyGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.ThirdParty, error) {
				return r.probo.ThirdParties.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.ThirdParty) gid.GID { return item.ID }, func(item *coredata.ThirdParty) types.Taggable {
				return types.NewThirdParty(item)
			})
		})
	case coredata.FrameworkEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionFrameworkGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Framework, error) {
				return r.probo.Frameworks.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Framework) gid.GID { return item.ID }, func(item *coredata.Framework) types.Taggable {
				return types.NewFramework(item)
			})
		})
	case coredata.InternalControlEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionInternalControlGet, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.InternalControl, error) {
				return r.probo.InternalControls.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.InternalControl) gid.GID { return item.ID }, func(item *coredata.InternalControl) types.Taggable {
				return types.NewInternalControl(item)
			})
		})
	case coredata.AuditEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionAuditList, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Audit, error) {
				return r.probo.Audits.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Audit) gid.GID { return item.ID }, func(item *coredata.Audit) types.Taggable {
				return types.NewAudit(item)
			})
		})
	case coredata.ControlEntityType:
		return authorizeSearchGroup(ctx, r, probo.ActionControlList, ids, func(ctx context.Context, scope *coredata.Scope) (map[gid.GID]types.Taggable, error) {
			return loadSearchGroup(ctx, ids, func(ctx context.Context, ids ...gid.GID) ([]*coredata.Control, error) {
				return r.probo.Controls.GetByIDs(ctx, scope, ids...)
			}, func(item *coredata.Control) gid.GID { return item.ID }, func(item *coredata.Control) types.Taggable {
				return types.NewControl(item)
			})
		})
	default:
		nodes := make(map[gid.GID]types.Taggable, len(ids))
		for _, id := range ids {
			node, err := r.Node(ctx, id)
			if err != nil {
				return nil, err
			}

			taggable, ok := node.(types.Taggable)
			if !ok || taggable == nil {
				return nil, gqlx.Internal(ctx)
			}

			nodes[id] = taggable
		}

		return nodes, nil
	}
}

func authorizeSearchGroup(
	ctx context.Context,
	r *queryResolver,
	action iam.Action,
	ids []gid.GID,
	load func(context.Context, *coredata.Scope) (map[gid.GID]types.Taggable, error),
) (map[gid.GID]types.Taggable, error) {
	scope, err := r.batchAuthorize(ctx, action, ids)
	if err != nil {
		return nil, err
	}

	return load(ctx, scope)
}

func loadSearchGroup[T any](
	ctx context.Context,
	ids []gid.GID,
	load func(context.Context, ...gid.GID) ([]*T, error),
	idOf func(*T) gid.GID,
	wrap func(*T) types.Taggable,
) (map[gid.GID]types.Taggable, error) {
	items, err := load(ctx, ids...)
	if err != nil {
		return nil, err
	}

	nodes := make(map[gid.GID]types.Taggable, len(items))
	for _, item := range items {
		nodes[idOf(item)] = wrap(item)
	}

	for _, id := range ids {
		if _, ok := nodes[id]; !ok {
			return nil, gqlx.NotFound(ctx, fmt.Errorf("node %q not found", id))
		}
	}

	return nodes, nil
}
