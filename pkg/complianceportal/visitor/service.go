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

package visitor

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/slack"
)

const DefaultNDAContactEmail = "security@probo.com"

func ndaConsentText(contactEmail string) string {
	if contactEmail == "" {
		contactEmail = DefaultNDAContactEmail
	}

	return fmt.Sprintf(
		"By clicking \"Review and sign\", I consent to sign this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature. If you have questions about the NDA, please contact %s.",
		contactEmail,
	)
}

type (
	// Service is the visitor-facing compliance portal service. It exposes the
	// public read operations for the compliance page and its related resources as
	// methods on a single type.
	Service struct {
		pg                *pg.Client
		s3                *s3.Client
		bucket            string
		baseURL           string
		esign             *esign.Service
		html2pdfConverter *html2pdf.Converter
		fileManager       *filemanager.Service
		logger            *log.Logger
		slack             *slack.Service
		resourceAlias     *resourcealias.Service
		management        *management.Service
	}
)

func NewService(
	pgClient *pg.Client,
	s3Client *s3.Client,
	bucket string,
	baseURL string,
	esignSvc *esign.Service,
	html2pdfConverter *html2pdf.Converter,
	fileManagerService *filemanager.Service,
	logger *log.Logger,
	slack *slack.Service,
	resourceAliasSvc *resourcealias.Service,
	managementSvc *management.Service,
) *Service {
	svc := &Service{
		pg:                pgClient,
		s3:                s3Client,
		bucket:            bucket,
		baseURL:           baseURL,
		esign:             esignSvc,
		html2pdfConverter: html2pdfConverter,
		fileManager:       fileManagerService,
		logger:            logger,
		slack:             slack,
		resourceAlias:     resourceAliasSvc,
		management:        managementSvc,
	}

	return svc
}

func (s *Service) GetPortalByID(
	ctx context.Context,
	id gid.GID,
) (*coredata.CompliancePortal, error) {
	compliancePage := &coredata.CompliancePortal{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return loadPortalByID(ctx, conn, coredata.NewNoScope(), id, compliancePage)
		},
	)
	if err != nil {
		return nil, err
	}

	return compliancePage, nil
}

func loadPortalByID(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	id gid.GID,
	compliancePage *coredata.CompliancePortal,
) error {
	if err := compliancePage.LoadByID(ctx, conn, scope, id); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return ErrPageNotFound
		}

		return fmt.Errorf("cannot load compliance page: %w", err)
	}

	return nil
}

func (s *Service) GetPortalEffectiveCanonicalHost(ctx context.Context, compliancePageID gid.GID) (string, error) {
	var host string

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			compliancePage := &coredata.CompliancePortal{}
			if err := compliancePage.LoadByID(ctx, conn, coredata.NewNoScope(), compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			domain, err := s.management.EffectiveDomainForCompliancePortal(ctx, conn, coredata.NewNoScope(), compliancePage)
			if err != nil {
				return err
			}

			if domain != nil {
				host = domain.Domain
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return host, nil
}

func (s *Service) GetPortalCanonicalBaseURL(
	ctx context.Context,
	compliancePageID gid.GID,
	currentBaseURL string,
) (string, error) {
	canonicalHost, err := s.GetPortalEffectiveCanonicalHost(ctx, compliancePageID)
	if err != nil {
		return "", fmt.Errorf("cannot resolve canonical host: %w", err)
	}

	if canonicalHost == "" {
		return currentBaseURL, nil
	}

	parsed, err := url.Parse(currentBaseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse portal base URL: %w", err)
	}

	parsed.Host = canonicalHost

	return parsed.String(), nil
}

func (s *Service) GetPortalByDomainName(ctx context.Context, domain string) (*coredata.CompliancePortal, error) {
	compliancePage := &coredata.CompliancePortal{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var customDomain coredata.CustomDomain
			if err := customDomain.LoadByDomain(ctx, conn, coredata.NewNoScope(), domain); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrPageNotFound
				}

				return fmt.Errorf("cannot load custom domain: %w", err)
			}

			compliancePage = &coredata.CompliancePortal{}
			if err := compliancePage.LoadByDomainID(ctx, conn, customDomain.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrPageNotFound
				}

				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return compliancePage, err
}

func (s *Service) IsVerifiedRedirectHost(ctx context.Context, host string) bool {
	if _, err := s.GetPortalByDomainName(ctx, host); err != nil {
		return false
	}

	verified, err := s.management.IsCustomDomainVerified(ctx, host)
	if err != nil {
		s.logger.ErrorCtx(ctx, "cannot check custom domain verification", log.Error(err), log.String("host", host))
		return false
	}

	return verified
}

func (s *Service) GetEmailPresenterConfigForSignature(
	ctx context.Context,
	signatureID gid.GID,
	organizationID gid.GID,
) (emails.PresenterConfig, error) {
	access := &coredata.CompliancePortalAccess{}
	scope := coredata.NewScopeFromObjectID(organizationID)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return access.LoadByElectronicSignatureID(ctx, conn, scope, signatureID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return emails.DefaultPresenterConfig(s.baseURL), nil
		}

		return emails.PresenterConfig{}, fmt.Errorf("cannot load compliance portal access for signature: %w", err)
	}

	return s.GetPortalEmailPresenterConfig(ctx, scope, access.CompliancePortalID)
}

