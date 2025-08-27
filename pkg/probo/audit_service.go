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

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"go.gearno.de/kit/pg"
)

type AuditService struct {
	svc *TenantService
}

type (
	CreateAuditRequest struct {
		OrganizationID gid.GID
		FrameworkID    gid.GID
		Name           *string
		ValidFrom      *time.Time
		ValidUntil     *time.Time
		State          *coredata.AuditState
	}

	UpdateAuditRequest struct {
		ID                gid.GID
		Name              **string
		ValidFrom         *time.Time
		ValidUntil        *time.Time
		State             *coredata.AuditState
		ShowOnTrustCenter *bool
	}

	UpdateAuditStateRequest struct {
		ID    gid.GID
		State coredata.AuditState
	}

	UploadAuditReportRequest struct {
		AuditID gid.GID
		File    File
	}

	DeleteAuditReportRequest struct {
		AuditID  gid.GID
		ReportID gid.GID
	}
)

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

func (s *AuditService) Create(
	ctx context.Context,
	req *CreateAuditRequest,
) (*coredata.Audit, error) {
	now := time.Now()

	audit := &coredata.Audit{
		ID:                gid.New(s.svc.scope.GetTenantID(), coredata.AuditEntityType),
		Name:              req.Name,
		OrganizationID:    req.OrganizationID,
		FrameworkID:       req.FrameworkID,
		ValidFrom:         req.ValidFrom,
		ValidUntil:        req.ValidUntil,
		State:             coredata.AuditStateNotStarted,
		ShowOnTrustCenter: false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if req.State != nil {
		audit.State = *req.State
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
			if req.ShowOnTrustCenter != nil {
				audit.ShowOnTrustCenter = *req.ShowOnTrustCenter
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
	audit := &coredata.Audit{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := audit.LoadByID(ctx, conn, s.svc.scope, req.AuditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			report, err := s.svc.Reports.Create(ctx, req.File)
			if err != nil {
				return fmt.Errorf("cannot create report: %w", err)
			}

			auditReport := &coredata.AuditReport{
				AuditID:   req.AuditID,
				ReportID:  report.ID,
				CreatedAt: time.Now(),
			}

			if err := auditReport.Upsert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot create audit report relationship: %w", err)
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

func (s AuditService) GetReports(
	ctx context.Context,
	auditID gid.GID,
) ([]*coredata.Report, error) {
	auditReports := &coredata.AuditReports{}
	
	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return auditReports.LoadByAuditID(ctx, conn, s.svc.scope, auditID)
		},
	)
	
	if err != nil {
		return nil, fmt.Errorf("cannot load audit reports: %w", err)
	}

	var reports []*coredata.Report
	for _, auditReport := range *auditReports {
		report := &coredata.Report{}
		err := s.svc.pg.WithConn(
			ctx,
			func(conn pg.Conn) error {
				return report.LoadByID(ctx, conn, s.svc.scope, auditReport.ReportID)
			},
		)
		if err != nil {
			return nil, fmt.Errorf("cannot load report %s: %w", auditReport.ReportID, err)
		}
		reports = append(reports, report)
	}

	return reports, nil
}

func (s AuditService) DeleteReport(
	ctx context.Context,
	req DeleteAuditReportRequest,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := audit.LoadByID(ctx, conn, s.svc.scope, req.AuditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			// Delete the report file
			report := &coredata.Report{ID: req.ReportID}
			if err := report.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete report: %w", err)
			}

			// Remove the audit-report relationship
			auditReport := &coredata.AuditReport{}
			if err := auditReport.Delete(ctx, conn, s.svc.scope, req.AuditID, req.ReportID); err != nil {
				return fmt.Errorf("cannot delete audit report relationship: %w", err)
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
