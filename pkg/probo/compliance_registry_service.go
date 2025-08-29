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

type ComplianceRegistryService struct {
	svc *TenantService
}

type (
	CreateComplianceRegistryRequest struct {
		OrganizationID         gid.GID
		ReferenceID            string
		Area                   *string
		Source                 *string
		Requirement            *string
		ActionsToBeImplemented *string
		Regulator              *string
		OwnerID                gid.GID
		LastReviewDate         *time.Time
		DueDate                *time.Time
		Status                 *coredata.ComplianceRegistryStatus
	}

	UpdateComplianceRegistryRequest struct {
		ID                     gid.GID
		ReferenceID            *string
		Area                   **string
		Source                 **string
		Requirement            **string
		ActionsToBeImplemented **string
		Regulator              **string
		OwnerID                *gid.GID
		LastReviewDate         **time.Time
		DueDate                **time.Time
		Status                 *coredata.ComplianceRegistryStatus
	}
)

func (s ComplianceRegistryService) Get(
	ctx context.Context,
	complianceRegistryID gid.GID,
) (*coredata.ComplianceRegistry, error) {
	registry := &coredata.ComplianceRegistry{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := registry.LoadByID(ctx, conn, s.svc.scope, complianceRegistryID); err != nil {
				return fmt.Errorf("cannot load compliance registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *ComplianceRegistryService) Create(
	ctx context.Context,
	req *CreateComplianceRegistryRequest,
) (*coredata.ComplianceRegistry, error) {
	now := time.Now()

	registry := &coredata.ComplianceRegistry{
		ID:                     gid.New(s.svc.scope.GetTenantID(), coredata.ComplianceRegistryEntityType),
		OrganizationID:         req.OrganizationID,
		ReferenceID:            req.ReferenceID,
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

			if err := registry.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert compliance registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *ComplianceRegistryService) Update(
	ctx context.Context,
	req *UpdateComplianceRegistryRequest,
) (*coredata.ComplianceRegistry, error) {
	registry := &coredata.ComplianceRegistry{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := registry.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance registry: %w", err)
			}

			if req.ReferenceID != nil {
				registry.ReferenceID = *req.ReferenceID
			}

			if req.Area != nil {
				registry.Area = *req.Area
			}

			if req.Source != nil {
				registry.Source = *req.Source
			}

			if req.Requirement != nil {
				registry.Requirement = *req.Requirement
			}

			if req.ActionsToBeImplemented != nil {
				registry.ActionsToBeImplemented = *req.ActionsToBeImplemented
			}

			if req.Regulator != nil {
				registry.Regulator = *req.Regulator
			}

			if req.OwnerID != nil {
				owner := &coredata.People{}
				if err := owner.LoadByID(ctx, conn, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner: %w", err)
				}
				registry.OwnerID = *req.OwnerID
			}

			if req.LastReviewDate != nil {
				registry.LastReviewDate = *req.LastReviewDate
			}

			if req.DueDate != nil {
				registry.DueDate = *req.DueDate
			}

			if req.Status != nil {
				registry.Status = *req.Status
			}

			registry.UpdatedAt = time.Now()

			if err := registry.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update compliance registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *ComplianceRegistryService) Delete(
	ctx context.Context,
	complianceRegistryID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			registry := &coredata.ComplianceRegistry{}
			if err := registry.LoadByID(ctx, conn, s.svc.scope, complianceRegistryID); err != nil {
				return fmt.Errorf("cannot load compliance registry: %w", err)
			}

			if err := registry.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete compliance registry: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s ComplianceRegistryService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.ComplianceRegistryFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			registries := coredata.ComplianceRegistries{}
			count, err = registries.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count compliance registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ComplianceRegistryService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ComplianceRegistryOrderField],
	filter *coredata.ComplianceRegistryFilter,
) (*page.Page[*coredata.ComplianceRegistry, coredata.ComplianceRegistryOrderField], error) {
	var registries coredata.ComplianceRegistries

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := registries.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load compliance registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(registries, cursor), nil
}
