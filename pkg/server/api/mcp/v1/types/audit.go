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
	"go.probo.inc/probo/pkg/page"
)

func NewAudit(a *coredata.Audit) *Audit {
	return &Audit{
		ID:                    a.ID,
		Name:                  a.Name,
		OrganizationID:        a.OrganizationID,
		FrameworkID:           a.FrameworkID,
		State:                 a.State,
		TrustCenterVisibility: a.TrustCenterVisibility,
		ValidFrom:             a.ValidFrom,
		ValidUntil:            a.ValidUntil,
		CreatedAt:             a.CreatedAt,
		UpdatedAt:             a.UpdatedAt,
	}
}

func NewListAuditsOutput(auditPage *page.Page[*coredata.Audit, coredata.AuditOrderField]) ListAuditsOutput {
	audits := make([]*Audit, 0, len(auditPage.Data))
	for _, v := range auditPage.Data {
		audits = append(audits, NewAudit(v))
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
