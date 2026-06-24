// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	RiskOrderBy OrderBy[coredata.RiskOrderField]

	RiskConnection struct {
		TotalCount int
		Edges      []*RiskEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.RiskFilter
	}
)

func NewRiskConnection(
	p *page.Page[*coredata.Risk, coredata.RiskOrderField],
	parentType any,
	parentID gid.GID,
	filters *coredata.RiskFilter,
) *RiskConnection {
	var edges = make([]*RiskEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewRiskEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &RiskConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filters:  filters,
	}
}

func NewRiskEdge(r *coredata.Risk, orderBy coredata.RiskOrderField) *RiskEdge {
	return &RiskEdge{
		Cursor: r.CursorKey(orderBy),
		Node:   NewRisk(r),
	}
}

func NewRisk(r *coredata.Risk) *Risk {
	risk := &Risk{
		ID:                 r.ID,
		Name:               r.Name,
		Description:        r.Description,
		Treatment:          r.Treatment,
		InherentLikelihood: r.InherentLikelihood,
		InherentImpact:     r.InherentImpact,
		InherentRiskScore:  r.InherentRiskScore,
		ResidualLikelihood: r.ResidualLikelihood,
		ResidualImpact:     r.ResidualImpact,
		ResidualRiskScore:  r.ResidualRiskScore,
		Organization: &Organization{
			ID: r.OrganizationID,
		},
		Category:  r.Category,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
		Note:      r.Note,
	}

	if r.OwnerID != nil {
		risk.Owner = &Profile{
			ID: *r.OwnerID,
		}
	}

	return risk
}
