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

type ContinualImprovementRegistriesService struct {
	svc *TenantService
}

type (
	CreateContinualImprovementRegistryRequest struct {
		OrganizationID gid.GID
		ReferenceID    string
		Description    *string
		AuditID        gid.GID
		Source         *string
		OwnerID        gid.GID
		TargetDate     *time.Time
		Status         *coredata.ContinualImprovementRegistriesStatus
		Priority       *coredata.ContinualImprovementRegistriesPriority
	}

	UpdateContinualImprovementRegistryRequest struct {
		ID          gid.GID
		ReferenceID *string
		Description **string
		AuditID     *gid.GID
		Source      **string
		OwnerID     *gid.GID
		TargetDate  **time.Time
		Status      *coredata.ContinualImprovementRegistriesStatus
		Priority    *coredata.ContinualImprovementRegistriesPriority
	}
)

func (s ContinualImprovementRegistriesService) Get(
	ctx context.Context,
	continualImprovementRegistryID gid.GID,
) (*coredata.ContinualImprovementRegistry, error) {
	registry := &coredata.ContinualImprovementRegistry{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := registry.LoadByID(ctx, conn, s.svc.scope, continualImprovementRegistryID); err != nil {
				return fmt.Errorf("cannot load continual improvement registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *ContinualImprovementRegistriesService) Create(
	ctx context.Context,
	req *CreateContinualImprovementRegistryRequest,
) (*coredata.ContinualImprovementRegistry, error) {
	now := time.Now()

	registry := &coredata.ContinualImprovementRegistry{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.ContinualImprovementRegistryEntityType),
		OrganizationID: req.OrganizationID,
		ReferenceID:    req.ReferenceID,
		Description:    req.Description,
		AuditID:        req.AuditID,
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

			audit := &coredata.Audit{}
			if err := audit.LoadByID(ctx, conn, s.svc.scope, req.AuditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			owner := &coredata.People{}
			if err := owner.LoadByID(ctx, conn, s.svc.scope, req.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner: %w", err)
			}

			if err := registry.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert continual improvement registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *ContinualImprovementRegistriesService) Update(
	ctx context.Context,
	req *UpdateContinualImprovementRegistryRequest,
) (*coredata.ContinualImprovementRegistry, error) {
	registry := &coredata.ContinualImprovementRegistry{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := registry.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load continual improvement registry: %w", err)
			}

			if req.ReferenceID != nil {
				registry.ReferenceID = *req.ReferenceID
			}

			if req.Description != nil {
				registry.Description = *req.Description
			}

			if req.AuditID != nil {
				audit := &coredata.Audit{}
				if err := audit.LoadByID(ctx, conn, s.svc.scope, *req.AuditID); err != nil {
					return fmt.Errorf("cannot load audit: %w", err)
				}
				registry.AuditID = *req.AuditID
			}

			if req.Source != nil {
				registry.Source = *req.Source
			}

			if req.OwnerID != nil {
				owner := &coredata.People{}
				if err := owner.LoadByID(ctx, conn, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner: %w", err)
				}
				registry.OwnerID = *req.OwnerID
			}

			if req.TargetDate != nil {
				registry.TargetDate = *req.TargetDate
			}

			if req.Status != nil {
				registry.Status = *req.Status
			}

			if req.Priority != nil {
				registry.Priority = *req.Priority
			}

			registry.UpdatedAt = time.Now()

			if err := registry.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update continual improvement registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *ContinualImprovementRegistriesService) Delete(
	ctx context.Context,
	continualImprovementRegistryID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			registry := &coredata.ContinualImprovementRegistry{}
			if err := registry.LoadByID(ctx, conn, s.svc.scope, continualImprovementRegistryID); err != nil {
				return fmt.Errorf("cannot load continual improvement registry: %w", err)
			}

			if err := registry.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete continual improvement registry: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s ContinualImprovementRegistriesService) CountByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			registries := coredata.ContinualImprovementRegistries{}
			count, err = registries.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count continual improvement registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ContinualImprovementRegistriesService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ContinualImprovementRegistriesOrderField],
) (*page.Page[*coredata.ContinualImprovementRegistry, coredata.ContinualImprovementRegistriesOrderField], error) {
	var registries coredata.ContinualImprovementRegistries

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := registries.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load continual improvement registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(registries, cursor), nil
}
