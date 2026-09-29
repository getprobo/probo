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

type AuditService struct {
	svc *Service
}

type (
	CreateAuditRequest struct {
		OrganizationID gid.GID
		FrameworkIDs   []gid.GID
		Name           *string
		Firm           *string
		ValidFrom      *time.Time
		ValidUntil     *time.Time
		AuditStartDate *time.Time
		AuditEndDate   *time.Time
		State          *coredata.AuditState
	}

	LinkAuditFrameworkRequest struct {
		AuditID     gid.GID
		FrameworkID gid.GID
	}

	UpdateAuditRequest struct {
		ID             gid.GID
		Name           **string
		Firm           **string
		ValidFrom      *time.Time
		ValidUntil     *time.Time
		AuditStartDate *time.Time
		AuditEndDate   *time.Time
		State          *coredata.AuditState
	}

	UploadAuditReportRequest struct {
		AuditID gid.GID
		File    File
	}
)

func (car *CreateAuditRequest) Validate() error {
	v := validator.New()

	v.Check(car.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(car.FrameworkIDs, "framework_ids", validator.Required(), validator.NoDuplicates())
	v.CheckEach(car.FrameworkIDs, "framework_ids", func(index int, item any) {
		v.Check(item, fmt.Sprintf("framework_ids[%d]", index), validator.Required(), validator.GID(coredata.FrameworkEntityType))
	})
	v.Check(car.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(car.Firm, "firm", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(car.ValidUntil, "valid_until", validator.After(car.ValidFrom))
	v.Check(car.AuditEndDate, "audit_end_date", validator.After(car.AuditStartDate))
	v.Check(car.State, "state", validator.OneOfSlice(coredata.AuditStates()))

	return v.Error()
}

func (uar *UpdateAuditRequest) Validate() error {
	v := validator.New()

	v.Check(uar.ID, "id", validator.Required(), validator.GID(coredata.AuditEntityType))
	v.Check(uar.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uar.Firm, "firm", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uar.ValidUntil, "valid_until", validator.After(uar.ValidFrom))
	v.Check(uar.AuditEndDate, "audit_end_date", validator.After(uar.AuditStartDate))
	v.Check(uar.State, "state", validator.OneOfSlice(coredata.AuditStates()))

	return v.Error()
}

func (uarr *UploadAuditReportRequest) Validate() error {
	v := validator.New()

	v.Check(uarr.AuditID, "audit_id", validator.Required(), validator.GID(coredata.AuditEntityType))

	if err := v.Error(); err != nil {
		return err
	}

	fv := filevalidation.NewValidator(
		filevalidation.WithCategories(filevalidation.CategoryDocument),
		filevalidation.WithMaxFileSize(25*1024*1024),
	)
	if err := fv.Validate(uarr.File.Filename, uarr.File.ContentType, uarr.File.Size); err != nil {
		return fmt.Errorf("invalid audit report file: %w", err)
	}

	return nil
}

func (r *LinkAuditFrameworkRequest) Validate() error {
	v := validator.New()

	v.Check(r.AuditID, "audit_id", validator.Required(), validator.GID(coredata.AuditEntityType))
	v.Check(r.FrameworkID, "framework_id", validator.Required(), validator.GID(coredata.FrameworkEntityType))

	return v.Error()
}

func (s AuditService) Get(
	ctx context.Context, scope coredata.Scoper,
	auditID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := audit.LoadByID(ctx, conn, scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s AuditService) GetByReportFileID(
	ctx context.Context, scope coredata.Scoper,
	fileID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := audit.LoadByReportFileID(ctx, conn, scope, fileID); err != nil {
				return fmt.Errorf("cannot load report file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s *AuditService) Create(
	ctx context.Context, scope coredata.Scoper,
	req *CreateAuditRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	audit := &coredata.Audit{
		ID:             gid.New(scope.GetTenantID(), coredata.AuditEntityType),
		Name:           req.Name,
		Firm:           req.Firm,
		OrganizationID: req.OrganizationID,
		ValidFrom:      req.ValidFrom,
		ValidUntil:     req.ValidUntil,
		AuditStartDate: req.AuditStartDate,
		AuditEndDate:   req.AuditEndDate,
		State:          coredata.AuditStateNotStarted,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if req.State != nil {
		audit.State = *req.State
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			for _, frameworkID := range req.FrameworkIDs {
				framework := &coredata.Framework{}
				if err := framework.LoadByID(ctx, conn, scope, frameworkID); err != nil {
					return fmt.Errorf("cannot load framework: %w", err)
				}

				if framework.OrganizationID != organization.ID {
					return fmt.Errorf("framework %q: %w", frameworkID, coredata.ErrResourceNotFound)
				}
			}

			if err := audit.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert audit: %w", err)
			}

			for _, frameworkID := range req.FrameworkIDs {
				link := &coredata.AuditFramework{
					AuditID:        audit.ID,
					FrameworkID:    frameworkID,
					OrganizationID: organization.ID,
					CreatedAt:      now,
				}
				if err := link.Insert(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot link framework: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s *AuditService) LinkFramework(
	ctx context.Context,
	scope coredata.Scoper,
	req *LinkAuditFrameworkRequest,
) (*coredata.Audit, *coredata.Framework, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid request: %w", err)
	}

	audit := &coredata.Audit{}
	framework := &coredata.Framework{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := audit.LoadByID(ctx, conn, scope, req.AuditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			if err := framework.LoadByID(ctx, conn, scope, req.FrameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if framework.OrganizationID != audit.OrganizationID {
				return fmt.Errorf("framework %q: %w", req.FrameworkID, coredata.ErrResourceNotFound)
			}

			link := &coredata.AuditFramework{
				AuditID:        audit.ID,
				FrameworkID:    framework.ID,
				OrganizationID: audit.OrganizationID,
				CreatedAt:      time.Now(),
			}
			if err := link.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot link framework: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return audit, framework, nil
}

func (s *AuditService) UnlinkFramework(
	ctx context.Context,
	scope coredata.Scoper,
	req *LinkAuditFrameworkRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	audit := &coredata.Audit{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := audit.LoadByID(ctx, conn, scope, req.AuditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, scope, req.FrameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if framework.OrganizationID != audit.OrganizationID {
				return fmt.Errorf("framework %q: %w", req.FrameworkID, coredata.ErrResourceNotFound)
			}

			link := &coredata.AuditFramework{
				AuditID:     audit.ID,
				FrameworkID: framework.ID,
			}
			if err := link.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot unlink framework: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s AuditService) ListFrameworksForAuditID(
	ctx context.Context,
	scope coredata.Scoper,
	auditID gid.GID,
	cursor *page.Cursor[coredata.FrameworkOrderField],
) (*page.Page[*coredata.Framework, coredata.FrameworkOrderField], error) {
	var frameworks coredata.Frameworks

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			audit := &coredata.Audit{}
			if err := audit.LoadByID(ctx, conn, scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			if err := frameworks.LoadByAuditID(ctx, conn, scope, audit.ID, cursor); err != nil {
				return fmt.Errorf("cannot load frameworks: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(frameworks, cursor), nil
}

func (s AuditService) CountFrameworksForAuditID(
	ctx context.Context,
	scope coredata.Scoper,
	auditID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			frameworks := coredata.Frameworks{}

			count, err = frameworks.CountByAuditID(ctx, conn, scope, auditID)
			if err != nil {
				return fmt.Errorf("cannot count frameworks: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s AuditService) ListFrameworkIDsByAuditIDs(
	ctx context.Context,
	scope coredata.Scoper,
	auditIDs []gid.GID,
) (map[gid.GID][]gid.GID, error) {
	frameworkIDs := make(map[gid.GID][]gid.GID, len(auditIDs))
	for _, auditID := range auditIDs {
		frameworkIDs[auditID] = []gid.GID{}
	}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var links coredata.AuditFrameworks
			if err := links.LoadByAuditIDs(ctx, conn, scope, auditIDs); err != nil {
				return fmt.Errorf("cannot load audit frameworks: %w", err)
			}

			for _, link := range links {
				frameworkIDs[link.AuditID] = append(frameworkIDs[link.AuditID], link.FrameworkID)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return frameworkIDs, nil
}

func (s *AuditService) Update(
	ctx context.Context, scope coredata.Scoper,
	req *UpdateAuditRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	audit := &coredata.Audit{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := audit.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			if req.Name != nil {
				audit.Name = *req.Name
			}

			if req.Firm != nil {
				audit.Firm = *req.Firm
			}

			if req.ValidFrom != nil {
				audit.ValidFrom = req.ValidFrom
			}

			if req.ValidUntil != nil {
				audit.ValidUntil = req.ValidUntil
			}

			if req.AuditStartDate != nil {
				audit.AuditStartDate = req.AuditStartDate
			}

			if req.AuditEndDate != nil {
				audit.AuditEndDate = req.AuditEndDate
			}

			if req.State != nil {
				audit.State = *req.State
			}

			audit.UpdatedAt = time.Now()

			if err := audit.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update audit: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s AuditService) Delete(
	ctx context.Context, scope coredata.Scoper,
	auditID gid.GID,
) error {
	audit := coredata.Audit{ID: auditID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			err := audit.Delete(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot delete audit: %w", err)
			}

			return nil
		},
	)
}

func (s AuditService) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			filter := coredata.NewAuditFilter()

			err := audits.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load audits: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(audits, cursor), nil
}

func (s AuditService) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			audits := coredata.Audits{}

			count, err = audits.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count audits: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s AuditService) UploadReport(
	ctx context.Context, scope coredata.Scoper,
	req *UploadAuditReportRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := audit.LoadByID(ctx, conn, scope, req.AuditID)
			if err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	fv := filevalidation.NewValidator(
		filevalidation.WithCategories(filevalidation.CategoryDocument),
		filevalidation.WithMaxFileSize(25*1024*1024),
	)

	file, err := s.svc.Files.UploadAndSaveFile(
		ctx, scope,
		fv,
		map[string]string{"organization-id": audit.OrganizationID.String()},
		&FileUpload{
			Content:     req.File.Content,
			Filename:    req.File.Filename,
			Size:        req.File.Size,
			ContentType: req.File.ContentType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot upload file: %w", err)
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := audit.LoadByID(ctx, conn, scope, req.AuditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			audit.ReportFileID = &file.ID
			audit.UpdatedAt = time.Now()

			return audit.Update(ctx, conn, scope)
		},
	)
	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s AuditService) GenerateReportURL(
	ctx context.Context, scope coredata.Scoper,
	auditID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	audit, err := s.Get(ctx, scope, auditID)
	if err != nil {
		return nil, fmt.Errorf("cannot get audit: %w", err)
	}

	if audit.ReportFileID == nil {
		return nil, fmt.Errorf("audit has no report")
	}

	url, err := s.svc.Files.GenerateFileURL(ctx, scope, *audit.ReportFileID, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate report download URL: %w", err)
	}

	return &url, nil
}

func (s AuditService) DeleteReport(
	ctx context.Context, scope coredata.Scoper,
	auditID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	return audit, s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := audit.LoadByID(ctx, conn, scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			if audit.ReportFileID != nil {
				file := coredata.File{ID: *audit.ReportFileID}

				if err := file.SoftDelete(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot soft-delete report file: %w", err)
				}

				audit.ReportFileID = nil
				audit.UpdatedAt = time.Now()

				if err := audit.Update(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot update audit: %w", err)
				}
			}

			return nil
		},
	)
}

func (s AuditService) ListForControlID(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits

	control := &coredata.Control{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := control.LoadByID(ctx, conn, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			err := audits.LoadByControlID(ctx, conn, scope, control.ID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load audits: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(audits, cursor), nil
}

func (s AuditService) CountForControlID(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			audits := coredata.Audits{}

			count, err = audits.CountByControlID(ctx, conn, scope, controlID)
			if err != nil {
				return fmt.Errorf("cannot count audits: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s AuditService) CountForFindingID(
	ctx context.Context, scope coredata.Scoper,
	findingID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			audits := coredata.Audits{}

			count, err = audits.CountByFindingID(ctx, conn, scope, findingID)
			if err != nil {
				return fmt.Errorf("cannot count audits: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s AuditService) ListForFindingID(
	ctx context.Context, scope coredata.Scoper,
	findingID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits

	finding := &coredata.Finding{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := finding.LoadByID(ctx, conn, scope, findingID); err != nil {
				return fmt.Errorf("cannot load finding: %w", err)
			}

			err := audits.LoadByFindingID(ctx, conn, scope, finding.ID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load audits: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(audits, cursor), nil
}

func (s AuditService) ListForFrameworkID(
	ctx context.Context,
	scope coredata.Scoper,
	frameworkID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, scope, frameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if err := audits.LoadByFrameworkID(ctx, conn, scope, framework.ID, cursor); err != nil {
				return fmt.Errorf("cannot load audits: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(audits, cursor), nil
}

func (s AuditService) CountForFrameworkID(
	ctx context.Context,
	scope coredata.Scoper,
	frameworkID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			audits := coredata.Audits{}

			count, err = audits.CountByFrameworkID(ctx, conn, scope, frameworkID)
			if err != nil {
				return fmt.Errorf("cannot count audits: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
