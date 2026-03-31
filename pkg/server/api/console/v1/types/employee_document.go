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
	"fmt"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type EmployeeDocumentFilterMode string

const (
	EmployeeDocumentFilterModeSignature EmployeeDocumentFilterMode = "SIGNATURE"
	EmployeeDocumentFilterModeApproval  EmployeeDocumentFilterMode = "APPROVAL"
)

type (
	EmployeeDocumentConnection struct {
		Edges    []*EmployeeDocumentEdge
		PageInfo *PageInfo
	}

	EmployeeDocumentEdge struct {
		Cursor page.CursorKey
		Node   *EmployeeDocument
	}

	EmployeeDocument struct {
		ID           gid.GID
		Title        string
		Description  *string
		DocumentType coredata.DocumentType
		CreatedAt    time.Time
		UpdatedAt    time.Time

		FilterMode EmployeeDocumentFilterMode
	}

	EmployeeDocumentVersionConnection struct {
		Edges    []*EmployeeDocumentVersionEdge
		PageInfo *PageInfo
	}

	EmployeeDocumentVersionEdge struct {
		Cursor page.CursorKey
		Node   *EmployeeDocumentVersion
	}

	EmployeeDocumentVersion struct {
		ID             gid.GID
		DocumentID     gid.GID
		OrganizationID gid.GID
		Major          int
		Minor          int
		Status         coredata.DocumentVersionStatus
		Classification coredata.DocumentClassification
		DocumentType   coredata.DocumentType
		PublishedAt    *time.Time
		CreatedAt      time.Time
		UpdatedAt      time.Time
	}
)

func NewEmployeeDocumentConnection(
	p *page.Page[*EmployeeDocument, coredata.DocumentOrderField],
) *EmployeeDocumentConnection {
	var edges = make([]*EmployeeDocumentEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewEmployeeDocumentEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &EmployeeDocumentConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewEmployeeDocumentEdge(document *EmployeeDocument, orderBy coredata.DocumentOrderField) *EmployeeDocumentEdge {
	return &EmployeeDocumentEdge{
		Cursor: document.CursorKey(orderBy),
		Node:   document,
	}
}

func (d EmployeeDocument) CursorKey(orderBy coredata.DocumentOrderField) page.CursorKey {
	switch orderBy {
	case coredata.DocumentOrderFieldCreatedAt:
		return page.NewCursorKey(d.ID, d.CreatedAt)
	case coredata.DocumentOrderFieldTitle:
		return page.NewCursorKey(d.ID, d.Title)
	case coredata.DocumentOrderFieldDocumentType:
		return page.NewCursorKey(d.ID, d.DocumentType)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func NewEmployeeDocumentVersionConnection(
	p *page.Page[*EmployeeDocumentVersion, coredata.DocumentVersionOrderField],
) *EmployeeDocumentVersionConnection {
	var edges = make([]*EmployeeDocumentVersionEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewEmployeeDocumentVersionEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &EmployeeDocumentVersionConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewEmployeeDocumentVersionEdge(version *EmployeeDocumentVersion, orderBy coredata.DocumentVersionOrderField) *EmployeeDocumentVersionEdge {
	return &EmployeeDocumentVersionEdge{
		Cursor: version.CursorKey(orderBy),
		Node:   version,
	}
}

func (v EmployeeDocumentVersion) CursorKey(orderBy coredata.DocumentVersionOrderField) page.CursorKey {
	switch orderBy {
	case coredata.DocumentVersionOrderFieldCreatedAt:
		return page.NewCursorKey(v.ID, v.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}
