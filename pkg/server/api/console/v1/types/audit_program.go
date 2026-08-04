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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AuditProgramOrderBy OrderBy[coredata.AuditProgramOrderField]

	AuditProgramConnection struct {
		TotalCount int
		Edges      []*AuditProgramEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewAuditProgramConnection(
	p *page.Page[*coredata.AuditProgram, coredata.AuditProgramOrderField],
	parentType any,
	parentID gid.GID,
) *AuditProgramConnection {
	edges := make([]*AuditProgramEdge, len(p.Data))
	for i, auditProgram := range p.Data {
		edges[i] = NewAuditProgramEdge(auditProgram, p.Cursor.OrderBy.Field)
	}

	return &AuditProgramConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewAuditProgramEdge(a *coredata.AuditProgram, orderField coredata.AuditProgramOrderField) *AuditProgramEdge {
	return &AuditProgramEdge{
		Node:   NewAuditProgram(a),
		Cursor: a.CursorKey(orderField),
	}
}

func NewAuditProgram(a *coredata.AuditProgram) *AuditProgram {
	return &AuditProgram{
		ID: a.ID,
		Organization: &Organization{
			ID: a.OrganizationID,
		},
		Framework: &Framework{
			ID: a.FrameworkID,
		},
		Name:       a.Name,
		ValidFrom:  a.ValidFrom,
		ValidUntil: a.ValidUntil,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}
