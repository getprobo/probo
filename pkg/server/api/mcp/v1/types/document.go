// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

func NewDocument(d *coredata.Document) *Document {
	return &Document{
		ID:                    d.ID,
		OrganizationID:        d.OrganizationID,
		Title:                 d.Title,
		CurrentPublishedMajor: d.CurrentPublishedMajor,
		CurrentPublishedMinor: d.CurrentPublishedMinor,
		TrustCenterVisibility: d.TrustCenterVisibility,
		Status:                d.Status,
		ArchivedAt:            d.ArchivedAt,
		CreatedAt:             d.CreatedAt,
		UpdatedAt:             d.UpdatedAt,
	}
}

func NewListControlDocumentsOutput(documentPage *page.Page[*coredata.Document, coredata.DocumentOrderField]) ListControlDocumentsOutput {
	documents := make([]*Document, 0, len(documentPage.Data))
	for _, d := range documentPage.Data {
		documents = append(documents, NewDocument(d))
	}

	var nextCursor *page.CursorKey
	if len(documentPage.Data) > 0 {
		cursorKey := documentPage.Data[len(documentPage.Data)-1].CursorKey(documentPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListControlDocumentsOutput{
		NextCursor: nextCursor,
		Documents:  documents,
	}
}

func NewListMeasureDocumentsOutput(documentPage *page.Page[*coredata.Document, coredata.DocumentOrderField]) ListMeasureDocumentsOutput {
	documents := make([]*Document, 0, len(documentPage.Data))
	for _, d := range documentPage.Data {
		documents = append(documents, NewDocument(d))
	}

	var nextCursor *page.CursorKey
	if len(documentPage.Data) > 0 {
		cursorKey := documentPage.Data[len(documentPage.Data)-1].CursorKey(documentPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListMeasureDocumentsOutput{
		NextCursor: nextCursor,
		Documents:  documents,
	}
}

func NewListDocumentsOutput(documentPage *page.Page[*coredata.Document, coredata.DocumentOrderField]) ListDocumentsOutput {
	documents := make([]*Document, 0, len(documentPage.Data))
	for _, d := range documentPage.Data {
		documents = append(documents, NewDocument(d))
	}

	var nextCursor *page.CursorKey
	if len(documentPage.Data) > 0 {
		cursorKey := documentPage.Data[len(documentPage.Data)-1].CursorKey(documentPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListDocumentsOutput{
		NextCursor: nextCursor,
		Documents:  documents,
	}
}

func NewAddDocumentOutput(doc *coredata.Document, docVersion *coredata.DocumentVersion) AddDocumentOutput {
	return AddDocumentOutput{
		Document:        NewDocument(doc),
		DocumentVersion: NewDocumentVersion(docVersion),
	}
}

func NewDocumentVersion(dv *coredata.DocumentVersion) *DocumentVersion {
	return &DocumentVersion{
		ID:             dv.ID,
		OrganizationID: dv.OrganizationID,
		DocumentID:     dv.DocumentID,
		Title:          dv.Title,
		Major:          dv.Major,
		Minor:          dv.Minor,
		Classification: dv.Classification,
		DocumentType:   dv.DocumentType,
		Content:        dv.Content,
		Changelog:      dv.Changelog,
		Status:         dv.Status,
		PublishedAt:    dv.PublishedAt,
		CreatedAt:      dv.CreatedAt,
		UpdatedAt:      dv.UpdatedAt,
	}
}

func NewListDocumentVersionsOutput(versionPage *page.Page[*coredata.DocumentVersion, coredata.DocumentVersionOrderField]) ListDocumentVersionsOutput {
	versions := make([]*DocumentVersion, 0, len(versionPage.Data))
	for _, v := range versionPage.Data {
		versions = append(versions, NewDocumentVersion(v))
	}

	var nextCursor *page.CursorKey
	if len(versionPage.Data) > 0 {
		cursorKey := versionPage.Data[len(versionPage.Data)-1].CursorKey(versionPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListDocumentVersionsOutput{
		NextCursor:       nextCursor,
		DocumentVersions: versions,
	}
}

func NewDocumentVersionSignature(dvs *coredata.DocumentVersionSignature) *DocumentVersionSignature {
	return &DocumentVersionSignature{
		ID:                dvs.ID,
		OrganizationID:    dvs.OrganizationID,
		DocumentVersionID: dvs.DocumentVersionID,
		State:             dvs.State,
		SignedBy:          dvs.SignedBy,
		SignedAt:          dvs.SignedAt,
		RequestedAt:       dvs.RequestedAt,
		CreatedAt:         dvs.CreatedAt,
		UpdatedAt:         dvs.UpdatedAt,
	}
}

func NewListDocumentVersionSignaturesOutput(signaturePage *page.Page[*coredata.DocumentVersionSignature, coredata.DocumentVersionSignatureOrderField]) ListDocumentVersionSignaturesOutput {
	signatures := make([]*DocumentVersionSignature, 0, len(signaturePage.Data))
	for _, s := range signaturePage.Data {
		signatures = append(signatures, NewDocumentVersionSignature(s))
	}

	var nextCursor *page.CursorKey
	if len(signaturePage.Data) > 0 {
		cursorKey := signaturePage.Data[len(signaturePage.Data)-1].CursorKey(signaturePage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListDocumentVersionSignaturesOutput{
		NextCursor:                nextCursor,
		DocumentVersionSignatures: signatures,
	}
}
