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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	StatementOfApplicabilityOrderBy OrderBy[coredata.StatementOfApplicabilityOrderField]

	StatementOfApplicabilityConnection struct {
		TotalCount int
		Edges      []*StatementOfApplicabilityEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewStatementOfApplicabilityConnection(
	p *page.Page[*coredata.StatementOfApplicability, coredata.StatementOfApplicabilityOrderField],
	parentType any,
	parentID gid.GID,
) *StatementOfApplicabilityConnection {
	var edges = make([]*StatementOfApplicabilityEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewStatementOfApplicabilityEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &StatementOfApplicabilityConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewStatementOfApplicabilityEdge(soa *coredata.StatementOfApplicability, orderBy coredata.StatementOfApplicabilityOrderField) *StatementOfApplicabilityEdge {
	return &StatementOfApplicabilityEdge{
		Cursor: soa.CursorKey(orderBy),
		Node:   NewStatementOfApplicability(soa),
	}
}

func NewStatementOfApplicability(soa *coredata.StatementOfApplicability) *StatementOfApplicability {
	s := &StatementOfApplicability{
		ID: soa.ID,
		Organization: &Organization{
			ID: soa.OrganizationID,
		},
		Name:      soa.Name,
		CreatedAt: soa.CreatedAt,
		UpdatedAt: soa.UpdatedAt,
	}

	if soa.DocumentID != nil {
		s.Document = &Document{
			ID: *soa.DocumentID,
		}
	}

	return s
}
