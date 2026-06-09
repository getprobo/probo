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
	RiskAssessmentProcessOrderBy OrderBy[coredata.RiskAssessmentProcessOrderField]

	RiskAssessmentProcessConnection struct {
		TotalCount int
		Edges      []*RiskAssessmentProcessConnectionEdge
		PageInfo   PageInfo
		Resolver   any
		ParentID   gid.GID
	}
)

func NewRiskAssessmentProcessConnection(
	p *page.Page[*coredata.RiskAssessmentProcess, coredata.RiskAssessmentProcessOrderField],
	parentType any,
	parentID gid.GID,
) *RiskAssessmentProcessConnection {
	edges := make([]*RiskAssessmentProcessConnectionEdge, len(p.Data))
	for i := range edges {
		edges[i] = &RiskAssessmentProcessConnectionEdge{
			Cursor: p.Data[i].CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewRiskAssessmentProcess(p.Data[i]),
		}
	}

	return &RiskAssessmentProcessConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewRiskAssessmentProcess(pr *coredata.RiskAssessmentProcess) *RiskAssessmentProcess {
	return &RiskAssessmentProcess{
		ID:                    pr.ID,
		RiskAssessmentScopeID: pr.RiskAssessmentScopeID,
		SourceNodeID:          pr.SourceNodeID,
		TargetNodeID:          pr.TargetNodeID,
		Name:                  pr.Name,
		CreatedAt:             pr.CreatedAt,
		UpdatedAt:             pr.UpdatedAt,
	}
}
