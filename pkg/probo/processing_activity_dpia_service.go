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

type ProcessingActivityDPIAService struct {
	svc *TenantService
}

type (
	CreateProcessingActivityDPIARequest struct {
		ProcessingActivityID        gid.GID
		Description                 *string
		NecessityAndProportionality *string
		PotentialRisk               *string
		Mitigations                 *string
		ResidualRisk                *coredata.ProcessingActivityDPIAResidualRisk
	}

	UpdateProcessingActivityDPIARequest struct {
		ID                          gid.GID
		Description                 **string
		NecessityAndProportionality **string
		PotentialRisk               **string
		Mitigations                 **string
		ResidualRisk                *coredata.ProcessingActivityDPIAResidualRisk
	}
)

func (req *CreateProcessingActivityDPIARequest) Validate() error {
	v := validator.New()

	v.Check(req.ProcessingActivityID, "processing_activity_id", validator.Required(), validator.GID(coredata.ProcessingActivityEntityType))
	v.Check(req.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(req.NecessityAndProportionality, "necessity_and_proportionality", validator.SafeText(ContentMaxLength))
	v.Check(req.PotentialRisk, "potential_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.Mitigations, "mitigations", validator.SafeText(ContentMaxLength))
	v.Check(req.ResidualRisk, "residual_risk", validator.OneOfSlice(coredata.ProcessingActivityDPIAResidualRisks()))

	return v.Error()
}

func (req *UpdateProcessingActivityDPIARequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.ProcessingActivityDPIAEntityType))
	v.Check(req.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(req.NecessityAndProportionality, "necessity_and_proportionality", validator.SafeText(ContentMaxLength))
	v.Check(req.PotentialRisk, "potential_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.Mitigations, "mitigations", validator.SafeText(ContentMaxLength))
	v.Check(req.ResidualRisk, "residual_risk", validator.OneOfSlice(coredata.ProcessingActivityDPIAResidualRisks()))

	return v.Error()
}

func (s ProcessingActivityDPIAService) Get(
	ctx context.Context,
	dpiaID gid.GID,
) (*coredata.ProcessingActivityDPIA, error) {
	dpia := &coredata.ProcessingActivityDPIA{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := dpia.LoadByID(ctx, conn, s.svc.scope, dpiaID); err != nil {
				return fmt.Errorf("cannot load processing activity dpia: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s ProcessingActivityDPIAService) GetByProcessingActivityID(
	ctx context.Context,
	processingActivityID gid.GID,
) (*coredata.ProcessingActivityDPIA, error) {
	dpia := &coredata.ProcessingActivityDPIA{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := dpia.LoadByProcessingActivityID(ctx, conn, s.svc.scope, processingActivityID); err != nil {
				return fmt.Errorf("cannot load processing activity dpia: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s ProcessingActivityDPIAService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ProcessingActivityDPIAOrderField],
	filter *coredata.ProcessingActivityDPIAFilter,
) (*page.Page[*coredata.ProcessingActivityDPIA, coredata.ProcessingActivityDPIAOrderField], error) {
	var dpias coredata.ProcessingActivityDPIAs

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := dpias.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load processing activity dpias: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(dpias, cursor), nil
}

func (s ProcessingActivityDPIAService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.ProcessingActivityDPIAFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			dpias := coredata.ProcessingActivityDPIAs{}
			count, err = dpias.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			return err
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *ProcessingActivityDPIAService) Create(
	ctx context.Context,
	req *CreateProcessingActivityDPIARequest,
) (*coredata.ProcessingActivityDPIA, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	dpia := &coredata.ProcessingActivityDPIA{
		ID:                          gid.New(s.svc.scope.GetTenantID(), coredata.ProcessingActivityDPIAEntityType),
		ProcessingActivityID:        req.ProcessingActivityID,
		Description:                 req.Description,
		NecessityAndProportionality: req.NecessityAndProportionality,
		PotentialRisk:               req.PotentialRisk,
		Mitigations:                 req.Mitigations,
		ResidualRisk:                req.ResidualRisk,
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			processingActivity := &coredata.ProcessingActivity{}
			if err := processingActivity.LoadByID(ctx, conn, s.svc.scope, req.ProcessingActivityID); err != nil {
				return fmt.Errorf("cannot load processing activity: %w", err)
			}

			dpia.OrganizationID = processingActivity.OrganizationID

			if err := dpia.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert processing activity dpia: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s *ProcessingActivityDPIAService) Update(
	ctx context.Context,
	req *UpdateProcessingActivityDPIARequest,
) (*coredata.ProcessingActivityDPIA, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	dpia := &coredata.ProcessingActivityDPIA{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := dpia.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load processing activity dpia: %w", err)
			}

			if req.Description != nil {
				dpia.Description = *req.Description
			}

			if req.NecessityAndProportionality != nil {
				dpia.NecessityAndProportionality = *req.NecessityAndProportionality
			}

			if req.PotentialRisk != nil {
				dpia.PotentialRisk = *req.PotentialRisk
			}

			if req.Mitigations != nil {
				dpia.Mitigations = *req.Mitigations
			}

			if req.ResidualRisk != nil {
				dpia.ResidualRisk = req.ResidualRisk
			}

			dpia.UpdatedAt = time.Now()

			if err := dpia.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update processing activity dpia: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s *ProcessingActivityDPIAService) Delete(
	ctx context.Context,
	dpiaID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			dpia := &coredata.ProcessingActivityDPIA{}
			if err := dpia.LoadByID(ctx, conn, s.svc.scope, dpiaID); err != nil {
				return fmt.Errorf("cannot load processing activity dpia: %w", err)
			}

			if err := dpia.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete processing activity dpia: %w", err)
			}

			return nil
		},
	)

	return err
}
