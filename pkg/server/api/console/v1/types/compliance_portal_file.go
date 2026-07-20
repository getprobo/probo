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

type CompliancePortalFileOrderBy = OrderBy[coredata.CompliancePortalFileOrderField]

type CompliancePortalFileConnection struct {
	TotalCount int                         `json:"totalCount"`
	Edges      []*CompliancePortalFileEdge `json:"edges"`
	PageInfo   *PageInfo                   `json:"pageInfo"`
	ParentID   gid.GID                     `json:"-"`
}

func NewCompliancePortalFile(tcf *coredata.CompliancePortalFile) *CompliancePortalFile {
	return &CompliancePortalFile{
		ID:                         tcf.ID,
		Name:                       tcf.Name,
		Category:                   tcf.Category,
		File:                       &File{ID: tcf.FileID},
		CompliancePortalVisibility: tcf.CompliancePortalVisibility,
		CreatedAt:                  tcf.CreatedAt,
		UpdatedAt:                  tcf.UpdatedAt,
		Organization: &Organization{
			ID: tcf.OrganizationID,
		},
	}
}

func NewCompliancePortalFileConnection(
	p *page.Page[*coredata.CompliancePortalFile, coredata.CompliancePortalFileOrderField],
	parentID gid.GID,
) *CompliancePortalFileConnection {
	var edges = make([]*CompliancePortalFileEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewCompliancePortalFileEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &CompliancePortalFileConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
		ParentID: parentID,
	}
}

func NewCompliancePortalFileEdge(tcf *coredata.CompliancePortalFile, orderBy coredata.CompliancePortalFileOrderField) *CompliancePortalFileEdge {
	return &CompliancePortalFileEdge{
		Cursor: tcf.CursorKey(orderBy),
		Node:   NewCompliancePortalFile(tcf),
	}
}
