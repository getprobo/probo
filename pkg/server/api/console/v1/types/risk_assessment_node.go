// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	RiskAssessmentNodeOrderBy OrderBy[coredata.RiskAssessmentNodeOrderField]

	RiskAssessmentNodeConnection struct {
		TotalCount int
		Edges      []*RiskAssessmentNodeConnectionEdge
		PageInfo   PageInfo
		Resolver   any
		ParentID   gid.GID
	}
)

func NewRiskAssessmentNodeConnection(
	p *page.Page[*coredata.RiskAssessmentNode, coredata.RiskAssessmentNodeOrderField],
	parentType any,
	parentID gid.GID,
) *RiskAssessmentNodeConnection {
	edges := make([]*RiskAssessmentNodeConnectionEdge, len(p.Data))
	for i := range edges {
		edges[i] = &RiskAssessmentNodeConnectionEdge{
			Cursor: p.Data[i].CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewRiskAssessmentNode(p.Data[i]),
		}
	}

	return &RiskAssessmentNodeConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewRiskAssessmentNode(n *coredata.RiskAssessmentNode) *RiskAssessmentNode {
	return &RiskAssessmentNode{
		ID:                    n.ID,
		RiskAssessmentScopeID: n.RiskAssessmentScopeID,
		BoundaryID:            n.BoundaryID,
		NodeType:              n.NodeType,
		Name:                  n.Name,
		CreatedAt:             n.CreatedAt,
		UpdatedAt:             n.UpdatedAt,
	}
}
