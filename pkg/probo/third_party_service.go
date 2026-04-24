// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
	"go.probo.inc/probo/pkg/vetting"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

// ErrVendorAssessmentDisabled is returned by VendorAssessor.Assess when the
// deployment has not configured an LLM provider for vendor assessment.
var ErrVendorAssessmentDisabled = errors.New("vendor assessment is not configured on this deployment")

// VendorAssessor produces a vendor assessment report from a website URL and
// an optional procedure description. Implementations that cannot perform
// assessment (missing LLM credentials, misconfigured provider) must return
// ErrVendorAssessmentDisabled from Assess so callers can surface a stable
// "feature unavailable" error instead of a generic internal error.
type VendorAssessor interface {
	Assess(
		ctx context.Context,
		websiteURL string,
		procedure string,
		reporter agent.ProgressReporter,
	) (*vetting.Result, error)
}

// DisabledVendorAssessor is the VendorAssessor implementation used when no
// LLM provider is configured for the vendor-assessor agent. Its Assess
// method always returns ErrVendorAssessmentDisabled.
type DisabledVendorAssessor struct{}

var _ VendorAssessor = DisabledVendorAssessor{}

func (DisabledVendorAssessor) Assess(
	_ context.Context,
	_ string,
	_ string,
	_ agent.ProgressReporter,
) (*vetting.Result, error) {
	return nil, ErrVendorAssessmentDisabled
}

type (
	ThirdPartyService struct {
		svc *TenantService
	}

	CreateThirdPartyRequest struct {
		OrganizationID                gid.GID
		Name                          string
		Description                   *string
		HeadquarterAddress            *string
		LegalName                     *string
		WebsiteURL                    *string
		Category                      *coredata.ThirdPartyCategory
		PrivacyPolicyURL              *string
		ServiceLevelAgreementURL      *string
		DataProcessingAgreementURL    *string
		BusinessAssociateAgreementURL *string
		SubprocessorsListURL          *string
		Certifications                []string
		Countries                     coredata.CountryCodes
		SecurityPageURL               *string
		TrustPageURL                  *string
		TermsOfServiceURL             *string
		StatusPageURL                 *string
		BusinessOwnerID               *gid.GID
		SecurityOwnerID               *gid.GID
	}

	UpdateThirdPartyRequest struct {
		ID                            gid.GID
		Name                          *string
		Description                   **string
		HeadquarterAddress            **string
		LegalName                     **string
		WebsiteURL                    **string
		TermsOfServiceURL             **string
		Category                      *coredata.ThirdPartyCategory
		PrivacyPolicyURL              **string
		ServiceLevelAgreementURL      **string
		DataProcessingAgreementURL    **string
		BusinessAssociateAgreementURL **string
		SubprocessorsListURL          **string
		Certifications                []string
		Countries                     coredata.CountryCodes
		SecurityPageURL               **string
		TrustPageURL                  **string
		StatusPageURL                 **string
		BusinessOwnerID               **gid.GID
		SecurityOwnerID               **gid.GID
		ShowOnTrustCenter             *bool
	}

	AssessThirdPartyRequest struct {
		ID         gid.GID
		WebsiteURL string
		Procedure  *string
	}

	AssessThirdPartyResult struct {
		ThirdParty    *coredata.ThirdParty
		Report        string
		Subprocessors []Subprocessor
	}

	Subprocessor struct {
		Name    string
		Country string
		Purpose string
	}

	CreateThirdPartyRiskAssessmentRequest struct {
		ThirdPartyID    gid.GID
		ExpiresAt       time.Time
		DataSensitivity coredata.DataSensitivity
		BusinessImpact  coredata.BusinessImpact
		Notes           *string
	}
)

