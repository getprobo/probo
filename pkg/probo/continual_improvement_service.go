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

type ContinualImprovementService struct {
	svc *TenantService
}

type (
	CreateContinualImprovementRequest struct {
		OrganizationID gid.GID
		ReferenceID    string
		Description    *string
		Source         *string
		OwnerID        gid.GID
		TargetDate     *time.Time
		Status         *coredata.ContinualImprovementStatus
		Priority       *coredata.ContinualImprovementPriority
	}

	UpdateContinualImprovementRequest struct {
		ID          gid.GID
		ReferenceID *string
		Description **string
		Source      **string
		OwnerID     *gid.GID
		TargetDate  **time.Time
		Status      *coredata.ContinualImprovementStatus
		Priority    *coredata.ContinualImprovementPriority
	}
)

func (ccir *CreateContinualImprovementRequest) Validate() error {
	v := validator.New()

	v.Check(ccir.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(ccir.ReferenceID, "reference_id", validator.SafeText(NameMaxLength))
	v.Check(ccir.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(ccir.Source, "source", validator.SafeText(ContentMaxLength))
	v.Check(ccir.OwnerID, "owner_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	v.Check(ccir.Status, "status", validator.OneOfSlice(coredata.ContinualImprovementStatuses()))
	v.Check(ccir.Priority, "priority", validator.OneOfSlice(coredata.ContinualImprovementPriorities()))

	return v.Error()
}

func (ucir *UpdateContinualImprovementRequest) Validate() error {
	v := validator.New()

	v.Check(ucir.ID, "id", validator.Required(), validator.GID(coredata.ContinualImprovementEntityType))
	v.Check(ucir.ReferenceID, "reference_id", validator.SafeText(NameMaxLength))
	v.Check(ucir.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(ucir.Source, "source", validator.SafeText(ContentMaxLength))
	v.Check(ucir.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(ucir.Status, "status", validator.OneOfSlice(coredata.ContinualImprovementStatuses()))
	v.Check(ucir.Priority, "priority", validator.OneOfSlice(coredata.ContinualImprovementPriorities()))

	return v.Error()
}

func (s ContinualImprovementService) Get(
	ctx context.Context,
	continualImprovementID gid.GID,
) (*coredata.ContinualImprovement, error) {
	improvement := &coredata.ContinualImprovement{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := improvement.LoadByID(ctx, conn, s.svc.scope, continualImprovementID); err != nil {
				return fmt.Errorf("cannot load continual improvement: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return improvement, nil
}

func (s *ContinualImprovementService) Create(
	ctx context.Context,
	req *CreateContinualImprovementRequest,
) (*coredata.ContinualImprovement, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	improvement := &coredata.ContinualImprovement{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.ContinualImprovementEntityType),
		OrganizationID: req.OrganizationID,
		ReferenceID:    req.ReferenceID,
		Description:    req.Description,
		Source:         req.Source,
		OwnerID:        req.OwnerID,
		TargetDate:     req.TargetDate,
		Status:         *req.Status,
		Priority:       *req.Priority,
		CreatedAt:      now,
		UpdatedAt:      now,
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

			if err := improvement.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert continual improvement: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return improvement, nil
}

func (s *ContinualImprovementService) Update(
	ctx context.Context,
	req *UpdateContinualImprovementRequest,
) (*coredata.ContinualImprovement, error) {
	improvement := &coredata.ContinualImprovement{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := improvement.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load continual improvement: %w", err)
			}

			if req.ReferenceID != nil {
				improvement.ReferenceID = *req.ReferenceID
			}

			if req.Description != nil {
				improvement.Description = *req.Description
			}

			if req.Source != nil {
				improvement.Source = *req.Source
			}

			if req.OwnerID != nil {
				owner := &coredata.MembershipProfile{}
				if err := owner.LoadByID(ctx, conn, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}
				improvement.OwnerID = *req.OwnerID
			}

			if req.TargetDate != nil {
				improvement.TargetDate = *req.TargetDate
			}

			if req.Status != nil {
				improvement.Status = *req.Status
			}

			if req.Priority != nil {
				improvement.Priority = *req.Priority
			}

			improvement.UpdatedAt = time.Now()

			if err := improvement.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update continual improvement: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return improvement, nil
}

func (s *ContinualImprovementService) Delete(
	ctx context.Context,
	continualImprovementID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			improvement := &coredata.ContinualImprovement{}
			if err := improvement.LoadByID(ctx, conn, s.svc.scope, continualImprovementID); err != nil {
				return fmt.Errorf("cannot load continual improvement: %w", err)
			}

			if err := improvement.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete continual improvement: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s ContinualImprovementService) CountByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.ContinualImprovementFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			improvements := coredata.ContinualImprovements{}
			count, err = improvements.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count continual improvements: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ContinualImprovementService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ContinualImprovementOrderField],
	filter *coredata.ContinualImprovementFilter,
) (*page.Page[*coredata.ContinualImprovement, coredata.ContinualImprovementOrderField], error) {
	var improvements coredata.ContinualImprovements

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := improvements.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load continual improvements: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(improvements, cursor), nil
}
