// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	SCIMEventOrderBy OrderBy[coredata.SCIMEventOrderField]

	SCIMEventConnection struct {
		TotalCount int
		Edges      []*SCIMEventEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewSCIMEventConnection(
	p *page.Page[*coredata.SCIMEvent, coredata.SCIMEventOrderField],
	resolver any,
	parentID gid.GID,
) *SCIMEventConnection {
	edges := make([]*SCIMEventEdge, len(p.Data))
	for i, scimEvent := range p.Data {
		edges[i] = NewSCIMEventEdge(scimEvent, p.Cursor.OrderBy.Field)
	}

	return &SCIMEventConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewSCIMEventEdge(scimEvent *coredata.SCIMEvent, orderField coredata.SCIMEventOrderField) *SCIMEventEdge {
	return &SCIMEventEdge{
		Node:   NewSCIMEvent(scimEvent),
		Cursor: scimEvent.CursorKey(orderField),
	}
}

func NewSCIMEvent(scimEvent *coredata.SCIMEvent) *SCIMEvent {
	event := &SCIMEvent{
		ID:           scimEvent.ID,
		Method:       scimEvent.Method,
		Path:         scimEvent.Path,
		UserName:     scimEvent.UserName,
		StatusCode:   scimEvent.StatusCode,
		RequestBody:  scimEvent.RequestBody,
		ResponseBody: scimEvent.ResponseBody,
		ErrorMessage: scimEvent.ErrorMessage,
		IPAddress:    scimEvent.IPAddress.String(),
		CreatedAt:    scimEvent.CreatedAt,
	}

	return event
}
