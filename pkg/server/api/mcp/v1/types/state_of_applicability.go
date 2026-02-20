// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHOR BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN
// ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewStateOfApplicability(s *coredata.StateOfApplicability) *StateOfApplicability {
	return &StateOfApplicability{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		Name:           s.Name,
		OwnerID:        s.OwnerID,
		SnapshotID:     s.SnapshotID,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func NewListStatesOfApplicabilityOutput(pg *page.Page[*coredata.StateOfApplicability, coredata.StateOfApplicabilityOrderField]) ListStatesOfApplicabilityOutput {
	items := make([]*StateOfApplicability, 0, len(pg.Data))
	for _, v := range pg.Data {
		items = append(items, NewStateOfApplicability(v))
	}
	var nextCursor *page.CursorKey
	if len(pg.Data) > 0 {
		cursorKey := pg.Data[len(pg.Data)-1].CursorKey(pg.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}
	return ListStatesOfApplicabilityOutput{
		NextCursor:            nextCursor,
		StatesOfApplicability: items,
	}
}

func NewApplicabilityStatement(a *coredata.ApplicabilityStatement) *ApplicabilityStatement {
	return &ApplicabilityStatement{
		ID:                     a.ID,
		StateOfApplicabilityID: a.StateOfApplicabilityID,
		ControlID:              a.ControlID,
		OrganizationID:         a.OrganizationID,
		SnapshotID:             a.SnapshotID,
		Applicability:          a.Applicability,
		Justification:          a.Justification,
		CreatedAt:              a.CreatedAt,
		UpdatedAt:              a.UpdatedAt,
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
