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
	RiskAssessmentThreatOrderBy OrderBy[coredata.RiskAssessmentThreatOrderField]

	RiskAssessmentThreatConnection struct {
		TotalCount int
		Edges      []*RiskAssessmentThreatConnectionEdge
		PageInfo   PageInfo
		Resolver   any
		ParentID   gid.GID
	}
)

func NewRiskAssessmentThreatConnection(
	p *page.Page[*coredata.RiskAssessmentThreat, coredata.RiskAssessmentThreatOrderField],
	parentType any,
	parentID gid.GID,
) *RiskAssessmentThreatConnection {
	edges := make([]*RiskAssessmentThreatConnectionEdge, len(p.Data))
	for i := range edges {
		edges[i] = &RiskAssessmentThreatConnectionEdge{
			Cursor: p.Data[i].CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewRiskAssessmentThreat(p.Data[i]),
		}
	}
	return &RiskAssessmentThreatConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewRiskAssessmentThreat(t *coredata.RiskAssessmentThreat) *RiskAssessmentThreat {
	return &RiskAssessmentThreat{
		ID:                    t.ID,
		RiskAssessmentScopeID: t.RiskAssessmentScopeID,
		ProcessID:             t.ProcessID,
		Name:                  t.Name,
		Category:              t.Category,
		CreatedAt:             t.CreatedAt,
		UpdatedAt:             t.UpdatedAt,
	}
}
