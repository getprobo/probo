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

package tag

import (
	"context"
	"errors"
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/itam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/riskmanagement"
	"go.probo.inc/probo/pkg/task"
)

// searchableActions is the permission Query.node uses to load each indexed type.
var searchableActions = map[uint16]iam.Action{
	coredata.AiSystemEntityType:                       probo.ActionAiSystemList,
	coredata.AssetEntityType:                          probo.ActionAssetList,
	coredata.AuditEntityType:                          probo.ActionAuditList,
	coredata.BusinessFunctionEntityType:               probo.ActionBusinessFunctionList,
	coredata.ControlEntityType:                        probo.ActionControlList,
	coredata.DataProtectionImpactAssessmentEntityType: probo.ActionDataProtectionImpactAssessmentGet,
	coredata.DatumEntityType:                          probo.ActionDatumList,
	coredata.DeviceEntityType:                         itam.ActionDeviceGet,
	coredata.DocumentEntityType:                       probo.ActionDocumentGet,
	coredata.EvidenceEntityType:                       probo.ActionEvidenceList,
	coredata.FindingEntityType:                        probo.ActionFindingList,
	coredata.FrameworkEntityType:                      probo.ActionFrameworkGet,
	coredata.InternalControlEntityType:                probo.ActionInternalControlGet,
	coredata.ObligationEntityType:                     probo.ActionObligationList,
	coredata.ProcessingActivityEntityType:             probo.ActionProcessingActivityList,
	coredata.RightsRequestEntityType:                  probo.ActionRightsRequestGet,
	coredata.RiskEntityType:                           probo.ActionRiskGet,
	coredata.RiskAnalysisEntityType:                   riskmanagement.ActionRiskAnalysisGet,
	coredata.StatementOfApplicabilityEntityType:       probo.ActionStatementOfApplicabilityGet,
	coredata.TaskEntityType:                           task.ActionTaskGet,
	coredata.ThirdPartyEntityType:                     probo.ActionThirdPartyGet,
	coredata.TransferImpactAssessmentEntityType:       probo.ActionTransferImpactAssessmentGet,
}

// ReadAction is the permission Query.node uses to load a searchable resource.
func ReadAction(entityType uint16) (iam.Action, bool) {
	action, ok := searchableActions[entityType]
	return action, ok
}

// ListableEntityTypes returns the indexed entity types the principal may list
// in organizationID. Each read action is checked with AuthorizeMulti against
// the organization. A denied action drops that type. Assumption failure fails
// the call.
func ListableEntityTypes(
	ctx context.Context,
	authorizer *iam.Authorizer,
	principal gid.GID,
	session *gid.GID,
	organizationID gid.GID,
) ([]uint16, error) {
	actions := make([]iam.Action, 0, len(searchableActions))
	seen := make(map[iam.Action]struct{}, len(searchableActions))
	items := make([]iam.MultiAuthorizeItem, 0, len(searchableActions))

	for _, action := range searchableActions {
		if _, ok := seen[action]; ok {
			continue
		}

		seen[action] = struct{}{}
		actions = append(actions, action)
		items = append(items, iam.MultiAuthorizeItem{
			Resource: organizationID,
			Action:   action,
			DryRun:   true,
		})
	}

	_, decisions, err := authorizer.AuthorizeMulti(
		ctx,
		iam.AuthorizeMultiParams{
			Principal: principal,
			Session:   session,
			Items:     items,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve listable entity types: %w", err)
	}

	allowed := make(map[iam.Action]struct{}, len(actions))
	for i, decision := range decisions {
		if decision == nil {
			allowed[actions[i]] = struct{}{}

			continue
		}

		if _, ok := errors.AsType[*iam.ErrAssumptionRequired](decision); ok {
			return nil, decision
		}

		if _, ok := errors.AsType[*iam.ErrInsufficientPermissions](decision); ok {
			continue
		}

		if _, ok := errors.AsType[*iam.ErrInsufficientOAuth2Scope](decision); ok {
			continue
		}

		return nil, fmt.Errorf("cannot resolve listable entity types: %w", decision)
	}

	entityTypes := make([]uint16, 0, len(searchableActions))
	for entityType, action := range searchableActions {
		if _, ok := allowed[action]; !ok {
			continue
		}

		entityTypes = append(entityTypes, entityType)
	}

	return entityTypes, nil
}
