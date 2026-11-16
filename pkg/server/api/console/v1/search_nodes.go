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
	"maps"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
	"go.probo.inc/probo/pkg/server/gqlx"
)

func (r *queryResolver) searchNodes(
	ctx context.Context,
	scope *coredata.Scope,
	results []*coredata.SearchResult,
) ([]types.Taggable, error) {
	grouped := map[uint16][]gid.GID{}

	for _, result := range results {
		entityType := result.ID.EntityType()
		grouped[entityType] = append(grouped[entityType], result.ID)
	}

	loaded := map[gid.GID]types.Taggable{}

	for entityType, ids := range grouped {
		nodes, err := r.loadSearchNodeGroup(ctx, scope, entityType, ids)
		if err != nil {
			return nil, err
		}

		maps.Copy(loaded, nodes)
	}

	nodes := make([]types.Taggable, len(results))
	for i, result := range results {
		nodes[i] = loaded[result.ID]
	}

	return nodes, nil
}

func (r *queryResolver) loadSearchNodeGroup(
	ctx context.Context,
	scope *coredata.Scope,
	entityType uint16,
	ids []gid.GID,
) (map[gid.GID]types.Taggable, error) {
	switch entityType {
	case coredata.AiSystemEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.AiSystems.GetByIDs,
			types.NewAiSystem,
		)
	case coredata.RegistryAssetEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.RegistryAssets.GetByIDs,
			types.NewRegistryAsset,
		)
	case coredata.AuditEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.Audits.GetByIDs,
			types.NewAudit,
		)
	case coredata.BusinessFunctionEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.BusinessFunctions.GetByIDs,
			types.NewBusinessFunction,
		)
	case coredata.ControlEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.Controls.GetByIDs,
			types.NewControl,
		)
	case coredata.DataProtectionImpactAssessmentEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.DataProtectionImpactAssessments.GetByIDs,
			types.NewDataProtectionImpactAssessment,
		)
	case coredata.DatumEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.Data.GetByIDs,
			types.NewDatum,
		)
	case coredata.DeviceEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.itam.GetDevicesByIDs,
			types.NewDevice,
		)
	case coredata.DocumentEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.Documents.GetByIDs,
			types.NewDocument,
		)
	case coredata.EvidenceEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.Evidences.GetByIDs,
			types.NewEvidence,
		)
	case coredata.FindingEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.Findings.GetByIDs,
			types.NewFinding,
		)
	case coredata.FrameworkEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.Frameworks.GetByIDs,
			types.NewFramework,
		)
	case coredata.InternalControlEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.InternalControls.GetByIDs,
			types.NewInternalControl,
		)
	case coredata.ObligationEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.Obligations.GetByIDs,
			types.NewObligation,
		)
	case coredata.ProcessingActivityEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.ProcessingActivities.GetByIDs,
			types.NewProcessingActivity,
		)
	case coredata.RightsRequestEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.RightsRequests.GetByIDs,
			types.NewRightsRequest,
		)
	case coredata.RiskEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.Risks.GetByIDs,
			types.NewRisk,
		)
	case coredata.RiskAnalysisEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.riskManagement.GetByIDs,
			types.NewRiskAnalysis,
		)
	case coredata.StatementOfApplicabilityEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.StatementsOfApplicability.GetByIDs,
			types.NewStatementOfApplicability,
		)
	case coredata.TaskEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.task.GetByIDs,
			types.NewTask,
		)
	case coredata.ThirdPartyEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.ThirdParties.GetByIDs,
			types.NewThirdParty,
		)
	case coredata.TransferImpactAssessmentEntityType:
		return loadSearchGroup(
			ctx,
			r.logger,
			scope,
			ids,
			r.probo.TransferImpactAssessments.GetByIDs,
			types.NewTransferImpactAssessment,
		)
	default:
		return nil, gqlx.Internal(ctx)
	}
}

func loadSearchGroup[T any, S ~[]*T, N types.Taggable](
	ctx context.Context,
	logger *log.Logger,
	scope coredata.Scoper,
	ids []gid.GID,
	load func(context.Context, coredata.Scoper, ...gid.GID) (S, error),
	wrap func(*T) N,
) (map[gid.GID]types.Taggable, error) {
	items, err := load(ctx, scope, ids...)
	if err != nil {
		return nil, gqlx.MapError(ctx, logger, "cannot load search nodes", err)
	}

	nodes := make(map[gid.GID]types.Taggable, len(items))
	for _, item := range items {
		node := wrap(item)
		nodes[node.GetID()] = node
	}

	for _, id := range ids {
		if _, ok := nodes[id]; !ok {
			return nil, gqlx.NotFoundf(ctx, "resource not found")
		}
	}

	return nodes, nil
}
