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

type ProcessingActivityRegistryService struct {
	svc *TenantService
}

type (
	CreateProcessingActivityRegistryRequest struct {
		OrganizationID                 gid.GID
		Name                           string
		Purpose                        *string
		DataSubjectCategory            *string
		PersonalDataCategory           *string
		SpecialOrCriminalData          coredata.ProcessingActivityRegistrySpecialOrCriminalData
		ConsentEvidenceLink            *string
		LawfulBasis                    coredata.ProcessingActivityRegistryLawfulBasis
		Recipients                     *string
		Location                       *string
		InternationalTransfers         bool
		TransferSafeguards             *coredata.ProcessingActivityRegistryTransferSafeguards
		RetentionPeriod                *string
		SecurityMeasures               *string
		DataProtectionImpactAssessment coredata.ProcessingActivityRegistryDataProtectionImpactAssessment
		TransferImpactAssessment       coredata.ProcessingActivityRegistryTransferImpactAssessment
	}

	UpdateProcessingActivityRegistryRequest struct {
		ID                             gid.GID
		Name                           *string
		Purpose                        **string
		DataSubjectCategory            **string
		PersonalDataCategory           **string
		SpecialOrCriminalData          *coredata.ProcessingActivityRegistrySpecialOrCriminalData
		ConsentEvidenceLink            **string
		LawfulBasis                    *coredata.ProcessingActivityRegistryLawfulBasis
		Recipients                     **string
		Location                       **string
		InternationalTransfers         *bool
		TransferSafeguards             **coredata.ProcessingActivityRegistryTransferSafeguards
		RetentionPeriod                **string
		SecurityMeasures               **string
		DataProtectionImpactAssessment *coredata.ProcessingActivityRegistryDataProtectionImpactAssessment
		TransferImpactAssessment       *coredata.ProcessingActivityRegistryTransferImpactAssessment
	}
)

func (s ProcessingActivityRegistryService) Get(
	ctx context.Context,
	processingActivityRegistryID gid.GID,
) (*coredata.ProcessingActivityRegistry, error) {
	processingActivityRegistry := &coredata.ProcessingActivityRegistry{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return processingActivityRegistry.LoadByID(ctx, conn, s.svc.scope, processingActivityRegistryID)
		},
	)

	if err != nil {
		return nil, err
	}

	return processingActivityRegistry, nil
}

func (s *ProcessingActivityRegistryService) Create(
	ctx context.Context,
	req *CreateProcessingActivityRegistryRequest,
) (*coredata.ProcessingActivityRegistry, error) {
	now := time.Now()

	processingActivityRegistry := &coredata.ProcessingActivityRegistry{
		ID:                             gid.New(s.svc.scope.GetTenantID(), coredata.ProcessingActivityRegistryEntityType),
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

			if err := processingActivityRegistry.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert processing activity registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return processingActivityRegistry, nil
}

func (s *ProcessingActivityRegistryService) Update(
	ctx context.Context,
	req *UpdateProcessingActivityRegistryRequest,
) (*coredata.ProcessingActivityRegistry, error) {
	processingActivityRegistry := &coredata.ProcessingActivityRegistry{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := processingActivityRegistry.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load processing activity registry: %w", err)
			}

			if req.Name != nil {
				processingActivityRegistry.Name = *req.Name
			}
			if req.Purpose != nil {
				processingActivityRegistry.Purpose = *req.Purpose
			}
			if req.DataSubjectCategory != nil {
				processingActivityRegistry.DataSubjectCategory = *req.DataSubjectCategory
			}
			if req.PersonalDataCategory != nil {
				processingActivityRegistry.PersonalDataCategory = *req.PersonalDataCategory
			}
			if req.SpecialOrCriminalData != nil {
				processingActivityRegistry.SpecialOrCriminalData = *req.SpecialOrCriminalData
			}
			if req.ConsentEvidenceLink != nil {
				processingActivityRegistry.ConsentEvidenceLink = *req.ConsentEvidenceLink
			}
			if req.LawfulBasis != nil {
				processingActivityRegistry.LawfulBasis = *req.LawfulBasis
			}
			if req.Recipients != nil {
				processingActivityRegistry.Recipients = *req.Recipients
			}
			if req.Location != nil {
				processingActivityRegistry.Location = *req.Location
			}
			if req.InternationalTransfers != nil {
				processingActivityRegistry.InternationalTransfers = *req.InternationalTransfers
			}
			if req.TransferSafeguards != nil {
				processingActivityRegistry.TransferSafeguards = *req.TransferSafeguards
			}
			if req.RetentionPeriod != nil {
				processingActivityRegistry.RetentionPeriod = *req.RetentionPeriod
			}
			if req.SecurityMeasures != nil {
				processingActivityRegistry.SecurityMeasures = *req.SecurityMeasures
			}
			if req.DataProtectionImpactAssessment != nil {
				processingActivityRegistry.DataProtectionImpactAssessment = *req.DataProtectionImpactAssessment
			}
			if req.TransferImpactAssessment != nil {
				processingActivityRegistry.TransferImpactAssessment = *req.TransferImpactAssessment
			}

			processingActivityRegistry.UpdatedAt = time.Now()

			if err := processingActivityRegistry.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update processing activity registry: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return processingActivityRegistry, nil
}

func (s ProcessingActivityRegistryService) Delete(
	ctx context.Context,
	processingActivityRegistryID gid.GID,
) error {
	processingActivityRegistry := coredata.ProcessingActivityRegistry{ID: processingActivityRegistryID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := processingActivityRegistry.Delete(ctx, conn, s.svc.scope)
			if err != nil {
				return fmt.Errorf("cannot delete processing activity registry: %w", err)
			}
			return nil
		},
	)
}

func (s ProcessingActivityRegistryService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ProcessingActivityRegistryOrderField],
	filter *coredata.ProcessingActivityRegistryFilter,
) (*page.Page[*coredata.ProcessingActivityRegistry, coredata.ProcessingActivityRegistryOrderField], error) {
	var processingActivityRegistries coredata.ProcessingActivityRegistries

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := processingActivityRegistries.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load processing activity registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(processingActivityRegistries, cursor), nil
}

func (s ProcessingActivityRegistryService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.ProcessingActivityRegistryFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			processingActivityRegistries := coredata.ProcessingActivityRegistries{}
			count, err = processingActivityRegistries.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count processing activity registries: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}
