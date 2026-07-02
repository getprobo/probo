// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

type DataProtectionImpactAssessmentService struct {
	svc *Service
}

type (
	CreateDataProtectionImpactAssessmentRequest struct {
		ProcessingActivityID        gid.GID
		Description                 *string
		NecessityAndProportionality *string
		PotentialRisk               *string
		Mitigations                 *string
		ResidualRisk                *coredata.DataProtectionImpactAssessmentResidualRisk
	}

	UpdateDataProtectionImpactAssessmentRequest struct {
		ID                          gid.GID
		Description                 **string
		NecessityAndProportionality **string
		PotentialRisk               **string
		Mitigations                 **string
		ResidualRisk                *coredata.DataProtectionImpactAssessmentResidualRisk
	}
)

func (req *CreateDataProtectionImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ProcessingActivityID, "processing_activity_id", validator.Required(), validator.GID(coredata.ProcessingActivityEntityType))
	v.Check(req.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(req.NecessityAndProportionality, "necessity_and_proportionality", validator.SafeText(ContentMaxLength))
	v.Check(req.PotentialRisk, "potential_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.Mitigations, "mitigations", validator.SafeText(ContentMaxLength))
	v.Check(req.ResidualRisk, "residual_risk", validator.OneOfSlice(coredata.DataProtectionImpactAssessmentResidualRisks()))

	return v.Error()
}

func (req *UpdateDataProtectionImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.DataProtectionImpactAssessmentEntityType))
	v.Check(req.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(req.NecessityAndProportionality, "necessity_and_proportionality", validator.SafeText(ContentMaxLength))
	v.Check(req.PotentialRisk, "potential_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.Mitigations, "mitigations", validator.SafeText(ContentMaxLength))
	v.Check(req.ResidualRisk, "residual_risk", validator.OneOfSlice(coredata.DataProtectionImpactAssessmentResidualRisks()))

	return v.Error()
}

func (s DataProtectionImpactAssessmentService) Get(
	ctx context.Context, scope coredata.Scoper,
	dpiaID gid.GID,
) (*coredata.DataProtectionImpactAssessment, error) {
	dpia := &coredata.DataProtectionImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := dpia.LoadByID(ctx, conn, scope, dpiaID); err != nil {
				return fmt.Errorf("cannot load data protection impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s DataProtectionImpactAssessmentService) GetByProcessingActivityID(
	ctx context.Context, scope coredata.Scoper,
	processingActivityID gid.GID,
) (*coredata.DataProtectionImpactAssessment, error) {
	dpia := &coredata.DataProtectionImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := dpia.LoadByProcessingActivityID(ctx, conn, scope, processingActivityID); err != nil {
				return fmt.Errorf("cannot load data protection impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s DataProtectionImpactAssessmentService) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DataProtectionImpactAssessmentOrderField],
) (*page.Page[*coredata.DataProtectionImpactAssessment, coredata.DataProtectionImpactAssessmentOrderField], error) {
	var dpias coredata.DataProtectionImpactAssessments

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := dpias.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load data protection impact assessments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(dpias, cursor), nil
}

func (s DataProtectionImpactAssessmentService) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			dpias := coredata.DataProtectionImpactAssessments{}
			count, err = dpias.CountByOrganizationID(ctx, conn, scope, organizationID)

			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DataProtectionImpactAssessmentService) Create(
	ctx context.Context, scope coredata.Scoper,
	req *CreateDataProtectionImpactAssessmentRequest,
) (*coredata.DataProtectionImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	dpia := &coredata.DataProtectionImpactAssessment{
		ID:                          gid.New(scope.GetTenantID(), coredata.DataProtectionImpactAssessmentEntityType),
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
		func(ctx context.Context, conn pg.Tx) error {
			processingActivity := &coredata.ProcessingActivity{}
			if err := processingActivity.LoadByID(ctx, conn, scope, req.ProcessingActivityID); err != nil {
				return fmt.Errorf("cannot load processing activity: %w", err)
			}

			dpia.OrganizationID = processingActivity.OrganizationID

			if err := dpia.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert data protection impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s *DataProtectionImpactAssessmentService) Update(
	ctx context.Context, scope coredata.Scoper,
	req *UpdateDataProtectionImpactAssessmentRequest,
) (*coredata.DataProtectionImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	dpia := &coredata.DataProtectionImpactAssessment{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := dpia.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load data protection impact assessment: %w", err)
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

			if err := dpia.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update data protection impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s *DataProtectionImpactAssessmentService) Delete(
	ctx context.Context, scope coredata.Scoper,
	dpiaID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			dpia := &coredata.DataProtectionImpactAssessment{}
			if err := dpia.LoadByID(ctx, conn, scope, dpiaID); err != nil {
				return fmt.Errorf("cannot load data protection impact assessment: %w", err)
			}

			if err := dpia.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete data protection impact assessment: %w", err)
			}

			return nil
		},
	)

	return err
}
