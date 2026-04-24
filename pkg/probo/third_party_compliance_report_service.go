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
	ThirdPartyComplianceReportService struct {
		svc           *TenantService
		fileValidator *filevalidation.FileValidator
	}

	ThirdPartyComplianceReportCreateRequest struct {
		File       FileUpload
		ReportDate time.Time
		ValidUntil *time.Time
		ReportName string
	}
)

func (vcrcr *ThirdPartyComplianceReportCreateRequest) Validate() error {
	v := validator.New()

	v.Check(vcrcr.ReportName, "report_name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (s ThirdPartyComplianceReportService) ListForThirdPartyID(
	ctx context.Context,
	thirdPartyID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyComplianceReportOrderField],
) (*page.Page[*coredata.ThirdPartyComplianceReport, coredata.ThirdPartyComplianceReportOrderField], error) {
	var thirdPartyComplianceReports coredata.ThirdPartyComplianceReports

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdPartyComplianceReports.LoadForThirdPartyID(ctx, conn, s.svc.scope, thirdPartyID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdPartyComplianceReports, cursor), nil
}

func (s ThirdPartyComplianceReportService) Upload(
	ctx context.Context,
	thirdPartyID gid.GID,
	req *ThirdPartyComplianceReportCreateRequest,
) (*coredata.ThirdPartyComplianceReport, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	thirdParty, err := s.svc.ThirdParties.Get(ctx, thirdPartyID)
	if err != nil {
		return nil, fmt.Errorf("cannot get thirdParty: %w", err)
	}

	f, err := s.svc.Files.UploadAndSaveFile(
		ctx,
		s.fileValidator,
		map[string]string{
			"type":            "thirdParty-compliance-report",
			"thirdParty-id":   thirdPartyID.String(),
			"organization-id": thirdParty.OrganizationID.String(),
		},
		&req.File)

	if err != nil {
		return nil, err
	}

	now := time.Now()

	thirdPartyComplianceReportID := gid.New(s.svc.scope.GetTenantID(), coredata.ThirdPartyComplianceReportEntityType)

	thirdPartyComplianceReport := &coredata.ThirdPartyComplianceReport{
		ID:             thirdPartyComplianceReportID,
		OrganizationID: thirdParty.OrganizationID,
		ThirdPartyID:   thirdPartyID,
		ReportDate:     req.ReportDate,
		ValidUntil:     req.ValidUntil,
		ReportName:     req.ReportName,
		ReportFileId:   &f.ID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return thirdPartyComplianceReport.Insert(ctx, tx, s.svc.scope)
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdPartyComplianceReport, nil
}

func (s ThirdPartyComplianceReportService) Get(
	ctx context.Context,
	thirdPartyComplianceReportID gid.GID,
) (*coredata.ThirdPartyComplianceReport, error) {
	thirdPartyComplianceReport := &coredata.ThirdPartyComplianceReport{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdPartyComplianceReport.LoadByID(ctx, conn, s.svc.scope, thirdPartyComplianceReportID)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot load thirdParty compliance report: %w", err)
	}

	return thirdPartyComplianceReport, nil
}

func (s ThirdPartyComplianceReportService) Delete(
	ctx context.Context,
	thirdPartyComplianceReportID gid.GID,
) error {
	thirdPartyComplianceReport := &coredata.ThirdPartyComplianceReport{ID: thirdPartyComplianceReportID}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := thirdPartyComplianceReport.Delete(ctx, tx, s.svc.scope); err != nil {
				return err
			}

			return nil
		},
	)

	if err != nil {
		return fmt.Errorf("cannot delete thirdParty compliance report: %w", err)
	}

	return nil
}