func (cvr *CreateThirdPartyRequest) Validate() error {
	v := validator.New()

	v.Check(cvr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(cvr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cvr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(cvr.HeadquarterAddress, "headquarter_address", validator.SafeText(ContentMaxLength))
	v.Check(cvr.LegalName, "cvr.LegalName", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cvr.WebsiteURL, "website_url", validator.SafeText(2048))
	v.Check(cvr.Category, "category", validator.OneOfSlice(coredata.ThirdPartyCategories()))
	v.Check(cvr.PrivacyPolicyURL, "privacy_policy_url", validator.SafeText(2048))
	v.Check(cvr.ServiceLevelAgreementURL, "service_level_agreement_url", validator.SafeText(2048))
	v.Check(cvr.DataProcessingAgreementURL, "data_processing_agreement_url", validator.SafeText(2048))
	v.Check(cvr.BusinessAssociateAgreementURL, "business_associate_agreement_url", validator.SafeText(2048))
	v.Check(cvr.SubprocessorsListURL, "subprocessors_list_url", validator.SafeText(2048))
	v.Check(cvr.SecurityPageURL, "security_page_url", validator.SafeText(2048))
	v.Check(cvr.TrustPageURL, "trust_page_url", validator.SafeText(2048))
	v.Check(cvr.TermsOfServiceURL, "terms_of_service_url", validator.SafeText(2048))
	v.Check(cvr.StatusPageURL, "status_page_url", validator.SafeText(2048))
	v.Check(cvr.BusinessOwnerID, "business_owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(cvr.SecurityOwnerID, "security_owner_id", validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (uvr *UpdateThirdPartyRequest) Validate() error {
	v := validator.New()

	v.Check(uvr.ID, "id", validator.Required(), validator.GID(coredata.ThirdPartyEntityType))
	v.Check(uvr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uvr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(uvr.HeadquarterAddress, "headquarter_address", validator.SafeText(ContentMaxLength))
	v.Check(uvr.LegalName, "uvr.LegalName", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uvr.WebsiteURL, "website_url", validator.SafeText(2048))
	v.Check(uvr.Category, "category", validator.OneOfSlice(coredata.ThirdPartyCategories()))
	v.Check(uvr.PrivacyPolicyURL, "privacy_policy_url", validator.SafeText(2048))
	v.Check(uvr.ServiceLevelAgreementURL, "service_level_agreement_url", validator.SafeText(2048))
	v.Check(uvr.DataProcessingAgreementURL, "data_processing_agreement_url", validator.SafeText(2048))
	v.Check(uvr.BusinessAssociateAgreementURL, "business_associate_agreement_url", validator.SafeText(2048))
	v.Check(uvr.SubprocessorsListURL, "subprocessors_list_url", validator.SafeText(2048))
	v.Check(uvr.SecurityPageURL, "security_page_url", validator.SafeText(2048))
	v.Check(uvr.TrustPageURL, "trust_page_url", validator.SafeText(2048))
	v.Check(uvr.TermsOfServiceURL, "terms_of_service_url", validator.SafeText(2048))
	v.Check(uvr.StatusPageURL, "status_page_url", validator.SafeText(2048))
	v.Check(uvr.BusinessOwnerID, "business_owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(uvr.SecurityOwnerID, "security_owner_id", validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (cvrar *CreateThirdPartyRiskAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(cvrar.ThirdPartyID, "third_party_id", validator.Required(), validator.GID(coredata.ThirdPartyEntityType))
	v.Check(cvrar.DataSensitivity, "data_sensitivity", validator.Required(), validator.OneOfSlice(coredata.DataSensitivities()))
	v.Check(cvrar.BusinessImpact, "business_impact", validator.Required(), validator.OneOfSlice(coredata.BusinessImpacts()))
	v.Check(cvrar.Notes, "notes", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s ThirdPartyService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			thirdParties := coredata.ThirdParties{}
			filter := &coredata.ThirdPartyFilter{}
			count, err = thirdParties.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count third_parties: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ThirdPartyService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
	filter *coredata.ThirdPartyFilter,
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			return thirdParties.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organization.ID,
				cursor,
				filter,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdParties, cursor), nil
}

func (s ThirdPartyService) CountForDatumID(
	ctx context.Context,
	datumID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			thirdParties := coredata.ThirdParties{}
			count, err = thirdParties.CountByDatumID(ctx, conn, s.svc.scope, datumID)
			if err != nil {
				return fmt.Errorf("cannot count third_parties: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ThirdPartyService) ListForDatumID(
	ctx context.Context,
	datumID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdParties.LoadByDatumID(
				ctx,
				conn,
				s.svc.scope,
				datumID,
				cursor,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdParties, cursor), nil
}

func (s ThirdPartyService) Update(
	ctx context.Context,
	req UpdateThirdPartyRequest,
) (*coredata.ThirdParty, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := thirdParty.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load thirdParty %q: %w", req.ID, err)
			}

			if req.Name != nil {
				thirdParty.Name = *req.Name
			}

			if req.Description != nil {
				thirdParty.Description = *req.Description
			}

			if req.StatusPageURL != nil {
				thirdParty.StatusPageURL = *req.StatusPageURL
			}

			if req.TermsOfServiceURL != nil {
				thirdParty.TermsOfServiceURL = *req.TermsOfServiceURL
			}

			if req.PrivacyPolicyURL != nil {
				thirdParty.PrivacyPolicyURL = *req.PrivacyPolicyURL
			}

			if req.ServiceLevelAgreementURL != nil {
				thirdParty.ServiceLevelAgreementURL = *req.ServiceLevelAgreementURL
			}

			if req.DataProcessingAgreementURL != nil {
				thirdParty.DataProcessingAgreementURL = *req.DataProcessingAgreementURL
			}

			if req.BusinessAssociateAgreementURL != nil {
				thirdParty.BusinessAssociateAgreementURL = *req.BusinessAssociateAgreementURL
			}

			if req.SubprocessorsListURL != nil {
				thirdParty.SubprocessorsListURL = *req.SubprocessorsListURL
			}

			if req.Category != nil {
				thirdParty.Category = *req.Category
			} else {
				thirdParty.Category = coredata.ThirdPartyCategoryOther
			}

			if req.SecurityPageURL != nil {
				thirdParty.SecurityPageURL = *req.SecurityPageURL
			}

			if req.ShowOnTrustCenter != nil {
				thirdParty.ShowOnTrustCenter = *req.ShowOnTrustCenter
			}

			if req.TrustPageURL != nil {
				thirdParty.TrustPageURL = *req.TrustPageURL
			}

			if req.HeadquarterAddress != nil {
				thirdParty.HeadquarterAddress = *req.HeadquarterAddress
			}

			if req.LegalName != nil {
				thirdParty.LegalName = *req.LegalName
			}

			if req.WebsiteURL != nil {
				thirdParty.WebsiteURL = *req.WebsiteURL
			}

			if req.TermsOfServiceURL != nil {
				thirdParty.TermsOfServiceURL = *req.TermsOfServiceURL
			}

			if req.Certifications != nil {
				thirdParty.Certifications = req.Certifications
			}

			if req.Countries != nil {
				thirdParty.Countries = req.Countries
			}

			if req.BusinessOwnerID != nil {
				if *req.BusinessOwnerID != nil {
					businessOwner := &coredata.MembershipProfile{}
					if err := businessOwner.LoadByID(ctx, conn, s.svc.scope, **req.BusinessOwnerID); err != nil {
						return fmt.Errorf("cannot load business owner profile: %w", err)
					}
					thirdParty.BusinessOwnerID = &businessOwner.ID
				} else {
					thirdParty.BusinessOwnerID = nil
				}
			}

			if req.SecurityOwnerID != nil {
				if *req.SecurityOwnerID != nil {
					securityOwner := &coredata.MembershipProfile{}
					if err := securityOwner.LoadByID(ctx, conn, s.svc.scope, **req.SecurityOwnerID); err != nil {
						return fmt.Errorf("cannot load security owner profile: %w", err)
					}
					thirdParty.SecurityOwnerID = &securityOwner.ID
				} else {
					thirdParty.SecurityOwnerID = nil
				}
			}

			thirdParty.UpdatedAt = time.Now()

			if err := thirdParty.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update thirdParty: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				conn,
				s.svc.scope,
				thirdParty.OrganizationID,
				coredata.WebhookEventTypeThirdPartyUpdated,
				webhooktypes.NewThirdParty(thirdParty),
			); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdParty, nil
}

func (s ThirdPartyService) Get(
	ctx context.Context,
	thirdPartyID gid.GID,
) (*coredata.ThirdParty, error) {
	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdParty.LoadByID(ctx, conn, s.svc.scope, thirdPartyID)
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdParty, nil
}

func (s ThirdPartyService) GetByIDs(
	ctx context.Context,
	thirdPartyIDs ...gid.GID,
) (coredata.ThirdParties, error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := thirdParties.LoadByIDs(
				ctx,
				conn,
				s.svc.scope,
				thirdPartyIDs,
			); err != nil {
				return fmt.Errorf("cannot load third_parties by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return thirdParties, nil
}

func (s ThirdPartyService) Delete(
	ctx context.Context,
	thirdPartyID gid.GID,
) error {
	thirdParty := &coredata.ThirdParty{}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := thirdParty.LoadByID(ctx, conn, s.svc.scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot load thirdParty: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				conn,
				s.svc.scope,
				thirdParty.OrganizationID,
				coredata.WebhookEventTypeThirdPartyDeleted,
				webhooktypes.NewThirdParty(thirdParty),
			); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return thirdParty.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s ThirdPartyService) Create(
	ctx context.Context,
	req CreateThirdPartyRequest,
) (*coredata.ThirdParty, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	thirdParty := &coredata.ThirdParty{
		ID:                            gid.New(s.svc.scope.GetTenantID(), coredata.ThirdPartyEntityType),
		Name:                          req.Name,
		CreatedAt:                     now,
		UpdatedAt:                     now,
		Description:                   req.Description,
		HeadquarterAddress:            req.HeadquarterAddress,
		LegalName:                     req.LegalName,
		WebsiteURL:                    req.WebsiteURL,
		PrivacyPolicyURL:              req.PrivacyPolicyURL,
		ServiceLevelAgreementURL:      req.ServiceLevelAgreementURL,
		DataProcessingAgreementURL:    req.DataProcessingAgreementURL,
		BusinessAssociateAgreementURL: req.BusinessAssociateAgreementURL,
		SubprocessorsListURL:          req.SubprocessorsListURL,
		Certifications:                req.Certifications,
		Countries:                     req.Countries,
		SecurityPageURL:               req.SecurityPageURL,
		TrustPageURL:                  req.TrustPageURL,
		StatusPageURL:                 req.StatusPageURL,
		TermsOfServiceURL:             req.TermsOfServiceURL,
		ShowOnTrustCenter:             false,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization %q: %w", req.OrganizationID, err)
			}

			thirdParty.OrganizationID = organization.ID

			if req.BusinessOwnerID != nil {
				businessOwner := &coredata.MembershipProfile{}
				if err := businessOwner.LoadByID(ctx, conn, s.svc.scope, *req.BusinessOwnerID); err != nil {
					return fmt.Errorf("cannot load business owner profile: %w", err)
				}
				thirdParty.BusinessOwnerID = &businessOwner.ID
			}

			if req.SecurityOwnerID != nil {
				securityOwner := &coredata.MembershipProfile{}
				if err := securityOwner.LoadByID(ctx, conn, s.svc.scope, *req.SecurityOwnerID); err != nil {
					return fmt.Errorf("cannot load security owner profile: %w", err)
				}
				thirdParty.SecurityOwnerID = &securityOwner.ID
			}

			if req.Category != nil {
				thirdParty.Category = *req.Category
			} else {
				thirdParty.Category = coredata.ThirdPartyCategoryOther
			}

			if err := thirdParty.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert thirdParty: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				conn,
				s.svc.scope,
				organization.ID,
				coredata.WebhookEventTypeThirdPartyCreated,
				webhooktypes.NewThirdParty(thirdParty),
			); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdParty, nil
}

func (s ThirdPartyService) CountForAssetID(
	ctx context.Context,
	assetID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			thirdParties := coredata.ThirdParties{}
			count, err = thirdParties.CountByAssetID(ctx, conn, s.svc.scope, assetID)
			if err != nil {
				return fmt.Errorf("cannot count third_parties: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ThirdPartyService) ListForAssetID(
	ctx context.Context,
	assetID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdParties.LoadByAssetID(ctx, conn, s.svc.scope, assetID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdParties, cursor), nil
}

func (s ThirdPartyService) ListForProcessingActivityID(
	ctx context.Context,
	processingActivityID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := thirdParties.LoadByProcessingActivityID(ctx, conn, s.svc.scope, processingActivityID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load third_parties by processing activity: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdParties, cursor), nil
}

func (s ThirdPartyService) ListRiskAssessments(
	ctx context.Context,
	thirdPartyID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyRiskAssessmentOrderField],
) (*page.Page[*coredata.ThirdPartyRiskAssessment, coredata.ThirdPartyRiskAssessmentOrderField], error) {
	var thirdPartyRiskAssessments coredata.ThirdPartyRiskAssessments

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdPartyRiskAssessments.LoadByThirdPartyID(ctx, conn, s.svc.scope, thirdPartyID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdPartyRiskAssessments, cursor), nil
}

func (s ThirdPartyService) CreateRiskAssessment(
	ctx context.Context,
	req CreateThirdPartyRiskAssessmentRequest,
) (*coredata.ThirdPartyRiskAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	thirdPartyRiskAssessmentID := gid.New(s.svc.scope.GetTenantID(), coredata.ThirdPartyRiskAssessmentEntityType)

	now := time.Now()

	thirdPartyRiskAssessment := &coredata.ThirdPartyRiskAssessment{
		ID:              thirdPartyRiskAssessmentID,
		ThirdPartyID:    req.ThirdPartyID,
		ExpiresAt:       req.ExpiresAt,
		DataSensitivity: req.DataSensitivity,
		BusinessImpact:  req.BusinessImpact,
		Notes:           req.Notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if !req.ExpiresAt.After(now) {
		return nil, fmt.Errorf("expiresAt %v must be in the future", req.ExpiresAt)
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			thirdParty := coredata.ThirdParty{}
			if err := thirdParty.LoadByID(ctx, tx, s.svc.scope, req.ThirdPartyID); err != nil {
				return fmt.Errorf("cannot load thirdParty: %w", err)
			}

			thirdPartyRiskAssessment.OrganizationID = thirdParty.OrganizationID

			if err := thirdParty.ExpireNonExpiredRiskAssessments(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot expire thirdParty risk assessments: %w", err)
			}

			if err := thirdPartyRiskAssessment.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert thirdParty risk assessment: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdPartyRiskAssessment, nil
}

func (s ThirdPartyService) GetRiskAssessment(
	ctx context.Context,
	thirdPartyRiskAssessmentID gid.GID,
) (*coredata.ThirdPartyRiskAssessment, error) {
	thirdPartyRiskAssessment := &coredata.ThirdPartyRiskAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdPartyRiskAssessment.LoadByID(ctx, conn, s.svc.scope, thirdPartyRiskAssessmentID)
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdPartyRiskAssessment, nil
}

func (s ThirdPartyService) GetByRiskAssessmentID(
	ctx context.Context,
	thirdPartyRiskAssessmentID gid.GID,
) (*coredata.ThirdParty, error) {
	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			thirdPartyRiskAssessment := &coredata.ThirdPartyRiskAssessment{}
			if err := thirdPartyRiskAssessment.LoadByID(ctx, conn, s.svc.scope, thirdPartyRiskAssessmentID); err != nil {
				return fmt.Errorf("cannot load thirdParty risk assessment: %w", err)
			}

			if err := thirdParty.LoadByID(ctx, conn, s.svc.scope, thirdPartyRiskAssessment.ThirdPartyID); err != nil {
				return fmt.Errorf("cannot load thirdParty: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdParty, nil
}

func (s ThirdPartyService) Assess(
	ctx context.Context,
	req AssessThirdPartyRequest,
) (*AssessThirdPartyResult, error) {
	result, err := s.svc.vendorAssessor.Assess(ctx, req.WebsiteURL, ref.UnrefOrZero(req.Procedure), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot assess third party: %w", err)
	}

	thirdParty := &coredata.ThirdParty{}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := thirdParty.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load third party %q: %w", req.ID, err)
			}

			info := result.Info

			if info.Name != "" {
				thirdParty.Name = info.Name
			}

			thirdParty.WebsiteURL = &req.WebsiteURL
			if info.Category != "" {
				thirdParty.Category = coredata.ThirdPartyCategory(info.Category)
			}
			thirdParty.UpdatedAt = time.Now()

			if info.Description != "" {
				thirdParty.Description = &info.Description
			}
			if info.HeadquarterAddress != "" {
				thirdParty.HeadquarterAddress = &info.HeadquarterAddress
			}
			if info.LegalName != "" {
				thirdParty.LegalName = &info.LegalName
			}
			if info.PrivacyPolicyURL != "" {
				thirdParty.PrivacyPolicyURL = &info.PrivacyPolicyURL
			}
			if info.ServiceLevelAgreementURL != "" {
				thirdParty.ServiceLevelAgreementURL = &info.ServiceLevelAgreementURL
			}
			if info.DataProcessingAgreementURL != "" {
				thirdParty.DataProcessingAgreementURL = &info.DataProcessingAgreementURL
			}
			if info.BusinessAssociateAgreementURL != "" {
				thirdParty.BusinessAssociateAgreementURL = &info.BusinessAssociateAgreementURL
			}
			if info.SubprocessorsListURL != "" {
				thirdParty.SubprocessorsListURL = &info.SubprocessorsListURL
			}
			if info.SecurityPageURL != "" {
				thirdParty.SecurityPageURL = &info.SecurityPageURL
			}
			if info.TrustPageURL != "" {
				thirdParty.TrustPageURL = &info.TrustPageURL
			}
			if info.TermsOfServiceURL != "" {
				thirdParty.TermsOfServiceURL = &info.TermsOfServiceURL
			}
			if info.StatusPageURL != "" {
				thirdParty.StatusPageURL = &info.StatusPageURL
			}

			if len(info.Certifications) > 0 {
				thirdParty.Certifications = info.Certifications
			}

			if err := thirdParty.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update third party: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				conn,
				s.svc.scope,
				thirdParty.OrganizationID,
				coredata.WebhookEventTypeThirdPartyUpdated,
				webhooktypes.NewThirdParty(thirdParty),
			); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	subprocessors := make([]Subprocessor, len(result.Info.Subprocessors))
	for i, sp := range result.Info.Subprocessors {
		subprocessors[i] = Subprocessor{
			Name:    sp.Name,
			Country: sp.Country,
			Purpose: sp.Purpose,
		}
	}

	return &AssessThirdPartyResult{
		ThirdParty:    thirdParty,
		Report:        result.Document,
		Subprocessors: subprocessors,
	}, nil
}
