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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

type (
	ThirdPartyService struct {
		svc *Service
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
		ParentThirdPartyID            *gid.GID
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

	CreateThirdPartyRiskAssessmentRequest struct {
		ThirdPartyID    gid.GID
		ExpiresAt       time.Time
		DataSensitivity coredata.DataSensitivity
		BusinessImpact  coredata.BusinessImpact
		Notes           *string
	}

	ImportThirdPartyFromCommonRequest struct {
		OrganizationID     gid.GID
		CommonThirdPartyID gid.GID
	}
)

func (r *ImportThirdPartyFromCommonRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.CommonThirdPartyID, "common_third_party_id", validator.Required(), validator.GID(coredata.CommonThirdPartyEntityType))

	return v.Error()
}

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
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.ThirdPartyFilter,
) (int, error) {
	var count int

	if filter == nil {
		filter = coredata.NewThirdPartyFilter(nil, nil, nil)
	}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			thirdParties := coredata.ThirdParties{}

			count, err = thirdParties.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count thirdParties: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
	filter *coredata.ThirdPartyFilter,
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			return thirdParties.LoadByOrganizationID(
				ctx,
				conn,
				scope,
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

func (s ThirdPartyService) CountForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			thirdParties := coredata.ThirdParties{}

			count, err = thirdParties.CountByMeasureID(ctx, conn, scope, measureID)
			if err != nil {
				return fmt.Errorf("cannot count thirdParties: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ThirdPartyService) ListForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	measure := &coredata.Measure{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := measure.LoadByID(ctx, conn, scope, measureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			if err := thirdParties.LoadByMeasureID(
				ctx,
				conn,
				scope,
				measure.ID,
				cursor,
			); err != nil {
				return fmt.Errorf("cannot load thirdParties: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdParties, cursor), nil
}

func (s ThirdPartyService) CountForDatumID(
	ctx context.Context, scope coredata.Scoper,
	datumID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			thirdParties := coredata.ThirdParties{}

			count, err = thirdParties.CountByDatumID(ctx, conn, scope, datumID)
			if err != nil {
				return fmt.Errorf("cannot count thirdParties: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
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
				scope,
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
	ctx context.Context, scope coredata.Scoper,
	req UpdateThirdPartyRequest,
) (*coredata.ThirdParty, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := thirdParty.LoadByID(ctx, conn, scope, req.ID); err != nil {
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
					if err := businessOwner.LoadByID(ctx, conn, scope, **req.BusinessOwnerID); err != nil {
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
					if err := securityOwner.LoadByID(ctx, conn, scope, **req.SecurityOwnerID); err != nil {
						return fmt.Errorf("cannot load security owner profile: %w", err)
					}

					thirdParty.SecurityOwnerID = &securityOwner.ID
				} else {
					thirdParty.SecurityOwnerID = nil
				}
			}

			thirdParty.UpdatedAt = time.Now()

			if err := thirdParty.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update thirdParty: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				conn,
				scope,
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
	ctx context.Context, scope coredata.Scoper,
	thirdPartyID gid.GID,
) (*coredata.ThirdParty, error) {
	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdParty.LoadByID(ctx, conn, scope, thirdPartyID)
		},
	)
	if err != nil {
		return nil, err
	}

	return thirdParty, nil
}

func (s ThirdPartyService) GetByIDs(
	ctx context.Context, scope coredata.Scoper,
	thirdPartyIDs ...gid.GID,
) (coredata.ThirdParties, error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := thirdParties.LoadByIDs(
				ctx,
				conn,
				scope,
				thirdPartyIDs,
			); err != nil {
				return fmt.Errorf("cannot load thirdParties by ids: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
	thirdPartyID gid.GID,
) error {
	thirdParty := &coredata.ThirdParty{}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := thirdParty.LoadByID(ctx, conn, scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot load thirdParty: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				conn,
				scope,
				thirdParty.OrganizationID,
				coredata.WebhookEventTypeThirdPartyDeleted,
				webhooktypes.NewThirdParty(thirdParty),
			); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return thirdParty.Delete(ctx, conn, scope)
		},
	)
}

func (s ThirdPartyService) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateThirdPartyRequest,
) (*coredata.ThirdParty, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	thirdParty := &coredata.ThirdParty{
		ID:                            gid.New(scope.GetTenantID(), coredata.ThirdPartyEntityType),
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
		Level:                         1,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization %q: %w", req.OrganizationID, err)
			}

			thirdParty.OrganizationID = organization.ID

			if req.ParentThirdPartyID != nil {
				parent := &coredata.ThirdParty{}
				if err := parent.LoadByID(ctx, conn, scope, *req.ParentThirdPartyID); err != nil {
					return fmt.Errorf("cannot load parent third party: %w", err)
				}

				if parent.OrganizationID != organization.ID {
					return fmt.Errorf("parent third party belongs to a different organization: %w", coredata.ErrResourceNotFound)
				}

				thirdParty.ParentThirdPartyID = &parent.ID
				// The level always follows the parent chain; ignore any
				// client-supplied level so it cannot desync from the hierarchy.
				thirdParty.Level = parent.Level + 1
			}

			levelValidator := validator.New()
			levelValidator.Check(thirdParty.Level, "level", validator.Max(coredata.MaxThirdPartyLevel))

			if err := levelValidator.Error(); err != nil {
				return err
			}

			if req.BusinessOwnerID != nil {
				businessOwner := &coredata.MembershipProfile{}
				if err := businessOwner.LoadByID(ctx, conn, scope, *req.BusinessOwnerID); err != nil {
					return fmt.Errorf("cannot load business owner profile: %w", err)
				}

				thirdParty.BusinessOwnerID = &businessOwner.ID
			}

			if req.SecurityOwnerID != nil {
				securityOwner := &coredata.MembershipProfile{}
				if err := securityOwner.LoadByID(ctx, conn, scope, *req.SecurityOwnerID); err != nil {
					return fmt.Errorf("cannot load security owner profile: %w", err)
				}

				thirdParty.SecurityOwnerID = &securityOwner.ID
			}

			if req.Category != nil {
				thirdParty.Category = *req.Category
			} else {
				thirdParty.Category = coredata.ThirdPartyCategoryOther
			}

			if err := thirdParty.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert thirdParty: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				conn,
				scope,
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

// ImportFromCommon creates an org ThirdParty seeded from the global
// CommonThirdParty catalog entry, or returns the existing one when the
// organization already imported it (idempotent on the
// (organization_id, common_third_party_id) pair). It then backfills
// tracker_patterns.third_party_id for the organization's unlinked
// patterns whose catalog row resolves to the same common third party, so
// the trackers UI and tracker-policy document surface the managed vendor
// instead of the catalog entry. The backfill runs on both the create and
// reuse paths because new patterns may have been detected since a prior
// import; it only touches patterns with no existing link. Returns the
// org ThirdParty and whether it was newly created.
func (s ThirdPartyService) ImportFromCommon(
	ctx context.Context, scope coredata.Scoper,
	req ImportThirdPartyFromCommonRequest,
) (*coredata.ThirdParty, bool, error) {
	if err := req.Validate(); err != nil {
		return nil, false, err
	}

	thirdParty := &coredata.ThirdParty{}
	created := false

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			existing := &coredata.ThirdParty{}

			err := existing.LoadByOrganizationIDAndCommonThirdPartyID(
				ctx,
				conn,
				scope,
				organization.ID,
				req.CommonThirdPartyID,
			)

			switch {
			case err == nil:
				*thirdParty = *existing
			case errors.Is(err, coredata.ErrResourceNotFound):
				commonParty := &coredata.CommonThirdParty{}
				if err := commonParty.LoadByID(ctx, conn, req.CommonThirdPartyID); err != nil {
					return fmt.Errorf("cannot load common third party: %w", err)
				}

				now := time.Now()
				commonID := commonParty.ID

				certifications := commonParty.Certifications
				if certifications == nil {
					certifications = []string{}
				}

				*thirdParty = coredata.ThirdParty{
					ID:                            gid.New(scope.GetTenantID(), coredata.ThirdPartyEntityType),
					OrganizationID:                organization.ID,
					CommonThirdPartyID:            &commonID,
					Name:                          commonParty.Name,
					Category:                      commonParty.Category,
					HeadquarterAddress:            commonParty.HeadquarterAddress,
					LegalName:                     commonParty.LegalName,
					WebsiteURL:                    commonParty.WebsiteURL,
					PrivacyPolicyURL:              commonParty.PrivacyPolicyURL,
					ServiceLevelAgreementURL:      commonParty.ServiceLevelAgreementURL,
					DataProcessingAgreementURL:    commonParty.DataProcessingAgreementURL,
					BusinessAssociateAgreementURL: commonParty.BusinessAssociateAgreementURL,
					SubprocessorsListURL:          commonParty.SubprocessorsListURL,
					Certifications:                certifications,
					Countries:                     coredata.CountryCodes{},
					StatusPageURL:                 commonParty.StatusPageURL,
					TermsOfServiceURL:             commonParty.TermsOfServiceURL,
					SecurityPageURL:               commonParty.SecurityPageURL,
					TrustPageURL:                  commonParty.TrustPageURL,
					ShowOnTrustCenter:             false,
					Level:                         1,
					CreatedAt:                     now,
					UpdatedAt:                     now,
				}

				if err := thirdParty.Insert(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot insert third party: %w", err)
				}

				created = true

				if err := webhook.InsertData(
					ctx,
					conn,
					scope,
					organization.ID,
					coredata.WebhookEventTypeThirdPartyCreated,
					webhooktypes.NewThirdParty(thirdParty),
				); err != nil {
					return fmt.Errorf("cannot insert webhook event: %w", err)
				}
			default:
				return fmt.Errorf("cannot load third party by common id: %w", err)
			}

			var patterns coredata.TrackerPatterns
			if err := patterns.LinkThirdPartyByCommonThirdPartyID(
				ctx,
				conn,
				scope,
				organization.ID,
				req.CommonThirdPartyID,
				thirdParty.ID,
			); err != nil {
				return fmt.Errorf("cannot link tracker patterns to imported third party: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, false, err
	}

	return thirdParty, created, nil
}

func (s ThirdPartyService) CountForAssetID(
	ctx context.Context, scope coredata.Scoper,
	assetID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			thirdParties := coredata.ThirdParties{}

			count, err = thirdParties.CountByAssetID(ctx, conn, scope, assetID)
			if err != nil {
				return fmt.Errorf("cannot count thirdParties: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
	assetID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdParties.LoadByAssetID(ctx, conn, scope, assetID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdParties, cursor), nil
}

func (s ThirdPartyService) ListForProcessingActivityID(
	ctx context.Context, scope coredata.Scoper,
	processingActivityID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := thirdParties.LoadByProcessingActivityID(ctx, conn, scope, processingActivityID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load thirdParties by processing activity: %w", err)
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
	ctx context.Context, scope coredata.Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyRiskAssessmentOrderField],
) (*page.Page[*coredata.ThirdPartyRiskAssessment, coredata.ThirdPartyRiskAssessmentOrderField], error) {
	var thirdPartyRiskAssessments coredata.ThirdPartyRiskAssessments

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdPartyRiskAssessments.LoadByThirdPartyID(ctx, conn, scope, thirdPartyID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdPartyRiskAssessments, cursor), nil
}

func (s ThirdPartyService) CreateRiskAssessment(
	ctx context.Context, scope coredata.Scoper,
	req CreateThirdPartyRiskAssessmentRequest,
) (*coredata.ThirdPartyRiskAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	thirdPartyRiskAssessmentID := gid.New(scope.GetTenantID(), coredata.ThirdPartyRiskAssessmentEntityType)

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
			if err := thirdParty.LoadByID(ctx, tx, scope, req.ThirdPartyID); err != nil {
				return fmt.Errorf("cannot load thirdParty: %w", err)
			}

			thirdPartyRiskAssessment.OrganizationID = thirdParty.OrganizationID

			if err := thirdParty.ExpireNonExpiredRiskAssessments(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot expire thirdParty risk assessments: %w", err)
			}

			if err := thirdPartyRiskAssessment.Insert(ctx, tx, scope); err != nil {
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
	ctx context.Context, scope coredata.Scoper,
	thirdPartyRiskAssessmentID gid.GID,
) (*coredata.ThirdPartyRiskAssessment, error) {
	thirdPartyRiskAssessment := &coredata.ThirdPartyRiskAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdPartyRiskAssessment.LoadByID(ctx, conn, scope, thirdPartyRiskAssessmentID)
		},
	)
	if err != nil {
		return nil, err
	}

	return thirdPartyRiskAssessment, nil
}

func (s ThirdPartyService) GetByRiskAssessmentID(
	ctx context.Context, scope coredata.Scoper,
	thirdPartyRiskAssessmentID gid.GID,
) (*coredata.ThirdParty, error) {
	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			thirdPartyRiskAssessment := &coredata.ThirdPartyRiskAssessment{}
			if err := thirdPartyRiskAssessment.LoadByID(ctx, conn, scope, thirdPartyRiskAssessmentID); err != nil {
				return fmt.Errorf("cannot load thirdParty risk assessment: %w", err)
			}

			if err := thirdParty.LoadByID(ctx, conn, scope, thirdPartyRiskAssessment.ThirdPartyID); err != nil {
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

func (s ThirdPartyService) GetAncestors(
	ctx context.Context,
	scope coredata.Scoper,
	thirdPartyID gid.GID,
) (coredata.ThirdParties, error) {
	var ancestors coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return ancestors.LoadAllAncestorsByThirdPartyID(ctx, conn, scope, thirdPartyID)
		},
	)
	if err != nil {
		return nil, err
	}

	return ancestors, nil
}

func (s ThirdPartyService) CountForParentThirdPartyID(
	ctx context.Context,
	scope coredata.Scoper,
	parentThirdPartyID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			thirdParties := coredata.ThirdParties{}

			count, err = thirdParties.CountByParentThirdPartyID(ctx, conn, scope, parentThirdPartyID)
			if err != nil {
				return fmt.Errorf("cannot count child third parties: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ThirdPartyService) ListForParentThirdPartyID(
	ctx context.Context,
	scope coredata.Scoper,
	parentThirdPartyID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyOrderField],
) (*page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField], error) {
	var thirdParties coredata.ThirdParties

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return thirdParties.LoadByParentThirdPartyID(ctx, conn, scope, parentThirdPartyID, cursor)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdParties, cursor), nil
}
