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
	ProcessingActivityTIAOrderBy OrderBy[coredata.ProcessingActivityTIAOrderField]

	ProcessingActivityTIAConnection struct {
		TotalCount int
		Edges      []*ProcessingActivityTIAEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewProcessingActivityTIAConnection(
	p *page.Page[*coredata.ProcessingActivityTIA, coredata.ProcessingActivityTIAOrderField],
	parentType any,
	parentID gid.GID,
) *ProcessingActivityTIAConnection {
	edges := make([]*ProcessingActivityTIAEdge, len(p.Data))
	for i, tia := range p.Data {
		edges[i] = NewProcessingActivityTIAEdge(tia, p.Cursor.OrderBy.Field)
	}

	return &ProcessingActivityTIAConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewProcessingActivityTIAEdge(tia *coredata.ProcessingActivityTIA, orderField coredata.ProcessingActivityTIAOrderField) *ProcessingActivityTIAEdge {
	return &ProcessingActivityTIAEdge{
		Node:   NewProcessingActivityTia(tia),
		Cursor: tia.CursorKey(orderField),
	}
}

func NewProcessingActivityTia(tia *coredata.ProcessingActivityTIA) *ProcessingActivityTia {
	return &ProcessingActivityTia{
		ID:                    tia.ID,
		DataSubjects:          tia.DataSubjects,
		LegalMechanism:        tia.LegalMechanism,
		Transfer:              tia.Transfer,
		LocalLawRisk:          tia.LocalLawRisk,
		SupplementaryMeasures: tia.SupplementaryMeasures,
		CreatedAt:             tia.CreatedAt,
		UpdatedAt:             tia.UpdatedAt,
	}
}
