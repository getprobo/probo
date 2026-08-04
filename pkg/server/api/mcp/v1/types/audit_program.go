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

func NewAuditProgram(a *coredata.AuditProgram) *AuditProgram {
	return &AuditProgram{
		ID:             a.ID,
		OrganizationID: a.OrganizationID,
		FrameworkID:    a.FrameworkID,
		Name:           a.Name,
		ValidFrom:      a.ValidFrom,
		ValidUntil:     a.ValidUntil,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

func NewListAuditProgramsOutput(auditProgramPage *page.Page[*coredata.AuditProgram, coredata.AuditProgramOrderField]) ListAuditProgramsOutput {
	auditPrograms := make([]*AuditProgram, 0, len(auditProgramPage.Data))
	for _, v := range auditProgramPage.Data {
		auditPrograms = append(auditPrograms, NewAuditProgram(v))
	}

	var nextCursor *page.CursorKey

	if len(auditProgramPage.Data) > 0 {
		cursorKey := auditProgramPage.Data[len(auditProgramPage.Data)-1].CursorKey(auditProgramPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAuditProgramsOutput{
		NextCursor:    nextCursor,
		AuditPrograms: auditPrograms,
	}
}
