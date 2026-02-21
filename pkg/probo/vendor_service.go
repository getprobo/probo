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
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

type (
	VendorService struct {
		svc *TenantService
	}

	CreateVendorRequest struct {
		OrganizationID                gid.GID
		Name                          string
		Description                   *string
		HeadquarterAddress            *string
		LegalName                     *string
		WebsiteURL                    *string
		Category                      *coredata.VendorCategory
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

	UpdateVendorRequest struct {
		ID                            gid.GID
		Name                          *string
		Description                   **string
		HeadquarterAddress            **string
		LegalName                     **string
		WebsiteURL                    **string
		TermsOfServiceURL             **string
		Category                      *coredata.VendorCategory
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

	AssessVendorRequest struct {
		ID         gid.GID
		WebsiteURL string
	}

	CreateVendorRiskAssessmentRequest struct {
		VendorID        gid.GID
		ExpiresAt       time.Time
		DataSensitivity coredata.DataSensitivity
		BusinessImpact  coredata.BusinessImpact
		Notes           *string
	}
)

func (cvr *CreateVendorRequest) Validate() error {
	v := validator.New()

	v.Check(cvr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(cvr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cvr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(cvr.HeadquarterAddress, "headquarter_address", validator.SafeText(ContentMaxLength))
	v.Check(cvr.LegalName, "cvr.LegalName", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cvr.WebsiteURL, "website_url", validator.SafeText(2048))
	v.Check(cvr.Category, "category", validator.OneOfSlice(coredata.VendorCategories()))
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

func (uvr *UpdateVendorRequest) Validate() error {
	v := validator.New()

	v.Check(uvr.ID, "id", validator.Required(), validator.GID(coredata.VendorEntityType))
	v.Check(uvr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uvr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(uvr.HeadquarterAddress, "headquarter_address", validator.SafeText(ContentMaxLength))
	v.Check(uvr.LegalName, "uvr.LegalName", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uvr.WebsiteURL, "website_url", validator.SafeText(2048))
	v.Check(uvr.Category, "category", validator.OneOfSlice(coredata.VendorCategories()))
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

func (cvrar *CreateVendorRiskAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(cvrar.VendorID, "vendor_id", validator.Required(), validator.GID(coredata.VendorEntityType))
	v.Check(cvrar.DataSensitivity, "data_sensitivity", validator.Required(), validator.OneOfSlice(coredata.DataSensitivities()))
	v.Check(cvrar.BusinessImpact, "business_impact", validator.Required(), validator.OneOfSlice(coredata.BusinessImpacts()))
	v.Check(cvrar.Notes, "notes", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s VendorService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			vendors := coredata.Vendors{}
			filter := &coredata.VendorFilter{}
			count, err = vendors.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count vendors: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s VendorService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.VendorOrderField],
	filter *coredata.VendorFilter,
) (*page.Page[*coredata.Vendor, coredata.VendorOrderField], error) {
	var vendors coredata.Vendors
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			return vendors.LoadByOrganizationID(
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

	return page.NewPage(vendors, cursor), nil
}

func (s VendorService) CountForDatumID(
	ctx context.Context,
	datumID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			vendors := coredata.Vendors{}
			count, err = vendors.CountByDatumID(ctx, conn, s.svc.scope, datumID)
			if err != nil {
				return fmt.Errorf("cannot count vendors: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s VendorService) ListForDatumID(
	ctx context.Context,
	datumID gid.GID,
	cursor *page.Cursor[coredata.VendorOrderField],
) (*page.Page[*coredata.Vendor, coredata.VendorOrderField], error) {
	var vendors coredata.Vendors

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendors.LoadByDatumID(
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

	return page.NewPage(vendors, cursor), nil
}

func (s VendorService) Update(
	ctx context.Context,
	req UpdateVendorRequest,
) (*coredata.Vendor, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	vendor := &coredata.Vendor{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := vendor.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load vendor %q: %w", req.ID, err)
			}

			if req.Name != nil {
				vendor.Name = *req.Name
			}

			if req.Description != nil {
				vendor.Description = *req.Description
			}

			if req.StatusPageURL != nil {
				vendor.StatusPageURL = *req.StatusPageURL
			}

			if req.TermsOfServiceURL != nil {
				vendor.TermsOfServiceURL = *req.TermsOfServiceURL
			}

			if req.PrivacyPolicyURL != nil {
				vendor.PrivacyPolicyURL = *req.PrivacyPolicyURL
			}

			if req.ServiceLevelAgreementURL != nil {
				vendor.ServiceLevelAgreementURL = *req.ServiceLevelAgreementURL
			}

			if req.DataProcessingAgreementURL != nil {
				vendor.DataProcessingAgreementURL = *req.DataProcessingAgreementURL
			}

			if req.BusinessAssociateAgreementURL != nil {
				vendor.BusinessAssociateAgreementURL = *req.BusinessAssociateAgreementURL
			}

			if req.SubprocessorsListURL != nil {
				vendor.SubprocessorsListURL = *req.SubprocessorsListURL
			}

			if req.Category != nil {
				vendor.Category = *req.Category
			} else {
				vendor.Category = coredata.VendorCategoryOther
			}

			if req.SecurityPageURL != nil {
				vendor.SecurityPageURL = *req.SecurityPageURL
			}

			if req.ShowOnTrustCenter != nil {
				vendor.ShowOnTrustCenter = *req.ShowOnTrustCenter
			}

			if req.TrustPageURL != nil {
				vendor.TrustPageURL = *req.TrustPageURL
			}

			if req.HeadquarterAddress != nil {
				vendor.HeadquarterAddress = *req.HeadquarterAddress
			}

			if req.LegalName != nil {
				vendor.LegalName = *req.LegalName
			}

			if req.WebsiteURL != nil {
				vendor.WebsiteURL = *req.WebsiteURL
			}

			if req.TermsOfServiceURL != nil {
				vendor.TermsOfServiceURL = *req.TermsOfServiceURL
			}

			if req.Certifications != nil {
				vendor.Certifications = req.Certifications
			}

			if req.Countries != nil {
				vendor.Countries = req.Countries
			}

			if req.BusinessOwnerID != nil {
				if *req.BusinessOwnerID != nil {
					businessOwner := &coredata.MembershipProfile{}
					if err := businessOwner.LoadByID(ctx, conn, s.svc.scope, **req.BusinessOwnerID); err != nil {
						return fmt.Errorf("cannot load business owner profile: %w", err)
					}
					vendor.BusinessOwnerID = &businessOwner.ID
				} else {
					vendor.BusinessOwnerID = nil
				}
			}

			if req.SecurityOwnerID != nil {
				if *req.SecurityOwnerID != nil {
					securityOwner := &coredata.MembershipProfile{}
					if err := securityOwner.LoadByID(ctx, conn, s.svc.scope, **req.SecurityOwnerID); err != nil {
						return fmt.Errorf("cannot load security owner profile: %w", err)
					}
					vendor.SecurityOwnerID = &securityOwner.ID
				} else {
					vendor.SecurityOwnerID = nil
				}
			}

			vendor.UpdatedAt = time.Now()

			if err := vendor.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update vendor: %w", err)
			}

			if err := webhook.InsertData(ctx, conn, s.svc.scope, vendor.OrganizationID, coredata.WebhookEventTypeVendorUpdated, webhooktypes.NewVendor(vendor)); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return vendor, nil
}

func (s VendorService) Get(
	ctx context.Context,
	vendorID gid.GID,
) (*coredata.Vendor, error) {
	vendor := &coredata.Vendor{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendor.LoadByID(ctx, conn, s.svc.scope, vendorID)
		},
	)

	if err != nil {
		return nil, err
	}

	return vendor, nil
}

func (s VendorService) Delete(
	ctx context.Context,
	vendorID gid.GID,
) error {
	vendor := &coredata.Vendor{}

	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := vendor.LoadByID(ctx, conn, s.svc.scope, vendorID); err != nil {
				return fmt.Errorf("cannot load vendor: %w", err)
			}

			if err := webhook.InsertData(ctx, conn, s.svc.scope, vendor.OrganizationID, coredata.WebhookEventTypeVendorDeleted, webhooktypes.NewVendor(vendor)); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return vendor.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s VendorService) Create(
	ctx context.Context,
	req CreateVendorRequest,
) (*coredata.Vendor, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	vendor := &coredata.Vendor{
		ID:                            gid.New(s.svc.scope.GetTenantID(), coredata.VendorEntityType),
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
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization %q: %w", req.OrganizationID, err)
			}

			vendor.OrganizationID = organization.ID

			if req.BusinessOwnerID != nil {
				businessOwner := &coredata.MembershipProfile{}
				if err := businessOwner.LoadByID(ctx, conn, s.svc.scope, *req.BusinessOwnerID); err != nil {
					return fmt.Errorf("cannot load business owner profile: %w", err)
				}
				vendor.BusinessOwnerID = &businessOwner.ID
			}

			if req.SecurityOwnerID != nil {
				securityOwner := &coredata.MembershipProfile{}
				if err := securityOwner.LoadByID(ctx, conn, s.svc.scope, *req.SecurityOwnerID); err != nil {
					return fmt.Errorf("cannot load security owner profile: %w", err)
				}
				vendor.SecurityOwnerID = &securityOwner.ID
			}

			if req.Category != nil {
				vendor.Category = *req.Category
			} else {
				vendor.Category = coredata.VendorCategoryOther
			}

			if err := vendor.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert vendor: %w", err)
			}

			if err := webhook.InsertData(ctx, conn, s.svc.scope, organization.ID, coredata.WebhookEventTypeVendorCreated, webhooktypes.NewVendor(vendor)); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return vendor, nil
}

func (s VendorService) CountForAssetID(
	ctx context.Context,
	assetID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			vendors := coredata.Vendors{}
			count, err = vendors.CountByAssetID(ctx, conn, s.svc.scope, assetID)
			if err != nil {
				return fmt.Errorf("cannot count vendors: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s VendorService) ListForAssetID(
	ctx context.Context,
	assetID gid.GID,
	cursor *page.Cursor[coredata.VendorOrderField],
) (*page.Page[*coredata.Vendor, coredata.VendorOrderField], error) {
	var vendors coredata.Vendors

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendors.LoadByAssetID(ctx, conn, s.svc.scope, assetID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(vendors, cursor), nil
}

func (s VendorService) ListForProcessingActivityID(
	ctx context.Context,
	processingActivityID gid.GID,
	cursor *page.Cursor[coredata.VendorOrderField],
) (*page.Page[*coredata.Vendor, coredata.VendorOrderField], error) {
	var vendors coredata.Vendors

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := vendors.LoadByProcessingActivityID(ctx, conn, s.svc.scope, processingActivityID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load vendors by processing activity: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(vendors, cursor), nil
}

func (s VendorService) ListRiskAssessments(
	ctx context.Context,
	vendorID gid.GID,
	cursor *page.Cursor[coredata.VendorRiskAssessmentOrderField],
) (*page.Page[*coredata.VendorRiskAssessment, coredata.VendorRiskAssessmentOrderField], error) {
	var vendorRiskAssessments coredata.VendorRiskAssessments

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendorRiskAssessments.LoadByVendorID(ctx, conn, s.svc.scope, vendorID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(vendorRiskAssessments, cursor), nil
}

func (s VendorService) CreateRiskAssessment(
	ctx context.Context,
	req CreateVendorRiskAssessmentRequest,
) (*coredata.VendorRiskAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	vendorRiskAssessmentID := gid.New(s.svc.scope.GetTenantID(), coredata.VendorRiskAssessmentEntityType)

	now := time.Now()

	vendorRiskAssessment := &coredata.VendorRiskAssessment{
		ID:              vendorRiskAssessmentID,
		VendorID:        req.VendorID,
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
		func(tx pg.Conn) error {
			vendor := coredata.Vendor{}
			if err := vendor.LoadByID(ctx, tx, s.svc.scope, req.VendorID); err != nil {
				return fmt.Errorf("cannot load vendor: %w", err)
			}

			vendorRiskAssessment.OrganizationID = vendor.OrganizationID

			if err := vendor.ExpireNonExpiredRiskAssessments(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot expire vendor risk assessments: %w", err)
			}

			if err := vendorRiskAssessment.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert vendor risk assessment: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return vendorRiskAssessment, nil
}

func (s VendorService) GetRiskAssessment(
	ctx context.Context,
	vendorRiskAssessmentID gid.GID,
) (*coredata.VendorRiskAssessment, error) {
	vendorRiskAssessment := &coredata.VendorRiskAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendorRiskAssessment.LoadByID(ctx, conn, s.svc.scope, vendorRiskAssessmentID)
		},
	)

	if err != nil {
		return nil, err
	}

	return vendorRiskAssessment, nil
}

func (s VendorService) GetByRiskAssessmentID(
	ctx context.Context,
	vendorRiskAssessmentID gid.GID,
) (*coredata.Vendor, error) {
	vendor := &coredata.Vendor{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			vendorRiskAssessment := &coredata.VendorRiskAssessment{}
			if err := vendorRiskAssessment.LoadByID(ctx, conn, s.svc.scope, vendorRiskAssessmentID); err != nil {
				return fmt.Errorf("cannot load vendor risk assessment: %w", err)
			}

			if err := vendor.LoadByID(ctx, conn, s.svc.scope, vendorRiskAssessment.VendorID); err != nil {
				return fmt.Errorf("cannot load vendor: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return vendor, nil
}

func (s VendorService) Assess(
	ctx context.Context,
	req AssessVendorRequest,
) (*coredata.Vendor, error) {
	vendorInfo, err := s.svc.agent.AssessVendor(ctx, req.WebsiteURL)
	if err != nil {
		return nil, fmt.Errorf("cannot assess vendor info: %w", err)
	}

	vendor := &coredata.Vendor{
		ID:                            req.ID,
		Name:                          vendorInfo.Name,
		WebsiteURL:                    &req.WebsiteURL,
		Description:                   &vendorInfo.Description,
		Category:                      coredata.VendorCategory(vendorInfo.Category),
		HeadquarterAddress:            &vendorInfo.HeadquarterAddress,
		LegalName:                     &vendorInfo.LegalName,
		PrivacyPolicyURL:              &vendorInfo.PrivacyPolicyURL,
		ServiceLevelAgreementURL:      &vendorInfo.ServiceLevelAgreementURL,
		DataProcessingAgreementURL:    &vendorInfo.DataProcessingAgreementURL,
		BusinessAssociateAgreementURL: &vendorInfo.BusinessAssociateAgreementURL,
		SubprocessorsListURL:          &vendorInfo.SubprocessorsListURL,
		SecurityPageURL:               &vendorInfo.SecurityPageURL,
		TrustPageURL:                  &vendorInfo.TrustPageURL,
		TermsOfServiceURL:             &vendorInfo.TermsOfServiceURL,
		StatusPageURL:                 &vendorInfo.StatusPageURL,
		Certifications:                vendorInfo.Certifications,
		UpdatedAt:                     time.Now(),
	}

	return vendor, nil
}
