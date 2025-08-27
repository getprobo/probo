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

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/agents"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/filevalidation"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/html2pdf"
	"github.com/getprobo/probo/pkg/usrmgr"
	"go.gearno.de/kit/pg"
)

type (
	TrustConfig struct {
		TokenSecret   string
		TokenDuration time.Duration
		TokenType     string
	}

	Service struct {
		pg                *pg.Client
		s3                *s3.Client
		bucket            string
		encryptionKey     cipher.EncryptionKey
		hostname          string
		tokenSecret       string
		trustConfig       TrustConfig
		agentConfig       agents.Config
		html2pdfConverter *html2pdf.Converter
		usrmgr            *usrmgr.Service
	}

	TenantService struct {
		pg                                *pg.Client
		s3                                *s3.Client
		bucket                            string
		encryptionKey                     cipher.EncryptionKey
		scope                             coredata.Scoper
		hostname                          string
		tokenSecret                       string
		trustConfig                       TrustConfig
		agent                             *agents.Agent
		Frameworks                        *FrameworkService
		Measures                          *MeasureService
		Tasks                             *TaskService
		Evidences                         *EvidenceService
		Organizations                     *OrganizationService
		Vendors                           *VendorService
		Peoples                           *PeopleService
		Documents                         *DocumentService
		Controls                          *ControlService
		Risks                             *RiskService
		VendorComplianceReports           *VendorComplianceReportService
		VendorBusinessAssociateAgreements *VendorBusinessAssociateAgreementService
		VendorContacts                    *VendorContactService
		VendorDataPrivacyAgreements       *VendorDataPrivacyAgreementService
		VendorServices                    *VendorServiceService
		Connectors                        *ConnectorService
		Assets                            *AssetService
		Data                              *DatumService
		Audits                            *AuditService
		Reports                           *ReportService
		TrustCenters                      *TrustCenterService
		TrustCenterAccesses               *TrustCenterAccessService
		NonconformityRegistries           *NonconformityRegistryService
		ComplianceRegistries              *ComplianceRegistryService
		Snapshots                         *SnapshotService
		ContinualImprovementRegistries    *ContinualImprovementRegistriesService
	}
)

func NewService(
	ctx context.Context,
	encryptionKey cipher.EncryptionKey,
	pgClient *pg.Client,
	s3Client *s3.Client,
	bucket string,
	hostname string,
	tokenSecret string,
	trustConfig TrustConfig,
	agentConfig agents.Config,
	html2pdfConverter *html2pdf.Converter,
	usrmgrService *usrmgr.Service,
) (*Service, error) {
	if bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	svc := &Service{
		pg:                pgClient,
		s3:                s3Client,
		bucket:            bucket,
		encryptionKey:     encryptionKey,
		hostname:          hostname,
		tokenSecret:       tokenSecret,
		trustConfig:       trustConfig,
		agentConfig:       agentConfig,
		html2pdfConverter: html2pdfConverter,
		usrmgr:            usrmgrService,
	}

	return svc, nil
}

func (s *Service) WithTenant(tenantID gid.TenantID) *TenantService {
	tenantService := &TenantService{
		pg:            s.pg,
		s3:            s.s3,
		bucket:        s.bucket,
		encryptionKey: s.encryptionKey,
		hostname:      s.hostname,
		scope:         coredata.NewScope(tenantID),
		tokenSecret:   s.tokenSecret,
		trustConfig:   s.trustConfig,
		agent:         agents.NewAgent(nil, s.agentConfig),
	}

	tenantService.Frameworks = &FrameworkService{svc: tenantService}
	tenantService.Measures = &MeasureService{svc: tenantService}
	tenantService.Tasks = &TaskService{svc: tenantService}
	tenantService.Evidences = &EvidenceService{
		svc: tenantService,
		fileValidator: filevalidation.NewValidator(
			filevalidation.CategoryDocument,
			filevalidation.CategorySpreadsheet,
			filevalidation.CategoryPresentation,
			filevalidation.CategoryData,
			filevalidation.CategoryText,
			filevalidation.CategoryImage,
			filevalidation.CategoryVideo,
		),
	}
	tenantService.Peoples = &PeopleService{svc: tenantService}
	tenantService.Vendors = &VendorService{svc: tenantService}
	tenantService.Documents = &DocumentService{
		svc:               tenantService,
		html2pdfConverter: s.html2pdfConverter,
	}
	tenantService.Organizations = &OrganizationService{
		svc: tenantService,
		fileValidator: filevalidation.NewValidator(
			filevalidation.CategoryImage,
		),
	}
	tenantService.Controls = &ControlService{svc: tenantService}
	tenantService.Risks = &RiskService{svc: tenantService}
	tenantService.VendorComplianceReports = &VendorComplianceReportService{svc: tenantService}
	tenantService.VendorBusinessAssociateAgreements = &VendorBusinessAssociateAgreementService{svc: tenantService}
	tenantService.VendorContacts = &VendorContactService{svc: tenantService}
	tenantService.VendorDataPrivacyAgreements = &VendorDataPrivacyAgreementService{svc: tenantService}
	tenantService.VendorServices = &VendorServiceService{svc: tenantService}
	tenantService.Connectors = &ConnectorService{svc: tenantService}
	tenantService.Assets = &AssetService{svc: tenantService}
	tenantService.Data = &DatumService{svc: tenantService}
	tenantService.Audits = &AuditService{svc: tenantService}
	tenantService.Reports = &ReportService{svc: tenantService}
	tenantService.TrustCenters = &TrustCenterService{svc: tenantService}
	tenantService.TrustCenterAccesses = &TrustCenterAccessService{svc: tenantService, usrmgr: s.usrmgr}
	tenantService.NonconformityRegistries = &NonconformityRegistryService{svc: tenantService}
	tenantService.ComplianceRegistries = &ComplianceRegistryService{svc: tenantService}
	tenantService.Snapshots = &SnapshotService{svc: tenantService}
	tenantService.ContinualImprovementRegistries = &ContinualImprovementRegistriesService{svc: tenantService}
	return tenantService
}
