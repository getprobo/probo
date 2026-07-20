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
	CreateCompliancePortalCommitmentGroupRequest struct {
		CompliancePortalID gid.GID
		Title              string
		Description        string
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

	v.Check(r.CompliancePortalID, "trust_center_id", validator.Required(), validator.GID(coredata.CompliancePortalEntityType))
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

func (s *Service) ListCommitmentGroups(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	cursor *page.Cursor[coredata.CompliancePortalCommitmentGroupOrderField],
) (*page.Page[*coredata.CompliancePortalCommitmentGroup, coredata.CompliancePortalCommitmentGroupOrderField], error) {
	var groups coredata.CompliancePortalCommitmentGroups

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := groups.LoadByCompliancePortalID(ctx, conn, scope, compliancePortalID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load compliance portal commitment groups: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(groups, cursor), nil
}

func (s *Service) CountCommitmentGroups(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			groups := coredata.CompliancePortalCommitmentGroups{}

			count, err = groups.CountByCompliancePortalID(ctx, conn, scope, compliancePortalID)
			if err != nil {
				return fmt.Errorf("cannot count compliance portal commitment groups: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) GetCommitmentGroup(
	ctx context.Context,
	scope coredata.Scoper,
	groupID gid.GID,
) (*coredata.CompliancePortalCommitmentGroup, error) {
	var group coredata.CompliancePortalCommitmentGroup

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := group.LoadByID(ctx, conn, scope, groupID)
			if err != nil {
				return fmt.Errorf("cannot load compliance portal commitment group: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (s *Service) CreateCommitmentGroup(
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

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			compliancePortal := &coredata.CompliancePortal{}
			if err := compliancePortal.LoadByID(ctx, tx, scope, req.CompliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			group = &coredata.CompliancePortalCommitmentGroup{
				ID:                 groupID,
				OrganizationID:     compliancePortal.OrganizationID,
				CompliancePortalID: req.CompliancePortalID,
				Title:              req.Title,
				Description:        req.Description,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if err := group.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert compliance portal commitment group: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *Service) UpdateCommitmentGroup(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateCompliancePortalCommitmentGroupRequest,
) (*coredata.CompliancePortalCommitmentGroup, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	var group *coredata.CompliancePortalCommitmentGroup

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
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
		},
	)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (s *Service) DeleteCommitmentGroup(
	ctx context.Context,
	scope coredata.Scoper,
	groupID gid.GID,
) error {
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			group := &coredata.CompliancePortalCommitmentGroup{}

			if err := group.LoadByID(ctx, tx, scope, groupID); err != nil {
				return fmt.Errorf("cannot load compliance portal commitment group: %w", err)
			}

			if err := group.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete compliance portal commitment group: %w", err)
			}

			return nil
		},
	)

	return err
}
