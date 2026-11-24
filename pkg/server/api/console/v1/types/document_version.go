// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
)

type (
	DocumentVersionOrderBy OrderBy[coredata.DocumentVersionOrderField]
)

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
		ID: documentVersion.ID,
		Document: &Document{
			ID: documentVersion.DocumentID,
		},
		Major:          documentVersion.Major,
		Minor:          documentVersion.Minor,
		Title:          documentVersion.Title,
		Content:        NewDocumentContent(documentVersion.Content),
		Status:         documentVersion.Status,
		Classification: documentVersion.Classification,
		DocumentType:   documentVersion.DocumentType,
		Orientation:    documentVersion.Orientation,
		PublishedAt:    documentVersion.PublishedAt,
		Changelog:      documentVersion.Changelog,
		CreatedAt:      documentVersion.CreatedAt,
		UpdatedAt:      documentVersion.UpdatedAt,
	}
}
