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
