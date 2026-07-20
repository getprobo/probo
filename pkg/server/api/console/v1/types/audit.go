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
	AuditOrderBy OrderBy[coredata.AuditOrderField]

	AuditConnection struct {
		TotalCount int
		Edges      []*AuditEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewAuditConnection(
	p *page.Page[*coredata.Audit, coredata.AuditOrderField],
	parentType any,
	parentID gid.GID,
) *AuditConnection {
	edges := make([]*AuditEdge, len(p.Data))
	for i, audit := range p.Data {
		edges[i] = NewAuditEdge(audit, p.Cursor.OrderBy.Field)
	}

	return &AuditConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAuditEdge(a *coredata.Audit, orderField coredata.AuditOrderField) *AuditEdge {
	return &AuditEdge{
		Node:   NewAudit(a),
		Cursor: a.CursorKey(orderField),
	}
}

func NewAudit(a *coredata.Audit) *Audit {
	node := &Audit{
		ID: a.ID,
		Organization: &Organization{
			ID: a.OrganizationID,
		},
		Framework: &Framework{
			ID: a.FrameworkID,
		},
		ValidFrom:                  a.ValidFrom,
		ValidUntil:                 a.ValidUntil,
		State:                      a.State,
		Name:                       a.Name,
		CompliancePortalVisibility: a.CompliancePortalVisibility,
		CreatedAt:                  a.CreatedAt,
		UpdatedAt:                  a.UpdatedAt,
	}

	if a.ReportFileID != nil {
		node.ReportFile = &File{
			ID: *a.ReportFileID,
		}
	}

	return node
}
