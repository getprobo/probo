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
	TransferImpactAssessmentOrderBy OrderBy[coredata.TransferImpactAssessmentOrderField]

	TransferImpactAssessmentConnection struct {
		TotalCount int
		Edges      []*TransferImpactAssessmentEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.TransferImpactAssessmentFilter
	}
)

func NewTransferImpactAssessmentConnection(
	p *page.Page[*coredata.TransferImpactAssessment, coredata.TransferImpactAssessmentOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.TransferImpactAssessmentFilter,
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
		Filter:   filter,
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
