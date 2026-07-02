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

type TransferImpactAssessmentService struct {
	svc *Service
}

type (
	CreateTransferImpactAssessmentRequest struct {
		ProcessingActivityID  gid.GID
		DataSubjects          *string
		LegalMechanism        *string
		Transfer              *string
		LocalLawRisk          *string
		SupplementaryMeasures *string
	}

	UpdateTransferImpactAssessmentRequest struct {
		ID                    gid.GID
		DataSubjects          **string
		LegalMechanism        **string
		Transfer              **string
		LocalLawRisk          **string
		SupplementaryMeasures **string
	}
)

func (req *CreateTransferImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ProcessingActivityID, "processing_activity_id", validator.Required(), validator.GID(coredata.ProcessingActivityEntityType))
	v.Check(req.DataSubjects, "data_subjects", validator.SafeText(ContentMaxLength))
	v.Check(req.LegalMechanism, "legal_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(req.Transfer, "transfer", validator.SafeText(ContentMaxLength))
	v.Check(req.LocalLawRisk, "local_law_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.SupplementaryMeasures, "supplementary_measures", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (req *UpdateTransferImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.TransferImpactAssessmentEntityType))
	v.Check(req.DataSubjects, "data_subjects", validator.SafeText(ContentMaxLength))
	v.Check(req.LegalMechanism, "legal_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(req.Transfer, "transfer", validator.SafeText(ContentMaxLength))
	v.Check(req.LocalLawRisk, "local_law_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.SupplementaryMeasures, "supplementary_measures", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s TransferImpactAssessmentService) Get(
	ctx context.Context, scope coredata.Scoper,
	tiaID gid.GID,
) (*coredata.TransferImpactAssessment, error) {
	tia := &coredata.TransferImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tia.LoadByID(ctx, conn, scope, tiaID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s TransferImpactAssessmentService) GetByProcessingActivityID(
	ctx context.Context, scope coredata.Scoper,
	processingActivityID gid.GID,
) (*coredata.TransferImpactAssessment, error) {
	tia := &coredata.TransferImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tia.LoadByProcessingActivityID(ctx, conn, scope, processingActivityID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s TransferImpactAssessmentService) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.TransferImpactAssessmentOrderField],
) (*page.Page[*coredata.TransferImpactAssessment, coredata.TransferImpactAssessmentOrderField], error) {
	var tias coredata.TransferImpactAssessments

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := tias.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load transfer impact assessments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(tias, cursor), nil
}

func (s TransferImpactAssessmentService) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			tias := coredata.TransferImpactAssessments{}
			count, err = tias.CountByOrganizationID(ctx, conn, scope, organizationID)

			return err
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *TransferImpactAssessmentService) Create(
	ctx context.Context, scope coredata.Scoper,
	req *CreateTransferImpactAssessmentRequest,
) (*coredata.TransferImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	tia := &coredata.TransferImpactAssessment{
		ID:                    gid.New(scope.GetTenantID(), coredata.TransferImpactAssessmentEntityType),
		ProcessingActivityID:  req.ProcessingActivityID,
		DataSubjects:          req.DataSubjects,
		LegalMechanism:        req.LegalMechanism,
		Transfer:              req.Transfer,
		LocalLawRisk:          req.LocalLawRisk,
		SupplementaryMeasures: req.SupplementaryMeasures,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			processingActivity := &coredata.ProcessingActivity{}
			if err := processingActivity.LoadByID(ctx, conn, scope, req.ProcessingActivityID); err != nil {
				return fmt.Errorf("cannot load processing activity: %w", err)
			}

			tia.OrganizationID = processingActivity.OrganizationID

			if err := tia.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert transfer impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s *TransferImpactAssessmentService) Update(
	ctx context.Context, scope coredata.Scoper,
	req *UpdateTransferImpactAssessmentRequest,
) (*coredata.TransferImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	tia := &coredata.TransferImpactAssessment{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := tia.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			if req.DataSubjects != nil {
				tia.DataSubjects = *req.DataSubjects
			}

			if req.LegalMechanism != nil {
				tia.LegalMechanism = *req.LegalMechanism
			}

			if req.Transfer != nil {
				tia.Transfer = *req.Transfer
			}

			if req.LocalLawRisk != nil {
				tia.LocalLawRisk = *req.LocalLawRisk
			}

			if req.SupplementaryMeasures != nil {
				tia.SupplementaryMeasures = *req.SupplementaryMeasures
			}

			tia.UpdatedAt = time.Now()

			if err := tia.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update transfer impact assessment: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s *TransferImpactAssessmentService) Delete(
	ctx context.Context, scope coredata.Scoper,
	tiaID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			tia := &coredata.TransferImpactAssessment{}
			if err := tia.LoadByID(ctx, conn, scope, tiaID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			if err := tia.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete transfer impact assessment: %w", err)
			}

			return nil
		},
	)

	return err
}
