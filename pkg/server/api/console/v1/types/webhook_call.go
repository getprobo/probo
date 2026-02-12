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
	WebhookCallOrderBy OrderBy[coredata.WebhookCallOrderField]

	WebhookCallConnection struct {
		TotalCount int
		Edges      []*WebhookCallEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewWebhookCallConnection(
	p *page.Page[*coredata.WebhookCall, coredata.WebhookCallOrderField],
	parentType any,
	parentID gid.GID,
) *WebhookCallConnection {
	var edges = make([]*WebhookCallEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewWebhookCallEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &WebhookCallConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewWebhookCallEdge(wc *coredata.WebhookCall, orderBy coredata.WebhookCallOrderField) *WebhookCallEdge {
	return &WebhookCallEdge{
		Cursor: wc.CursorKey(orderBy),
		Node:   NewWebhookCall(wc),
	}
}

func NewWebhookCall(wc *coredata.WebhookCall) *WebhookCall {
	var response *string
	if len(wc.Response) > 0 {
		s := string(wc.Response)
		response = &s
	}

	return &WebhookCall{
		ID:                     wc.ID,
		WebhookEventID:        wc.WebhookEventID,
		WebhookConfigurationID: wc.WebhookConfigurationID,
		EndpointURL:            wc.EndpointURL,
		Status:                 wc.Status,
		Response:               response,
		CreatedAt:              wc.CreatedAt,
	}
}
