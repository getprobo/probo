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

func NewMailingListUpdate(mlu *coredata.MailingListUpdate) *MailingListUpdate {
	return &MailingListUpdate{
		ID:        mlu.ID,
		Title:     mlu.Title,
		Body:      mlu.Body,
		UpdatedAt: mlu.UpdatedAt,
	}
}

func NewMailingListUpdateEdge(mlu *coredata.MailingListUpdate) *MailingListUpdateEdge {
	return &MailingListUpdateEdge{
		Cursor: mlu.CursorKey(coredata.MailingListUpdateOrderFieldUpdatedAt),
		Node:   NewMailingListUpdate(mlu),
	}
}

func NewMailingListUpdateConnection(
	p *page.Page[*coredata.MailingListUpdate, coredata.MailingListUpdateOrderField],
) *MailingListUpdateConnection {
	edges := make([]*MailingListUpdateEdge, len(p.Data))
	for i, mlu := range p.Data {
		edges[i] = NewMailingListUpdateEdge(mlu)
	}

	return &MailingListUpdateConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}
