// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type AuditProgramService struct {
	svc *Service
}

type (
	CreateAuditProgramRequest struct {
		OrganizationID gid.GID
		FrameworkID    gid.GID
		Name           string
		ValidFrom      *time.Time
		ValidUntil     *time.Time
	}

	UpdateAuditProgramRequest struct {
		ID         gid.GID
		Name       *string
		ValidFrom  **time.Time
		ValidUntil **time.Time
	}
)

func (r *CreateAuditProgramRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.FrameworkID, "framework_id", validator.Required(), validator.GID(coredata.FrameworkEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.ValidUntil, "valid_until", validator.After(r.ValidFrom))

	return v.Error()
}

func (r *UpdateAuditProgramRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.AuditProgramEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.ValidUntil, "valid_until", validator.After(r.ValidFrom))

	return v.Error()
}

func (s AuditProgramService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	auditProgramID gid.GID,
) (*coredata.AuditProgram, error) {
	auditProgram := &coredata.AuditProgram{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := auditProgram.LoadByID(ctx, conn, scope, auditProgramID); err != nil {
				return fmt.Errorf("cannot load audit program: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return auditProgram, nil
}

func (s AuditProgramService) GetByIDs(
	ctx context.Context,
	scope coredata.Scoper,
	auditProgramIDs ...gid.GID,
) (coredata.AuditPrograms, error) {
	var auditPrograms coredata.AuditPrograms

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := auditPrograms.LoadByIDs(
				ctx,
				conn,
				scope,
				auditProgramIDs,
			); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load audit programs by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return auditPrograms, nil
}

func (s *AuditProgramService) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateAuditProgramRequest,
) (*coredata.AuditProgram, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	auditProgram := &coredata.AuditProgram{
		ID:             gid.New(scope.GetTenantID(), coredata.AuditProgramEntityType),
		OrganizationID: req.OrganizationID,
		FrameworkID:    req.FrameworkID,
		Name:           req.Name,
		ValidFrom:      req.ValidFrom,
		ValidUntil:     req.ValidUntil,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, scope, req.FrameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if framework.OrganizationID != organization.ID {
				return fmt.Errorf("framework belongs to a different organization: %w", coredata.ErrResourceNotFound)
			}

			if err := auditProgram.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert audit program: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return auditProgram, nil
}

func (s *AuditProgramService) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateAuditProgramRequest,
) (*coredata.AuditProgram, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	auditProgram := &coredata.AuditProgram{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := auditProgram.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load audit program: %w", err)
			}

			if req.Name != nil {
				auditProgram.Name = *req.Name
			}

			if req.ValidFrom != nil {
				auditProgram.ValidFrom = *req.ValidFrom
			}

			if req.ValidUntil != nil {
				auditProgram.ValidUntil = *req.ValidUntil
			}

			v := validator.New()
			v.Check(auditProgram.ValidUntil, "valid_until", validator.After(auditProgram.ValidFrom))

			if err := v.Error(); err != nil {
				return err
			}

			auditProgram.UpdatedAt = time.Now()

			if err := auditProgram.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update audit program: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return auditProgram, nil
}

func (s AuditProgramService) Delete(
	ctx context.Context,
	scope coredata.Scoper,
	auditProgramID gid.GID,
) error {
	auditProgram := coredata.AuditProgram{ID: auditProgramID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := auditProgram.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete audit program: %w", err)
			}

			return nil
		},
	)
}

func (s AuditProgramService) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AuditProgramOrderField],
) (*page.Page[*coredata.AuditProgram, coredata.AuditProgramOrderField], error) {
	var auditPrograms coredata.AuditPrograms

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := auditPrograms.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
				return fmt.Errorf("cannot load audit programs: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(auditPrograms, cursor), nil
}

func (s AuditProgramService) CountForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			auditPrograms := coredata.AuditPrograms{}

			count, err = auditPrograms.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count audit programs: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
