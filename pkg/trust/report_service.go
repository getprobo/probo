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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/watermarkpdf"
)

type ReportService struct {
	svc *TenantService
}

func (s ReportService) Get(
	ctx context.Context,
	reportID gid.GID,
) (*coredata.Report, error) {
	report := &coredata.Report{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := report.LoadByID(ctx, conn, s.svc.scope, reportID)
			if err != nil {
				return fmt.Errorf("cannot load report: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return report, nil
}

func (s ReportService) GenerateDownloadURL(
	ctx context.Context,
	reportID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	report, err := s.Get(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("cannot get report: %w", err)
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(report.ObjectKey),
		ResponseCacheControl:       aws.String("max-age=3600, public"),
		ResponseContentType:        aws.String(report.MimeType),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", report.Filename)),
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
	reportID gid.GID,
	email mail.Addr,
) ([]byte, error) {
	pdfData, err := s.exportPDFData(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("cannot export report PDF: %w", err)
	}

	watermarkedPDF, err := watermarkpdf.AddConfidentialWithTimestamp(pdfData, email)
	if err != nil {
		return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
	}

	return watermarkedPDF, nil
}

func (s ReportService) ExportPDFWithoutWatermark(
	ctx context.Context,
	reportID gid.GID,
) ([]byte, error) {
	return s.exportPDFData(ctx, reportID)
}

func (s ReportService) exportPDFData(
	ctx context.Context,
	reportID gid.GID,
) ([]byte, error) {
	report, err := s.Get(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("cannot get report: %w", err)
	}

	result, err := s.svc.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(report.ObjectKey),
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
