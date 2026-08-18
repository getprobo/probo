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
// Format: core:<entity>:<action>
const (
	// RiskAnalysis actions
	ActionRiskAnalysisGet    = "core:risk-analysis:get"
	ActionRiskAnalysisList   = "core:risk-analysis:list"
	ActionRiskAnalysisCreate = "core:risk-analysis:create"
	ActionRiskAnalysisUpdate = "core:risk-analysis:update"
	ActionRiskAnalysisDelete = "core:risk-analysis:delete"

	// RiskAnalysisDiagram actions
	ActionRiskAnalysisDiagramGet    = "core:risk-analysis-diagram:get"
	ActionRiskAnalysisDiagramList   = "core:risk-analysis-diagram:list"
	ActionRiskAnalysisDiagramCreate = "core:risk-analysis-diagram:create"
	ActionRiskAnalysisDiagramUpdate = "core:risk-analysis-diagram:update"
	ActionRiskAnalysisDiagramDelete = "core:risk-analysis-diagram:delete"

	// RiskAnalysisNode actions
	ActionRiskAnalysisNodeGet    = "core:risk-analysis-node:get"
	ActionRiskAnalysisNodeList   = "core:risk-analysis-node:list"
	ActionRiskAnalysisNodeCreate = "core:risk-analysis-node:create"
	ActionRiskAnalysisNodeUpdate = "core:risk-analysis-node:update"
	ActionRiskAnalysisNodeDelete = "core:risk-analysis-node:delete"

	// RiskAnalysisBoundary actions
	ActionRiskAnalysisBoundaryGet    = "core:risk-analysis-boundary:get"
	ActionRiskAnalysisBoundaryList   = "core:risk-analysis-boundary:list"
	ActionRiskAnalysisBoundaryCreate = "core:risk-analysis-boundary:create"
	ActionRiskAnalysisBoundaryUpdate = "core:risk-analysis-boundary:update"
	ActionRiskAnalysisBoundaryDelete = "core:risk-analysis-boundary:delete"

	// RiskAnalysisProcess actions
	ActionRiskAnalysisProcessGet    = "core:risk-analysis-process:get"
	ActionRiskAnalysisProcessList   = "core:risk-analysis-process:list"
	ActionRiskAnalysisProcessCreate = "core:risk-analysis-process:create"
	ActionRiskAnalysisProcessUpdate = "core:risk-analysis-process:update"
	ActionRiskAnalysisProcessDelete = "core:risk-analysis-process:delete"

	// RiskAnalysisThreat actions
	ActionRiskAnalysisThreatGet    = "core:risk-analysis-threat:get"
	ActionRiskAnalysisThreatList   = "core:risk-analysis-threat:list"
	ActionRiskAnalysisThreatCreate = "core:risk-analysis-threat:create"
	ActionRiskAnalysisThreatUpdate = "core:risk-analysis-threat:update"
	ActionRiskAnalysisThreatDelete = "core:risk-analysis-threat:delete"

	// RiskAnalysisScenario actions
	ActionRiskAnalysisScenarioGet    = "core:risk-analysis-scenario:get"
	ActionRiskAnalysisScenarioList   = "core:risk-analysis-scenario:list"
	ActionRiskAnalysisScenarioCreate = "core:risk-analysis-scenario:create"
	ActionRiskAnalysisScenarioUpdate = "core:risk-analysis-scenario:update"
	ActionRiskAnalysisScenarioDelete = "core:risk-analysis-scenario:delete"

	// RiskAnalysisScenarioThreat actions
	ActionRiskAnalysisScenarioThreatLink   = "core:risk-analysis-scenario-threat:create"
	ActionRiskAnalysisScenarioThreatUnlink = "core:risk-analysis-scenario-threat:delete"

	// RiskAnalysisScenarioRisk actions
	ActionRiskAnalysisScenarioRiskLink   = "core:risk-analysis-scenario-risk:create"
	ActionRiskAnalysisScenarioRiskUnlink = "core:risk-analysis-scenario-risk:delete"

	// TreatmentPlan actions
	ActionTreatmentPlanGet    = "core:treatment-plan:get"
	ActionTreatmentPlanList   = "core:treatment-plan:list"
	ActionTreatmentPlanCreate = "core:treatment-plan:create"
	ActionTreatmentPlanUpdate = "core:treatment-plan:update"
	ActionTreatmentPlanDelete = "core:treatment-plan:delete"
)
