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
	ProcessingActivityOrderBy OrderBy[coredata.ProcessingActivityOrderField]

	ProcessingActivityConnection struct {
		TotalCount int
		Edges      []*ProcessingActivityEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewProcessingActivityConnection(
	p *page.Page[*coredata.ProcessingActivity, coredata.ProcessingActivityOrderField],
	parentType any,
	parentID gid.GID,
) *ProcessingActivityConnection {
	edges := make([]*ProcessingActivityEdge, len(p.Data))
	for i, processingActivity := range p.Data {
		edges[i] = NewProcessingActivityEdge(processingActivity, p.Cursor.OrderBy.Field)
	}

	return &ProcessingActivityConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewProcessingActivityEdge(par *coredata.ProcessingActivity, orderField coredata.ProcessingActivityOrderField) *ProcessingActivityEdge {
	return &ProcessingActivityEdge{
		Node:   NewProcessingActivity(par),
		Cursor: par.CursorKey(orderField),
	}
}

func NewProcessingActivity(par *coredata.ProcessingActivity) *ProcessingActivity {
	object := &ProcessingActivity{
		ID: par.ID,
		Organization: &Organization{
			ID: par.OrganizationID,
		},
		Name:                                 par.Name,
		Purpose:                              par.Purpose,
		DataSubjectCategory:                  par.DataSubjectCategory,
		PersonalDataCategory:                 par.PersonalDataCategory,
		SpecialOrCriminalData:                par.SpecialOrCriminalData,
		ConsentEvidenceLink:                  par.ConsentEvidenceLink,
		LawfulBasis:                          par.LawfulBasis,
		Recipients:                           par.Recipients,
		Location:                             par.Location,
		InternationalTransfers:               par.InternationalTransfers,
		TransferSafeguards:                   par.TransferSafeguard,
		RetentionPeriod:                      par.RetentionPeriod,
		SecurityMeasures:                     par.SecurityMeasures,
		DataProtectionImpactAssessmentNeeded: par.DataProtectionImpactAssessmentNeeded,
		TransferImpactAssessmentNeeded:       par.TransferImpactAssessmentNeeded,
		LastReviewDate:                       par.LastReviewDate,
		NextReviewDate:                       par.NextReviewDate,
		Role:                                 par.Role,
		CreatedAt:                            par.CreatedAt,
		UpdatedAt:                            par.UpdatedAt,
	}

	if par.DataProtectionOfficerID != nil {
		object.DataProtectionOfficer = &Profile{
			ID: *par.DataProtectionOfficerID,
		}
	}

	return object
}
