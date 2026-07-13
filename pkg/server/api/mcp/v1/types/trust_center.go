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

func NewTrustCenter(tc *coredata.TrustCenter) *TrustCenter {
	return &TrustCenter{
		ID:                   tc.ID,
		OrganizationID:       tc.OrganizationID,
		Active:               tc.Active,
		SearchEngineIndexing: tc.SearchEngineIndexing,
		Title:                tc.Title,
		Description:          tc.Description,
		WebsiteURL:           tc.WebsiteURL,
		Email:                tc.Email,
		HeadquarterAddress:   tc.HeadquarterAddress,
		CreatedAt:            tc.CreatedAt,
		UpdatedAt:            tc.UpdatedAt,
	}
}

func NewTrustCenterReference(r *coredata.TrustCenterReference) *TrustCenterReference {
	return &TrustCenterReference{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		WebsiteURL:  &r.WebsiteURL,
		Rank:        r.Rank,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func NewListTrustCenterReferencesOutput(
	refs []*TrustCenterReference,
	p *page.Page[*coredata.TrustCenterReference, coredata.TrustCenterReferenceOrderField],
) ListTrustCenterReferencesOutput {
	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListTrustCenterReferencesOutput{
		NextCursor:            nextCursor,
		TrustCenterReferences: refs,
	}
}

func NewTrustCenterFile(f *coredata.TrustCenterFile, file *File) *TrustCenterFile {
	return &TrustCenterFile{
		ID:                    f.ID,
		OrganizationID:        f.OrganizationID,
		Name:                  f.Name,
		Category:              f.Category,
		File:                  file,
		TrustCenterVisibility: f.TrustCenterVisibility,
		CreatedAt:             f.CreatedAt,
		UpdatedAt:             f.UpdatedAt,
	}
}

func NewListTrustCenterFilesOutput(files []*TrustCenterFile, p *page.Page[*coredata.TrustCenterFile, coredata.TrustCenterFileOrderField]) ListTrustCenterFilesOutput {
	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListTrustCenterFilesOutput{
		NextCursor:       nextCursor,
		TrustCenterFiles: files,
	}
}
