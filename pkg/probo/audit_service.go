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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type AuditService struct {
	svc *TenantService
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
	ctx context.Context,
	auditID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return audit.LoadByID(ctx, conn, s.svc.scope, auditID)
		},
	)

	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s AuditService) GetByReportID(
	ctx context.Context,
	reportID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return audit.LoadByReportID(ctx, conn, s.svc.scope, reportID)
		},
	)

	if err != nil {
		return nil, err
	}

	return audit, nil
}

func (s *AuditService) Create(
	ctx context.Context,
	req *CreateAuditRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	audit := &coredata.Audit{
		ID:                    gid.New(s.svc.scope.GetTenantID(), coredata.AuditEntityType),
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
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, s.svc.scope, req.FrameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if err := audit.Insert(ctx, conn, s.svc.scope); err != nil {
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
	ctx context.Context,
	req *UpdateAuditRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	audit := &coredata.Audit{}
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := audit.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
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

			if err := audit.Update(ctx, conn, s.svc.scope); err != nil {
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
	ctx context.Context,
	auditID gid.GID,
) error {
	audit := coredata.Audit{ID: auditID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := audit.Delete(ctx, conn, s.svc.scope)
			if err != nil {
				return fmt.Errorf("cannot delete audit: %w", err)
			}
			return nil
		},
	)
}

func (s AuditService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			filter := coredata.NewAuditFilter()
			err := audits.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
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
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			audits := coredata.Audits{}
			count, err = audits.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
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
	ctx context.Context,
	req UploadAuditReportRequest,
) (*coredata.Audit, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	audit := &coredata.Audit{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := audit.LoadByID(ctx, conn, s.svc.scope, req.AuditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			reportID := gid.New(s.svc.scope.GetTenantID(), coredata.ReportEntityType)
			now := time.Now()

			objectKey, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("cannot generate object key: %w", err)
			}

			_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      aws.String(s.svc.bucket),
				Key:         aws.String(objectKey.String()),
				Body:        req.File.Content,
				ContentType: aws.String(req.File.ContentType),
				Metadata: map[string]string{
					"type":            "report",
					"report-id":       reportID.String(),
					"organization-id": audit.OrganizationID.String(),
				},
			})
			if err != nil {
				return fmt.Errorf("cannot upload report to S3: %w", err)
			}

			report := &coredata.Report{
				ID:             reportID,
				OrganizationID: audit.OrganizationID,
				ObjectKey:      objectKey.String(),
				MimeType:       req.File.ContentType,
				Filename:       req.File.Filename,
				Size:           req.File.Size,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := report.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert report: %w", err)
			}

			audit.ReportID = &report.ID
			audit.UpdatedAt = time.Now()

			if err := audit.Update(ctx, conn, s.svc.scope); err != nil {
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

func (s AuditService) GenerateReportURL(
	ctx context.Context,
	auditID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	audit, err := s.Get(ctx, auditID)
	if err != nil {
		return nil, fmt.Errorf("cannot get audit: %w", err)
	}

	if audit.ReportID == nil {
		return nil, fmt.Errorf("audit has no report")
	}

	url, err := s.svc.Reports.GenerateDownloadURL(ctx, *audit.ReportID, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate report download URL: %w", err)
	}

	return url, nil
}

func (s AuditService) DeleteReport(
	ctx context.Context,
	auditID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := audit.LoadByID(ctx, conn, s.svc.scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			if audit.ReportID != nil {
				report := &coredata.Report{ID: *audit.ReportID}

				if err := report.Delete(ctx, conn, s.svc.scope); err != nil {
					return fmt.Errorf("cannot delete report: %w", err)
				}

				audit.ReportID = nil
				audit.UpdatedAt = time.Now()

				if err := audit.Update(ctx, conn, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update audit: %w", err)
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

func (s AuditService) ListForControlID(
	ctx context.Context,
	controlID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits
	control := &coredata.Control{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := control.LoadByID(ctx, conn, s.svc.scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			err := audits.LoadByControlID(ctx, conn, s.svc.scope, control.ID, cursor)
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
