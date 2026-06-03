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
	RiskAssessmentOrderBy OrderBy[coredata.RiskAssessmentOrderField]

	RiskAssessmentConnection struct {
		TotalCount int
		Edges      []*RiskAssessmentConnectionEdge
		PageInfo   PageInfo
		Resolver   any
		ParentID   gid.GID
	}
)

func NewRiskAssessmentConnection(
	p *page.Page[*coredata.RiskAssessment, coredata.RiskAssessmentOrderField],
	parentType any,
	parentID gid.GID,
) *RiskAssessmentConnection {
	edges := make([]*RiskAssessmentConnectionEdge, len(p.Data))
	for i := range edges {
		edges[i] = &RiskAssessmentConnectionEdge{
			Cursor: p.Data[i].CursorKey(p.Cursor.OrderBy.Field),
			Node:   NewRiskAssessment(p.Data[i]),
		}
	}

	return &RiskAssessmentConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewRiskAssessmentConnectionEdge(ra *coredata.RiskAssessment, orderBy coredata.RiskAssessmentOrderField) *RiskAssessmentConnectionEdge {
	return &RiskAssessmentConnectionEdge{
		Cursor: ra.CursorKey(orderBy),
		Node:   NewRiskAssessment(ra),
	}
}

func NewRiskAssessment(ra *coredata.RiskAssessment) *RiskAssessment {
	return &RiskAssessment{
		ID:          ra.ID,
		Name:        ra.Name,
		Description: ra.Description,
		Organization: &Organization{
			ID: ra.OrganizationID,
		},
		CreatedAt: ra.CreatedAt,
		UpdatedAt: ra.UpdatedAt,
	}
}
