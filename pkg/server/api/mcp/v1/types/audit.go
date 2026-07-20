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
	"go.probo.inc/probo/pkg/page"
)

func NewAudit(a *coredata.Audit, file *coredata.File) *Audit {
	audit := &Audit{
		ID:                         a.ID,
		Name:                       a.Name,
		OrganizationID:             a.OrganizationID,
		FrameworkID:                a.FrameworkID,
		State:                      a.State,
		CompliancePortalVisibility: a.CompliancePortalVisibility,
		HasReport:                  a.ReportFileID != nil,
		ValidFrom:                  a.ValidFrom,
		ValidUntil:                 a.ValidUntil,
		CreatedAt:                  a.CreatedAt,
		UpdatedAt:                  a.UpdatedAt,
	}

	if file != nil {
		audit.ReportFilename = &file.FileName
		audit.ReportMimeType = &file.MimeType
	}

	return audit
}

func NewListControlAuditsOutput(auditPage *page.Page[*coredata.Audit, coredata.AuditOrderField]) ListControlAuditsOutput {
	audits := make([]*Audit, 0, len(auditPage.Data))
	for _, v := range auditPage.Data {
		audits = append(audits, NewAudit(v, nil))
	}

	var nextCursor *page.CursorKey

	if len(auditPage.Data) > 0 {
		cursorKey := auditPage.Data[len(auditPage.Data)-1].CursorKey(auditPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListControlAuditsOutput{
		NextCursor: nextCursor,
		Audits:     audits,
	}
}

func NewListAuditsOutput(auditPage *page.Page[*coredata.Audit, coredata.AuditOrderField]) ListAuditsOutput {
	audits := make([]*Audit, 0, len(auditPage.Data))
	for _, v := range auditPage.Data {
		audits = append(audits, NewAudit(v, nil))
	}

	var nextCursor *page.CursorKey

	if len(auditPage.Data) > 0 {
		cursorKey := auditPage.Data[len(auditPage.Data)-1].CursorKey(auditPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAuditsOutput{
		NextCursor: nextCursor,
		Audits:     audits,
	}
}

func NewListFindingAuditsOutput(auditPage *page.Page[*coredata.Audit, coredata.AuditOrderField]) ListFindingAuditsOutput {
	audits := make([]*Audit, 0, len(auditPage.Data))
	for _, v := range auditPage.Data {
		audits = append(audits, NewAudit(v, nil))
	}

	var nextCursor *page.CursorKey

	if len(auditPage.Data) > 0 {
		cursorKey := auditPage.Data[len(auditPage.Data)-1].CursorKey(auditPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListFindingAuditsOutput{
		NextCursor: nextCursor,
		Audits:     audits,
	}
}
