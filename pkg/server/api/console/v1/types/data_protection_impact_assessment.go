// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	DataProtectionImpactAssessmentOrderBy OrderBy[coredata.DataProtectionImpactAssessmentOrderField]

	DataProtectionImpactAssessmentConnection struct {
		TotalCount int
		Edges      []*DataProtectionImpactAssessmentEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.DataProtectionImpactAssessmentFilter
	}
)

func NewDataProtectionImpactAssessmentConnection(
	p *page.Page[*coredata.DataProtectionImpactAssessment, coredata.DataProtectionImpactAssessmentOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.DataProtectionImpactAssessmentFilter,
) *DataProtectionImpactAssessmentConnection {
	edges := make([]*DataProtectionImpactAssessmentEdge, len(p.Data))
	for i, dpia := range p.Data {
		edges[i] = NewDataProtectionImpactAssessmentEdge(dpia, p.Cursor.OrderBy.Field)
	}

	return &DataProtectionImpactAssessmentConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewDataProtectionImpactAssessmentEdge(dpia *coredata.DataProtectionImpactAssessment, orderField coredata.DataProtectionImpactAssessmentOrderField) *DataProtectionImpactAssessmentEdge {
	return &DataProtectionImpactAssessmentEdge{
		Node:   NewDataProtectionImpactAssessment(dpia),
		Cursor: dpia.CursorKey(orderField),
	}
}

func NewDataProtectionImpactAssessment(dpia *coredata.DataProtectionImpactAssessment) *DataProtectionImpactAssessment {
	return &DataProtectionImpactAssessment{
		ID: dpia.ID,
		ProcessingActivity: &ProcessingActivity{
			ID: dpia.ProcessingActivityID,
		},
		Organization: &Organization{
			ID: dpia.OrganizationID,
		},
		Description:                 dpia.Description,
		NecessityAndProportionality: dpia.NecessityAndProportionality,
		PotentialRisk:               dpia.PotentialRisk,
		Mitigations:                 dpia.Mitigations,
		ResidualRisk:                dpia.ResidualRisk,
		CreatedAt:                   dpia.CreatedAt,
		UpdatedAt:                   dpia.UpdatedAt,
	}
}
