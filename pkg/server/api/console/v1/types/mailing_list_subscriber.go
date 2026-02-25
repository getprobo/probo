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
	MailingListSubscriberOrderBy OrderBy[coredata.MailingListSubscriberOrderField]

	MailingListSubscriberConnection struct {
		TotalCount int
		Edges      []*MailingListSubscriberEdge
		PageInfo   *PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewMailingListSubscriber(s *coredata.MailingListSubscriber) *MailingListSubscriber {
	return &MailingListSubscriber{
		ID:        s.ID,
		FullName:  s.FullName,
		Email:     s.Email,
		Status:    s.Status,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func NewMailingListSubscriberEdge(s *coredata.MailingListSubscriber, orderBy coredata.MailingListSubscriberOrderField) *MailingListSubscriberEdge {
	return &MailingListSubscriberEdge{
		Cursor: s.CursorKey(orderBy),
		Node:   NewMailingListSubscriber(s),
	}
}

func NewMailingListSubscriberConnection(
	p *page.Page[*coredata.MailingListSubscriber, coredata.MailingListSubscriberOrderField],
	resolver any,
	mailingListID gid.GID,
) *MailingListSubscriberConnection {
	var edges = make([]*MailingListSubscriberEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewMailingListSubscriberEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &MailingListSubscriberConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
		Resolver: resolver,
		ParentID: mailingListID,
	}
}
