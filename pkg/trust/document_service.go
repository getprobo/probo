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

package trust

import (
	"context"
	"fmt"
	"io"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/watermarkpdf"
)

type (
	DocumentService struct {
		svc               *TenantService
		html2pdfConverter *html2pdf.Converter
	}
)

// ListVersions lists all versions of a document
func (s *DocumentService) ListVersions(
	ctx context.Context,
	documentID gid.GID,
	cursor *page.Cursor[coredata.DocumentVersionOrderField],
) (*page.Page[*coredata.DocumentVersion, coredata.DocumentVersionOrderField], error) {
	var documentVersions coredata.DocumentVersions

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			filter := coredata.NewDocumentVersionFilter()
			return documentVersions.LoadByDocumentID(ctx, conn, s.svc.scope, documentID, cursor, filter)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(documentVersions, cursor), nil
}

func (s *DocumentService) ListForOrganizationId(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DocumentOrderField],
) (*page.Page[*coredata.Document, coredata.DocumentOrderField], error) {
	var documents coredata.Documents

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			filter := coredata.NewDocumentTrustCenterFilter()
			err := documents.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load documents: %w", err)
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

	watermarkedPDF, err := watermarkpdf.AddConfidentialWithTimestamp(pdfData, email)
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
	documentID gid.GID,
) (*coredata.Document, error) {
	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := document.LoadByID(ctx, conn, s.svc.scope, documentID)
			if err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return document, nil
}

func (s *DocumentService) exportPDFData(
	ctx context.Context,
	documentID gid.GID,
) ([]byte, error) {
	document := &coredata.Document{}
	version := &coredata.DocumentVersion{}
	organization := &coredata.Organization{}
	var approverNames []string

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			if document.TrustCenterVisibility == coredata.TrustCenterVisibilityNone {
				return fmt.Errorf("document not visible on trust center")
			}

			if err := version.LoadLatestPublishedVersion(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load latest published document version: %w", err)
			}

			// Load approvers
			docApprovers := &coredata.DocumentApprovers{}
			if err := docApprovers.LoadByDocumentID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document approvers: %w", err)
			}

			profiles := coredata.MembershipProfiles{}
			if err := profiles.LoadByIDs(ctx, conn, s.svc.scope, docApprovers.ApproverProfileIDs()); err != nil {
				return fmt.Errorf("cannot load document approver profiles: %w", err)
			}

			for _, p := range profiles {
				approverNames = append(approverNames, p.FullName)
			}

			if err := organization.LoadByID(ctx, conn, s.svc.scope, document.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	classification := docgen.ClassificationInternal
	switch document.DocumentType {
	case coredata.DocumentTypePolicy:
		classification = docgen.ClassificationConfidential
	case coredata.DocumentTypeISMS:
		classification = docgen.ClassificationSecret
	}

	horizontalLogoBase64 := ""
	if organization.HorizontalLogoFileID != nil {
		fileRecord := &coredata.File{}
		fileErr := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
			return fileRecord.LoadByID(ctx, conn, s.svc.scope, *organization.HorizontalLogoFileID)
		})
		if fileErr == nil {
			base64Data, mimeType, logoErr := s.svc.fileManager.GetFileBase64(ctx, fileRecord)
			if logoErr == nil {
				horizontalLogoBase64 = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
			}
		}
	}

	docData := docgen.DocumentData{
		Title:                       version.Title,
		Content:                     version.Content,
		Version:                     version.VersionNumber,
		Classification:              classification,
		Approvers:                   approverNames,
		PublishedAt:                 version.PublishedAt,
		CompanyHorizontalLogoBase64: horizontalLogoBase64,
	}

	htmlContent, err := docgen.RenderHTML(docData)
	if err != nil {
		return nil, fmt.Errorf("cannot generate HTML: %w", err)
	}

	cfg := html2pdf.RenderConfig{
		PageFormat:      html2pdf.PageFormatA4,
		Orientation:     html2pdf.OrientationPortrait,
		MarginTop:       html2pdf.NewMarginInches(1.0),
		MarginBottom:    html2pdf.NewMarginInches(1.0),
		MarginLeft:      html2pdf.NewMarginInches(1.0),
		MarginRight:     html2pdf.NewMarginInches(1.0),
		PrintBackground: true,
		Scale:           1.0,
	}

	pdfReader, err := s.html2pdfConverter.GeneratePDF(ctx, htmlContent, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot generate PDF: %w", err)
	}

	pdfData, err := io.ReadAll(pdfReader)
	if err != nil {
		return nil, fmt.Errorf("cannot read PDF data: %w", err)
	}

	return pdfData, nil
}
