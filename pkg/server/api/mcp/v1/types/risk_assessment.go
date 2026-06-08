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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewRiskAssessment(ra *coredata.RiskAssessment) *RiskAssessment {
	return &RiskAssessment{
		ID:             ra.ID,
		OrganizationID: ra.OrganizationID,
		Name:           ra.Name,
		Description:    ra.Description,
		CreatedAt:      ra.CreatedAt,
		UpdatedAt:      ra.UpdatedAt,
	}
}

func NewListRiskAssessmentsOutput(
	p *page.Page[*coredata.RiskAssessment, coredata.RiskAssessmentOrderField],
) ListRiskAssessmentsOutput {
	items := make([]*RiskAssessment, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAssessment(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAssessmentsOutput{
		NextCursor:      nextCursor,
		RiskAssessments: items,
	}
}

func NewRiskAssessmentScope(s *coredata.RiskAssessmentScope) *RiskAssessmentScope {
	return &RiskAssessmentScope{
		ID:               s.ID,
		OrganizationID:   s.OrganizationID,
		RiskAssessmentID: s.RiskAssessmentID,
		Name:             s.Name,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}

func NewListRiskAssessmentScopesOutput(
	p *page.Page[*coredata.RiskAssessmentScope, coredata.RiskAssessmentScopeOrderField],
) ListRiskAssessmentScopesOutput {
	items := make([]*RiskAssessmentScope, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAssessmentScope(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAssessmentScopesOutput{
		NextCursor:           nextCursor,
		RiskAssessmentScopes: items,
	}
}

func NewRiskAssessmentNode(n *coredata.RiskAssessmentNode) *RiskAssessmentNode {
	return &RiskAssessmentNode{
		ID:                    n.ID,
		OrganizationID:        n.OrganizationID,
		RiskAssessmentScopeID: n.RiskAssessmentScopeID,
		BoundaryID:            n.BoundaryID,
		NodeType:              n.NodeType,
		Name:                  n.Name,
		CreatedAt:             n.CreatedAt,
		UpdatedAt:             n.UpdatedAt,
	}
}

func NewRiskAssessmentBoundary(b *coredata.RiskAssessmentBoundary) *RiskAssessmentBoundary {
	return &RiskAssessmentBoundary{
		ID:                    b.ID,
		OrganizationID:        b.OrganizationID,
		RiskAssessmentScopeID: b.RiskAssessmentScopeID,
		ParentBoundaryID:      b.ParentBoundaryID,
		Name:                  b.Name,
		CreatedAt:             b.CreatedAt,
		UpdatedAt:             b.UpdatedAt,
	}
}

func NewListRiskAssessmentBoundariesOutput(
	p *page.Page[*coredata.RiskAssessmentBoundary, coredata.RiskAssessmentBoundaryOrderField],
) ListRiskAssessmentBoundariesOutput {
	items := make([]*RiskAssessmentBoundary, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAssessmentBoundary(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAssessmentBoundariesOutput{
		NextCursor:               nextCursor,
		RiskAssessmentBoundaries: items,
	}
}

func NewListRiskAssessmentNodesOutput(
	p *page.Page[*coredata.RiskAssessmentNode, coredata.RiskAssessmentNodeOrderField],
) ListRiskAssessmentNodesOutput {
	items := make([]*RiskAssessmentNode, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAssessmentNode(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAssessmentNodesOutput{
		NextCursor:          nextCursor,
		RiskAssessmentNodes: items,
	}
}

func NewRiskAssessmentProcess(p *coredata.RiskAssessmentProcess) *RiskAssessmentProcess {
	return &RiskAssessmentProcess{
		ID:                    p.ID,
		OrganizationID:        p.OrganizationID,
		RiskAssessmentScopeID: p.RiskAssessmentScopeID,
		SourceNodeID:          p.SourceNodeID,
		TargetNodeID:          p.TargetNodeID,
		Name:                  p.Name,
		CreatedAt:             p.CreatedAt,
		UpdatedAt:             p.UpdatedAt,
	}
}

func NewListRiskAssessmentProcessesOutput(
	p *page.Page[*coredata.RiskAssessmentProcess, coredata.RiskAssessmentProcessOrderField],
) ListRiskAssessmentProcessesOutput {
	items := make([]*RiskAssessmentProcess, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAssessmentProcess(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAssessmentProcessesOutput{
		NextCursor:              nextCursor,
		RiskAssessmentProcesses: items,
	}
}

func NewRiskAssessmentThreat(t *coredata.RiskAssessmentThreat) *RiskAssessmentThreat {
	return &RiskAssessmentThreat{
		ID:                    t.ID,
		OrganizationID:        t.OrganizationID,
		RiskAssessmentScopeID: t.RiskAssessmentScopeID,
		ProcessID:             t.ProcessID,
		Name:                  t.Name,
		Category:              t.Category,
		CreatedAt:             t.CreatedAt,
		UpdatedAt:             t.UpdatedAt,
	}
}

func NewListRiskAssessmentThreatsOutput(
	p *page.Page[*coredata.RiskAssessmentThreat, coredata.RiskAssessmentThreatOrderField],
) ListRiskAssessmentThreatsOutput {
	items := make([]*RiskAssessmentThreat, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAssessmentThreat(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAssessmentThreatsOutput{
		NextCursor:            nextCursor,
		RiskAssessmentThreats: items,
	}
}

func NewRiskAssessmentScenario(s *coredata.RiskAssessmentScenario) *RiskAssessmentScenario {
	return &RiskAssessmentScenario{
		ID:                    s.ID,
		OrganizationID:        s.OrganizationID,
		RiskAssessmentScopeID: s.RiskAssessmentScopeID,
		Name:                  s.Name,
		Description:           s.Description,
		CreatedAt:             s.CreatedAt,
		UpdatedAt:             s.UpdatedAt,
	}
}

func NewListRiskAssessmentScenariosOutput(
	p *page.Page[*coredata.RiskAssessmentScenario, coredata.RiskAssessmentScenarioOrderField],
) ListRiskAssessmentScenariosOutput {
	items := make([]*RiskAssessmentScenario, 0, len(p.Data))
	for _, v := range p.Data {
		items = append(items, NewRiskAssessmentScenario(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRiskAssessmentScenariosOutput{
		NextCursor:              nextCursor,
		RiskAssessmentScenarios: items,
	}
}
