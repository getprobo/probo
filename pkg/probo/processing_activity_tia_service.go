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

type ProcessingActivityTIAService struct {
	svc *TenantService
}

type (
	CreateProcessingActivityTIARequest struct {
		ProcessingActivityID  gid.GID
		DataSubjects          *string
		LegalMechanism        *string
		Transfer              *string
		LocalLawRisk          *string
		SupplementaryMeasures *string
	}

	UpdateProcessingActivityTIARequest struct {
		ID                    gid.GID
		DataSubjects          **string
		LegalMechanism        **string
		Transfer              **string
		LocalLawRisk          **string
		SupplementaryMeasures **string
	}
)

func (req *CreateProcessingActivityTIARequest) Validate() error {
	v := validator.New()

	v.Check(req.ProcessingActivityID, "processing_activity_id", validator.Required(), validator.GID(coredata.ProcessingActivityEntityType))
	v.Check(req.DataSubjects, "data_subjects", validator.SafeText(ContentMaxLength))
	v.Check(req.LegalMechanism, "legal_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(req.Transfer, "transfer", validator.SafeText(ContentMaxLength))
	v.Check(req.LocalLawRisk, "local_law_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.SupplementaryMeasures, "supplementary_measures", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (req *UpdateProcessingActivityTIARequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.ProcessingActivityTIAEntityType))
	v.Check(req.DataSubjects, "data_subjects", validator.SafeText(ContentMaxLength))
	v.Check(req.LegalMechanism, "legal_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(req.Transfer, "transfer", validator.SafeText(ContentMaxLength))
	v.Check(req.LocalLawRisk, "local_law_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.SupplementaryMeasures, "supplementary_measures", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s ProcessingActivityTIAService) Get(
	ctx context.Context,
	tiaID gid.GID,
) (*coredata.ProcessingActivityTIA, error) {
	tia := &coredata.ProcessingActivityTIA{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := tia.LoadByID(ctx, conn, s.svc.scope, tiaID); err != nil {
				return fmt.Errorf("cannot load processing activity tia: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s ProcessingActivityTIAService) GetByProcessingActivityID(
	ctx context.Context,
	processingActivityID gid.GID,
) (*coredata.ProcessingActivityTIA, error) {
	tia := &coredata.ProcessingActivityTIA{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := tia.LoadByProcessingActivityID(ctx, conn, s.svc.scope, processingActivityID); err != nil {
				return fmt.Errorf("cannot load processing activity tia: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s ProcessingActivityTIAService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ProcessingActivityTIAOrderField],
) (*page.Page[*coredata.ProcessingActivityTIA, coredata.ProcessingActivityTIAOrderField], error) {
	var tias coredata.ProcessingActivityTIAs

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := tias.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load processing activity tias: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(tias, cursor), nil
}

func (s ProcessingActivityTIAService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			tias := coredata.ProcessingActivityTIAs{}
			count, err = tias.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			return err
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *ProcessingActivityTIAService) Create(
	ctx context.Context,
	req *CreateProcessingActivityTIARequest,
) (*coredata.ProcessingActivityTIA, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	tia := &coredata.ProcessingActivityTIA{
		ID:                    gid.New(s.svc.scope.GetTenantID(), coredata.ProcessingActivityTIAEntityType),
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
		func(conn pg.Conn) error {
			processingActivity := &coredata.ProcessingActivity{}
			if err := processingActivity.LoadByID(ctx, conn, s.svc.scope, req.ProcessingActivityID); err != nil {
				return fmt.Errorf("cannot load processing activity: %w", err)
			}

			tia.OrganizationID = processingActivity.OrganizationID

			if err := tia.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert processing activity tia: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s *ProcessingActivityTIAService) Update(
	ctx context.Context,
	req *UpdateProcessingActivityTIARequest,
) (*coredata.ProcessingActivityTIA, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	tia := &coredata.ProcessingActivityTIA{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := tia.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load processing activity tia: %w", err)
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

			if err := tia.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update processing activity tia: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s *ProcessingActivityTIAService) Delete(
	ctx context.Context,
	tiaID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			tia := &coredata.ProcessingActivityTIA{}
			if err := tia.LoadByID(ctx, conn, s.svc.scope, tiaID); err != nil {
				return fmt.Errorf("cannot load processing activity tia: %w", err)
			}

			if err := tia.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete processing activity tia: %w", err)
			}

			return nil
		},
	)

	return err
}
