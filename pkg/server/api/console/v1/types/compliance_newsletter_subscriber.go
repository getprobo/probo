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
	"go.probo.inc/probo/pkg/page"
)

func EmptyComplianceNewsletterSubscriberConnection() *ComplianceNewsletterSubscriberConnection {
	return &ComplianceNewsletterSubscriberConnection{
		Edges:    []*ComplianceNewsletterSubscriberEdge{},
		PageInfo: &PageInfo{},
	}
}

func NewComplianceNewsletterSubscriberConnection(
	p *page.Page[*coredata.ComplianceNewsletterSubscriber, coredata.ComplianceNewsletterSubscriberOrderField],
) *ComplianceNewsletterSubscriberConnection {
	edges := make([]*ComplianceNewsletterSubscriberEdge, len(p.Data))

	for i, subscriber := range p.Data {
		edges[i] = &ComplianceNewsletterSubscriberEdge{
			Cursor: subscriber.CursorKey(p.Cursor.OrderBy.Field),
			Node: &ComplianceNewsletterSubscriber{
				ID:        subscriber.ID,
				Email:     subscriber.Email,
				CreatedAt: subscriber.CreatedAt,
				UpdatedAt: subscriber.UpdatedAt,
			},
		}
	}

	return &ComplianceNewsletterSubscriberConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}
