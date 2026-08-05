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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewRiskAnalysis(ra *coredata.RiskAnalysis) *RiskAnalysis {
	return &RiskAnalysis{
		ID:             ra.ID,
		OrganizationID: ra.OrganizationID,
		Name:           ra.Name,
		Description:    ra.Description,
		CreatedAt:      ra.CreatedAt,
		UpdatedAt:      ra.UpdatedAt,
	}
}

func NewListRiskAnalysesOutput(
	p *page.Page[*coredata.RiskAnalysis, coredata.RiskAnalysisOrderField],
) ListRiskAnalysesOutput {
	items := make([]*RiskAnalysis, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAnalysis(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAnalysesOutput{
		NextCursor:   nextCursor,
		RiskAnalyses: items,
	}
}

func NewRiskAnalysisScope(s *coredata.RiskAnalysisScope) *RiskAnalysisScope {
	return &RiskAnalysisScope{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		RiskAnalysisID: s.RiskAnalysisID,
		Name:           s.Name,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func NewListRiskAnalysisScopesOutput(
	p *page.Page[*coredata.RiskAnalysisScope, coredata.RiskAnalysisScopeOrderField],
) ListRiskAnalysisScopesOutput {
	items := make([]*RiskAnalysisScope, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAnalysisScope(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAnalysisScopesOutput{
		NextCursor:         nextCursor,
		RiskAnalysisScopes: items,
	}
}

func NewRiskAnalysisNode(n *coredata.RiskAnalysisNode) *RiskAnalysisNode {
	return &RiskAnalysisNode{
		ID:                  n.ID,
		OrganizationID:      n.OrganizationID,
		RiskAnalysisScopeID: n.RiskAnalysisScopeID,
		BoundaryID:          n.BoundaryID,
		NodeType:            n.NodeType,
		Name:                n.Name,
		CreatedAt:           n.CreatedAt,
		UpdatedAt:           n.UpdatedAt,
	}
}

func NewRiskAnalysisBoundary(b *coredata.RiskAnalysisBoundary) *RiskAnalysisBoundary {
	return &RiskAnalysisBoundary{
		ID:                  b.ID,
		OrganizationID:      b.OrganizationID,
		RiskAnalysisScopeID: b.RiskAnalysisScopeID,
		ParentBoundaryID:    b.ParentBoundaryID,
		Name:                b.Name,
		CreatedAt:           b.CreatedAt,
		UpdatedAt:           b.UpdatedAt,
	}
}

func NewListRiskAnalysisBoundariesOutput(
	p *page.Page[*coredata.RiskAnalysisBoundary, coredata.RiskAnalysisBoundaryOrderField],
) ListRiskAnalysisBoundariesOutput {
	items := make([]*RiskAnalysisBoundary, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAnalysisBoundary(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAnalysisBoundariesOutput{
		NextCursor:             nextCursor,
		RiskAnalysisBoundaries: items,
	}
}

func NewListRiskAnalysisNodesOutput(
	p *page.Page[*coredata.RiskAnalysisNode, coredata.RiskAnalysisNodeOrderField],
) ListRiskAnalysisNodesOutput {
	items := make([]*RiskAnalysisNode, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAnalysisNode(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAnalysisNodesOutput{
		NextCursor:        nextCursor,
		RiskAnalysisNodes: items,
	}
}

func NewRiskAnalysisProcess(p *coredata.RiskAnalysisProcess) *RiskAnalysisProcess {
	return &RiskAnalysisProcess{
		ID:                  p.ID,
		OrganizationID:      p.OrganizationID,
		RiskAnalysisScopeID: p.RiskAnalysisScopeID,
		SourceNodeID:        p.SourceNodeID,
		TargetNodeID:        p.TargetNodeID,
		Name:                p.Name,
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
}

func NewListRiskAnalysisProcessesOutput(
	p *page.Page[*coredata.RiskAnalysisProcess, coredata.RiskAnalysisProcessOrderField],
) ListRiskAnalysisProcessesOutput {
	items := make([]*RiskAnalysisProcess, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAnalysisProcess(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAnalysisProcessesOutput{
		NextCursor:            nextCursor,
		RiskAnalysisProcesses: items,
	}
}

func NewRiskAnalysisThreat(t *coredata.RiskAnalysisThreat) *RiskAnalysisThreat {
	return &RiskAnalysisThreat{
		ID:                  t.ID,
		OrganizationID:      t.OrganizationID,
		RiskAnalysisScopeID: t.RiskAnalysisScopeID,
		ProcessID:           t.ProcessID,
		Name:                t.Name,
		Category:            t.Category,
		CreatedAt:           t.CreatedAt,
		UpdatedAt:           t.UpdatedAt,
	}
}

func NewListRiskAnalysisThreatsOutput(
	p *page.Page[*coredata.RiskAnalysisThreat, coredata.RiskAnalysisThreatOrderField],
) ListRiskAnalysisThreatsOutput {
	items := make([]*RiskAnalysisThreat, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAnalysisThreat(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAnalysisThreatsOutput{
		NextCursor:          nextCursor,
		RiskAnalysisThreats: items,
	}
}

func NewRiskAnalysisScenario(s *coredata.RiskAnalysisScenario) *RiskAnalysisScenario {
	return &RiskAnalysisScenario{
		ID:                  s.ID,
		OrganizationID:      s.OrganizationID,
		RiskAnalysisScopeID: s.RiskAnalysisScopeID,
		Name:                s.Name,
		Description:         s.Description,
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
	}
}

func NewListRiskAnalysisScenariosOutput(
	p *page.Page[*coredata.RiskAnalysisScenario, coredata.RiskAnalysisScenarioOrderField],
) ListRiskAnalysisScenariosOutput {
	items := make([]*RiskAnalysisScenario, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAnalysisScenario(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAnalysisScenariosOutput{
		NextCursor:            nextCursor,
		RiskAnalysisScenarios: items,
	}
}
