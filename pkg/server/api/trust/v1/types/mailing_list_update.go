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
