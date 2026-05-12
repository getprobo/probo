// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

type MailingListUpdateConnection struct {
	TotalCount int
	Edges      []*MailingListUpdateEdge
	PageInfo   *PageInfo

	Resolver any
	ParentID gid.GID
}

func NewMailingListUpdate(mlu *coredata.MailingListUpdate) *MailingListUpdate {
	return &MailingListUpdate{
		ID:        mlu.ID,
		Title:     mlu.Title,
		Body:      mlu.Body,
		Status:    mlu.Status,
		CreatedAt: mlu.CreatedAt,
		UpdatedAt: mlu.UpdatedAt,
	}
}

func NewMailingListUpdateEdge(mlu *coredata.MailingListUpdate, orderBy coredata.MailingListUpdateOrderField) *MailingListUpdateEdge {
	return &MailingListUpdateEdge{
		Cursor: mlu.CursorKey(orderBy),
		Node:   NewMailingListUpdate(mlu),
	}
}

func NewMailingListUpdateConnection(
	p *page.Page[*coredata.MailingListUpdate, coredata.MailingListUpdateOrderField],
	resolver any,
	mailingListID gid.GID,
) *MailingListUpdateConnection {
	edges := make([]*MailingListUpdateEdge, len(p.Data))
	for i := range edges {
		edges[i] = NewMailingListUpdateEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &MailingListUpdateConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
		Resolver: resolver,
		ParentID: mailingListID,
	}
}
