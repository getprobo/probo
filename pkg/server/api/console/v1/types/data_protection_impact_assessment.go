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
	DataProtectionImpactAssessmentOrderBy OrderBy[coredata.DataProtectionImpactAssessmentOrderField]

	DataProtectionImpactAssessmentConnection struct {
		TotalCount int
		Edges      []*DataProtectionImpactAssessmentEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewDataProtectionImpactAssessmentConnection(
	p *page.Page[*coredata.DataProtectionImpactAssessment, coredata.DataProtectionImpactAssessmentOrderField],
	parentType any,
	parentID gid.GID,
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
