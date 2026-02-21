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
	ProcessingActivityOrderBy OrderBy[coredata.ProcessingActivityOrderField]

	ProcessingActivityConnection struct {
		TotalCount int
		Edges      []*ProcessingActivityEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *ProcessingActivityFilter
	}
)

func NewProcessingActivityConnection(
	p *page.Page[*coredata.ProcessingActivity, coredata.ProcessingActivityOrderField],
	parentType any,
	parentID gid.GID,
	filter *ProcessingActivityFilter,
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
		Filter:   filter,
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
		SnapshotID:                           par.SnapshotID,
		SourceID:                             par.SourceID,
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
