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
		ID:                     we.ID,
		WebhookSubscriptionID: we.WebhookSubscriptionID,
		Status:                 we.Status,
		Response:               response,
		CreatedAt:              we.CreatedAt,
	}
}
