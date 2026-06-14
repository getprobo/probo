// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type AuditService struct {
	svc *Service
}

type (
	CreateAuditRequest struct {
		OrganizationID        gid.GID
		FrameworkID           gid.GID
		Name                  *string
		ValidFrom             *time.Time
		ValidUntil            *time.Time
		State                 *coredata.AuditState
		TrustCenterVisibility *coredata.TrustCenterVisibility
	}

	UpdateAuditRequest struct {
		ID                    gid.GID
		Name                  **string
		ValidFrom             *time.Time
		ValidUntil            *time.Time
		State                 *coredata.AuditState
		TrustCenterVisibility *coredata.TrustCenterVisibility
	}

	UploadAuditReportRequest struct {
		AuditID gid.GID
		File    File
	}
)

func (car *CreateAuditRequest) Validate() error {
	v := validator.New()

	v.Check(car.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(car.FrameworkID, "framework_id", validator.Required(), validator.GID(coredata.FrameworkEntityType))
	v.Check(car.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(car.ValidUntil, "valid_until", validator.After(car.ValidFrom))
	v.Check(car.State, "state", validator.OneOfSlice(coredata.AuditStates()))
	v.Check(car.TrustCenterVisibility, "trust_center_visibility", validator.OneOfSlice(coredata.TrustCenterVisibilities()))

	return v.Error()
}

func (uar *UpdateAuditRequest) Validate() error {
	v := validator.New()

	v.Check(uar.ID, "id", validator.Required(), validator.GID(coredata.AuditEntityType))
	v.Check(uar.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uar.ValidUntil, "valid_until", validator.After(uar.ValidFrom))
	v.Check(uar.State, "state", validator.OneOfSlice(coredata.AuditStates()))
	v.Check(uar.TrustCenterVisibility, "trust_center_visibility", validator.OneOfSlice(coredata.TrustCenterVisibilities()))

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

func (s AuditService) Get(
	ctx context.Context, predicate coredata.Predicater,
	auditID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := audit.LoadByID(ctx, conn, predicate, auditID); err != nil {
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
	ctx context.Context, predicate coredata.Predicater,
	fileID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := audit.LoadByReportFileID(ctx, conn, predicate, fileID); err != nil {
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
	ctx context.Context, predicate coredata.Predicater,
	req *CreateAuditRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	audit := &coredata.Audit{
		ID:                    gid.New(predicate.GetTenantID(), coredata.AuditEntityType),
		Name:                  req.Name,
		OrganizationID:        req.OrganizationID,
		FrameworkID:           req.FrameworkID,
		ValidFrom:             req.ValidFrom,
		ValidUntil:            req.ValidUntil,
		State:                 coredata.AuditStateNotStarted,
		TrustCenterVisibility: coredata.TrustCenterVisibilityNone,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if req.State != nil {
		audit.State = *req.State
	}

	if req.TrustCenterVisibility != nil {
		audit.TrustCenterVisibility = *req.TrustCenterVisibility
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, predicate, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, predicate, req.FrameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if err := audit.Insert(ctx, conn, predicate); err != nil {
				return fmt.Errorf("cannot insert audit: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s *AuditService) Update(
	ctx context.Context, predicate coredata.Predicater,
	req *UpdateAuditRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	audit := &coredata.Audit{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := audit.LoadByID(ctx, conn, predicate, req.ID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			if req.Name != nil {
				audit.Name = *req.Name
			}

			if req.ValidFrom != nil {
				audit.ValidFrom = req.ValidFrom
			}

			if req.ValidUntil != nil {
				audit.ValidUntil = req.ValidUntil
			}

			if req.State != nil {
				audit.State = *req.State
			}

			if req.TrustCenterVisibility != nil {
				audit.TrustCenterVisibility = *req.TrustCenterVisibility
			}

			audit.UpdatedAt = time.Now()

			if err := audit.Update(ctx, conn, predicate); err != nil {
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
	ctx context.Context, predicate coredata.Predicater,
	auditID gid.GID,
) error {
	audit := coredata.Audit{ID: auditID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			err := audit.Delete(ctx, tx, predicate)
			if err != nil {
				return fmt.Errorf("cannot delete audit: %w", err)
			}

			return nil
		},
	)
}

func (s AuditService) ListForOrganizationID(
	ctx context.Context, predicate coredata.Predicater,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			filter := coredata.NewAuditFilter()

			err := audits.LoadByOrganizationID(ctx, conn, predicate, organizationID, cursor, filter)
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
	ctx context.Context, predicate coredata.Predicater,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			audits := coredata.Audits{}

			count, err = audits.CountByOrganizationID(ctx, conn, predicate, organizationID)
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
	ctx context.Context, predicate coredata.Predicater,
	req *UploadAuditReportRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := audit.LoadByID(ctx, conn, predicate, req.AuditID)
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
		ctx, predicate, fv,
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
			if err := audit.LoadByID(ctx, conn, predicate, req.AuditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			audit.ReportFileID = &file.ID
			audit.UpdatedAt = time.Now()

			return audit.Update(ctx, conn, predicate)
		},
	)
	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s AuditService) GenerateReportURL(
	ctx context.Context, predicate coredata.Predicater,
	auditID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	audit, err := s.Get(ctx, predicate, auditID)
	if err != nil {
		return nil, fmt.Errorf("cannot get audit: %w", err)
	}

	if audit.ReportFileID == nil {
		return nil, fmt.Errorf("audit has no report")
	}

	url, err := s.svc.Files.GenerateFileURL(ctx, predicate, *audit.ReportFileID, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate report download URL: %w", err)
	}

	return &url, nil
}

func (s AuditService) DeleteReport(
	ctx context.Context, predicate coredata.Predicater,
	auditID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	return audit, s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := audit.LoadByID(ctx, conn, predicate, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			if audit.ReportFileID != nil {
				file := coredata.File{ID: *audit.ReportFileID}

				if err := file.SoftDelete(ctx, conn, predicate); err != nil {
					return fmt.Errorf("cannot soft-delete report file: %w", err)
				}

				audit.ReportFileID = nil
				audit.UpdatedAt = time.Now()

				if err := audit.Update(ctx, conn, predicate); err != nil {
					return fmt.Errorf("cannot update audit: %w", err)
				}
			}

			return nil
		},
	)
}

func (s AuditService) ListForControlID(
	ctx context.Context, predicate coredata.Predicater,
	controlID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits

	control := &coredata.Control{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := control.LoadByID(ctx, conn, predicate, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			err := audits.LoadByControlID(ctx, conn, predicate, control.ID, cursor)
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
	ctx context.Context, predicate coredata.Predicater,
	controlID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			audits := coredata.Audits{}

			count, err = audits.CountByControlID(ctx, conn, predicate, controlID)
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
	ctx context.Context, predicate coredata.Predicater,
	findingID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			audits := coredata.Audits{}

			count, err = audits.CountByFindingID(ctx, conn, predicate, findingID)
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
	ctx context.Context, predicate coredata.Predicater,
	findingID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits

	finding := &coredata.Finding{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := finding.LoadByID(ctx, conn, predicate, findingID); err != nil {
				return fmt.Errorf("cannot load finding: %w", err)
			}

			err := audits.LoadByFindingID(ctx, conn, predicate, finding.ID, cursor)
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
