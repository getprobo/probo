// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

type AiSystemService struct {
	svc *Service
}

type (
	CreateAiSystemRequest struct {
		OrganizationID          gid.GID
		Name                    string
		Version                 *string
		CompanyRoles            []coredata.AiSystemCompanyRole
		Status                  coredata.AiSystemStatus
		OwnerID                 *gid.GID
		Source                  *string
		Purpose                 *string
		IntendedUseCases        *string
		AutonomyLevel           *string
		HumanOversightMechanism *string
		RiskClassification      *coredata.AiSystemRiskClassification
		KeyStakeholders         *string
		DataSourcesAndType      *string
		DeploymentDate          *time.Time
		LastReviewDate          *time.Time
		NextReviewDate          *time.Time
		Notes                   *string
	}

	UpdateAiSystemRequest struct {
		ID                      gid.GID
		Name                    *string
		Version                 **string
		CompanyRoles            *[]coredata.AiSystemCompanyRole
		Status                  *coredata.AiSystemStatus
		OwnerID                 **gid.GID
		Source                  **string
		Purpose                 **string
		IntendedUseCases        **string
		AutonomyLevel           **string
		HumanOversightMechanism **string
		RiskClassification      **coredata.AiSystemRiskClassification
		KeyStakeholders         **string
		DataSourcesAndType      **string
		DeploymentDate          **time.Time
		LastReviewDate          **time.Time
		NextReviewDate          **time.Time
		Notes                   **string
	}
)

