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
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/policy"
)

var organizationCondition = policy.Equals("principal.organization_id", "resource.organization_id")

var readActions = []string{
	ActionRiskAnalysisGet, ActionRiskAnalysisList,
	ActionRiskAnalysisDiagramGet, ActionRiskAnalysisDiagramList,
	ActionRiskAnalysisNodeGet, ActionRiskAnalysisNodeList,
	ActionRiskAnalysisBoundaryGet, ActionRiskAnalysisBoundaryList,
	ActionRiskAnalysisProcessGet, ActionRiskAnalysisProcessList,
	ActionRiskAnalysisThreatGet, ActionRiskAnalysisThreatList,
	ActionRiskAnalysisScenarioGet, ActionRiskAnalysisScenarioList,
	ActionTreatmentPlanGet, ActionTreatmentPlanList,
}

var writeActions = []string{
	ActionRiskAnalysisCreate, ActionRiskAnalysisUpdate, ActionRiskAnalysisDelete,
	ActionRiskAnalysisDiagramCreate, ActionRiskAnalysisDiagramUpdate, ActionRiskAnalysisDiagramDelete,
	ActionRiskAnalysisNodeCreate, ActionRiskAnalysisNodeUpdate, ActionRiskAnalysisNodeDelete,
	ActionRiskAnalysisBoundaryCreate, ActionRiskAnalysisBoundaryUpdate, ActionRiskAnalysisBoundaryDelete,
	ActionRiskAnalysisProcessCreate, ActionRiskAnalysisProcessUpdate, ActionRiskAnalysisProcessDelete,
	ActionRiskAnalysisThreatCreate, ActionRiskAnalysisThreatUpdate, ActionRiskAnalysisThreatDelete,
	ActionRiskAnalysisScenarioCreate, ActionRiskAnalysisScenarioUpdate, ActionRiskAnalysisScenarioDelete,
	ActionRiskAnalysisScenarioThreatLink, ActionRiskAnalysisScenarioThreatUnlink,
	ActionRiskAnalysisScenarioRiskLink, ActionRiskAnalysisScenarioRiskUnlink,
	ActionTreatmentPlanCreate, ActionTreatmentPlanUpdate, ActionTreatmentPlanDelete,
}

// FullAccessPolicy grants complete risk-management access to organization
// owners and admins, including catalog risk actions that share this prefix.
var FullAccessPolicy = policy.NewPolicy(
	"risk-management:full-access",
	"Risk Management Full Access",
	policy.Allow("risk-management:*").
		WithSID("risk-management-full-access").
		When(organizationCondition),
).WithDescription("Full risk-management access")

// ReadAccessPolicy grants read-only risk-analysis and treatment-plan access to
// viewers and auditors.
var ReadAccessPolicy = policy.NewPolicy(
	"risk-management:read-access",
	"Risk Management Read Access",
	policy.Allow(readActions...).
		WithSID("risk-management-read-access").
		When(organizationCondition),
).WithDescription("Read-only risk-analysis and treatment-plan access")

// PolicySet returns the PolicySet for the risk-management service. It is owned
// by this package and registered into the authorizer at composition time so
// risk-analysis and treatment-plan authorization rules live alongside the
// domain logic instead of in the core probo policy set.
func PolicySet() *iam.PolicySet {
	return iam.NewPolicySet().
		AddRolePolicy("OWNER", FullAccessPolicy).
		AddRolePolicy("ADMIN", FullAccessPolicy).
		AddRolePolicy("VIEWER", ReadAccessPolicy).
		AddRolePolicy("AUDITOR", ReadAccessPolicy)
}
