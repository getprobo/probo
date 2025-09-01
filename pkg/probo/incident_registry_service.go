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

type IncidentRegistryService struct {
	svc *TenantService
}

type (
	CreateIncidentRegistryRequest struct {
		OrganizationID gid.GID
		ReferenceID    string
		Title          string
		Description    *string
		Source         *string
		OwnerID        gid.GID
		IncidentDate   *time.Time
		Status         *coredata.IncidentRegistryStatus
		Priority       *coredata.IncidentRegistryPriority
		Severity       *coredata.IncidentRegistrySeverity
		Category       *string
	}

	UpdateIncidentRegistryRequest struct {
		ID           gid.GID
		ReferenceID  *string
		Title        *string
		Description  **string
		Source       **string
		OwnerID      *gid.GID
		IncidentDate **time.Time
		ResolvedDate **time.Time
		Status       *coredata.IncidentRegistryStatus
		Priority     *coredata.IncidentRegistryPriority
		Severity     *coredata.IncidentRegistrySeverity
		Category     **string
	}
)

func (s IncidentRegistryService) Get(
	ctx context.Context,
	incidentRegistryID gid.GID,
) (*coredata.IncidentRegistry, error) {
	registry := &coredata.IncidentRegistry{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := registry.LoadByID(ctx, conn, s.svc.scope, incidentRegistryID); err != nil {
				return fmt.Errorf("cannot load incident registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *IncidentRegistryService) Create(
	ctx context.Context,
	req *CreateIncidentRegistryRequest,
) (*coredata.IncidentRegistry, error) {
	now := time.Now()

	status := coredata.IncidentRegistryStatusOpen
	if req.Status != nil {
		status = *req.Status
	}

	priority := coredata.IncidentRegistryPriorityMedium
	if req.Priority != nil {
		priority = *req.Priority
	}

	severity := coredata.IncidentRegistrySeverityMedium
	if req.Severity != nil {
		severity = *req.Severity
	}

	registry := &coredata.IncidentRegistry{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.IncidentRegistryEntityType),
		OrganizationID: req.OrganizationID,
		ReferenceID:    req.ReferenceID,
		Title:          req.Title,
		Description:    req.Description,
		Source:         req.Source,
		OwnerID:        req.OwnerID,
		IncidentDate:   req.IncidentDate,
		Status:         status,
		Priority:       priority,
		Severity:       severity,
		Category:       req.Category,
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

			owner := &coredata.People{}
			if err := owner.LoadByID(ctx, conn, s.svc.scope, req.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner: %w", err)
			}

			if err := registry.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert incident registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *IncidentRegistryService) Update(
	ctx context.Context,
	req *UpdateIncidentRegistryRequest,
) (*coredata.IncidentRegistry, error) {
	registry := &coredata.IncidentRegistry{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := registry.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load incident registry: %w", err)
			}

			if req.ReferenceID != nil {
				registry.ReferenceID = *req.ReferenceID
			}

			if req.Title != nil {
				registry.Title = *req.Title
			}

			if req.Description != nil {
				registry.Description = *req.Description
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

			if req.IncidentDate != nil {
				registry.IncidentDate = *req.IncidentDate
			}

			if req.ResolvedDate != nil {
				registry.ResolvedDate = *req.ResolvedDate
			}

			if req.Status != nil {
				registry.Status = *req.Status
				// Auto-set resolved date when status changes to RESOLVED or CLOSED
				if (*req.Status == coredata.IncidentRegistryStatusResolved || *req.Status == coredata.IncidentRegistryStatusClosed) && registry.ResolvedDate == nil {
					now := time.Now()
					registry.ResolvedDate = &now
				}
				// Clear resolved date if status is not RESOLVED or CLOSED
				if *req.Status != coredata.IncidentRegistryStatusResolved && *req.Status != coredata.IncidentRegistryStatusClosed {
					registry.ResolvedDate = nil
				}
			}

			if req.Priority != nil {
				registry.Priority = *req.Priority
			}

			if req.Severity != nil {
				registry.Severity = *req.Severity
			}

			if req.Category != nil {
				registry.Category = *req.Category
			}

			registry.UpdatedAt = time.Now()

			if err := registry.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update incident registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *IncidentRegistryService) Delete(
	ctx context.Context,
	incidentRegistryID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			registry := &coredata.IncidentRegistry{}
			if err := registry.LoadByID(ctx, conn, s.svc.scope, incidentRegistryID); err != nil {
				return fmt.Errorf("cannot load incident registry: %w", err)
			}

			if err := registry.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete incident registry: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s IncidentRegistryService) CountByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			registries := coredata.IncidentRegistries{}
			count, err = registries.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count incident registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s IncidentRegistryService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.IncidentRegistryOrderField],
) (*page.Page[*coredata.IncidentRegistry, coredata.IncidentRegistryOrderField], error) {
	var registries coredata.IncidentRegistries

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := registries.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load incident registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(registries, cursor), nil
}