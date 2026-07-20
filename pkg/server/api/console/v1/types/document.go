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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	DocumentOrderBy OrderBy[coredata.DocumentOrderField]

	DocumentConnection struct {
		TotalCount int
		Edges      []*DocumentEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filters  *coredata.DocumentFilter
	}
)

func NewDocumentConnection(
	p *page.Page[*coredata.Document, coredata.DocumentOrderField],
	parentType any,
	parentID gid.GID,
	filters *coredata.DocumentFilter,
) *DocumentConnection {
	var edges = make([]*DocumentEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewDocumentEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &DocumentConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filters:  filters,
	}
}

func NewDocumentEdges(documents []*coredata.Document, orderBy coredata.DocumentOrderField) []*DocumentEdge {
	edges := make([]*DocumentEdge, len(documents))

	for i := range edges {
		edges[i] = NewDocumentEdge(documents[i], orderBy)
	}

	return edges
}

func NewDocumentEdge(document *coredata.Document, orderBy coredata.DocumentOrderField) *DocumentEdge {
	return &DocumentEdge{
		Cursor: document.CursorKey(orderBy),
		Node:   NewDocument(document),
	}
}

func NewDocument(document *coredata.Document) *Document {
	return &Document{
		ID: document.ID,
		Organization: &Organization{
			ID: document.OrganizationID,
		},
		CurrentPublishedMajor:      document.CurrentPublishedMajor,
		CurrentPublishedMinor:      document.CurrentPublishedMinor,
		WriteMode:                  document.WriteMode,
		CompliancePortalVisibility: document.CompliancePortalVisibility,
		Status:                     document.Status,
		ArchivedAt:                 document.ArchivedAt,
		CreatedAt:                  document.CreatedAt,
		UpdatedAt:                  document.UpdatedAt,
	}
}
