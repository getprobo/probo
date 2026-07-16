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
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	CompliancePortalCommitmentGroupService struct {
		svc *Service
	}

	CreateCompliancePortalCommitmentGroupRequest struct {
		TrustCenterID gid.GID
		Title         string
		Description   string
	}

	UpdateCompliancePortalCommitmentGroupRequest struct {
		ID          gid.GID
		Title       *string
		Description *string
		Rank        *int
	}
)

func (r *CreateCompliancePortalCommitmentGroupRequest) Validate() error {
	v := validator.New()

	v.Check(r.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(r.Title, "title", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (r *UpdateCompliancePortalCommitmentGroupRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.CompliancePortalCommitmentGroupEntityType))
	v.Check(r.Title, "title", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s CompliancePortalCommitmentGroupService) ListForTrustCenterID(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.CompliancePortalCommitmentGroupOrderField],
) (*page.Page[*coredata.CompliancePortalCommitmentGroup, coredata.CompliancePortalCommitmentGroupOrderField], error) {
	var groups coredata.CompliancePortalCommitmentGroups

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		err := groups.LoadByTrustCenterID(ctx, conn, scope, trustCenterID, cursor)
		if err != nil {
			return fmt.Errorf("cannot load compliance portal commitment groups: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return page.NewPage(groups, cursor), nil
}

func (s CompliancePortalCommitmentGroupService) CountForTrustCenterID(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) (err error) {
		groups := coredata.CompliancePortalCommitmentGroups{}

		count, err = groups.CountByTrustCenterID(ctx, conn, scope, trustCenterID)
		if err != nil {
			return fmt.Errorf("cannot count compliance portal commitment groups: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s CompliancePortalCommitmentGroupService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	groupID gid.GID,
) (*coredata.CompliancePortalCommitmentGroup, error) {
	var group coredata.CompliancePortalCommitmentGroup

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		err := group.LoadByID(ctx, conn, scope, groupID)
		if err != nil {
			return fmt.Errorf("cannot load compliance portal commitment group: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (s CompliancePortalCommitmentGroupService) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateCompliancePortalCommitmentGroupRequest,
) (*coredata.CompliancePortalCommitmentGroup, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	groupID := gid.New(scope.GetTenantID(), coredata.CompliancePortalCommitmentGroupEntityType)

	var group *coredata.CompliancePortalCommitmentGroup

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		trustCenter := &coredata.TrustCenter{}
		if err := trustCenter.LoadByID(ctx, tx, scope, req.TrustCenterID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}

		group = &coredata.CompliancePortalCommitmentGroup{
			ID:             groupID,
			OrganizationID: trustCenter.OrganizationID,
			TrustCenterID:  req.TrustCenterID,
			Title:          req.Title,
			Description:    req.Description,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := group.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert compliance portal commitment group: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s CompliancePortalCommitmentGroupService) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateCompliancePortalCommitmentGroupRequest,
) (*coredata.CompliancePortalCommitmentGroup, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	var group *coredata.CompliancePortalCommitmentGroup

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		group = &coredata.CompliancePortalCommitmentGroup{}

		if err := group.LoadByID(ctx, tx, scope, req.ID); err != nil {
			return fmt.Errorf("cannot load compliance portal commitment group: %w", err)
		}

		if req.Title != nil {
			group.Title = *req.Title
		}

		if req.Description != nil {
			group.Description = *req.Description
		}

		group.UpdatedAt = now

		if req.Rank != nil {
			group.Rank = *req.Rank
			if err := group.UpdateRank(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update rank: %w", err)
			}
		}

		if err := group.Update(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot update compliance portal commitment group: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s CompliancePortalCommitmentGroupService) Delete(
	ctx context.Context,
	scope coredata.Scoper,
	groupID gid.GID,
) error {
	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		group := &coredata.CompliancePortalCommitmentGroup{}

		if err := group.LoadByID(ctx, tx, scope, groupID); err != nil {
			return fmt.Errorf("cannot load compliance portal commitment group: %w", err)
		}

		if err := group.Delete(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot delete compliance portal commitment group: %w", err)
		}

		return nil
	})

	return err
}
