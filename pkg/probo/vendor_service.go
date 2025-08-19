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
		Description                   *string
		HeadquarterAddress            *string
		LegalName                     *string
		WebsiteURL                    *string
		TermsOfServiceURL             *string
		Category                      *coredata.VendorCategory
		PrivacyPolicyURL              *string
		ServiceLevelAgreementURL      *string
		DataProcessingAgreementURL    *string
		BusinessAssociateAgreementURL *string
		SubprocessorsListURL          *string
		Certifications                []string
		SecurityPageURL               *string
		TrustPageURL                  *string
		StatusPageURL                 *string
		BusinessOwnerID               *gid.GID
		SecurityOwnerID               *gid.GID
		ShowOnTrustCenter             *bool
	}

	AssessVendorRequest struct {
		ID         gid.GID
		WebsiteURL string
	}

	CreateVendorRiskAssessmentRequest struct {
		VendorID        gid.GID
		AssessedByID    gid.GID
		ExpiresAt       time.Time
		DataSensitivity coredata.DataSensitivity
		BusinessImpact  coredata.BusinessImpact
		Notes           *string
	}
)

func (s VendorService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			vendors := coredata.Vendors{}
			count, err = vendors.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
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
				vendor.Description = req.Description
			}

			if req.StatusPageURL != nil {
				vendor.StatusPageURL = req.StatusPageURL
			}

			if req.TermsOfServiceURL != nil {
				vendor.TermsOfServiceURL = req.TermsOfServiceURL
			}

			if req.PrivacyPolicyURL != nil {
				vendor.PrivacyPolicyURL = req.PrivacyPolicyURL
			}

			if req.ServiceLevelAgreementURL != nil {
				vendor.ServiceLevelAgreementURL = req.ServiceLevelAgreementURL
			}

			if req.DataProcessingAgreementURL != nil {
				vendor.DataProcessingAgreementURL = req.DataProcessingAgreementURL
			}

			if req.BusinessAssociateAgreementURL != nil {
				vendor.BusinessAssociateAgreementURL = req.BusinessAssociateAgreementURL
			}

			if req.SubprocessorsListURL != nil {
				vendor.SubprocessorsListURL = req.SubprocessorsListURL
			}

			if req.Category != nil {
				vendor.Category = *req.Category
			} else {
				vendor.Category = coredata.VendorCategoryOther
			}

			if req.SecurityPageURL != nil {
				vendor.SecurityPageURL = req.SecurityPageURL
			}

			if req.ShowOnTrustCenter != nil {
				vendor.ShowOnTrustCenter = *req.ShowOnTrustCenter
			}

			if req.TrustPageURL != nil {
				vendor.TrustPageURL = req.TrustPageURL
			}

			if req.HeadquarterAddress != nil {
				vendor.HeadquarterAddress = req.HeadquarterAddress
			}

			if req.LegalName != nil {
				vendor.LegalName = req.LegalName
			}

			if req.WebsiteURL != nil {
				vendor.WebsiteURL = req.WebsiteURL
			}

			if req.TermsOfServiceURL != nil {
				vendor.TermsOfServiceURL = req.TermsOfServiceURL
			}

			if req.Certifications != nil {
				vendor.Certifications = req.Certifications
			}

			if req.StatusPageURL != nil {
				vendor.StatusPageURL = req.StatusPageURL
			}

			if req.SecurityPageURL != nil {
				vendor.SecurityPageURL = req.SecurityPageURL
			}

			if req.TrustPageURL != nil {
				vendor.TrustPageURL = req.TrustPageURL
			}

			if req.BusinessOwnerID != nil {
				businessOwner := &coredata.People{}
				if err := businessOwner.LoadByID(ctx, conn, s.svc.scope, *req.BusinessOwnerID); err != nil {
					return fmt.Errorf("cannot load business owner: %w", err)
				}
				vendor.BusinessOwnerID = &businessOwner.ID
			}

			if req.SecurityOwnerID != nil {
				securityOwner := &coredata.People{}
				if err := securityOwner.LoadByID(ctx, conn, s.svc.scope, *req.SecurityOwnerID); err != nil {
					return fmt.Errorf("cannot load security owner: %w", err)
				}
				vendor.SecurityOwnerID = &securityOwner.ID
			}

			vendor.UpdatedAt = time.Now()

			if err := vendor.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update vendor: %w", err)
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
	vendor := coredata.Vendor{ID: vendorID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendor.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s VendorService) Create(
	ctx context.Context,
	req CreateVendorRequest,
) (*coredata.Vendor, error) {
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
				businessOwner := &coredata.People{}
				if err := businessOwner.LoadByID(ctx, conn, s.svc.scope, *req.BusinessOwnerID); err != nil {
					return fmt.Errorf("cannot load business owner: %w", err)
				}
				vendor.BusinessOwnerID = &businessOwner.ID
			}

			if req.SecurityOwnerID != nil {
				securityOwner := &coredata.People{}
				if err := securityOwner.LoadByID(ctx, conn, s.svc.scope, *req.SecurityOwnerID); err != nil {
					return fmt.Errorf("cannot load security owner: %w", err)
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
	vendorRiskAssessmentID := gid.New(s.svc.scope.GetTenantID(), coredata.VendorRiskAssessmentEntityType)

	now := time.Now()

	vendorRiskAssessment := &coredata.VendorRiskAssessment{
		ID:              vendorRiskAssessmentID,
		VendorID:        req.VendorID,
		AssessedBy:      req.AssessedByID,
		AssessedAt:      now,
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
			vendor := coredata.Vendor{ID: req.VendorID}
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

func (s VendorService) Assess(
	ctx context.Context,
	req AssessVendorRequest,
) (*coredata.Vendor, error) {
	vendorInfo, err := s.svc.agent.AssessVendor(ctx, req.WebsiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to assess vendor info: %w", err)
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
