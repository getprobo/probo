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

package trust

import (
	"context"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/pdfutils"
)

type (
	DocumentService struct {
		svc *TenantService
	}

	ErrDocumentArchived struct{}
)

func (e ErrDocumentArchived) Error() string {
	return "cannot access an archived document"
}

func (s *DocumentService) ListForOrganizationId(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			filter := coredata.NewDocumentTrustCenterFilter()

			if err := documents.LoadPublishedByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter); err != nil {
				return fmt.Errorf("cannot load published documents: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documents, cursor), nil
}

func (s *DocumentService) ExportPDF(
	ctx context.Context,
	documentID gid.GID,
	email mail.Addr,
) ([]byte, error) {
	pdfData, err := s.exportPDFData(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("cannot export document PDF: %w", err)
	}

	watermarkedPDF, err := pdfutils.AddConfidentialWithTimestamp(pdfData, email)
	if err != nil {
		return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
	}

	return watermarkedPDF, nil
}

func (s *DocumentService) ExportPDFWithoutWatermark(
	ctx context.Context,
	documentID gid.GID,
) ([]byte, error) {
	return s.exportPDFData(ctx, documentID)
}

func (s DocumentService) Get(
	ctx context.Context,
	organizationID gid.GID,
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := document.LoadByID(ctx, conn, s.svc.scope, documentID)
			if err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	if document.OrganizationID != organizationID {
		return nil, ErrDocumentNotFound
	}

	if document.TrustCenterVisibility == coredata.TrustCenterVisibilityNone {
		return nil, ErrDocumentNotVisible
	}

	return document, nil
}

func (s *DocumentService) exportPDFData(
	ctx context.Context,
	documentID gid.GID,
) ([]byte, error) {
	document := &coredata.Document{}
	version := &coredata.DocumentVersion{}
	fileRecord := &coredata.File{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.ArchivedAt != nil {
				return &ErrDocumentArchived{}
			}

			if document.TrustCenterVisibility == coredata.TrustCenterVisibilityNone {
				return fmt.Errorf("document not visible on trust center")
			}

			if err := version.LoadLatestPublishedVersion(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load latest published document version: %w", err)
			}

			if version.FileID == nil {
				return fmt.Errorf("cannot export document: publication PDF not yet generated")
			}

			if err := fileRecord.LoadByID(ctx, conn, s.svc.scope, *version.FileID); err != nil {
				return fmt.Errorf("cannot load document version file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	pdfData, err := s.svc.fileManager.GetFileBytes(ctx, fileRecord)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch document PDF file: %w", err)
	}

	return pdfData, nil
}
