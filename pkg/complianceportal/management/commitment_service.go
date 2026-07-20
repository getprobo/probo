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
	CreateCompliancePortalCommitmentRequest struct {
		GroupID     gid.GID
		Icon        coredata.CompliancePortalCommitmentIcon
		Eyebrow     string
		Title       string
		Description string
	}

	UpdateCompliancePortalCommitmentRequest struct {
		ID          gid.GID
		Icon        *coredata.CompliancePortalCommitmentIcon
		Eyebrow     *string
		Title       *string
		Description *string
		Rank        *int
	}
)

func (r *CreateCompliancePortalCommitmentRequest) Validate() error {
	v := validator.New()

	v.Check(r.GroupID, "group_id", validator.Required(), validator.GID(coredata.CompliancePortalCommitmentGroupEntityType))
	v.Check(r.Icon, "icon", validator.Required(), validator.OneOfSlice(coredata.CompliancePortalCommitmentIcons()))
	v.Check(r.Eyebrow, "eyebrow", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Title, "title", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (r *UpdateCompliancePortalCommitmentRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.CompliancePortalCommitmentEntityType))

	if r.Icon != nil {
		v.Check(*r.Icon, "icon", validator.OneOfSlice(coredata.CompliancePortalCommitmentIcons()))
	}

	v.Check(r.Eyebrow, "eyebrow", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Title, "title", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s *Service) ListCommitments(
	ctx context.Context,
	scope coredata.Scoper,
	groupID gid.GID,
	cursor *page.Cursor[coredata.CompliancePortalCommitmentOrderField],
) (*page.Page[*coredata.CompliancePortalCommitment, coredata.CompliancePortalCommitmentOrderField], error) {
	var commitments coredata.CompliancePortalCommitments

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := commitments.LoadByGroupID(ctx, conn, scope, groupID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load compliance portal commitments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(commitments, cursor), nil
}

func (s *Service) CountCommitments(
	ctx context.Context,
	scope coredata.Scoper,
	groupID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			commitments := coredata.CompliancePortalCommitments{}

			count, err = commitments.CountByGroupID(ctx, conn, scope, groupID)
			if err != nil {
				return fmt.Errorf("cannot count compliance portal commitments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) GetCommitment(
	ctx context.Context,
	scope coredata.Scoper,
	commitmentID gid.GID,
) (*coredata.CompliancePortalCommitment, error) {
	var commitment coredata.CompliancePortalCommitment

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := commitment.LoadByID(ctx, conn, scope, commitmentID)
			if err != nil {
				return fmt.Errorf("cannot load compliance portal commitment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &commitment, nil
}

func (s *Service) CreateCommitment(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateCompliancePortalCommitmentRequest,
) (*coredata.CompliancePortalCommitment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	commitmentID := gid.New(scope.GetTenantID(), coredata.CompliancePortalCommitmentEntityType)

	var commitment *coredata.CompliancePortalCommitment

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			group := &coredata.CompliancePortalCommitmentGroup{}
			if err := group.LoadByID(ctx, tx, scope, req.GroupID); err != nil {
				return fmt.Errorf("cannot load compliance portal commitment group: %w", err)
			}

			commitment = &coredata.CompliancePortalCommitment{
				ID:                 commitmentID,
				OrganizationID:     group.OrganizationID,
				CompliancePortalID: group.CompliancePortalID,
				GroupID:            req.GroupID,
				Icon:               req.Icon,
				Eyebrow:            req.Eyebrow,
				Title:              req.Title,
				Description:        req.Description,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if err := commitment.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert compliance portal commitment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return commitment, nil
}

func (s *Service) UpdateCommitment(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateCompliancePortalCommitmentRequest,
) (*coredata.CompliancePortalCommitment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	var commitment *coredata.CompliancePortalCommitment

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			commitment = &coredata.CompliancePortalCommitment{}

			if err := commitment.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance portal commitment: %w", err)
			}

			if req.Icon != nil {
				commitment.Icon = *req.Icon
			}

			if req.Eyebrow != nil {
				commitment.Eyebrow = *req.Eyebrow
			}

			if req.Title != nil {
				commitment.Title = *req.Title
			}

			if req.Description != nil {
				commitment.Description = *req.Description
			}

			commitment.UpdatedAt = now

			if req.Rank != nil {
				commitment.Rank = *req.Rank
				if err := commitment.UpdateRank(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update rank: %w", err)
				}
			}

			if err := commitment.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update compliance portal commitment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return commitment, nil
}

func (s *Service) DeleteCommitment(
	ctx context.Context,
	scope coredata.Scoper,
	commitmentID gid.GID,
) error {
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			commitment := &coredata.CompliancePortalCommitment{}

			if err := commitment.LoadByID(ctx, tx, scope, commitmentID); err != nil {
				return fmt.Errorf("cannot load compliance portal commitment: %w", err)
			}

			if err := commitment.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete compliance portal commitment: %w", err)
			}

			return nil
		},
	)

	return err
}
