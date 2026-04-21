// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

func NewTrustCenter(tc *coredata.TrustCenter) *TrustCenter {
	return &TrustCenter{
		ID:                   tc.ID,
		OrganizationID:       tc.OrganizationID,
		Active:               tc.Active,
		SearchEngineIndexing: tc.SearchEngineIndexing,
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

func NewListTrustCenterReferencesOutput(p *page.Page[*coredata.TrustCenterReference, coredata.TrustCenterReferenceOrderField]) ListTrustCenterReferencesOutput {
	refs := make([]*TrustCenterReference, 0, len(p.Data))
	for _, r := range p.Data {
		refs = append(refs, NewTrustCenterReference(r))
	}

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

func NewTrustCenterFile(f *coredata.TrustCenterFile, fileURL string) *TrustCenterFile {
	return &TrustCenterFile{
		ID:                    f.ID,
		OrganizationID:        f.OrganizationID,
		Name:                  f.Name,
		Category:              f.Category,
		FileURL:               fileURL,
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
