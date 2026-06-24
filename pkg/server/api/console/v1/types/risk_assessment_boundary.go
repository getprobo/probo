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
	RiskAssessmentBoundaryOrderBy OrderBy[coredata.RiskAssessmentBoundaryOrderField]

	RiskAssessmentBoundaryConnection struct {
		TotalCount int
		Edges      []*RiskAssessmentBoundaryConnectionEdge
		PageInfo   PageInfo
		Resolver   any
		ParentID   gid.GID
	}
)

func NewRiskAssessmentBoundaryConnection(
	p *page.Page[*coredata.RiskAssessmentBoundary, coredata.RiskAssessmentBoundaryOrderField],
	parentType any,
	parentID gid.GID,
) *RiskAssessmentBoundaryConnection {
	edges := make([]*RiskAssessmentBoundaryConnectionEdge, len(p.Data))
	for i := range edges {
		edges[i] = &RiskAssessmentBoundaryConnectionEdge{
			Cursor: p.Data[i].CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewRiskAssessmentBoundary(p.Data[i]),
		}
	}

	return &RiskAssessmentBoundaryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewRiskAssessmentBoundary(b *coredata.RiskAssessmentBoundary) *RiskAssessmentBoundary {
	return &RiskAssessmentBoundary{
		ID:                    b.ID,
		RiskAssessmentScopeID: b.RiskAssessmentScopeID,
		ParentBoundaryID:      b.ParentBoundaryID,
		Name:                  b.Name,
		CreatedAt:             b.CreatedAt,
		UpdatedAt:             b.UpdatedAt,
	}
}
