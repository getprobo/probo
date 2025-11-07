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
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	VendorComplianceReportService struct {
		svc           *TenantService
		fileValidator *filevalidation.FileValidator
	}

	VendorComplianceReportCreateRequest struct {
		File       FileUpload
		ReportDate time.Time
		ValidUntil *time.Time
		ReportName string
	}
)

func (vcrcr *VendorComplianceReportCreateRequest) Validate() error {
	v := validator.New()

	v.Check(vcrcr.ReportName, "report_name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

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
	if err := req.Validate(); err != nil {
		return nil, err
	}

	vendor, err := s.svc.Vendors.Get(ctx, vendorID)
	if err != nil {
		return nil, fmt.Errorf("cannot get vendor: %w", err)
	}

	f, err := s.svc.Files.UploadAndSaveFile(
		ctx,
		s.fileValidator,
		map[string]string{
			"type":            "vendor-compliance-report",
			"vendor-id":       vendorID.String(),
			"organization-id": vendor.OrganizationID.String(),
		},
		&req.File)

	if err != nil {
		return nil, err
	}

	now := time.Now()

	vendorComplianceReportID := gid.New(s.svc.scope.GetTenantID(), coredata.VendorComplianceReportEntityType)

	vendorComplianceReport := &coredata.VendorComplianceReport{
		ID:             vendorComplianceReportID,
		OrganizationID: vendor.OrganizationID,
		VendorID:       vendorID,
		ReportDate:     req.ReportDate,
		ValidUntil:     req.ValidUntil,
		ReportName:     req.ReportName,
		ReportFileId:   &f.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
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

func (s VendorComplianceReportService) Delete(
	ctx context.Context,
	vendorComplianceReportID gid.GID,
) error {
	vendorComplianceReport := &coredata.VendorComplianceReport{ID: vendorComplianceReportID}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := vendorComplianceReport.Delete(ctx, conn, s.svc.scope); err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		return fmt.Errorf("cannot delete vendor compliance report: %w", err)
	}

	return nil
}
