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
	WebhookSubscriptionOrderBy OrderBy[coredata.WebhookSubscriptionOrderField]

	WebhookSubscriptionConnection struct {
		TotalCount int
		Edges      []*WebhookSubscriptionEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewWebhookSubscriptionConnection(
	p *page.Page[*coredata.WebhookSubscription, coredata.WebhookSubscriptionOrderField],
	parentType any,
	parentID gid.GID,
) *WebhookSubscriptionConnection {
	var edges = make([]*WebhookSubscriptionEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewWebhookSubscriptionEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &WebhookSubscriptionConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewWebhookSubscriptionEdge(wc *coredata.WebhookSubscription, orderBy coredata.WebhookSubscriptionOrderField) *WebhookSubscriptionEdge {
	return &WebhookSubscriptionEdge{
		Cursor: wc.CursorKey(orderBy),
		Node:   NewWebhookSubscription(wc),
	}
}

func NewWebhookSubscription(wc *coredata.WebhookSubscription) *WebhookSubscription {
	return &WebhookSubscription{
		ID: wc.ID,
		Organization: &Organization{
			ID: wc.OrganizationID,
		},
		EndpointURL:    wc.EndpointURL,
		SelectedEvents: wc.SelectedEvents,
		CreatedAt:      wc.CreatedAt,
		UpdatedAt:      wc.UpdatedAt,
	}
}
