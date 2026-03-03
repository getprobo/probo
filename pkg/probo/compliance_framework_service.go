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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	ComplianceFrameworkService struct {
		svc *TenantService
	}

	CreateComplianceFrameworkRequest struct {
		TrustCenterID gid.GID
		FrameworkID   gid.GID
	}

	UpdateComplianceFrameworkRequest struct {
		ID   gid.GID
		Rank int
	}

	DeleteComplianceFrameworkRequest struct {
		ID gid.GID
	}
)

func (r *CreateComplianceFrameworkRequest) Validate() error {
	v := validator.New()

	v.Check(r.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(r.FrameworkID, "framework_id", validator.Required(), validator.GID(coredata.FrameworkEntityType))

	return v.Error()
}

func (r *UpdateComplianceFrameworkRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.ComplianceFrameworkEntityType))

	return v.Error()
}

func (r *DeleteComplianceFrameworkRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.ComplianceFrameworkEntityType))

	return v.Error()
}

func (s ComplianceFrameworkService) ListWithHiddenForTrustCenterID(
	ctx context.Context,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.ComplianceFrameworkOrderField],
) (*page.Page[*coredata.ComplianceFramework, coredata.ComplianceFrameworkOrderField], error) {
	var cfs coredata.ComplianceFrameworks

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := cfs.LoadWithHiddenByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID, cursor); err != nil {
				return fmt.Errorf("cannot load compliance frameworks with hidden: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(cfs, cursor), nil
}

func (s ComplianceFrameworkService) Create(
	ctx context.Context,
	req *CreateComplianceFrameworkRequest,
) (*coredata.ComplianceFramework, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	cfID := gid.New(s.svc.scope.GetTenantID(), coredata.ComplianceFrameworkEntityType)

	var cf *coredata.ComplianceFramework

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			cf = &coredata.ComplianceFramework{
				ID:             cfID,
				OrganizationID: trustCenter.OrganizationID,
				TrustCenterID:  req.TrustCenterID,
				FrameworkID:    req.FrameworkID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := cf.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert compliance framework: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return cf, nil
}

func (s ComplianceFrameworkService) Update(
	ctx context.Context,
	req *UpdateComplianceFrameworkRequest,
) (*coredata.ComplianceFramework, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var cf *coredata.ComplianceFramework

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			cf = &coredata.ComplianceFramework{}

			if err := cf.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance framework: %w", err)
			}

			cf.Rank = req.Rank
			cf.UpdatedAt = time.Now()

			if err := cf.UpdateRank(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update compliance framework rank: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return cf, nil
}

func (s ComplianceFrameworkService) Delete(
	ctx context.Context,
	req *DeleteComplianceFrameworkRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			cf := &coredata.ComplianceFramework{}

			if err := cf.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance framework: %w", err)
			}

			if err := cf.Delete(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete compliance framework: %w", err)
			}

			return nil
		},
	)
}
