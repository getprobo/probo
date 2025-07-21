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
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/page"
)

type (
	DocumentVersionSignatureOrderBy OrderBy[coredata.DocumentVersionSignatureOrderField]
)

func NewDocumentVersionSignatureConnection(page *page.Page[*coredata.DocumentVersionSignature, coredata.DocumentVersionSignatureOrderField]) *DocumentVersionSignatureConnection {
	edges := make([]*DocumentVersionSignatureEdge, len(page.Data))
	for i, documentVersionSignature := range page.Data {
		edges[i] = NewDocumentVersionSignatureEdge(documentVersionSignature, page.Cursor.OrderBy.Field)
	}

	return &DocumentVersionSignatureConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(page),
	}
}

func NewDocumentVersionSignatureEdges(documentVersionSignatures []*coredata.DocumentVersionSignature, orderBy coredata.DocumentVersionSignatureOrderField) []*DocumentVersionSignatureEdge {
	edges := make([]*DocumentVersionSignatureEdge, len(documentVersionSignatures))
	for i := range edges {
		edges[i] = NewDocumentVersionSignatureEdge(documentVersionSignatures[i], orderBy)
	}

	return edges
}

func NewDocumentVersionSignatureEdge(documentVersionSignature *coredata.DocumentVersionSignature, orderBy coredata.DocumentVersionSignatureOrderField) *DocumentVersionSignatureEdge {
	return &DocumentVersionSignatureEdge{
		Cursor: documentVersionSignature.CursorKey(orderBy),
		Node:   NewDocumentVersionSignature(documentVersionSignature),
	}
}

func NewDocumentVersionSignature(documentVersionSignature *coredata.DocumentVersionSignature) *DocumentVersionSignature {
	return &DocumentVersionSignature{
		ID:          documentVersionSignature.ID,
		State:       documentVersionSignature.State,
		SignedAt:    documentVersionSignature.SignedAt,
		RequestedAt: documentVersionSignature.RequestedAt,
		CreatedAt:   documentVersionSignature.CreatedAt,
		UpdatedAt:   documentVersionSignature.UpdatedAt,
	}
}
