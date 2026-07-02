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
	WebhookEventOrderBy OrderBy[coredata.WebhookEventOrderField]

	WebhookEventConnection struct {
		TotalCount int
		Edges      []*WebhookEventEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewWebhookEventConnection(
	p *page.Page[*coredata.WebhookEvent, coredata.WebhookEventOrderField],
	parentType any,
	parentID gid.GID,
) *WebhookEventConnection {
	var edges = make([]*WebhookEventEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewWebhookEventEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &WebhookEventConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewWebhookEventEdge(we *coredata.WebhookEvent, orderBy coredata.WebhookEventOrderField) *WebhookEventEdge {
	return &WebhookEventEdge{
		Cursor: we.CursorKey(orderBy),
		Node:   NewWebhookEvent(we),
	}
}

func NewWebhookEvent(we *coredata.WebhookEvent) *WebhookEvent {
	var response *string

	if len(we.Response) > 0 {
		s := string(we.Response)
		response = &s
	}

	return &WebhookEvent{
		ID:                    we.ID,
		WebhookSubscriptionID: we.WebhookSubscriptionID,
		Status:                we.Status,
		Response:              response,
		CreatedAt:             we.CreatedAt,
	}
}
