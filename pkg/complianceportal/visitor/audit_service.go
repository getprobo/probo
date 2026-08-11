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

package visitor

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func (s *Service) GetAudit(
	ctx context.Context,
	scope coredata.Scoper,
	auditID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := audit.LoadByID(ctx, conn, scope, auditID)
			if err != nil {
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

func (s *Service) GetAuditForCompliancePortalID(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	auditID gid.GID,
) (*coredata.Audit, error) {
	audit := &coredata.Audit{}
	portalAudit := &coredata.CompliancePortalAudit{}
	compliancePortal := &coredata.CompliancePortal{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := loadPortalByID(ctx, conn, scope, compliancePortalID, compliancePortal); err != nil {
				return err
			}

			err := portalAudit.LoadByCompliancePortalIDAndAuditID(ctx, conn, scope, compliancePortalID, auditID)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrAuditNotFound
				}

				return fmt.Errorf("cannot load compliance portal audit: %w", err)
			}

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

func (s *Service) GetAuditByReportFileID(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	fileID gid.GID,
) (*coredata.Audit, *coredata.CompliancePortalAudit, error) {
	var (
		audit       *coredata.Audit
		portalAudit *coredata.CompliancePortalAudit
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			audit, portalAudit, err = loadAuditByReportFileID(ctx, conn, scope, compliancePortalID, fileID)

			return err
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return audit, portalAudit, nil
}

func loadAuditByReportFileID(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	fileID gid.GID,
) (*coredata.Audit, *coredata.CompliancePortalAudit, error) {
	audit := &coredata.Audit{}
	portalAudit := &coredata.CompliancePortalAudit{}
	compliancePortal := &coredata.CompliancePortal{}

	if err := loadPortalByID(ctx, conn, scope, compliancePortalID, compliancePortal); err != nil {
		return nil, nil, err
	}

	if err := portalAudit.LoadByCompliancePortalIDAndReportFileID(ctx, conn, scope, compliancePortalID, fileID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil, ErrReportNotFound
		}

		return nil, nil, fmt.Errorf("cannot load compliance portal audit: %w", err)
	}

	if err := audit.LoadByID(ctx, conn, scope, portalAudit.AuditID); err != nil {
		return nil, nil, fmt.Errorf("cannot load audit: %w", err)
	}

	return audit, portalAudit, nil
}

func (s *Service) ListAuditsForCompliancePortalID(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	cursor *page.Cursor[coredata.AuditOrderField],
	filter *coredata.AuditFilter,
) (*page.Page[*coredata.Audit, coredata.AuditOrderField], error) {
	var audits coredata.Audits

	if filter == nil {
		filter = coredata.NewAuditCompliancePortalFilter()
	}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			compliancePortal := &coredata.CompliancePortal{}
			if err := loadPortalByID(ctx, conn, scope, compliancePortalID, compliancePortal); err != nil {
				return err
			}

			err := audits.LoadByCompliancePortalID(ctx, conn, scope, compliancePortalID, compliancePortal.OrganizationID, cursor, filter)
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
