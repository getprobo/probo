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
		svc           *Service
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
	ctx context.Context, scope coredata.Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyComplianceReportOrderField],
) (*page.Page[*coredata.ThirdPartyComplianceReport, coredata.ThirdPartyComplianceReportOrderField], error) {
	var thirdPartyComplianceReports coredata.ThirdPartyComplianceReports

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdPartyComplianceReports.LoadForThirdPartyID(ctx, conn, scope, thirdPartyID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdPartyComplianceReports, cursor), nil
}

func (s ThirdPartyComplianceReportService) Upload(
	ctx context.Context, scope coredata.Scoper,
	thirdPartyID gid.GID,
	req *ThirdPartyComplianceReportCreateRequest,
) (*coredata.ThirdPartyComplianceReport, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	thirdParty, err := s.svc.ThirdParties.Get(ctx, scope, thirdPartyID)
	if err != nil {
		return nil, fmt.Errorf("cannot get thirdParty: %w", err)
	}

	f, err := s.svc.Files.UploadAndSaveFile(
		ctx,
		scope,
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

	thirdPartyComplianceReportID := gid.New(scope.GetTenantID(), coredata.ThirdPartyComplianceReportEntityType)

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
			return thirdPartyComplianceReport.Insert(ctx, tx, scope)
		},
	)
	if err != nil {
		return nil, err
	}

	return thirdPartyComplianceReport, nil
}

func (s ThirdPartyComplianceReportService) Get(
	ctx context.Context, scope coredata.Scoper,
	thirdPartyComplianceReportID gid.GID,
) (*coredata.ThirdPartyComplianceReport, error) {
	thirdPartyComplianceReport := &coredata.ThirdPartyComplianceReport{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdPartyComplianceReport.LoadByID(ctx, conn, scope, thirdPartyComplianceReportID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load thirdParty compliance report: %w", err)
	}

	return thirdPartyComplianceReport, nil
}

func (s ThirdPartyComplianceReportService) Delete(
	ctx context.Context, scope coredata.Scoper,
	thirdPartyComplianceReportID gid.GID,
) error {
	thirdPartyComplianceReport := &coredata.ThirdPartyComplianceReport{ID: thirdPartyComplianceReportID}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := thirdPartyComplianceReport.Delete(ctx, tx, scope); err != nil {
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
