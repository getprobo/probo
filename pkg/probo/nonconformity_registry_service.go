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

type NonconformityRegistryService struct {
	svc *TenantService
}

type (
	CreateNonconformityRegistryRequest struct {
		OrganizationID     gid.GID
		ReferenceID        string
		Description        *string
		AuditID            gid.GID
		DateIdentified     *time.Time
		RootCause          string
		CorrectiveAction   *string
		OwnerID            gid.GID
		DueDate            *time.Time
		Status             *coredata.NonconformityRegistryStatus
		EffectivenessCheck *string
	}

	UpdateNonconformityRegistryRequest struct {
		ID                 gid.GID
		ReferenceID        *string
		Description        **string
		DateIdentified     **time.Time
		RootCause          *string
		CorrectiveAction   **string
		OwnerID            *gid.GID
		AuditID            *gid.GID
		DueDate            **time.Time
		Status             *coredata.NonconformityRegistryStatus
		EffectivenessCheck **string
	}
)

func (s NonconformityRegistryService) Get(
	ctx context.Context,
	nonconformityRegistryID gid.GID,
) (*coredata.NonconformityRegistry, error) {
	registry := &coredata.NonconformityRegistry{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return registry.LoadByID(ctx, conn, s.svc.scope, nonconformityRegistryID)
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *NonconformityRegistryService) Create(
	ctx context.Context,
	req *CreateNonconformityRegistryRequest,
) (*coredata.NonconformityRegistry, error) {
	now := time.Now()

	registry := &coredata.NonconformityRegistry{
		ID:                 gid.New(s.svc.scope.GetTenantID(), coredata.NonconformityRegistryEntityType),
		OrganizationID:     req.OrganizationID,
		ReferenceID:        req.ReferenceID,
		Description:        req.Description,
		AuditID:            req.AuditID,
		DateIdentified:     req.DateIdentified,
		RootCause:          req.RootCause,
		CorrectiveAction:   req.CorrectiveAction,
		OwnerID:            req.OwnerID,
		DueDate:            req.DueDate,
		Status:             coredata.NonconformityRegistryStatusOpen,
		EffectivenessCheck: req.EffectivenessCheck,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if req.Status != nil {
		registry.Status = *req.Status
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

			people := &coredata.People{}
			if err := people.LoadByID(ctx, conn, s.svc.scope, req.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner: %w", err)
			}

			if err := registry.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert nonconformity registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s *NonconformityRegistryService) Update(
	ctx context.Context,
	req *UpdateNonconformityRegistryRequest,
) (*coredata.NonconformityRegistry, error) {
	registry := &coredata.NonconformityRegistry{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := registry.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load nonconformity registry: %w", err)
			}

			if req.ReferenceID != nil {
				registry.ReferenceID = *req.ReferenceID
			}
			if req.Description != nil {
				registry.Description = *req.Description
			}
			if req.DateIdentified != nil {
				registry.DateIdentified = *req.DateIdentified
			}
			if req.RootCause != nil {
				registry.RootCause = *req.RootCause
			}
			if req.CorrectiveAction != nil {
				registry.CorrectiveAction = *req.CorrectiveAction
			}
			if req.OwnerID != nil {
				registry.OwnerID = *req.OwnerID
			}
			if req.AuditID != nil {
				registry.AuditID = *req.AuditID
			}
			if req.DueDate != nil {
				registry.DueDate = *req.DueDate
			}
			if req.Status != nil {
				registry.Status = *req.Status
			}
			if req.EffectivenessCheck != nil {
				registry.EffectivenessCheck = *req.EffectivenessCheck
			}

			registry.UpdatedAt = time.Now()

			if err := registry.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update nonconformity registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return registry, nil
}

func (s NonconformityRegistryService) Delete(
	ctx context.Context,
	nonconformityRegistryID gid.GID,
) error {
	registry := coredata.NonconformityRegistry{ID: nonconformityRegistryID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := registry.Delete(ctx, conn, s.svc.scope)
			if err != nil {
				return fmt.Errorf("cannot delete nonconformity registry: %w", err)
			}
			return nil
		},
	)
}

func (s NonconformityRegistryService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.NonconformityRegistryOrderField],
	filter *coredata.NonconformityRegistryFilter,
) (*page.Page[*coredata.NonconformityRegistry, coredata.NonconformityRegistryOrderField], error) {
	var registries coredata.NonconformityRegistries

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := registries.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load nonconformity registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(registries, cursor), nil
}

func (s NonconformityRegistryService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.NonconformityRegistryFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			registries := coredata.NonconformityRegistries{}
			count, err = registries.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count nonconformity registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}
