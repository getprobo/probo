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
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/pdfutils"
)

func (s *Service) GetReport(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	fileID gid.GID,
) (*coredata.File, error) {
	file := &coredata.File{}

	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			_, _, err := loadAuditByReportFileID(ctx, conn, scope, compliancePortalID, fileID)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) || errors.Is(err, ErrReportNotFound) {
					return ErrReportNotFound
				}

				return fmt.Errorf("cannot verify report file: %w", err)
			}

			if err := file.LoadActiveByID(ctx, conn, scope, fileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return file, nil
}

func (s *Service) loadReportByID(
	ctx context.Context,
	scope coredata.Scoper,
	fileID gid.GID,
) (*coredata.File, error) {
	file := &coredata.File{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadActiveByID(ctx, conn, scope, fileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *Service) ExportReportPDF(
	ctx context.Context,
	scope coredata.Scoper,
	reportID gid.GID,
	email mail.Addr,
) ([]byte, error) {
	pdfData, err := s.exportReportPDFData(ctx, scope, reportID)
	if err != nil {
		return nil, fmt.Errorf("cannot export report PDF: %w", err)
	}

	watermarkedPDF, err := pdfutils.AddConfidentialWithTimestamp(pdfData, email.String())
	if err != nil {
		return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
	}

	return watermarkedPDF, nil
}

func (s *Service) ExportReportPDFWithoutWatermark(
	ctx context.Context,
	scope coredata.Scoper,
	reportID gid.GID,
) ([]byte, error) {
	return s.exportReportPDFData(ctx, scope, reportID)
}

func (s *Service) exportReportPDFData(
	ctx context.Context,
	scope coredata.Scoper,
	fileID gid.GID,
) ([]byte, error) {
	file, err := s.loadReportByID(ctx, scope, fileID)
	if err != nil {
		return nil, fmt.Errorf("cannot get file: %w", err)
	}

	result, err := s.s3.GetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: new(s.bucket),
			Key:    &file.FileKey,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot download PDF from S3: %w", err)
	}

	defer func() { _ = result.Body.Close() }()

	pdfData, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read PDF data: %w", err)
	}

	return pdfData, nil
}
