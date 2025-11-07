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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type ProcessingActivity struct {
	ID                             gid.GID                                                   `json:"id"`
	OrganizationID                 gid.GID                                                   `json:"-"`
	SnapshotID                     *gid.GID                                                  `json:"snapshotId,omitempty"`
	SourceID                       *gid.GID                                                  `json:"sourceId,omitempty"`
	Organization                   *Organization                                             `json:"organization"`
	Name                           string                                                    `json:"name"`
	Purpose                        *string                                                   `json:"purpose,omitempty"`
	DataSubjectCategory            *string                                                   `json:"dataSubjectCategory,omitempty"`
	PersonalDataCategory           *string                                                   `json:"personalDataCategory,omitempty"`
	SpecialOrCriminalData          coredata.ProcessingActivitySpecialOrCriminalDatum         `json:"specialOrCriminalData"`
	ConsentEvidenceLink            *string                                                   `json:"consentEvidenceLink,omitempty"`
	LawfulBasis                    coredata.ProcessingActivityLawfulBasis                    `json:"lawfulBasis"`
	Recipients                     *string                                                   `json:"recipients,omitempty"`
	Location                       *string                                                   `json:"location,omitempty"`
	InternationalTransfers         bool                                                      `json:"internationalTransfers"`
	TransferSafeguards             *coredata.ProcessingActivityTransferSafeguard             `json:"transferSafeguards,omitempty"`
	RetentionPeriod                *string                                                   `json:"retentionPeriod,omitempty"`
	SecurityMeasures               *string                                                   `json:"securityMeasures,omitempty"`
	DataProtectionImpactAssessment coredata.ProcessingActivityDataProtectionImpactAssessment `json:"dataProtectionImpactAssessment"`
	TransferImpactAssessment       coredata.ProcessingActivityTransferImpactAssessment       `json:"transferImpactAssessment"`
	Vendors                        *VendorConnection                                         `json:"vendors"`
	CreatedAt                      time.Time                                                 `json:"createdAt"`
	UpdatedAt                      time.Time                                                 `json:"updatedAt"`
}

func (ProcessingActivity) IsNode()             {}
func (this ProcessingActivity) GetID() gid.GID { return this.ID }

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

func NewProcessingActivity(par *coredata.ProcessingActivity) *ProcessingActivity {
	return &ProcessingActivity{
		ID:                             par.ID,
		OrganizationID:                 par.OrganizationID,
		SnapshotID:                     par.SnapshotID,
		SourceID:                       par.SourceID,
		Name:                           par.Name,
		Purpose:                        par.Purpose,
		DataSubjectCategory:            par.DataSubjectCategory,
		PersonalDataCategory:           par.PersonalDataCategory,
		SpecialOrCriminalData:          par.SpecialOrCriminalData,
		ConsentEvidenceLink:            par.ConsentEvidenceLink,
		LawfulBasis:                    par.LawfulBasis,
		Recipients:                     par.Recipients,
		Location:                       par.Location,
		InternationalTransfers:         par.InternationalTransfers,
		TransferSafeguards:             par.TransferSafeguard,
		RetentionPeriod:                par.RetentionPeriod,
		SecurityMeasures:               par.SecurityMeasures,
		DataProtectionImpactAssessment: par.DataProtectionImpactAssessment,
		TransferImpactAssessment:       par.TransferImpactAssessment,
		CreatedAt:                      par.CreatedAt,
		UpdatedAt:                      par.UpdatedAt,
	}
}

func NewProcessingActivityEdge(par *coredata.ProcessingActivity, orderField coredata.ProcessingActivityOrderField) *ProcessingActivityEdge {
	return &ProcessingActivityEdge{
		Node:   NewProcessingActivity(par),
		Cursor: par.CursorKey(orderField),
	}
}
