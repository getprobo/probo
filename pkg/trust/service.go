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

package trust

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/slack"
)

const NDAConsentText = "By clicking \"Review and sign\", I consent to sign this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature. If you have questions about the NDA, please contact security@probo.com."

type (
	Service struct {
		pg                    *pg.Client
		s3                    *s3.Client
		bucket                string
		proboSvc              *probo.Service
		slackSigningSecret    string
		baseURL               string
		iam                   *iam.Service
		esign                 *esign.Service
		html2pdfConverter     *html2pdf.Converter
		fileManager           *filemanager.Service
		logger                *log.Logger
		slack                 *slack.Service
		TrustCenters          *TrustCenterService
		Documents             *DocumentService
		Audits                *AuditService
		ThirdParties          *ThirdPartyService
		Frameworks            *FrameworkService
		ComplianceFrameworks  *ComplianceFrameworkService
		TrustCenterAccesses   *TrustCenterAccessService
		TrustCenterReferences *TrustCenterReferenceService

		CompliancePortalCommitmentGroups *CompliancePortalCommitmentGroupService
		CompliancePortalCommitments      *CompliancePortalCommitmentService

		TrustCenterFiles       *TrustCenterFileService
		Reports                *ReportService
		Organizations          *OrganizationService
		ComplianceExternalURLs *ComplianceExternalURLService
		resourceAlias          *resourcealias.Service
	}
)

func NewService(
	pgClient *pg.Client,
	s3Client *s3.Client,
	bucket string,
	baseURL string,
	slackSigningSecret string,
	iam *iam.Service,
	esignSvc *esign.Service,
	html2pdfConverter *html2pdf.Converter,
	fileManagerService *filemanager.Service,
	logger *log.Logger,
	slack *slack.Service,
	resourceAliasSvc *resourcealias.Service,
) *Service {
	svc := &Service{
		pg:                 pgClient,
		s3:                 s3Client,
		bucket:             bucket,
		slackSigningSecret: slackSigningSecret,
		baseURL:            baseURL,
		iam:                iam,
		esign:              esignSvc,
		html2pdfConverter:  html2pdfConverter,
		fileManager:        fileManagerService,
		logger:             logger,
		slack:              slack,
		resourceAlias:      resourceAliasSvc,
	}
	svc.TrustCenters = &TrustCenterService{svc: svc}
	svc.Documents = &DocumentService{svc: svc, html2pdfConverter: html2pdfConverter}
	svc.Audits = &AuditService{svc: svc}
	svc.ThirdParties = &ThirdPartyService{svc: svc}
	svc.Frameworks = &FrameworkService{svc: svc}
	svc.ComplianceFrameworks = &ComplianceFrameworkService{svc: svc}
	svc.TrustCenterAccesses = &TrustCenterAccessService{svc: svc, iamSvc: iam, logger: logger}
	svc.TrustCenterReferences = &TrustCenterReferenceService{svc: svc}
	svc.CompliancePortalCommitmentGroups = &CompliancePortalCommitmentGroupService{svc: svc}
	svc.CompliancePortalCommitments = &CompliancePortalCommitmentService{svc: svc}
	svc.TrustCenterFiles = &TrustCenterFileService{svc: svc}
	svc.Reports = &ReportService{svc: svc}
	svc.Organizations = &OrganizationService{svc: svc}
	svc.ComplianceExternalURLs = &ComplianceExternalURLService{svc: svc}

	return svc
}

