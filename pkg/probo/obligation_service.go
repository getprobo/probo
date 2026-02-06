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
		Status                 coredata.ObligationStatus
		Type                   coredata.ObligationType
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
		Type                   *coredata.ObligationType
	}
)

func (cor *CreateObligationRequest) Validate() error {
	v := validator.New()

	v.Check(cor.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(cor.Area, "area", validator.SafeText(TitleMaxLength))
	v.Check(cor.Source, "source", validator.SafeText(TitleMaxLength))
	v.Check(cor.Requirement, "requirement", validator.SafeText(ContentMaxLength))
	v.Check(cor.ActionsToBeImplemented, "actions_to_be_implemented", validator.SafeText(ContentMaxLength))
	v.Check(cor.Regulator, "regulator", validator.SafeText(TitleMaxLength))
	v.Check(cor.OwnerID, "owner_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	v.Check(cor.Status, "status", validator.OneOfSlice(coredata.ObligationStatuses()))
	v.Check(cor.Type, "type", validator.OneOfSlice(coredata.ObligationTypes()))

	return v.Error()
}

func (uor *UpdateObligationRequest) Validate() error {
	v := validator.New()

	v.Check(uor.ID, "id", validator.Required(), validator.GID(coredata.ObligationEntityType))
	v.Check(uor.Area, "area", validator.SafeText(NameMaxLength))
	v.Check(uor.Source, "source", validator.SafeText(NameMaxLength))
	v.Check(uor.Requirement, "requirement", validator.SafeText(ContentMaxLength))
	v.Check(uor.ActionsToBeImplemented, "actions_to_be_implemented", validator.SafeText(ContentMaxLength))
	v.Check(uor.Regulator, "regulator", validator.SafeText(NameMaxLength))
	v.Check(uor.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(uor.Status, "status", validator.OneOfSlice(coredata.ObligationStatuses()))
	v.Check(uor.Type, "type", validator.OneOfSlice(coredata.ObligationTypes()))

	return v.Error()
}

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
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

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
		Status:                 req.Status,
		Type:                   req.Type,
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

			owner := &coredata.MembershipProfile{}
			if err := owner.LoadByID(ctx, conn, s.svc.scope, req.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner profile: %w", err)
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
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

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
				owner := &coredata.MembershipProfile{}
				if err := owner.LoadByID(ctx, conn, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
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

			if req.Type != nil {
				obligation.Type = *req.Type
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

func (s ObligationService) ListForControlID(
	ctx context.Context,
	controlID gid.GID,
	cursor *page.Cursor[coredata.ObligationOrderField],
	filter *coredata.ObligationFilter,
) (*page.Page[*coredata.Obligation, coredata.ObligationOrderField], error) {
	var obligations coredata.Obligations
	control := &coredata.Control{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := control.LoadByID(ctx, conn, s.svc.scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			err := obligations.LoadByControlID(ctx, conn, s.svc.scope, control.ID, cursor, filter)
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