func (s *Service) GetPortalOrganization(
	ctx context.Context,
	compliancePageID gid.GID,
) (*coredata.Organization, error) {
	compliancePage, err := s.GetPortalByID(ctx, compliancePageID)
	if err != nil {
		return nil, fmt.Errorf("cannot load compliance page: %w", err)
	}

	org := &coredata.Organization{}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return org.LoadByID(ctx, conn, coredata.NewNoScope(), compliancePage.OrganizationID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load organization: %w", err)
	}

	return org, nil
}

func (s *Service) GetPortalMembership(ctx context.Context, compliancePageID gid.GID, identityID gid.GID) (*coredata.CompliancePortalAccess, error) {
	membership := &coredata.CompliancePortalAccess{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return membership.LoadByCompliancePortalIDAndIdentityID(
				ctx,
				conn,
				coredata.NewScopeFromObjectID(compliancePageID),
				compliancePageID,
				identityID,
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrMembershipNotFound
		}

		return nil, err
	}

	return membership, nil
}

func (s *Service) ProvisionPortalMember(
	ctx context.Context,
	compliancePageID gid.GID,
	identityID gid.GID,
) (*coredata.CompliancePortalAccess, error) {
	var (
		access *coredata.CompliancePortalAccess
		now    = time.Now()
		scope  = coredata.NewScopeFromObjectID(compliancePageID)
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			compliancePage := &coredata.CompliancePortal{}
			if err := compliancePage.LoadByID(ctx, tx, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			identity := &coredata.Identity{}
			if err := identity.LoadByID(ctx, tx, identityID); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			access = &coredata.CompliancePortalAccess{}
			if err := access.LoadByCompliancePortalIDAndIdentityID(ctx, tx, scope, compliancePageID, identityID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load compliance page access: %w", err)
				}

				access = &coredata.CompliancePortalAccess{
					ID:                 gid.New(scope.GetTenantID(), coredata.CompliancePortalAccessEntityType),
					OrganizationID:     compliancePage.OrganizationID,
					TenantID:           scope.GetTenantID(),
					IdentityID:         identityID,
					CompliancePortalID: compliancePageID,
					CreatedAt:          now,
					UpdatedAt:          now,
				}

				var sig *coredata.ElectronicSignature

				if compliancePage.NonDisclosureAgreementFileID != nil && s.esign != nil {
					var err error

					sig, err = s.esign.CreateSignature(
						ctx,
						tx,
						&esign.CreateSignatureRequest{
							OrganizationID: access.OrganizationID,
							DocumentType:   coredata.ElectronicSignatureDocumentTypeNDA,
							FileID:         *compliancePage.NonDisclosureAgreementFileID,
							SignerEmail:    identity.EmailAddress,
							ConsentText:    ndaConsentText(ref.UnrefOrZero(compliancePage.Email)),
						},
					)
					if err != nil {
						return fmt.Errorf("cannot create pending signature: %w", err)
					}
				}

				if sig != nil {
					access.ElectronicSignatureID = &sig.ID
				}

				if err := access.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert compliance page access: %w", err)
				}
			}

			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByIdentityIDAndOrganizationID(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(access.ID),
				identityID,
				access.OrganizationID,
			); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load profile: %w", err)
				}

				profile = &coredata.MembershipProfile{
					ID:             gid.New(access.TenantID, coredata.MembershipProfileEntityType),
					IdentityID:     identityID,
					OrganizationID: access.OrganizationID,
					EmailAddress:   identity.EmailAddress,
					Source:         coredata.ProfileSourceManual,
					State:          coredata.ProfileStateActive,
					ActivatedAt:    &now,
					FullName:       identity.FullName,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := profile.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert profile: %w", err)
				}
			}

			return nil
		},
	)

	return access, err
}
