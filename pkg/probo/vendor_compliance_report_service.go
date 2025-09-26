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

package probo

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
)

type (
	VendorComplianceReportService struct {
		svc *TenantService
	}

	VendorComplianceReportCreateRequest struct {
		File       io.Reader
		ReportDate time.Time
		ValidUntil *time.Time
		ReportName string
	}
)

func (s VendorComplianceReportService) ListForVendorID(
	ctx context.Context,
	vendorID gid.GID,
	cursor *page.Cursor[coredata.VendorComplianceReportOrderField],
) (*page.Page[*coredata.VendorComplianceReport, coredata.VendorComplianceReportOrderField], error) {
	var vendorComplianceReports coredata.VendorComplianceReports

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendorComplianceReports.LoadForVendorID(ctx, conn, s.svc.scope, vendorID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(vendorComplianceReports, cursor), nil
}

func (s VendorComplianceReportService) Upload(
	ctx context.Context,
	vendorID gid.GID,
	req *VendorComplianceReportCreateRequest,
) (*coredata.VendorComplianceReport, error) {
	objectKey, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("cannot generate object key: %w", err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(req.ReportName))

	_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.svc.bucket,
		Key:         aws.String(objectKey.String()),
		Body:        req.File,
		ContentType: &mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot upload file to S3: %w", err)
	}

	headOutput, err := s.svc.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(objectKey.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get object metadata: %w", err)
	}

	now := time.Now()

	vendorComplianceReportID := gid.New(s.svc.scope.GetTenantID(), coredata.VendorComplianceReportEntityType)

	vendorComplianceReport := &coredata.VendorComplianceReport{
		ID:         vendorComplianceReportID,
		VendorID:   vendorID,
		ReportDate: req.ReportDate,
		ValidUntil: req.ValidUntil,
		ReportName: req.ReportName,
		FileKey:    objectKey.String(),
		FileSize:   *headOutput.ContentLength,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err = s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendorComplianceReport.Insert(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, err
	}

	return vendorComplianceReport, nil
}

func (s VendorComplianceReportService) Get(
	ctx context.Context,
	vendorComplianceReportID gid.GID,
) (*coredata.VendorComplianceReport, error) {
	vendorComplianceReport := &coredata.VendorComplianceReport{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendorComplianceReport.LoadByID(ctx, conn, s.svc.scope, vendorComplianceReportID)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot load vendor compliance report: %w", err)
	}

	return vendorComplianceReport, nil
}

func (s VendorComplianceReportService) GenerateFileURL(
	ctx context.Context,
	vendorComplianceReportID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	vendorComplianceReport, err := s.Get(ctx, vendorComplianceReportID)
	if err != nil {
		return "", fmt.Errorf("cannot get vendor compliance report: %w", err)
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	// Use RFC 6266/5987 encoding for filename with UTF-8 support
	encodedFilename := url.QueryEscape(vendorComplianceReport.ReportName)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(vendorComplianceReport.FileKey),
		ResponseCacheControl:       aws.String("max-age=3600, public"),
		ResponseContentDisposition: aws.String(contentDisposition),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}

func (s VendorComplianceReportService) Delete(
	ctx context.Context,
	vendorComplianceReportID gid.GID,
) error {
	vendorComplianceReport := &coredata.VendorComplianceReport{ID: vendorComplianceReportID}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendorComplianceReport.Delete(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return fmt.Errorf("cannot delete vendor compliance report: %w", err)
	}

	return nil
}
