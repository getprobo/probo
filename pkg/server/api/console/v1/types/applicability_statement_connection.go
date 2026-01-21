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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ApplicabilityStatementOrderBy OrderBy[coredata.ApplicabilityStatementOrderField]

	ApplicabilityStatementConnection struct {
		TotalCount int
		Edges      []*ApplicabilityStatementEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewApplicabilityStatementConnection(
	p *page.Page[*coredata.ApplicabilityStatement, coredata.ApplicabilityStatementOrderField],
	parentType any,
	parentID gid.GID,
) *ApplicabilityStatementConnection {
	edges := make([]*ApplicabilityStatementEdge, len(p.Data))
	for i, statement := range p.Data {
		edges[i] = NewApplicabilityStatementEdge(statement, p.Cursor.OrderBy.Field)
	}

	return &ApplicabilityStatementConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewApplicabilityStatementEdge(
	statement *coredata.ApplicabilityStatement,
	orderBy coredata.ApplicabilityStatementOrderField,
) *ApplicabilityStatementEdge {
	return &ApplicabilityStatementEdge{
		Cursor: statement.CursorKey(orderBy),
		Node:   NewApplicabilityStatement(statement),
	}
}

func NewApplicabilityStatement(as *coredata.ApplicabilityStatement) *ApplicabilityStatement {
	justification := ""
	if as.Justification != nil {
		justification = *as.Justification
	}

	return &ApplicabilityStatement{
		ID: as.ID,
		StateOfApplicability: &StateOfApplicability{
			ID: as.StateOfApplicabilityID,
		},
		Control: &Control{
			ID: as.ControlID,
		},
		Applicability: as.Applicability,
		Justification: justification,
		CreatedAt:     as.CreatedAt,
		UpdatedAt:     as.UpdatedAt,
	}
}
