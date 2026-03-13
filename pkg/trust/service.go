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
	"go.probo.inc/probo/pkg/slack"
)

type (
	Service struct {
		pg                 *pg.Client
		s3                 *s3.Client
		bucket             string
		proboSvc           *probo.Service
		slackSigningSecret string
		baseURL            string
		iam                *iam.Service
		esign              *esign.Service
		html2pdfConverter  *html2pdf.Converter
		fileManager        *filemanager.Service
		logger             *log.Logger
		slack              *slack.Service
	}

	TenantService struct {
		pg                     *pg.Client
		s3                     *s3.Client
		bucket                 string
		scope                  coredata.Scoper
		proboSvc               *probo.Service
		baseURL                string
		iam                    *iam.Service
		esign                  *esign.Service
		html2pdfConverter      *html2pdf.Converter
		fileManager            *filemanager.Service
		logger                 *log.Logger
		TrustCenters           *TrustCenterService
		Documents              *DocumentService
		Audits                 *AuditService
		Vendors                *VendorService
		Frameworks             *FrameworkService
		ComplianceFrameworks   *ComplianceFrameworkService
		TrustCenterAccesses    *TrustCenterAccessService
		TrustCenterReferences  *TrustCenterReferenceService
		TrustCenterFiles       *TrustCenterFileService
		Reports                *ReportService
		Organizations          *OrganizationService
		ComplianceExternalURLs *ComplianceExternalURLService
		SlackMessages          *slack.SlackMessageService
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
) *Service {
	return &Service{
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
	}
}

func (s *Service) WithTenant(tenantID gid.TenantID) *TenantService {
	tenantService := &TenantService{
		pg:                s.pg,
		s3:                s.s3,
		bucket:            s.bucket,
		scope:             coredata.NewScope(tenantID),
		proboSvc:          s.proboSvc,
		baseURL:           s.baseURL,
		iam:               s.iam,
		esign:             s.esign,
		html2pdfConverter: s.html2pdfConverter,
		fileManager:       s.fileManager,
		logger:            s.logger,
	}

	tenantService.TrustCenters = &TrustCenterService{svc: tenantService}
	tenantService.Documents = &DocumentService{svc: tenantService, html2pdfConverter: s.html2pdfConverter}
	tenantService.Audits = &AuditService{svc: tenantService}
	tenantService.Vendors = &VendorService{svc: tenantService}
	tenantService.Frameworks = &FrameworkService{svc: tenantService}
	tenantService.ComplianceFrameworks = &ComplianceFrameworkService{svc: tenantService}
	tenantService.TrustCenterAccesses = &TrustCenterAccessService{svc: tenantService, iamSvc: s.iam, logger: s.logger}
	tenantService.TrustCenterReferences = &TrustCenterReferenceService{svc: tenantService}
	tenantService.TrustCenterFiles = &TrustCenterFileService{svc: tenantService}
	tenantService.Reports = &ReportService{svc: tenantService}
	tenantService.Organizations = &OrganizationService{svc: tenantService}
	tenantService.ComplianceExternalURLs = &ComplianceExternalURLService{svc: tenantService}
	tenantService.SlackMessages = s.slack.WithTenant(tenantID).SlackMessages

	return tenantService
}

func (s *Service) Get(
	ctx context.Context,
	id gid.GID,
) (*coredata.TrustCenter, error) {
	trustCenter := &coredata.TrustCenter{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
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
		func(conn pg.Conn) error {
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
		func(conn pg.Conn) error {
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
		func(conn pg.Conn) error {
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
	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return trustCenter.LoadByOrganizationID(ctx, conn, scope, orgID)
	})
	if err != nil {
		return emails.PresenterConfig{}, fmt.Errorf("cannot load trust center for org %s: %w", orgID, err)
	}
	return s.WithTenant(orgID.TenantID()).TrustCenters.EmailPresenterConfig(ctx, trustCenter.ID)
}

func (s *Service) GetMembershipByCompliancePageIDAndIdentityID(ctx context.Context, compliancePageID gid.GID, identityID gid.GID) (*coredata.TrustCenterAccess, error) {
	membership := &coredata.TrustCenterAccess{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
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
		func(conn pg.Conn) error {
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
		func(tx pg.Conn) error {
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
