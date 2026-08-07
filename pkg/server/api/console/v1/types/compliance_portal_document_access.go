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
	CompliancePortalDocumentAccessOrderBy = OrderBy[coredata.CompliancePortalDocumentAccessOrderField]

	CompliancePortalDocumentAccessConnection struct {
		TotalCount int
		Edges      []*CompliancePortalDocumentAccessEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewCompliancePortalDocumentAccess(tcda *coredata.CompliancePortalDocumentAccess) *CompliancePortalDocumentAccess {
	object := &CompliancePortalDocumentAccess{
		ID:     tcda.ID,
		Status: tcda.Status,
	}

	if tcda.DocumentID != nil {
		object.Document = &Document{
			ID: *tcda.DocumentID,
		}
	}

	if tcda.ReportFileID != nil {
		object.ReportFile = &File{
			ID: *tcda.ReportFileID,
		}
	}

	if tcda.CompliancePortalFileID != nil {
		object.CompliancePortalFile = &CompliancePortalFile{
			ID: *tcda.CompliancePortalFileID,
		}
	}

	return object
}

func NewCompliancePortalDocumentAccessConnection(
	p *page.Page[*coredata.CompliancePortalDocumentAccess, coredata.CompliancePortalDocumentAccessOrderField],
	parentType any,
	parentID gid.GID,
) *CompliancePortalDocumentAccessConnection {
	var edges = make([]*CompliancePortalDocumentAccessEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewCompliancePortalDocumentAccessEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &CompliancePortalDocumentAccessConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewCompliancePortalDocumentAccessEdge(access *coredata.CompliancePortalDocumentAccess, orderBy coredata.CompliancePortalDocumentAccessOrderField) *CompliancePortalDocumentAccessEdge {
	return &CompliancePortalDocumentAccessEdge{
		Cursor: access.CursorKey(orderBy),
		Node:   NewCompliancePortalDocumentAccess(access),
	}
}
