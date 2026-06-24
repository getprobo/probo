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
	"go.probo.inc/probo/pkg/page"
)

func NewStatementOfApplicability(s *coredata.StatementOfApplicability) *StatementOfApplicability {
	return &StatementOfApplicability{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		Name:           s.Name,
		DocumentID:     s.DocumentID,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func NewListStatementsOfApplicabilityOutput(pg *page.Page[*coredata.StatementOfApplicability, coredata.StatementOfApplicabilityOrderField]) ListStatementsOfApplicabilityOutput {
	items := make([]*StatementOfApplicability, 0, len(pg.Data))
	for _, v := range pg.Data {
		items = append(items, NewStatementOfApplicability(v))
	}

	var nextCursor *page.CursorKey

	if len(pg.Data) > 0 {
		cursorKey := pg.Data[len(pg.Data)-1].CursorKey(pg.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListStatementsOfApplicabilityOutput{
		NextCursor:                nextCursor,
		StatementsOfApplicability: items,
	}
}

func NewApplicabilityStatement(a *coredata.ApplicabilityStatement) *ApplicabilityStatement {
	return &ApplicabilityStatement{
		ID:                         a.ID,
		StatementOfApplicabilityID: a.StatementOfApplicabilityID,
		ControlID:                  a.ControlID,
		OrganizationID:             a.OrganizationID,
		Applicability:              a.Applicability,
		Justification:              a.Justification,
		CreatedAt:                  a.CreatedAt,
		UpdatedAt:                  a.UpdatedAt,
	}
}

func NewListApplicabilityStatementsOutput(pg *page.Page[*coredata.ApplicabilityStatement, coredata.ApplicabilityStatementOrderField]) ListApplicabilityStatementsOutput {
	items := make([]*ApplicabilityStatement, 0, len(pg.Data))
	for _, v := range pg.Data {
		items = append(items, NewApplicabilityStatement(v))
	}

	var nextCursor *page.CursorKey

	if len(pg.Data) > 0 {
		cursorKey := pg.Data[len(pg.Data)-1].CursorKey(pg.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListApplicabilityStatementsOutput{
		NextCursor:              nextCursor,
		ApplicabilityStatements: items,
	}
}
