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

type ObligationService struct {
	svc *TenantService
}

type (
	CreateObligationRequest struct {
		OrganizationID         gid.GID
		Area                   *string
		Source                 *string
		Requirement            *string
		ActionsToBeImplemented *string
		Regulator              *string
		OwnerID                gid.GID
		LastReviewDate         *time.Time
		DueDate                *time.Time
		Status                 *coredata.ObligationStatus
	}

	UpdateObligationRequest struct {
		ID                     gid.GID
		Area                   **string
		Source                 **string
		Requirement            **string
		ActionsToBeImplemented **string
		Regulator              **string
		OwnerID                *gid.GID
		LastReviewDate         **time.Time
		DueDate                **time.Time
		Status                 *coredata.ObligationStatus
	}
)

func (s ObligationService) Get(
	ctx context.Context,
	obligationID gid.GID,
) (*coredata.Obligation, error) {
	obligation := &coredata.Obligation{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := obligation.LoadByID(ctx, conn, s.svc.scope, obligationID); err != nil {
				return fmt.Errorf("cannot load obligation: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return obligation, nil
}

func (s *ObligationService) Create(
	ctx context.Context,
	req *CreateObligationRequest,
) (*coredata.Obligation, error) {
	now := time.Now()

	obligation := &coredata.Obligation{
		ID:                     gid.New(s.svc.scope.GetTenantID(), coredata.ObligationEntityType),
		OrganizationID:         req.OrganizationID,
		Area:                   req.Area,
		Source:                 req.Source,
		Requirement:            req.Requirement,
		ActionsToBeImplemented: req.ActionsToBeImplemented,
		Regulator:              req.Regulator,
		OwnerID:                req.OwnerID,
		LastReviewDate:         req.LastReviewDate,
		DueDate:                req.DueDate,
		Status:                 *req.Status,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			owner := &coredata.People{}
			if err := owner.LoadByID(ctx, conn, s.svc.scope, req.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner: %w", err)
			}

			if err := obligation.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert obligation: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return obligation, nil
}

func (s *ObligationService) Update(
	ctx context.Context,
	req *UpdateObligationRequest,
) (*coredata.Obligation, error) {
	obligation := &coredata.Obligation{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := obligation.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load obligation: %w", err)
			}

			if req.Area != nil {
				obligation.Area = *req.Area
			}

			if req.Source != nil {
				obligation.Source = *req.Source
			}

			if req.Requirement != nil {
				obligation.Requirement = *req.Requirement
			}

			if req.ActionsToBeImplemented != nil {
				obligation.ActionsToBeImplemented = *req.ActionsToBeImplemented
			}

			if req.Regulator != nil {
				obligation.Regulator = *req.Regulator
			}

			if req.OwnerID != nil {
				owner := &coredata.People{}
				if err := owner.LoadByID(ctx, conn, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner: %w", err)
				}
				obligation.OwnerID = *req.OwnerID
			}

			if req.LastReviewDate != nil {
				obligation.LastReviewDate = *req.LastReviewDate
			}

			if req.DueDate != nil {
				obligation.DueDate = *req.DueDate
			}

			if req.Status != nil {
				obligation.Status = *req.Status
			}

			obligation.UpdatedAt = time.Now()

			if err := obligation.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update obligation: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return obligation, nil
}

func (s *ObligationService) Delete(
	ctx context.Context,
	obligationID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			obligation := &coredata.Obligation{}
			if err := obligation.LoadByID(ctx, conn, s.svc.scope, obligationID); err != nil {
				return fmt.Errorf("cannot load obligation: %w", err)
			}

			if err := obligation.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete obligation: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s ObligationService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.ObligationFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			obligations := coredata.Obligations{}
			count, err = obligations.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count obligations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ObligationService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ObligationOrderField],
	filter *coredata.ObligationFilter,
) (*page.Page[*coredata.Obligation, coredata.ObligationOrderField], error) {
	var obligations coredata.Obligations

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := obligations.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load obligations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(obligations, cursor), nil
}

func (s ObligationService) CountForRiskID(
	ctx context.Context,
	riskID gid.GID,
	filter *coredata.ObligationFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			obligations := &coredata.Obligations{}
			count, err = obligations.CountByRiskID(ctx, conn, s.svc.scope, riskID, filter)
			if err != nil {
				return fmt.Errorf("cannot count obligations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ObligationService) ListForRiskID(
	ctx context.Context,
	riskID gid.GID,
	cursor *page.Cursor[coredata.ObligationOrderField],
	filter *coredata.ObligationFilter,
) (*page.Page[*coredata.Obligation, coredata.ObligationOrderField], error) {
	var obligations coredata.Obligations

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := obligations.LoadByRiskID(ctx, conn, s.svc.scope, riskID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load obligations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(obligations, cursor), nil
}
