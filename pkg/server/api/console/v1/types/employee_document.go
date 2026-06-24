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
	case coredata.DocumentOrderFieldUpdatedAt:
		return page.NewCursorKey(d.ID, d.UpdatedAt)
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
