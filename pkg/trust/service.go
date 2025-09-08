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
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/filemanager"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/html2pdf"
	"github.com/getprobo/probo/pkg/probo"
	"go.gearno.de/kit/pg"
)

type (
	Service struct {
		pg                *pg.Client
		s3                *s3.Client
		bucket            string
		proboSvc          *probo.Service
		encryptionKey     cipher.EncryptionKey
		tokenSecret       string
		auth              *auth.Service
		html2pdfConverter *html2pdf.Converter
		fileManager       *filemanager.Service
	}

	TenantService struct {
		pg                    *pg.Client
		s3                    *s3.Client
		bucket                string
		scope                 coredata.Scoper
		proboSvc              *probo.Service
		encryptionKey         cipher.EncryptionKey
		tokenSecret           string
		auth                  *auth.Service
		html2pdfConverter     *html2pdf.Converter
		fileManager           *filemanager.Service
		TrustCenters          *TrustCenterService
		Documents             *DocumentService
		Audits                *AuditService
		Vendors               *VendorService
		Frameworks            *FrameworkService
		TrustCenterAccesses   *TrustCenterAccessService
		TrustCenterReferences *TrustCenterReferenceService
		Reports               *ReportService
		Organizations         *OrganizationService
	}
)

func NewService(
	pgClient *pg.Client,
	s3Client *s3.Client,
	bucket string,
	encryptionKey cipher.EncryptionKey,
	tokenSecret string,
	auth *auth.Service,
	html2pdfConverter *html2pdf.Converter,
	fileManagerService *filemanager.Service,
) *Service {
	return &Service{
		pg:                pgClient,
		s3:                s3Client,
		bucket:            bucket,
		encryptionKey:     encryptionKey,
		tokenSecret:       tokenSecret,
		auth:              auth,
		html2pdfConverter: html2pdfConverter,
		fileManager:       fileManagerService,
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
		tokenSecret:       s.tokenSecret,
		auth:              s.auth,
		html2pdfConverter: s.html2pdfConverter,
		fileManager:       s.fileManager,
	}

	tenantService.TrustCenters = &TrustCenterService{svc: tenantService}
	tenantService.Documents = &DocumentService{svc: tenantService, html2pdfConverter: s.html2pdfConverter}
	tenantService.Audits = &AuditService{svc: tenantService}
	tenantService.Vendors = &VendorService{svc: tenantService}
	tenantService.Frameworks = &FrameworkService{svc: tenantService}
	tenantService.TrustCenterAccesses = &TrustCenterAccessService{svc: tenantService, auth: s.auth}
	tenantService.TrustCenterReferences = &TrustCenterReferenceService{svc: tenantService}
	tenantService.Reports = &ReportService{svc: tenantService}
	tenantService.Organizations = &OrganizationService{svc: tenantService}

	return tenantService
}

func (s *Service) GetTokenSecret() string {
	return s.tokenSecret
}
