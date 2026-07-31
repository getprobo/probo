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
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewCompliancePortalCatalogDocument(entry *management.CatalogDocument) *CompliancePortalCatalogDocument {
	return &CompliancePortalCatalogDocument{
		ID:         entry.ID,
		Document:   NewDocument(entry.Document),
		Visibility: entry.Visibility,
	}
}

func NewCompliancePortalCatalogAudit(entry *management.CatalogAudit, reportFile *coredata.File) *CompliancePortalCatalogAudit {
	return &CompliancePortalCatalogAudit{
		ID:         entry.ID,
		Audit:      NewAudit(entry.Audit, reportFile),
		Visibility: entry.Visibility,
	}
}

func NewCompliancePortalCatalogThirdParty(entry *management.CatalogThirdParty) *CompliancePortalCatalogThirdParty {
	return &CompliancePortalCatalogThirdParty{
		ID:         entry.ID,
		ThirdParty: NewThirdParty(entry.ThirdParty, nil),
	}
}

func NewListCompliancePortalDocumentsOutput(
	entries []*CompliancePortalCatalogDocument,
	p *page.Page[*management.CatalogDocument, coredata.DocumentOrderField],
) ListCompliancePortalDocumentsOutput {
	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCompliancePortalDocumentsOutput{
		NextCursor:                nextCursor,
		CompliancePortalDocuments: entries,
	}
}

func NewListCompliancePortalAuditsOutput(
	entries []*CompliancePortalCatalogAudit,
	p *page.Page[*management.CatalogAudit, coredata.AuditOrderField],
) ListCompliancePortalAuditsOutput {
	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCompliancePortalAuditsOutput{
		NextCursor:             nextCursor,
		CompliancePortalAudits: entries,
	}
}

func NewListCompliancePortalThirdPartiesOutput(
	entries []*CompliancePortalCatalogThirdParty,
	p *page.Page[*management.CatalogThirdParty, coredata.ThirdPartyOrderField],
) ListCompliancePortalThirdPartiesOutput {
	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCompliancePortalThirdPartiesOutput{
		NextCursor:                   nextCursor,
		CompliancePortalThirdParties: entries,
	}
}
