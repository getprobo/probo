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

type (
	RiskService struct {
		svc *TenantService
	}

	CreateRiskRequest struct {
		OrganizationID     gid.GID
		Name               string
		Description        string
		Category           string
		Treatment          coredata.RiskTreatment
		OwnerID            *gid.GID
		InherentLikelihood int
		InherentImpact     int
		ResidualLikelihood *int
		ResidualImpact     *int
	}

	UpdateRiskRequest struct {
		ID                 gid.GID
		Name               *string
		Description        *string
		Category           *string
		Treatment          *coredata.RiskTreatment
		OwnerID            *gid.GID
		InherentLikelihood *int
		InherentImpact     *int
		ResidualLikelihood *int
		ResidualImpact     *int
	}
)

func (s RiskService) ListForMesureID(
	ctx context.Context,
	mesureID gid.GID,
	cursor *page.Cursor[coredata.RiskOrderField],
) (*page.Page[*coredata.Risk, coredata.RiskOrderField], error) {
	var risks coredata.Risks

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return risks.LoadByMesureID(ctx, conn, s.svc.scope, mesureID, cursor)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot list risks: %w", err)
	}

	return page.NewPage(risks, cursor), nil
}

func (s RiskService) CreatePolicyMapping(
	ctx context.Context,
	riskID gid.GID,
	policyID gid.GID,
) error {
	riskPolicy := &coredata.RiskPolicy{
		RiskID:    riskID,
		PolicyID:  policyID,
		TenantID:  s.svc.scope.GetTenantID(),
		CreatedAt: time.Now(),
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return riskPolicy.Insert(ctx, conn, s.svc.scope)
		},
	)
}

func (s RiskService) DeletePolicyMapping(
	ctx context.Context,
	riskID gid.GID,
	policyID gid.GID,
) error {
	riskPolicy := &coredata.RiskPolicy{
		RiskID:    riskID,
		PolicyID:  policyID,
		TenantID:  s.svc.scope.GetTenantID(),
		CreatedAt: time.Now(),
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return riskPolicy.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s RiskService) CreateMesureMapping(
	ctx context.Context,
	riskID gid.GID,
	mesureID gid.GID,
) error {
	riskMesure := &coredata.RiskMesure{
		RiskID:    riskID,
		MesureID:  mesureID,
		TenantID:  s.svc.scope.GetTenantID(),
		CreatedAt: time.Now(),
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return riskMesure.Insert(ctx, conn, s.svc.scope)
		},
	)
}

func (s RiskService) DeleteMesureMapping(
	ctx context.Context,
	riskID gid.GID,
	mesureID gid.GID,
) error {
	riskMesure := &coredata.RiskMesure{
		RiskID:    riskID,
		MesureID:  mesureID,
		TenantID:  s.svc.scope.GetTenantID(),
		CreatedAt: time.Now(),
	}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return riskMesure.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s RiskService) Create(
	ctx context.Context,
	req CreateRiskRequest,
) (*coredata.Risk, error) {
	now := time.Now()
	riskID, err := gid.NewGID(s.svc.scope.GetTenantID(), coredata.RiskEntityType)
	if err != nil {
		return nil, fmt.Errorf("cannot create global id: %w", err)
	}

	risk := &coredata.Risk{
		ID:                 riskID,
		OrganizationID:     req.OrganizationID,
		Name:               req.Name,
		Description:        req.Description,
		Category:           req.Category,
		OwnerID:            req.OwnerID,
		InherentLikelihood: req.InherentLikelihood,
		InherentImpact:     req.InherentImpact,
		Treatment:          req.Treatment,
		ResidualLikelihood: req.InherentLikelihood,
		ResidualImpact:     req.InherentImpact,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if req.ResidualLikelihood != nil {
		risk.ResidualLikelihood = *req.ResidualLikelihood
	}

	if req.ResidualImpact != nil {
		risk.ResidualImpact = *req.ResidualImpact
	}

	err = s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return risk.Insert(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot create risk: %w", err)
	}

	return risk, nil
}

func (s RiskService) Get(
	ctx context.Context,
	riskID gid.GID,
) (*coredata.Risk, error) {
	risk := &coredata.Risk{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return risk.LoadByID(ctx, conn, s.svc.scope, riskID)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot get risk: %w", err)
	}

	return risk, nil
}

func (s RiskService) Update(
	ctx context.Context,
	req UpdateRiskRequest,
) (*coredata.Risk, error) {
	risk := &coredata.Risk{ID: req.ID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := risk.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			if req.Name != nil {
				risk.Name = *req.Name
			}

			if req.Description != nil {
				risk.Description = *req.Description
			}

			if req.InherentLikelihood != nil {
				risk.InherentLikelihood = *req.InherentLikelihood
			}

			if req.InherentImpact != nil {
				risk.InherentImpact = *req.InherentImpact
			}

			if req.ResidualLikelihood != nil {
				risk.ResidualLikelihood = *req.ResidualLikelihood
			}

			if req.ResidualImpact != nil {
				risk.ResidualImpact = *req.ResidualImpact
			}

			if req.Treatment != nil {
				risk.Treatment = *req.Treatment
			}

			if req.OwnerID != nil {
				risk.OwnerID = req.OwnerID
			}

			if req.Category != nil {
				risk.Category = *req.Category
			}

			risk.UpdatedAt = time.Now()

			if err := risk.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update risk: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot update risk: %w", err)
	}

	return risk, nil
}

func (s RiskService) Delete(
	ctx context.Context,
	riskID gid.GID,
) error {
	risk := &coredata.Risk{ID: riskID}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return risk.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s RiskService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.RiskOrderField],
) (*page.Page[*coredata.Risk, coredata.RiskOrderField], error) {
	var risks coredata.Risks

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return risks.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
				cursor,
			)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot list risks: %w", err)
	}

	return page.NewPage(risks, cursor), nil
}
