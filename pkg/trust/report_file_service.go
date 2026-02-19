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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/watermarkpdf"
)

type ReportFileService struct {
	svc *TenantService
}

func (s ReportFileService) GetFile(
	ctx context.Context,
	fileID gid.GID,
) (*coredata.File, error) {
	file := &coredata.File{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := file.LoadByID(ctx, conn, s.svc.scope, fileID)
			if err != nil {
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

func (s ReportFileService) ExportPDF(
	ctx context.Context,
	fileID gid.GID,
	email mail.Addr,
) ([]byte, error) {
	pdfData, err := s.exportPDFData(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("cannot export report PDF: %w", err)
	}

	watermarkedPDF, err := watermarkpdf.AddConfidentialWithTimestamp(pdfData, email)
	if err != nil {
		return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
	}

	return watermarkedPDF, nil
}

func (s ReportFileService) ExportPDFWithoutWatermark(
	ctx context.Context,
	fileID gid.GID,
) ([]byte, error) {
	return s.exportPDFData(ctx, fileID)
}

func (s ReportFileService) exportPDFData(
	ctx context.Context,
	fileID gid.GID,
) ([]byte, error) {
	file, err := s.GetFile(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("cannot get file: %w", err)
	}

	result, err := s.svc.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(file.BucketName),
		Key:    aws.String(file.FileKey),
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