func (s *Service) Get(
	ctx context.Context,
	id gid.GID,
) (*coredata.TrustCenter, error) {
	trustCenter := &coredata.TrustCenter{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := trustCenter.LoadByID(ctx, conn, coredata.NewNoScope(), id)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrPageNotFound
				}

				return fmt.Errorf("cannot load trust center: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return trustCenter, nil
}

func (s *Service) GetBySlug(
	ctx context.Context,
	slug string,
) (*coredata.TrustCenter, error) {
	trustCenter := &coredata.TrustCenter{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := trustCenter.LoadBySlug(ctx, conn, slug)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrPageNotFound
				}

				return fmt.Errorf("cannot load trust center: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return trustCenter, nil
}

func (s *Service) GetByDomainName(ctx context.Context, domain string) (*coredata.TrustCenter, error) {
	trustCenter := &coredata.TrustCenter{}

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

			var org coredata.Organization
			if err := org.LoadByCustomDomainID(ctx, conn, coredata.NewNoScope(), customDomain.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrPageNotFound
				}

				return fmt.Errorf("cannot load organization: %w", err)
			}

			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByOrganizationID(ctx, conn, coredata.NewNoScope(), org.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrPageNotFound
				}

				return fmt.Errorf("cannot load trust center: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return trustCenter, err
}

func (s *Service) GetCustomDomainByOrganizationID(ctx context.Context, organizationID gid.GID) (*coredata.CustomDomain, error) {
	customDomain := &coredata.CustomDomain{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return customDomain.LoadByOrganizationID(ctx, conn, coredata.NewNoScope(), organizationID)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrCustomDomainNotFound
		}

		return nil, err
	}

	return customDomain, err
}

// EmailPresenterConfigByOrganizationID resolves the emails.PresenterConfig for
// the trust center that belongs to the given organization. This is used by the
// esign certificate worker which needs per-org branding at render time.
func (s *Service) EmailPresenterConfigByOrganizationID(ctx context.Context, orgID gid.GID) (emails.PresenterConfig, error) {
	var trustCenter coredata.TrustCenter

	scope := coredata.NewScopeFromObjectID(orgID)

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return trustCenter.LoadByOrganizationID(ctx, conn, scope, orgID)
	})
	if err != nil {
		return emails.PresenterConfig{}, fmt.Errorf("cannot load trust center for org %s: %w", orgID, err)
	}

	return s.TrustCenters.EmailPresenterConfig(ctx, scope, trustCenter.ID)
}

func (s *Service) GetOrganizationByTrustCenterID(
	ctx context.Context,
	trustCenterID gid.GID,
) (*coredata.Organization, error) {
	trustCenter, err := s.Get(ctx, trustCenterID)
	if err != nil {
		return nil, fmt.Errorf("cannot load trust center: %w", err)
	}

	org := &coredata.Organization{}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return org.LoadByID(ctx, conn, coredata.NewNoScope(), trustCenter.OrganizationID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load organization: %w", err)
	}

	return org, nil
}

func (s *Service) GetMembershipByCompliancePageIDAndIdentityID(ctx context.Context, compliancePageID gid.GID, identityID gid.GID) (*coredata.TrustCenterAccess, error) {
	membership := &coredata.TrustCenterAccess{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return membership.LoadByTrustCenterIDAndIdentityID(
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

func (s *Service) GetNDAFile(
	ctx context.Context,
	compliancePageID gid.GID,
) (*coredata.File, error) {
	var (
		file  *coredata.File
		scope = coredata.NewScopeFromObjectID(compliancePageID)
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID == nil {
				return ErrNDAFileNotFound
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrNDAFileNotFound
				}

				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *Service) ProvisionMember(
	ctx context.Context,
	compliancePageID gid.GID,
	identityID gid.GID,
) (*coredata.TrustCenterAccess, error) {
	var (
		access *coredata.TrustCenterAccess
		now    = time.Now()
		scope  = coredata.NewScopeFromObjectID(compliancePageID)
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			compliancePage := &coredata.TrustCenter{}
			if err := compliancePage.LoadByID(ctx, tx, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			identity := &coredata.Identity{}
			if err := identity.LoadByID(ctx, tx, identityID); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			access = &coredata.TrustCenterAccess{}
			if err := access.LoadByTrustCenterIDAndIdentityID(ctx, tx, scope, compliancePageID, identityID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load trust center access: %w", err)
				}

				access = &coredata.TrustCenterAccess{
					ID:             gid.New(scope.GetTenantID(), coredata.TrustCenterAccessEntityType),
					OrganizationID: compliancePage.OrganizationID,
					TenantID:       scope.GetTenantID(),
					IdentityID:     identityID,
					TrustCenterID:  compliancePageID,
					CreatedAt:      now,
					UpdatedAt:      now,
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
							ConsentText:    NDAConsentText,
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
					return fmt.Errorf("cannot insert trust center access: %w", err)
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
