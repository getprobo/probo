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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	RiskAssessmentScopeOrderBy OrderBy[coredata.RiskAssessmentScopeOrderField]

	RiskAssessmentScopeConnection struct {
		TotalCount int
		Edges      []*RiskAssessmentScopeConnectionEdge
		PageInfo   PageInfo
		Resolver   any
		ParentID   gid.GID
	}
)

func NewRiskAssessmentScopeConnection(
	p *page.Page[*coredata.RiskAssessmentScope, coredata.RiskAssessmentScopeOrderField],
	parentType any,
	parentID gid.GID,
) *RiskAssessmentScopeConnection {
	edges := make([]*RiskAssessmentScopeConnectionEdge, len(p.Data))
	for i := range edges {
		edges[i] = &RiskAssessmentScopeConnectionEdge{
			Cursor: p.Data[i].CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewRiskAssessmentScope(p.Data[i]),
		}
	}

	return &RiskAssessmentScopeConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewRiskAssessmentScopeConnectionEdge(s *coredata.RiskAssessmentScope, orderBy coredata.RiskAssessmentScopeOrderField) *RiskAssessmentScopeConnectionEdge {
	return &RiskAssessmentScopeConnectionEdge{
		Cursor: s.CursorKey(orderBy),
		Node:   NewRiskAssessmentScope(s),
	}
}

func NewRiskAssessmentScope(s *coredata.RiskAssessmentScope) *RiskAssessmentScope {
	return &RiskAssessmentScope{
		ID:               s.ID,
		RiskAssessmentID: s.RiskAssessmentID,
		Name:             s.Name,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
