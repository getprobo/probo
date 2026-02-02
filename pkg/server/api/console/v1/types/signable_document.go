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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	SignableDocumentConnection struct {
		Edges    []*SignableDocumentEdge
		PageInfo *PageInfo
	}

	SignableDocumentEdge struct {
		Cursor page.CursorKey
		Node   *SignableDocument
	}

	SignableDocument struct {
		ID             gid.GID
		Title          string
		Description    *string
		DocumentType   coredata.DocumentType
		Classification coredata.DocumentClassification
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}
)

func (SignableDocument) IsNode()          {}
func (d SignableDocument) GetID() gid.GID { return d.ID }

func NewSignableDocumentConnection(
	p *page.Page[*SignableDocument, coredata.DocumentOrderField],
) *SignableDocumentConnection {
	var edges = make([]*SignableDocumentEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewSignableDocumentEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &SignableDocumentConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewSignableDocumentEdge(document *SignableDocument, orderBy coredata.DocumentOrderField) *SignableDocumentEdge {
	return &SignableDocumentEdge{
		Cursor: document.CursorKey(orderBy),
		Node:   document,
	}
}

func (d SignableDocument) CursorKey(orderBy coredata.DocumentOrderField) page.CursorKey {
	switch orderBy {
	case coredata.DocumentOrderFieldCreatedAt:
		return page.NewCursorKey(d.ID, d.CreatedAt)
	case coredata.DocumentOrderFieldTitle:
		return page.NewCursorKey(d.ID, d.Title)
	case coredata.DocumentOrderFieldDocumentType:
		return page.NewCursorKey(d.ID, d.DocumentType)
	}

	panic("unsupported order by")
}
