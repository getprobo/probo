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

package visitor

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/pdfutils"
)

func (s *Service) GetPortalFile(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	compliancePortalFileID gid.GID,
) (*coredata.CompliancePortalFile, error) {
	compliancePortalFile := &coredata.CompliancePortalFile{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := compliancePortalFile.LoadByID(ctx, conn, scope, compliancePortalFileID)
			if err != nil {
				return fmt.Errorf("cannot load compliance page file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if compliancePortalFile.OrganizationID != organizationID {
		return nil, ErrPortalFileNotFound
	}

	if compliancePortalFile.CompliancePortalVisibility == coredata.CompliancePortalVisibilityNone {
		return nil, ErrPortalFileNotVisible
	}

	return compliancePortalFile, nil
}

func (s *Service) ListPortalFilesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.CompliancePortalFileOrderField],
	filter *coredata.CompliancePortalFileFilter,
) (*page.Page[*coredata.CompliancePortalFile, coredata.CompliancePortalFileOrderField], error) {
	var compliancePortalFiles coredata.CompliancePortalFiles

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := compliancePortalFiles.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load compliance page files: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(compliancePortalFiles, cursor), nil
}

func (s *Service) ExportPortalFile(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalFileID gid.GID,
	email mail.Addr,
) ([]byte, string, error) {
	fileData, mimeType, err := s.exportPortalFileData(ctx, scope, compliancePortalFileID)
	if err != nil {
		return nil, "", fmt.Errorf("cannot export compliance page file: %w", err)
	}

	if mimeType == "application/pdf" {
		watermarkedPDF, err := pdfutils.AddConfidentialWithTimestamp(fileData, email)
		if err != nil {
			return nil, "", fmt.Errorf("cannot add watermark to PDF: %w", err)
		}

		return watermarkedPDF, mimeType, nil
	}

	return fileData, mimeType, nil
}

func (s *Service) ExportPortalFileWithoutWatermark(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalFileID gid.GID,
) ([]byte, string, error) {
	return s.exportPortalFileData(ctx, scope, compliancePortalFileID)
}

func (s *Service) exportPortalFileData(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalFileID gid.GID,
) ([]byte, string, error) {
	var (
		compliancePortalFile *coredata.CompliancePortalFile
		file                 *coredata.File
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			compliancePortalFile = &coredata.CompliancePortalFile{}
			if err := compliancePortalFile.LoadByID(ctx, conn, scope, compliancePortalFileID); err != nil {
				return fmt.Errorf("cannot load compliance page file: %w", err)
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, scope, compliancePortalFile.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, "", err
	}

	result, err := s.s3.GetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: new(s.bucket),
			Key:    new(file.FileKey),
		},
	)
	if err != nil {
		return nil, "", fmt.Errorf("cannot download file from S3: %w", err)
	}

	defer func() { _ = result.Body.Close() }()

	fileData, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, "", fmt.Errorf("cannot read file data: %w", err)
	}

	return fileData, file.MimeType, nil
}
