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
	DocumentVersionOrderBy OrderBy[coredata.DocumentVersionOrderField]
)

func NewDocumentVersionConnection(page *page.Page[*coredata.DocumentVersion, coredata.DocumentVersionOrderField]) *DocumentVersionConnection {
	edges := make([]*DocumentVersionEdge, len(page.Data))
	for i, documentVersion := range page.Data {
		edges[i] = NewDocumentVersionEdge(documentVersion, page.Cursor.OrderBy.Field)
	}

	return &DocumentVersionConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(page),
	}
}

func NewDocumentVersionEdges(documentVersions []*coredata.DocumentVersion, orderBy coredata.DocumentVersionOrderField) []*DocumentVersionEdge {
	edges := make([]*DocumentVersionEdge, len(documentVersions))

	for i := range edges {
		edges[i] = NewDocumentVersionEdge(documentVersions[i], orderBy)
	}

	return edges
}

func NewDocumentVersionEdge(documentVersion *coredata.DocumentVersion, orderBy coredata.DocumentVersionOrderField) *DocumentVersionEdge {
	return &DocumentVersionEdge{
		Cursor: documentVersion.CursorKey(orderBy),
		Node:   NewDocumentVersion(documentVersion),
	}
}

func NewDocumentVersion(documentVersion *coredata.DocumentVersion) *DocumentVersion {
	return &DocumentVersion{
		ID:          documentVersion.ID,
		Version:     documentVersion.VersionNumber,
		Title:       documentVersion.Title,
		Content:     documentVersion.Content,
		Status:      documentVersion.Status,
		PublishedAt: documentVersion.PublishedAt,
		Changelog:   documentVersion.Changelog,
		CreatedAt:   documentVersion.CreatedAt,
		UpdatedAt:   documentVersion.UpdatedAt,
	}
}
