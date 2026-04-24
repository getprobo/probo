// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

type (
	ThirdPartyRiskAssessmentOrderBy OrderBy[coredata.ThirdPartyRiskAssessmentOrderField]
)

func NewThirdPartyRiskAssessmentConnection(p *page.Page[*coredata.ThirdPartyRiskAssessment, coredata.ThirdPartyRiskAssessmentOrderField]) *ThirdPartyRiskAssessmentConnection {
	var edges = make([]*ThirdPartyRiskAssessmentEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewThirdPartyRiskAssessmentEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &ThirdPartyRiskAssessmentConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewThirdPartyRiskAssessmentEdge(c *coredata.ThirdPartyRiskAssessment, orderBy coredata.ThirdPartyRiskAssessmentOrderField) *ThirdPartyRiskAssessmentEdge {
	return &ThirdPartyRiskAssessmentEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewThirdPartyRiskAssessment(c),
	}
}

func NewThirdPartyRiskAssessment(c *coredata.ThirdPartyRiskAssessment) *ThirdPartyRiskAssessment {
	return &ThirdPartyRiskAssessment{
		ID: c.ID,
		ThirdParty: &ThirdParty{
			ID: c.ThirdPartyID,
		},
		ExpiresAt:       c.ExpiresAt,
		DataSensitivity: c.DataSensitivity,
		BusinessImpact:  c.BusinessImpact,
		Notes:           c.Notes,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
	}
}