func (r *CreateAiSystemRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeText(TitleMaxLength))
	v.Check(r.Version, "version", validator.SafeText(TitleMaxLength))
	v.Check(r.Status, "status", validator.Required(), validator.OneOfSlice(coredata.AiSystemStatuses()))
	v.CheckEach(
		r.CompanyRoles,
		"company_roles",
		func(index int, item any) {
			v.Check(item, fmt.Sprintf("company_roles[%d]", index), validator.Required(), validator.OneOfSlice(coredata.AiSystemCompanyRoles()))
		},
	)
	v.Check(r.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(r.Source, "source", validator.SafeText(ContentMaxLength))
	v.Check(r.Purpose, "purpose", validator.SafeText(ContentMaxLength))
	v.Check(r.IntendedUseCases, "intended_use_cases", validator.SafeText(ContentMaxLength))
	v.Check(r.AutonomyLevel, "autonomy_level", validator.SafeText(ContentMaxLength))
	v.Check(r.HumanOversightMechanism, "human_oversight_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(r.RiskClassification, "risk_classification", validator.Required(), validator.OneOfSlice(coredata.AiSystemRiskClassifications()))
	v.Check(r.KeyStakeholders, "key_stakeholders", validator.SafeText(ContentMaxLength))
	v.Check(r.DataSourcesAndType, "data_sources_and_type", validator.SafeText(ContentMaxLength))
	v.Check(r.Notes, "notes", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (r *UpdateAiSystemRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.AiSystemEntityType))
	v.Check(r.Name, "name", validator.SafeText(TitleMaxLength))
	v.Check(r.Version, "version", validator.SafeText(TitleMaxLength))
	v.Check(r.Status, "status", validator.OneOfSlice(coredata.AiSystemStatuses()))
	v.CheckEach(
		r.CompanyRoles,
		"company_roles",
		func(index int, item any) {
			if item == nil {
				return
			}

			v.Check(item, fmt.Sprintf("company_roles[%d]", index), validator.OneOfSlice(coredata.AiSystemCompanyRoles()))
		},
	)
	v.Check(r.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(r.Source, "source", validator.SafeText(ContentMaxLength))
	v.Check(r.Purpose, "purpose", validator.SafeText(ContentMaxLength))
	v.Check(r.IntendedUseCases, "intended_use_cases", validator.SafeText(ContentMaxLength))
	v.Check(r.AutonomyLevel, "autonomy_level", validator.SafeText(ContentMaxLength))
	v.Check(r.HumanOversightMechanism, "human_oversight_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(r.RiskClassification, "risk_classification", validator.OneOfSlice(coredata.AiSystemRiskClassifications()))
	v.Check(r.KeyStakeholders, "key_stakeholders", validator.SafeText(ContentMaxLength))
	v.Check(r.DataSourcesAndType, "data_sources_and_type", validator.SafeText(ContentMaxLength))
	v.Check(r.Notes, "notes", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s AiSystemService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	aiSystemID gid.GID,
) (*coredata.AiSystem, error) {
	aiSystem := &coredata.AiSystem{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return aiSystem.LoadByID(ctx, conn, scope, aiSystemID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get ai system: %w", err)
	}

	return aiSystem, nil
}

func (s *AiSystemService) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateAiSystemRequest,
) (*coredata.AiSystem, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	req.CompanyRoles = normalizeAiSystemCompanyRoles(req.CompanyRoles)

	now := time.Now()

	aiSystem := &coredata.AiSystem{
		ID:                      gid.New(scope.GetTenantID(), coredata.AiSystemEntityType),
		OrganizationID:          req.OrganizationID,
		Name:                    req.Name,
		Version:                 req.Version,
		CompanyRoles:            req.CompanyRoles,
		Status:                  req.Status,
		OwnerID:                 req.OwnerID,
		Source:                  req.Source,
		Purpose:                 req.Purpose,
		IntendedUseCases:        req.IntendedUseCases,
		AutonomyLevel:           req.AutonomyLevel,
		HumanOversightMechanism: req.HumanOversightMechanism,
		RiskClassification:      req.RiskClassification,
		KeyStakeholders:         req.KeyStakeholders,
		DataSourcesAndType:      req.DataSourcesAndType,
		DeploymentDate:          req.DeploymentDate,
		LastReviewDate:          req.LastReviewDate,
		NextReviewDate:          req.NextReviewDate,
		Notes:                   req.Notes,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if req.OwnerID != nil {
				owner := &coredata.MembershipProfile{}
				if err := owner.LoadByID(ctx, conn, scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}

				if owner.OrganizationID != req.OrganizationID {
					return fmt.Errorf("cannot load owner profile: %w", coredata.ErrResourceNotFound)
				}
			}

			if err := aiSystem.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert ai system: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return aiSystem, nil
}

func (s *AiSystemService) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateAiSystemRequest,
) (*coredata.AiSystem, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if req.CompanyRoles != nil {
		normalized := normalizeAiSystemCompanyRoles(*req.CompanyRoles)
		req.CompanyRoles = &normalized
	}

	aiSystem := &coredata.AiSystem{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := aiSystem.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load ai system: %w", err)
			}

			if req.Name != nil {
				aiSystem.Name = *req.Name
			}

			if req.Version != nil {
				aiSystem.Version = *req.Version
			}

			if req.CompanyRoles != nil {
				aiSystem.CompanyRoles = *req.CompanyRoles
			}

			if req.Status != nil {
				aiSystem.Status = *req.Status
			}

			if req.OwnerID != nil {
				if *req.OwnerID != nil {
					owner := &coredata.MembershipProfile{}
					if err := owner.LoadByID(ctx, conn, scope, **req.OwnerID); err != nil {
						return fmt.Errorf("cannot load owner profile: %w", err)
					}

					if owner.OrganizationID != aiSystem.OrganizationID {
						return fmt.Errorf("cannot load owner profile: %w", coredata.ErrResourceNotFound)
					}
				}

				aiSystem.OwnerID = *req.OwnerID
			}

			if req.Source != nil {
				aiSystem.Source = *req.Source
			}

			if req.Purpose != nil {
				aiSystem.Purpose = *req.Purpose
			}

			if req.IntendedUseCases != nil {
				aiSystem.IntendedUseCases = *req.IntendedUseCases
			}

			if req.AutonomyLevel != nil {
				aiSystem.AutonomyLevel = *req.AutonomyLevel
			}

			if req.HumanOversightMechanism != nil {
				aiSystem.HumanOversightMechanism = *req.HumanOversightMechanism
			}

			if req.RiskClassification != nil {
				aiSystem.RiskClassification = *req.RiskClassification
			}

			if req.KeyStakeholders != nil {
				aiSystem.KeyStakeholders = *req.KeyStakeholders
			}

			if req.DataSourcesAndType != nil {
				aiSystem.DataSourcesAndType = *req.DataSourcesAndType
			}

			if req.DeploymentDate != nil {
				aiSystem.DeploymentDate = *req.DeploymentDate
			}

			if req.LastReviewDate != nil {
				aiSystem.LastReviewDate = *req.LastReviewDate
			}

			if req.NextReviewDate != nil {
				aiSystem.NextReviewDate = *req.NextReviewDate
			}

			if req.Notes != nil {
				aiSystem.Notes = *req.Notes
			}

			aiSystem.UpdatedAt = time.Now()

			if err := aiSystem.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update ai system: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return aiSystem, nil
}

func (s AiSystemService) Delete(
	ctx context.Context,
	scope coredata.Scoper,
	aiSystemID gid.GID,
) error {
	aiSystem := coredata.AiSystem{ID: aiSystemID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			err := aiSystem.Delete(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot delete ai system: %w", err)
			}

			return nil
		},
	)
}

func (s AiSystemService) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.AiSystemOrderField],
	filter *coredata.AiSystemFilter,
) (*page.Page[*coredata.AiSystem, coredata.AiSystemOrderField], error) {
	var aiSystems coredata.AiSystems

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := aiSystems.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load ai systems: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(aiSystems, cursor), nil
}

func (s AiSystemService) CountForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.AiSystemFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			aiSystems := coredata.AiSystems{}

			count, err = aiSystems.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count ai systems: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func normalizeAiSystemCompanyRoles(
	roles []coredata.AiSystemCompanyRole,
) []coredata.AiSystemCompanyRole {
	if len(roles) == 0 {
		return roles
	}

	seen := make(map[coredata.AiSystemCompanyRole]struct{}, len(roles))
	normalized := make([]coredata.AiSystemCompanyRole, 0, len(roles))

	for _, role := range roles {
		if _, ok := seen[role]; ok {
			continue
		}

		seen[role] = struct{}{}
		normalized = append(normalized, role)
	}

	return normalized
}
