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
	"go.probo.inc/probo/pkg/page"
)

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

func NewRightsRequestEdge(
	rr *coredata.RightsRequest,
	orderBy coredata.RightsRequestOrderField,
) *RightsRequestEdge {
	return &RightsRequestEdge{
		Cursor: rr.CursorKey(orderBy),
		Node:   NewRightsRequest(rr),
	}
}

func NewRightsRequestConnection(
	p *page.Page[*coredata.RightsRequest, coredata.RightsRequestOrderField],
) *RightsRequestConnection {
	edges := make([]*RightsRequestEdge, len(p.Data))

	for i, item := range p.Data {
		edges[i] = NewRightsRequestEdge(item, p.Cursor.OrderBy.Field)
	}

	return &RightsRequestConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}
