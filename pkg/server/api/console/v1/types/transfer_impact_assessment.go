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
	TransferImpactAssessmentOrderBy OrderBy[coredata.TransferImpactAssessmentOrderField]

	TransferImpactAssessmentConnection struct {
		TotalCount int
		Edges      []*TransferImpactAssessmentEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewTransferImpactAssessmentConnection(
	p *page.Page[*coredata.TransferImpactAssessment, coredata.TransferImpactAssessmentOrderField],
	parentType any,
	parentID gid.GID,
) *TransferImpactAssessmentConnection {
	edges := make([]*TransferImpactAssessmentEdge, len(p.Data))
	for i, tia := range p.Data {
		edges[i] = NewTransferImpactAssessmentEdge(tia, p.Cursor.OrderBy.Field)
	}

	return &TransferImpactAssessmentConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTransferImpactAssessmentEdge(tia *coredata.TransferImpactAssessment, orderField coredata.TransferImpactAssessmentOrderField) *TransferImpactAssessmentEdge {
	return &TransferImpactAssessmentEdge{
		Node:   NewTransferImpactAssessment(tia),
		Cursor: tia.CursorKey(orderField),
	}
}

func NewTransferImpactAssessment(tia *coredata.TransferImpactAssessment) *TransferImpactAssessment {
	return &TransferImpactAssessment{
		ID:                    tia.ID,
		DataSubjects:          tia.DataSubjects,
		LegalMechanism:        tia.LegalMechanism,
		Transfer:              tia.Transfer,
		LocalLawRisk:          tia.LocalLawRisk,
		SupplementaryMeasures: tia.SupplementaryMeasures,
		CreatedAt:             tia.CreatedAt,
		UpdatedAt:             tia.UpdatedAt,
		ProcessingActivity: &ProcessingActivity{
			ID: tia.ProcessingActivityID,
		},
		Organization: &Organization{
			ID: tia.OrganizationID,
		},
	}
}
