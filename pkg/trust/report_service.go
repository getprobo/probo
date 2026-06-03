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
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/pdfutils"
)

type ReportService struct {
	svc *Service
}

func (s ReportService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	fileID gid.GID,
) (*coredata.File, error) {
	file, err := s.loadByID(ctx, scope, fileID)
	if err != nil {
		return nil, err
	}

	if file.OrganizationID != organizationID {
		return nil, ErrReportNotFound
	}

	// check the given report file ID is linked to an audit in order to avoid
	// being able to get any file from the report request.
	_, err = s.svc.Audits.GetByReportFileID(ctx, scope, fileID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrReportNotFound
		}

		return nil, fmt.Errorf("cannot verify report file: %w", err)
	}

	return file, nil
}

func (s ReportService) loadByID(
	ctx context.Context,
	scope coredata.Scoper,
	fileID gid.GID,
) (*coredata.File, error) {
	file := &coredata.File{}

	err := s.svc.pg.WithConn(
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

func (s ReportService) GenerateDownloadURL(
	ctx context.Context,
	scope coredata.Scoper,
	fileID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file, err := s.loadByID(ctx, scope, fileID)
	if err != nil {
		return nil, fmt.Errorf("cannot get file: %w", err)
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     new(s.svc.bucket),
		Key:                        new(file.FileKey),
		ResponseCacheControl:       new("max-age=3600, public"),
		ResponseContentType:        new(file.MimeType),
		ResponseContentDisposition: new(fmt.Sprintf("attachment; filename=\"%s\"", file.FileName)),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return nil, fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return &presignedReq.URL, nil
}

func (s ReportService) ExportPDF(
	ctx context.Context,
	scope coredata.Scoper,
	reportID gid.GID,
	email mail.Addr,
) ([]byte, error) {
	pdfData, err := s.exportPDFData(ctx, scope, reportID)
	if err != nil {
		return nil, fmt.Errorf("cannot export report PDF: %w", err)
	}

	watermarkedPDF, err := pdfutils.AddConfidentialWithTimestamp(pdfData, email)
	if err != nil {
		return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
	}

	return watermarkedPDF, nil
}

func (s ReportService) ExportPDFWithoutWatermark(
	ctx context.Context,
	scope coredata.Scoper,
	reportID gid.GID,
) ([]byte, error) {
	return s.exportPDFData(ctx, scope, reportID)
}

func (s ReportService) exportPDFData(
	ctx context.Context,
	scope coredata.Scoper,
	fileID gid.GID,
) ([]byte, error) {
	file, err := s.loadByID(ctx, scope, fileID)
	if err != nil {
		return nil, fmt.Errorf("cannot get file: %w", err)
	}

	result, err := s.svc.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: new(s.svc.bucket),
		Key:    new(file.FileKey),
	})
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
