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

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/slack"
)

type (
	Service struct {
		pg                 *pg.Client
		s3                 *s3.Client
		bucket             string
		proboSvc           *probo.Service
		encryptionKey      cipher.EncryptionKey
		slackSigningSecret string
		baseURL            string
		iam                *iam.Service
		html2pdfConverter  *html2pdf.Converter
		fileManager        *filemanager.Service
		logger             *log.Logger
		slack              *slack.Service
	}

	TenantService struct {
		pg                    *pg.Client
		s3                    *s3.Client
		bucket                string
		scope                 coredata.Scoper
		proboSvc              *probo.Service
		encryptionKey         cipher.EncryptionKey
		baseURL               string
		iam                   *iam.Service
		html2pdfConverter     *html2pdf.Converter
		fileManager           *filemanager.Service
		logger                *log.Logger
		TrustCenters          *TrustCenterService
		Documents             *DocumentService
		Audits                *AuditService
		Vendors               *VendorService
		Frameworks            *FrameworkService
		TrustCenterAccesses   *TrustCenterAccessService
		TrustCenterReferences *TrustCenterReferenceService
		TrustCenterFiles      *TrustCenterFileService
		Reports               *ReportService
		Organizations         *OrganizationService
		SlackMessages         *slack.SlackMessageService
	}
)

func NewService(
	pgClient *pg.Client,
	s3Client *s3.Client,
	bucket string,
	baseURL string,
	encryptionKey cipher.EncryptionKey,
	slackSigningSecret string,
	iam *iam.Service,
	html2pdfConverter *html2pdf.Converter,
	fileManagerService *filemanager.Service,
	logger *log.Logger,
	slack *slack.Service,
) *Service {
	return &Service{
		pg:                 pgClient,
		s3:                 s3Client,
		bucket:             bucket,
		encryptionKey:      encryptionKey,
		slackSigningSecret: slackSigningSecret,
		baseURL:            baseURL,
		iam:                iam,
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
		encryptionKey:     s.encryptionKey,
		baseURL:           s.baseURL,
		iam:               s.iam,
		html2pdfConverter: s.html2pdfConverter,
		fileManager:       s.fileManager,
		logger:            s.logger,
	}

	tenantService.TrustCenters = &TrustCenterService{svc: tenantService}
	tenantService.Documents = &DocumentService{svc: tenantService, html2pdfConverter: s.html2pdfConverter}
	tenantService.Audits = &AuditService{svc: tenantService}
	tenantService.Vendors = &VendorService{svc: tenantService}
	tenantService.Frameworks = &FrameworkService{svc: tenantService}
	tenantService.TrustCenterAccesses = &TrustCenterAccessService{svc: tenantService, iamSvc: s.iam, logger: s.logger}
	tenantService.TrustCenterReferences = &TrustCenterReferenceService{svc: tenantService}
	tenantService.TrustCenterFiles = &TrustCenterFileService{svc: tenantService}
	tenantService.Reports = &ReportService{svc: tenantService}
	tenantService.Organizations = &OrganizationService{svc: tenantService}
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
			if err := customDomain.LoadByDomain(ctx, conn, coredata.NewNoScope(), s.encryptionKey, domain); err != nil {
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
			return customDomain.LoadByOrganizationID(ctx, conn, coredata.NewNoScope(), s.encryptionKey, organizationID)
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

func (s *Service) GetMembershipByCompliancePageIDAndEmail(ctx context.Context, compliancePageID gid.GID, email mail.Addr) (*coredata.TrustCenterAccess, error) {
	membership := &coredata.TrustCenterAccess{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return membership.LoadByTrustCenterIDAndEmail(
				ctx,
				conn,
				coredata.NewScopeFromObjectID(compliancePageID),
				compliancePageID,
				email,
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
