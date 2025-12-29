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
	RightsRequestOrderBy OrderBy[coredata.RightsRequestOrderField]

	RightsRequestConnection struct {
		TotalCount int
		Edges      []*RightsRequestEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewRightsRequestConnection(
	p *page.Page[*coredata.RightsRequest, coredata.RightsRequestOrderField],
	parentType any,
	parentID gid.GID,
) *RightsRequestConnection {
	edges := make([]*RightsRequestEdge, len(p.Data))
	for i, request := range p.Data {
		edges[i] = NewRightsRequestEdge(request, p.Cursor.OrderBy.Field)
	}

	return &RightsRequestConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewRightsRequest(rr *coredata.RightsRequest) *RightsRequest {
	return &RightsRequest{
		ID:           rr.ID,
		RequestType:  rr.RequestType,
		RequestState: rr.RequestState,
		DataSubject:  rr.DataSubject,
		Contact:      rr.Contact,
		Details:      rr.Details,
		Deadline:     rr.Deadline,
		ActionTaken:  rr.ActionTaken,
		CreatedAt:    rr.CreatedAt,
		UpdatedAt:    rr.UpdatedAt,
	}
}

func NewRightsRequestEdge(rr *coredata.RightsRequest, orderField coredata.RightsRequestOrderField) *RightsRequestEdge {
	return &RightsRequestEdge{
		Node:   NewRightsRequest(rr),
		Cursor: rr.CursorKey(orderField),
	}
}
