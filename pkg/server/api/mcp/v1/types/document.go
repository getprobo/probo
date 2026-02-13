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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func NewDocument(d *coredata.Document, approverIDs []gid.GID) *Document {
	return &Document{
		ID:                      d.ID,
		OrganizationID:          d.OrganizationID,
		ApproverIds:             approverIDs,
		Title:                   d.Title,
		DocumentType:            d.DocumentType,
		Classification:          d.Classification,
		CurrentPublishedVersion: d.CurrentPublishedVersion,
		TrustCenterVisibility:   d.TrustCenterVisibility,
		CreatedAt:               d.CreatedAt,
		UpdatedAt:               d.UpdatedAt,
	}
}

func NewListDocumentsOutput(documentPage *page.Page[*coredata.Document, coredata.DocumentOrderField], approverIDsMap map[gid.GID][]gid.GID) ListDocumentsOutput {
	documents := make([]*Document, 0, len(documentPage.Data))
	for _, d := range documentPage.Data {
		documents = append(documents, NewDocument(d, approverIDsMap[d.ID]))
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

func NewAddDocumentOutput(doc *coredata.Document, docVersion *coredata.DocumentVersion, docApproverIDs []gid.GID, versionApproverIDs []gid.GID) AddDocumentOutput {
	return AddDocumentOutput{
		Document:        NewDocument(doc, docApproverIDs),
		DocumentVersion: NewDocumentVersion(docVersion, versionApproverIDs),
	}
}

func NewDocumentVersion(dv *coredata.DocumentVersion, approverIDs []gid.GID) *DocumentVersion {
	return &DocumentVersion{
		ID:             dv.ID,
		OrganizationID: dv.OrganizationID,
		DocumentID:     dv.DocumentID,
		Title:          dv.Title,
		ApproverIds:    approverIDs,
		VersionNumber:  dv.VersionNumber,
		Classification: dv.Classification,
		Content:        dv.Content,
		Changelog:      dv.Changelog,
		Status:         dv.Status,
		PublishedAt:    dv.PublishedAt,
		CreatedAt:      dv.CreatedAt,
		UpdatedAt:      dv.UpdatedAt,
	}
}

func NewListDocumentVersionsOutput(versionPage *page.Page[*coredata.DocumentVersion, coredata.DocumentVersionOrderField], approverIDsMap map[gid.GID][]gid.GID) ListDocumentVersionsOutput {
	versions := make([]*DocumentVersion, 0, len(versionPage.Data))
	for _, v := range versionPage.Data {
		versions = append(versions, NewDocumentVersion(v, approverIDsMap[v.ID]))
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
