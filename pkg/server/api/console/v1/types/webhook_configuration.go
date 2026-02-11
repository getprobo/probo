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
	WebhookConfigurationOrderBy OrderBy[coredata.WebhookConfigurationOrderField]

	WebhookConfigurationConnection struct {
		TotalCount int
		Edges      []*WebhookConfigurationEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewWebhookConfigurationConnection(
	p *page.Page[*coredata.WebhookConfiguration, coredata.WebhookConfigurationOrderField],
	parentType any,
	parentID gid.GID,
) *WebhookConfigurationConnection {
	var edges = make([]*WebhookConfigurationEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewWebhookConfigurationEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &WebhookConfigurationConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewWebhookConfigurationEdge(wc *coredata.WebhookConfiguration, orderBy coredata.WebhookConfigurationOrderField) *WebhookConfigurationEdge {
	return &WebhookConfigurationEdge{
		Cursor: wc.CursorKey(orderBy),
		Node:   NewWebhookConfiguration(wc),
	}
}

func NewWebhookConfiguration(wc *coredata.WebhookConfiguration) *WebhookConfiguration {
	return &WebhookConfiguration{
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
