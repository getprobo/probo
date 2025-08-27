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
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
)

type (
	ProcessingActivityRegistryOrderBy OrderBy[coredata.ProcessingActivityRegistryOrderField]

	ProcessingActivityRegistryConnection struct {
		TotalCount int
		Edges      []*ProcessingActivityRegistryEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewProcessingActivityRegistryConnection(
	p *page.Page[*coredata.ProcessingActivityRegistry, coredata.ProcessingActivityRegistryOrderField],
	parentType any,
	parentID gid.GID,
) *ProcessingActivityRegistryConnection {
	edges := make([]*ProcessingActivityRegistryEdge, len(p.Data))
	for i, registry := range p.Data {
		edges[i] = NewProcessingActivityRegistryEdge(registry, p.Cursor.OrderBy.Field)
	}

	return &ProcessingActivityRegistryConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewProcessingActivityRegistry(par *coredata.ProcessingActivityRegistry) *ProcessingActivityRegistry {
	return &ProcessingActivityRegistry{
		ID:                             par.ID,
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
		TransferSafeguards:             par.TransferSafeguards,
		RetentionPeriod:                par.RetentionPeriod,
		SecurityMeasures:               par.SecurityMeasures,
		DataProtectionImpactAssessment: par.DataProtectionImpactAssessment,
		TransferImpactAssessment:       par.TransferImpactAssessment,
		CreatedAt:                      par.CreatedAt,
		UpdatedAt:                      par.UpdatedAt,
	}
}

func NewProcessingActivityRegistryEdge(par *coredata.ProcessingActivityRegistry, orderField coredata.ProcessingActivityRegistryOrderField) *ProcessingActivityRegistryEdge {
	return &ProcessingActivityRegistryEdge{
		Node:   NewProcessingActivityRegistry(par),
		Cursor: par.CursorKey(orderField),
	}
}
