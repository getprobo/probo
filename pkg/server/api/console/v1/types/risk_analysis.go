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
	RiskAnalysisOrderBy OrderBy[coredata.RiskAnalysisOrderField]

	RiskAnalysisConnection struct {
		TotalCount int
		Edges      []*RiskAnalysisConnectionEdge
		PageInfo   PageInfo
		Resolver   any
		ParentID   gid.GID
	}
)

func NewRiskAnalysisConnection(
	p *page.Page[*coredata.RiskAnalysis, coredata.RiskAnalysisOrderField],
	parentType any,
	parentID gid.GID,
) *RiskAnalysisConnection {
	edges := make([]*RiskAnalysisConnectionEdge, len(p.Data))
	for i := range edges {
		edges[i] = &RiskAnalysisConnectionEdge{
			Cursor: p.Data[i].CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewRiskAnalysis(p.Data[i]),
		}
	}

	return &RiskAnalysisConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewRiskAnalysisConnectionEdge(ra *coredata.RiskAnalysis, orderBy coredata.RiskAnalysisOrderField) *RiskAnalysisConnectionEdge {
	return &RiskAnalysisConnectionEdge{
		Cursor: ra.CursorKey(orderBy),
		Node:   NewRiskAnalysis(ra),
	}
}

func NewRiskAnalysis(ra *coredata.RiskAnalysis) *RiskAnalysis {
	return &RiskAnalysis{
		ID:          ra.ID,
		Name:        ra.Name,
		Description: ra.Description,
		Period:      NewPeriod(ra.PeriodStart, ra.PeriodEnd),
		MatrixSize:  NewMatrixSize(ra.MatrixRows, ra.MatrixCols),
		Organization: &Organization{
			ID: ra.OrganizationID,
		},
		CreatedAt: ra.CreatedAt,
		UpdatedAt: ra.UpdatedAt,
	}
}
