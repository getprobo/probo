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
	ProcessingActivityDPIAOrderBy OrderBy[coredata.ProcessingActivityDPIAOrderField]

	ProcessingActivityDPIAConnection struct {
		TotalCount int
		Edges      []*ProcessingActivityDPIAEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.ProcessingActivityDPIAFilter
	}
)

func NewProcessingActivityDPIAConnection(
	p *page.Page[*coredata.ProcessingActivityDPIA, coredata.ProcessingActivityDPIAOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.ProcessingActivityDPIAFilter,
) *ProcessingActivityDPIAConnection {
	edges := make([]*ProcessingActivityDPIAEdge, len(p.Data))
	for i, dpia := range p.Data {
		edges[i] = NewProcessingActivityDPIAEdge(dpia, p.Cursor.OrderBy.Field)
	}

	return &ProcessingActivityDPIAConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewProcessingActivityDPIAEdge(dpia *coredata.ProcessingActivityDPIA, orderField coredata.ProcessingActivityDPIAOrderField) *ProcessingActivityDPIAEdge {
	return &ProcessingActivityDPIAEdge{
		Node:   NewProcessingActivityDpia(dpia),
		Cursor: dpia.CursorKey(orderField),
	}
}

func NewProcessingActivityDpia(dpia *coredata.ProcessingActivityDPIA) *ProcessingActivityDpia {
	return &ProcessingActivityDpia{
		ID:                          dpia.ID,
		Description:                 dpia.Description,
		NecessityAndProportionality: dpia.NecessityAndProportionality,
		PotentialRisk:               dpia.PotentialRisk,
		Mitigations:                 dpia.Mitigations,
		ResidualRisk:                dpia.ResidualRisk,
		CreatedAt:                   dpia.CreatedAt,
		UpdatedAt:                   dpia.UpdatedAt,
	}
}
