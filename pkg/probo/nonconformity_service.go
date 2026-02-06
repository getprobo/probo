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

type NonconformityService struct {
	svc *TenantService
}

type (
	CreateNonconformityRequest struct {
		OrganizationID     gid.GID
		ReferenceID        string
		Description        *string
		AuditID            *gid.GID
		DateIdentified     *time.Time
		RootCause          string
		CorrectiveAction   *string
		OwnerID            gid.GID
		DueDate            *time.Time
		Status             *coredata.NonconformityStatus
		EffectivenessCheck *string
	}

	UpdateNonconformityRequest struct {
		ID                 gid.GID
		ReferenceID        *string
		Description        **string
		DateIdentified     **time.Time
		RootCause          *string
		CorrectiveAction   **string
		OwnerID            *gid.GID
		AuditID            **gid.GID
		DueDate            **time.Time
		Status             *coredata.NonconformityStatus
		EffectivenessCheck **string
	}
)

func (cnr *CreateNonconformityRequest) Validate() error {
	v := validator.New()

	v.Check(cnr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(cnr.ReferenceID, "reference_id", validator.Required(), validator.SafeText(NameMaxLength))
	v.Check(cnr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(cnr.AuditID, "audit_id", validator.GID(coredata.AuditEntityType))
	v.Check(cnr.RootCause, "root_cause", validator.Required(), validator.SafeText(ContentMaxLength))
	v.Check(cnr.CorrectiveAction, "corrective_action", validator.SafeText(ContentMaxLength))
	v.Check(cnr.OwnerID, "owner_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))
	v.Check(cnr.Status, "status", validator.OneOfSlice(coredata.NonconformityStatuses()))
	v.Check(cnr.EffectivenessCheck, "effectiveness_check", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (unr *UpdateNonconformityRequest) Validate() error {
	v := validator.New()

	v.Check(unr.ID, "id", validator.Required(), validator.GID(coredata.NonconformityEntityType))
	v.Check(unr.ReferenceID, "reference_id", validator.SafeText(NameMaxLength))
	v.Check(unr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(unr.RootCause, "root_cause", validator.SafeText(ContentMaxLength))
	v.Check(unr.CorrectiveAction, "corrective_action", validator.SafeText(ContentMaxLength))
	v.Check(unr.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(unr.Status, "status", validator.OneOfSlice(coredata.NonconformityStatuses()))
	v.Check(unr.EffectivenessCheck, "effectiveness_check", validator.SafeText(ContentMaxLength))

	return v.Error()
}
func (s NonconformityService) Get(
	ctx context.Context,
	nonconformityID gid.GID,
) (*coredata.Nonconformity, error) {
	nonconformity := &coredata.Nonconformity{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return nonconformity.LoadByID(ctx, conn, s.svc.scope, nonconformityID)
		},
	)

	if err != nil {
		return nil, err
	}

	return nonconformity, nil
}

func (s *NonconformityService) Create(
	ctx context.Context,
	req *CreateNonconformityRequest,
) (*coredata.Nonconformity, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()

	nonconformity := &coredata.Nonconformity{
		ID:                 gid.New(s.svc.scope.GetTenantID(), coredata.NonconformityEntityType),
		OrganizationID:     req.OrganizationID,
		ReferenceID:        req.ReferenceID,
		Description:        req.Description,
		AuditID:            req.AuditID,
		DateIdentified:     req.DateIdentified,
		RootCause:          req.RootCause,
		CorrectiveAction:   req.CorrectiveAction,
		OwnerID:            req.OwnerID,
		DueDate:            req.DueDate,
		Status:             coredata.NonconformityStatusOpen,
		EffectivenessCheck: req.EffectivenessCheck,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if req.Status != nil {
		nonconformity.Status = *req.Status
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if req.AuditID != nil {
				audit := &coredata.Audit{}
				if err := audit.LoadByID(ctx, conn, s.svc.scope, *req.AuditID); err != nil {
					return fmt.Errorf("cannot load audit: %w", err)
				}
			}

			owner := &coredata.MembershipProfile{}
			if err := owner.LoadByID(ctx, conn, s.svc.scope, req.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner profile: %w", err)
			}

			if err := nonconformity.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert nonconformity: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return nonconformity, nil
}

func (s *NonconformityService) Update(
	ctx context.Context,
	req *UpdateNonconformityRequest,
) (*coredata.Nonconformity, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	nonconformity := &coredata.Nonconformity{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := nonconformity.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load nonconformity: %w", err)
			}

			if req.ReferenceID != nil {
				nonconformity.ReferenceID = *req.ReferenceID
			}
			if req.Description != nil {
				nonconformity.Description = *req.Description
			}
			if req.DateIdentified != nil {
				nonconformity.DateIdentified = *req.DateIdentified
			}
			if req.RootCause != nil {
				nonconformity.RootCause = *req.RootCause
			}
			if req.CorrectiveAction != nil {
				nonconformity.CorrectiveAction = *req.CorrectiveAction
			}
			if req.OwnerID != nil {
				owner := &coredata.MembershipProfile{}
				if err := owner.LoadByID(ctx, conn, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}
				nonconformity.OwnerID = *req.OwnerID
			}
			if req.AuditID != nil {
				if *req.AuditID != nil {
					audit := &coredata.Audit{}
					if err := audit.LoadByID(ctx, conn, s.svc.scope, **req.AuditID); err != nil {
						return fmt.Errorf("cannot load audit: %w", err)
					}
				}
				nonconformity.AuditID = *req.AuditID
			}
			if req.DueDate != nil {
				nonconformity.DueDate = *req.DueDate
			}
			if req.Status != nil {
				nonconformity.Status = *req.Status
			}
			if req.EffectivenessCheck != nil {
				nonconformity.EffectivenessCheck = *req.EffectivenessCheck
			}

			nonconformity.UpdatedAt = time.Now()

			if err := nonconformity.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update nonconformity: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return nonconformity, nil
}

func (s NonconformityService) Delete(
	ctx context.Context,
	nonconformityID gid.GID,
) error {
	nonconformity := coredata.Nonconformity{ID: nonconformityID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := nonconformity.Delete(ctx, conn, s.svc.scope)
			if err != nil {
				return fmt.Errorf("cannot delete nonconformity: %w", err)
			}
			return nil
		},
	)
}

func (s NonconformityService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.NonconformityOrderField],
	filter *coredata.NonconformityFilter,
) (*page.Page[*coredata.Nonconformity, coredata.NonconformityOrderField], error) {
	var nonconformities coredata.Nonconformities

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := nonconformities.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load nonconformities: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(nonconformities, cursor), nil
}

func (s NonconformityService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.NonconformityFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			nonconformities := coredata.Nonconformities{}
			count, err = nonconformities.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count nonconformities: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}
