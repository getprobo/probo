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
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
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

func NewAudit(a *coredata.Audit) *Audit {
	audit := &Audit{
		ID:           a.ID,
		ValidFrom:    a.ValidFrom,
		ValidUntil:   a.ValidUntil,
		State:        a.State,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
		Organization: &Organization{ID: a.OrganizationID},
		Framework:    &Framework{ID: a.FrameworkID},
	}

	if a.ReportID != nil {
		audit.Report = &Report{ID: *a.ReportID}
	}

	return audit
}

func NewAuditEdge(a *coredata.Audit, orderField coredata.AuditOrderField) *AuditEdge {
	return &AuditEdge{
		Node:   NewAudit(a),
		Cursor: a.CursorKey(orderField),
	}
}
