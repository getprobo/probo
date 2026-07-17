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

package management

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
	CreateFrameworkRequest struct {
		TrustCenterID gid.GID
		FrameworkID   gid.GID
	}

	UpdateFrameworkRequest struct {
		ID   gid.GID
		Rank int
	}

	DeleteFrameworkRequest struct {
		ID gid.GID
	}
)

func (r *CreateFrameworkRequest) Validate() error {
	v := validator.New()

	v.Check(r.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(r.FrameworkID, "framework_id", validator.Required(), validator.GID(coredata.FrameworkEntityType))

	return v.Error()
}

func (r *UpdateFrameworkRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.ComplianceFrameworkEntityType))

	return v.Error()
}

func (r *DeleteFrameworkRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.ComplianceFrameworkEntityType))

	return v.Error()
}

func (s *Service) ListFrameworksWithHidden(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	cursor *page.Cursor[coredata.ComplianceFrameworkOrderField],
) (*page.Page[*coredata.ComplianceFramework, coredata.ComplianceFrameworkOrderField], error) {
	var cfs coredata.ComplianceFrameworks

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := cfs.LoadWithHiddenByTrustCenterID(ctx, conn, scope, compliancePageID, cursor); err != nil {
				return fmt.Errorf("cannot load frameworks with hidden: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(cfs, cursor), nil
}

func (s *Service) CreateFramework(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateFrameworkRequest,
) (*coredata.ComplianceFramework, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	cfID := gid.New(scope.GetTenantID(), coredata.ComplianceFrameworkEntityType)

	var cf *coredata.ComplianceFramework

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			compliancePage := &coredata.TrustCenter{}
			if err := compliancePage.LoadByID(ctx, tx, scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, tx, scope, req.FrameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			cf = &coredata.ComplianceFramework{
				ID:             cfID,
				OrganizationID: compliancePage.OrganizationID,
				TrustCenterID:  req.TrustCenterID,
				FrameworkID:    req.FrameworkID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := cf.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert framework: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cf, nil
}

func (s *Service) UpdateFramework(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateFrameworkRequest,
) (*coredata.ComplianceFramework, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var cf *coredata.ComplianceFramework

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cf = &coredata.ComplianceFramework{}

			if err := cf.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			cf.Rank = req.Rank
			cf.UpdatedAt = time.Now()

			if err := cf.UpdateRank(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update framework rank: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cf, nil
}

func (s *Service) DeleteFramework(
	ctx context.Context,
	scope coredata.Scoper,
	req *DeleteFrameworkRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cf := &coredata.ComplianceFramework{}

			if err := cf.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if err := cf.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete framework: %w", err)
			}

			return nil
		},
	)
}
