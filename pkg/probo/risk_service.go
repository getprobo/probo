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

type (
	RiskService struct {
		svc *TenantService
	}

	CreateRiskRequest struct {
		OrganizationID     gid.GID
		Name               string
		Description        *string
		Category           string
		Treatment          coredata.RiskTreatment
		OwnerID            *gid.GID
		InherentLikelihood int
		InherentImpact     int
		ResidualLikelihood *int
		ResidualImpact     *int
		Note               *string
	}

	UpdateRiskRequest struct {
		ID                 gid.GID
		Name               *string
		Description        **string
		Category           *string
		Treatment          *coredata.RiskTreatment
		OwnerID            **gid.GID
		InherentLikelihood *int
		InherentImpact     *int
		ResidualLikelihood *int
		ResidualImpact     *int
		Note               *string
	}
)

func (crr *CreateRiskRequest) Validate() error {
	v := validator.New()

	v.Check(crr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(crr.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(crr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(crr.Category, "category", validator.Required(), validator.SafeText(TitleMaxLength))
	v.Check(crr.Treatment, "treatment", validator.Required(), validator.OneOfSlice(coredata.RiskTreatments()))
	v.Check(crr.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(crr.InherentLikelihood, "inherent_likelihood", validator.Required(), validator.Min(1), validator.Max(5))
	v.Check(crr.InherentImpact, "inherent_impact", validator.Required(), validator.Min(1), validator.Max(5))
	v.Check(crr.ResidualLikelihood, "residual_likelihood", validator.Min(1), validator.Max(5))
	v.Check(crr.ResidualImpact, "residual_impact", validator.Min(1), validator.Max(5))
	v.Check(crr.Note, "note", validator.SafeText(TitleMaxLength))

	return v.Error()
}

func (urr *UpdateRiskRequest) Validate() error {
	v := validator.New()

	v.Check(urr.ID, "id", validator.Required(), validator.GID(coredata.RiskEntityType))
	v.Check(urr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(urr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(urr.Category, "category", validator.SafeText(TitleMaxLength))
	v.Check(urr.Treatment, "treatment", validator.OneOfSlice(coredata.RiskTreatments()))
	v.Check(urr.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(urr.InherentLikelihood, "inherent_likelihood", validator.Min(1), validator.Max(5))
	v.Check(urr.InherentImpact, "inherent_impact", validator.Min(1), validator.Max(5))
	v.Check(urr.ResidualLikelihood, "residual_likelihood", validator.Min(1), validator.Max(5))
	v.Check(urr.ResidualImpact, "residual_impact", validator.Min(1), validator.Max(5))
	v.Check(urr.Note, "note", validator.SafeText(TitleMaxLength))

	return v.Error()
}

func (s RiskService) CountForMeasureID(
	ctx context.Context,
	measureID gid.GID,
	filter *coredata.RiskFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			risks := &coredata.Risks{}
			count, err = risks.CountByMeasureID(ctx, conn, s.svc.scope, measureID, filter)
			if err != nil {
				return fmt.Errorf("cannot count risks: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, fmt.Errorf("cannot count risks: %w", err)
	}

	return count, nil
}

func (s RiskService) ListForMeasureID(
	ctx context.Context,
	measureID gid.GID,
	cursor *page.Cursor[coredata.RiskOrderField],
	filter *coredata.RiskFilter,
) (*page.Page[*coredata.Risk, coredata.RiskOrderField], error) {
	var risks coredata.Risks

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return risks.LoadByMeasureID(ctx, conn, s.svc.scope, measureID, cursor, filter)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot list risks: %w", err)
	}

	return page.NewPage(risks, cursor), nil
}

func (s RiskService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.RiskFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			risks := &coredata.Risks{}
			count, err = risks.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count risks: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, fmt.Errorf("cannot count risks: %w", err)
	}

	return count, nil
}

func (s RiskService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.RiskOrderField],
	filter *coredata.RiskFilter,
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
				filter,
			)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot list risks: %w", err)
	}

	return page.NewPage(risks, cursor), nil
}

func (s RiskService) CreateDocumentMapping(
	ctx context.Context,
	riskID gid.GID,
	documentID gid.GID,
) (*coredata.Risk, *coredata.Document, error) {
	risk := &coredata.Risk{}
	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := risk.LoadByID(ctx, conn, s.svc.scope, riskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			riskDocument := &coredata.RiskDocument{
				RiskID:         risk.ID,
				DocumentID:     document.ID,
				OrganizationID: risk.OrganizationID,
				CreatedAt:      time.Now(),
			}

			return riskDocument.Insert(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot create risk document mapping: %w", err)
	}

	return risk, document, nil
}

func (s RiskService) DeleteDocumentMapping(
	ctx context.Context,
	riskID gid.GID,
	documentID gid.GID,
) (*coredata.Risk, *coredata.Document, error) {
	riskDocument := &coredata.RiskDocument{}
	risk := &coredata.Risk{}
	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := risk.LoadByID(ctx, conn, s.svc.scope, riskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			if err := document.LoadByID(ctx, conn, s.svc.scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			return riskDocument.Delete(ctx, conn, s.svc.scope, risk.ID, document.ID)
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot delete risk document mapping: %w", err)
	}

	return risk, document, nil
}

func (s RiskService) CreateMeasureMapping(
	ctx context.Context,
	riskID gid.GID,
	measureID gid.GID,
) (*coredata.Risk, *coredata.Measure, error) {
	measure := &coredata.Measure{}
	risk := &coredata.Risk{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := risk.LoadByID(ctx, conn, s.svc.scope, riskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			if err := measure.LoadByID(ctx, conn, s.svc.scope, measureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			riskMeasure := &coredata.RiskMeasure{
				RiskID:         risk.ID,
				MeasureID:      measure.ID,
				OrganizationID: risk.OrganizationID,
				CreatedAt:      time.Now(),
			}

			return riskMeasure.Insert(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot create risk measure mapping: %w", err)
	}

	return risk, measure, nil
}

func (s RiskService) DeleteMeasureMapping(
	ctx context.Context,
	riskID gid.GID,
	measureID gid.GID,
) (*coredata.Risk, *coredata.Measure, error) {
	risk := &coredata.Risk{}
	measure := &coredata.Measure{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := risk.LoadByID(ctx, conn, s.svc.scope, riskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			if err := measure.LoadByID(ctx, conn, s.svc.scope, measureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			riskMeasure := &coredata.RiskMeasure{
				RiskID:         riskID,
				MeasureID:      measureID,
				OrganizationID: risk.OrganizationID,
				CreatedAt:      time.Now(),
			}

			return riskMeasure.Delete(ctx, conn, s.svc.scope, risk.ID, measure.ID)
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot delete risk measure mapping: %w", err)
	}

	return risk, measure, nil
}

func (s RiskService) CreateObligationMapping(
	ctx context.Context,
	riskID gid.GID,
	obligationID gid.GID,
) (*coredata.Risk, *coredata.Obligation, error) {
	risk := &coredata.Risk{}
	obligation := &coredata.Obligation{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := risk.LoadByID(ctx, conn, s.svc.scope, riskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			if err := obligation.LoadByID(ctx, conn, s.svc.scope, obligationID); err != nil {
				return fmt.Errorf("cannot load obligation: %w", err)
			}

			riskObligation := &coredata.RiskObligation{
				RiskID:         risk.ID,
				ObligationID:   obligation.ID,
				OrganizationID: risk.OrganizationID,
				CreatedAt:      time.Now(),
			}

			return riskObligation.Insert(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot create risk obligation mapping: %w", err)
	}

	return risk, obligation, nil
}

func (s RiskService) DeleteObligationMapping(
	ctx context.Context,
	riskID gid.GID,
	obligationID gid.GID,
) (*coredata.Risk, *coredata.Obligation, error) {
	riskObligation := &coredata.RiskObligation{}
	risk := &coredata.Risk{}
	obligation := &coredata.Obligation{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := risk.LoadByID(ctx, conn, s.svc.scope, riskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			if err := obligation.LoadByID(ctx, conn, s.svc.scope, obligationID); err != nil {
				return fmt.Errorf("cannot load obligation: %w", err)
			}

			riskObligation.RiskID = risk.ID
			riskObligation.ObligationID = obligation.ID

			return riskObligation.Delete(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot delete risk obligation mapping: %w", err)
	}

	return risk, obligation, nil
}

func (s RiskService) Create(
	ctx context.Context,
	req CreateRiskRequest,
) (*coredata.Risk, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	owner := coredata.MembershipProfile{}
	organization := coredata.Organization{}

	risk := &coredata.Risk{
		ID:                 gid.New(s.svc.scope.GetTenantID(), coredata.RiskEntityType),
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

	if req.Note != nil {
		risk.Note = *req.Note
	}

	if req.ResidualLikelihood != nil {
		risk.ResidualLikelihood = *req.ResidualLikelihood
	}

	if req.ResidualImpact != nil {
		risk.ResidualImpact = *req.ResidualImpact
	}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if req.OwnerID != nil {
				if err := owner.LoadByID(ctx, conn, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}
			}

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
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

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
				if *req.OwnerID != nil {
					owner := coredata.MembershipProfile{}
					if err := owner.LoadByID(ctx, conn, s.svc.scope, **req.OwnerID); err != nil {
						return fmt.Errorf("cannot load owner profile: %w", err)
					}
					risk.OwnerID = *req.OwnerID
				} else {
					risk.OwnerID = nil
				}
			}

			if req.Category != nil {
				risk.Category = *req.Category
			}

			if req.Note != nil {
				risk.Note = *req.Note
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
	risk := &coredata.Risk{}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return risk.Delete(ctx, conn, s.svc.scope, riskID)
		},
	)
}
