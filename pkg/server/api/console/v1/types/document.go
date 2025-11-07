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

type Document struct {
	ID                      gid.GID                         `json:"id"`
	OrganizationID          gid.GID                         `json:"-"`
	Title                   string                          `json:"title"`
	Description             *string                         `json:"description,omitempty"`
	DocumentType            coredata.DocumentType           `json:"documentType"`
	Classification          coredata.DocumentClassification `json:"classification"`
	CurrentPublishedVersion *int                            `json:"currentPublishedVersion,omitempty"`
	TrustCenterVisibility   coredata.TrustCenterVisibility  `json:"trustCenterVisibility"`
	Owner                   *People                         `json:"owner"`
	Organization            *Organization                   `json:"organization"`
	Versions                *DocumentVersionConnection      `json:"versions"`
	Controls                *ControlConnection              `json:"controls"`
	CreatedAt               time.Time                       `json:"createdAt"`
	UpdatedAt               time.Time                       `json:"updatedAt"`
}

func (Document) IsNode()             {}
func (this Document) GetID() gid.GID { return this.ID }

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
		ID:                      document.ID,
		OrganizationID:          document.OrganizationID,
		Title:                   document.Title,
		DocumentType:            document.DocumentType,
		Classification:          document.Classification,
		CurrentPublishedVersion: document.CurrentPublishedVersion,
		TrustCenterVisibility:   document.TrustCenterVisibility,
		CreatedAt:               document.CreatedAt,
		UpdatedAt:               document.UpdatedAt,
	}
}
