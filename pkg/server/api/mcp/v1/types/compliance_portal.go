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

func NewCompliancePortal(tc *coredata.CompliancePortal) *CompliancePortal {
	return &CompliancePortal{
		ID:                   tc.ID,
		OrganizationID:       tc.OrganizationID,
		Active:               tc.Active,
		SearchEngineIndexing: tc.SearchEngineIndexing,
		EntityName:           tc.EntityName,
		Description:          tc.Description,
		WebsiteURL:           tc.WebsiteURL,
		Email:                tc.Email,
		HeadquarterAddress:   tc.HeadquarterAddress,
		CreatedAt:            tc.CreatedAt,
		UpdatedAt:            tc.UpdatedAt,
	}
}

func NewCompliancePortalReference(r *coredata.CompliancePortalReference) *CompliancePortalReference {
	return &CompliancePortalReference{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		WebsiteURL:  &r.WebsiteURL,
		Rank:        r.Rank,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func NewListCompliancePortalReferencesOutput(
	refs []*CompliancePortalReference,
	p *page.Page[*coredata.CompliancePortalReference, coredata.CompliancePortalReferenceOrderField],
) ListCompliancePortalReferencesOutput {
	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCompliancePortalReferencesOutput{
		NextCursor:                 nextCursor,
		CompliancePortalReferences: refs,
	}
}

func NewCompliancePortalFile(f *coredata.CompliancePortalFile, file *File) *CompliancePortalFile {
	return &CompliancePortalFile{
		ID:                         f.ID,
		OrganizationID:             f.OrganizationID,
		Name:                       f.Name,
		Category:                   f.Category,
		File:                       file,
		CompliancePortalVisibility: f.CompliancePortalVisibility,
		CreatedAt:                  f.CreatedAt,
		UpdatedAt:                  f.UpdatedAt,
	}
}

func NewListCompliancePortalFilesOutput(files []*CompliancePortalFile, p *page.Page[*coredata.CompliancePortalFile, coredata.CompliancePortalFileOrderField]) ListCompliancePortalFilesOutput {
	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListCompliancePortalFilesOutput{
		NextCursor:            nextCursor,
		CompliancePortalFiles: files,
	}
}
