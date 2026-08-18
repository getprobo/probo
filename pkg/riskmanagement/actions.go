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

// Risk-management service actions.
// Format: risk-management:<entity>:<action>
const (
	// RiskAnalysis actions
	ActionRiskAnalysisGet    = "risk-management:risk-analysis:get"
	ActionRiskAnalysisList   = "risk-management:risk-analysis:list"
	ActionRiskAnalysisCreate = "risk-management:risk-analysis:create"
	ActionRiskAnalysisUpdate = "risk-management:risk-analysis:update"
	ActionRiskAnalysisDelete = "risk-management:risk-analysis:delete"

	// RiskAnalysisDiagram actions
	ActionRiskAnalysisDiagramGet    = "risk-management:diagram:get"
	ActionRiskAnalysisDiagramList   = "risk-management:diagram:list"
	ActionRiskAnalysisDiagramCreate = "risk-management:diagram:create"
	ActionRiskAnalysisDiagramUpdate = "risk-management:diagram:update"
	ActionRiskAnalysisDiagramDelete = "risk-management:diagram:delete"

	// RiskAnalysisNode actions
	ActionRiskAnalysisNodeGet    = "risk-management:node:get"
	ActionRiskAnalysisNodeList   = "risk-management:node:list"
	ActionRiskAnalysisNodeCreate = "risk-management:node:create"
	ActionRiskAnalysisNodeUpdate = "risk-management:node:update"
	ActionRiskAnalysisNodeDelete = "risk-management:node:delete"

	// RiskAnalysisBoundary actions
	ActionRiskAnalysisBoundaryGet    = "risk-management:boundary:get"
	ActionRiskAnalysisBoundaryList   = "risk-management:boundary:list"
	ActionRiskAnalysisBoundaryCreate = "risk-management:boundary:create"
	ActionRiskAnalysisBoundaryUpdate = "risk-management:boundary:update"
	ActionRiskAnalysisBoundaryDelete = "risk-management:boundary:delete"

	// RiskAnalysisProcess actions
	ActionRiskAnalysisProcessGet    = "risk-management:process:get"
	ActionRiskAnalysisProcessList   = "risk-management:process:list"
	ActionRiskAnalysisProcessCreate = "risk-management:process:create"
	ActionRiskAnalysisProcessUpdate = "risk-management:process:update"
	ActionRiskAnalysisProcessDelete = "risk-management:process:delete"

	// RiskAnalysisThreat actions
	ActionRiskAnalysisThreatGet    = "risk-management:threat:get"
	ActionRiskAnalysisThreatList   = "risk-management:threat:list"
	ActionRiskAnalysisThreatCreate = "risk-management:threat:create"
	ActionRiskAnalysisThreatUpdate = "risk-management:threat:update"
	ActionRiskAnalysisThreatDelete = "risk-management:threat:delete"

	// RiskAnalysisScenario actions
	ActionRiskAnalysisScenarioGet    = "risk-management:scenario:get"
	ActionRiskAnalysisScenarioList   = "risk-management:scenario:list"
	ActionRiskAnalysisScenarioCreate = "risk-management:scenario:create"
	ActionRiskAnalysisScenarioUpdate = "risk-management:scenario:update"
	ActionRiskAnalysisScenarioDelete = "risk-management:scenario:delete"

	// RiskAnalysisScenarioThreat actions
	ActionRiskAnalysisScenarioThreatLink   = "risk-management:scenario-threat:create"
	ActionRiskAnalysisScenarioThreatUnlink = "risk-management:scenario-threat:delete"

	// RiskAnalysisScenarioRisk actions
	ActionRiskAnalysisScenarioRiskLink   = "risk-management:scenario-risk:create"
	ActionRiskAnalysisScenarioRiskUnlink = "risk-management:scenario-risk:delete"

	// TreatmentPlan actions
	ActionTreatmentPlanGet    = "risk-management:treatment-plan:get"
	ActionTreatmentPlanList   = "risk-management:treatment-plan:list"
	ActionTreatmentPlanCreate = "risk-management:treatment-plan:create"
	ActionTreatmentPlanUpdate = "risk-management:treatment-plan:update"
	ActionTreatmentPlanDelete = "risk-management:treatment-plan:delete"
)
