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

type ProcessingActivityService struct {
	svc *TenantService
}

type (
	CreateProcessingActivityRequest struct {
		OrganizationID                 gid.GID
		Name                           string
		Purpose                        *string
		DataSubjectCategory            *string
		PersonalDataCategory           *string
		SpecialOrCriminalData          coredata.ProcessingActivitySpecialOrCriminalData
		ConsentEvidenceLink            *string
		LawfulBasis                    coredata.ProcessingActivityLawfulBasis
		Recipients                     *string
		Location                       *string
		InternationalTransfers         bool
		TransferSafeguards             *coredata.ProcessingActivityTransferSafeguards
		RetentionPeriod                *string
		SecurityMeasures               *string
		DataProtectionImpactAssessment coredata.ProcessingActivityDataProtectionImpactAssessment
		TransferImpactAssessment       coredata.ProcessingActivityTransferImpactAssessment
	}

	UpdateProcessingActivityRequest struct {
		ID                             gid.GID
		Name                           *string
		Purpose                        **string
		DataSubjectCategory            **string
		PersonalDataCategory           **string
		SpecialOrCriminalData          *coredata.ProcessingActivitySpecialOrCriminalData
		ConsentEvidenceLink            **string
		LawfulBasis                    *coredata.ProcessingActivityLawfulBasis
		Recipients                     **string
		Location                       **string
		InternationalTransfers         *bool
		TransferSafeguards             **coredata.ProcessingActivityTransferSafeguards
		RetentionPeriod                **string
		SecurityMeasures               **string
		DataProtectionImpactAssessment *coredata.ProcessingActivityDataProtectionImpactAssessment
		TransferImpactAssessment       *coredata.ProcessingActivityTransferImpactAssessment
	}
)

func (s ProcessingActivityService) Get(
	ctx context.Context,
	processingActivityID gid.GID,
) (*coredata.ProcessingActivity, error) {
	processingActivity := &coredata.ProcessingActivity{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return processingActivity.LoadByID(ctx, conn, s.svc.scope, processingActivityID)
		},
	)

	if err != nil {
		return nil, err
	}

	return processingActivity, nil
}

func (s *ProcessingActivityService) Create(
	ctx context.Context,
	req *CreateProcessingActivityRequest,
) (*coredata.ProcessingActivity, error) {
	now := time.Now()

	processingActivity := &coredata.ProcessingActivity{
		ID:                             gid.New(s.svc.scope.GetTenantID(), coredata.ProcessingActivityEntityType),
		OrganizationID:                 req.OrganizationID,
		Name:                           req.Name,
		Purpose:                        req.Purpose,
		DataSubjectCategory:            req.DataSubjectCategory,
		PersonalDataCategory:           req.PersonalDataCategory,
		SpecialOrCriminalData:          req.SpecialOrCriminalData,
		ConsentEvidenceLink:            req.ConsentEvidenceLink,
		LawfulBasis:                    req.LawfulBasis,
		Recipients:                     req.Recipients,
		Location:                       req.Location,
		InternationalTransfers:         req.InternationalTransfers,
		TransferSafeguards:             req.TransferSafeguards,
		RetentionPeriod:                req.RetentionPeriod,
		SecurityMeasures:               req.SecurityMeasures,
		DataProtectionImpactAssessment: req.DataProtectionImpactAssessment,
		TransferImpactAssessment:       req.TransferImpactAssessment,
		CreatedAt:                      now,
		UpdatedAt:                      now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := processingActivity.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert processing activity: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return processingActivity, nil
}

func (s *ProcessingActivityService) Update(
	ctx context.Context,
	req *UpdateProcessingActivityRequest,
) (*coredata.ProcessingActivity, error) {
	processingActivity := &coredata.ProcessingActivity{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := processingActivity.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load processing activity: %w", err)
			}

			if req.Name != nil {
				processingActivity.Name = *req.Name
			}
			if req.Purpose != nil {
				processingActivity.Purpose = *req.Purpose
			}
			if req.DataSubjectCategory != nil {
				processingActivity.DataSubjectCategory = *req.DataSubjectCategory
			}
			if req.PersonalDataCategory != nil {
				processingActivity.PersonalDataCategory = *req.PersonalDataCategory
			}
			if req.SpecialOrCriminalData != nil {
				processingActivity.SpecialOrCriminalData = *req.SpecialOrCriminalData
			}
			if req.ConsentEvidenceLink != nil {
				processingActivity.ConsentEvidenceLink = *req.ConsentEvidenceLink
			}
			if req.LawfulBasis != nil {
				processingActivity.LawfulBasis = *req.LawfulBasis
			}
			if req.Recipients != nil {
				processingActivity.Recipients = *req.Recipients
			}
			if req.Location != nil {
				processingActivity.Location = *req.Location
			}
			if req.InternationalTransfers != nil {
				processingActivity.InternationalTransfers = *req.InternationalTransfers
			}
			if req.TransferSafeguards != nil {
				processingActivity.TransferSafeguards = *req.TransferSafeguards
			}
			if req.RetentionPeriod != nil {
				processingActivity.RetentionPeriod = *req.RetentionPeriod
			}
			if req.SecurityMeasures != nil {
				processingActivity.SecurityMeasures = *req.SecurityMeasures
			}
			if req.DataProtectionImpactAssessment != nil {
				processingActivity.DataProtectionImpactAssessment = *req.DataProtectionImpactAssessment
			}
			if req.TransferImpactAssessment != nil {
				processingActivity.TransferImpactAssessment = *req.TransferImpactAssessment
			}

			processingActivity.UpdatedAt = time.Now()

			if err := processingActivity.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update processing activity: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return processingActivity, nil
}

func (s ProcessingActivityService) Delete(
	ctx context.Context,
	processingActivityID gid.GID,
) error {
	processingActivity := coredata.ProcessingActivity{ID: processingActivityID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := processingActivity.Delete(ctx, conn, s.svc.scope)
			if err != nil {
				return fmt.Errorf("cannot delete processing activity: %w", err)
			}
			return nil
		},
	)
}

func (s ProcessingActivityService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ProcessingActivityOrderField],
	filter *coredata.ProcessingActivityFilter,
) (*page.Page[*coredata.ProcessingActivity, coredata.ProcessingActivityOrderField], error) {
	var processingActivities coredata.ProcessingActivities

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := processingActivities.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load processing activities: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(processingActivities, cursor), nil
}

func (s ProcessingActivityService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.ProcessingActivityFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			processingActivities := coredata.ProcessingActivities{}
			count, err = processingActivities.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count processing activities: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}
