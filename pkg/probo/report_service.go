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

type ReportService struct {
	svc *TenantService
}

type (
	CreateReportRequest struct {
		OrganizationID        gid.GID
		FrameworkID           gid.GID
		Name                  *string
		FrameworkType         *string
		ValidFrom             *time.Time
		ValidUntil            *time.Time
		State                 *coredata.ReportState
		TrustCenterVisibility *coredata.TrustCenterVisibility
	}

	UpdateReportRequest struct {
		ID                    gid.GID
		Name                  **string
		FrameworkType         **string
		ValidFrom             *time.Time
		ValidUntil            *time.Time
		State                 *coredata.ReportState
		TrustCenterVisibility *coredata.TrustCenterVisibility
	}

	UploadReportFileRequest struct {
		ReportID gid.GID
		File     File
	}
)

func (crr *CreateReportRequest) Validate() error {
	v := validator.New()

	v.Check(crr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(crr.FrameworkID, "framework_id", validator.Required(), validator.GID(coredata.FrameworkEntityType))
	v.Check(crr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(crr.FrameworkType, "framework_type", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(crr.ValidUntil, "valid_until", validator.After(crr.ValidFrom))
	v.Check(crr.State, "state", validator.OneOfSlice(coredata.ReportStates()))
	v.Check(crr.TrustCenterVisibility, "trust_center_visibility", validator.OneOfSlice(coredata.TrustCenterVisibilities()))

	return v.Error()
}

func (urr *UpdateReportRequest) Validate() error {
	v := validator.New()

	v.Check(urr.ID, "id", validator.Required(), validator.GID(coredata.ReportEntityType))
	v.Check(urr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(urr.FrameworkType, "framework_type", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(urr.ValidUntil, "valid_until", validator.After(urr.ValidFrom))
	v.Check(urr.State, "state", validator.OneOfSlice(coredata.ReportStates()))
	v.Check(urr.TrustCenterVisibility, "trust_center_visibility", validator.OneOfSlice(coredata.TrustCenterVisibilities()))

	return v.Error()
}

func (urfr *UploadReportFileRequest) Validate() error {
	v := validator.New()

	v.Check(urfr.ReportID, "report_id", validator.Required(), validator.GID(coredata.ReportEntityType))
	if err := v.Error(); err != nil {
		return err
	}

	fv := filevalidation.NewValidator(
		filevalidation.WithCategories(filevalidation.CategoryDocument),
		filevalidation.WithMaxFileSize(25*1024*1024),
	)
	if err := fv.Validate(urfr.File.Filename, urfr.File.ContentType, urfr.File.Size); err != nil {
		return fmt.Errorf("invalid report file: %w", err)
	}

	return nil
}

func (s ReportService) Get(
	ctx context.Context,
	reportID gid.GID,
) (*coredata.Report, error) {
	report := &coredata.Report{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return report.LoadByID(ctx, conn, s.svc.scope, reportID)
		},
	)

	if err != nil {
		return nil, err
	}

	return report, nil
}

func (s ReportService) GetByFileID(
	ctx context.Context,
	fileID gid.GID,
) (*coredata.Report, error) {
	report := &coredata.Report{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return report.LoadByFileID(ctx, conn, s.svc.scope, fileID)
		},
	)

	if err != nil {
		return nil, err
	}

	return report, nil
}

func (s *ReportService) Create(
	ctx context.Context,
	req *CreateReportRequest,
) (*coredata.Report, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	report := &coredata.Report{
		ID:                    gid.New(s.svc.scope.GetTenantID(), coredata.ReportEntityType),
		Name:                  req.Name,
		OrganizationID:        req.OrganizationID,
		FrameworkID:           req.FrameworkID,
		FrameworkType:         req.FrameworkType,
		ValidFrom:             req.ValidFrom,
		ValidUntil:            req.ValidUntil,
		State:                 coredata.ReportStateNotStarted,
		TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if req.State != nil {
		report.State = *req.State
	}

	if req.TrustCenterVisibility != nil {
		report.TrustCenterVisibility = *req.TrustCenterVisibility
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, s.svc.scope, req.FrameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if err := report.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert report: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return report, nil
}

func (s *ReportService) Update(
	ctx context.Context,
	req *UpdateReportRequest,
) (*coredata.Report, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	report := &coredata.Report{}
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := report.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load report: %w", err)
			}

			if req.Name != nil {
				report.Name = *req.Name
			}
			if req.FrameworkType != nil {
				report.FrameworkType = *req.FrameworkType
			}
			if req.ValidFrom != nil {
				report.ValidFrom = req.ValidFrom
			}
			if req.ValidUntil != nil {
				report.ValidUntil = req.ValidUntil
			}
			if req.State != nil {
				report.State = *req.State
			}
			if req.TrustCenterVisibility != nil {
				report.TrustCenterVisibility = *req.TrustCenterVisibility
			}

			report.UpdatedAt = time.Now()

			if err := report.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update report: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return report, nil
}

func (s ReportService) Delete(
	ctx context.Context,
	reportID gid.GID,
) error {
	report := coredata.Report{ID: reportID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := report.Delete(ctx, conn, s.svc.scope)
			if err != nil {
				return fmt.Errorf("cannot delete report: %w", err)
			}
			return nil
		},
	)
}

func (s ReportService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ReportOrderField],
) (*page.Page[*coredata.Report, coredata.ReportOrderField], error) {
	var reports coredata.Reports

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			filter := coredata.NewReportFilter()
			err := reports.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load reports: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(reports, cursor), nil
}

func (s ReportService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			reports := coredata.Reports{}
			count, err = reports.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count reports: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ReportService) UploadFile(
	ctx context.Context,
	req UploadReportFileRequest,
) (*coredata.Report, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	report, err := s.Get(ctx, req.ReportID)
	if err != nil {
		return nil, fmt.Errorf("cannot load report: %w", err)
	}

	file, err := s.svc.Files.UploadAndSaveFile(
		ctx,
		filevalidation.NewValidator(
			filevalidation.WithCategories(filevalidation.CategoryDocument),
			filevalidation.WithMaxFileSize(25*1024*1024),
		),
		map[string]string{
			"type":            "report",
			"organization-id": report.OrganizationID.String(),
		},
		&FileUpload{
			Content:     req.File.Content,
			Filename:    req.File.Filename,
			Size:        req.File.Size,
			ContentType: req.File.ContentType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot upload report file: %w", err)
	}

	report.FileID = &file.ID
	report.UpdatedAt = time.Now()

	err = s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		return report.Update(ctx, conn, s.svc.scope)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot update report: %w", err)
	}

	return report, nil
}

func (s ReportService) GenerateFileURL(
	ctx context.Context,
	reportID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	report, err := s.Get(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("cannot get report: %w", err)
	}

	if report.FileID == nil {
		return nil, fmt.Errorf("report has no file")
	}

	url, err := s.svc.Files.GenerateFileTempURL(ctx, *report.FileID, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file download URL: %w", err)
	}

	return &url, nil
}

func (s ReportService) DeleteFile(
	ctx context.Context,
	reportID gid.GID,
) (*coredata.Report, error) {
	report := &coredata.Report{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := report.LoadByID(ctx, conn, s.svc.scope, reportID); err != nil {
				return fmt.Errorf("cannot load report: %w", err)
			}

			if report.FileID != nil {
				file := coredata.File{ID: *report.FileID}

				if err := file.SoftDelete(ctx, conn, s.svc.scope); err != nil {
					return fmt.Errorf("cannot delete report file: %w", err)
				}

				report.FileID = nil
				report.UpdatedAt = time.Now()

				if err := report.Update(ctx, conn, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update report: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return report, nil
}

func (s ReportService) ListForControlID(
	ctx context.Context,
	controlID gid.GID,
	cursor *page.Cursor[coredata.ReportOrderField],
) (*page.Page[*coredata.Report, coredata.ReportOrderField], error) {
	var reports coredata.Reports
	control := &coredata.Control{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := control.LoadByID(ctx, conn, s.svc.scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			err := reports.LoadByControlID(ctx, conn, s.svc.scope, control.ID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load reports: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(reports, cursor), nil
}
