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
		StatementOfApplicability: &StatementOfApplicability{
			ID: as.StatementOfApplicabilityID,
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
