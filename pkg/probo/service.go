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
	"github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/certmanager"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/filemanager"
	"github.com/getprobo/probo/pkg/filevalidation"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/html2pdf"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
)

type ExportService interface {
	BuildAndUploadExport(ctx context.Context, exportJobID gid.GID) (*coredata.ExportJob, error)
	SendExportEmail(ctx context.Context, fileID gid.GID, recipientName, recipientEmail string) error
}

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
		baseURL           string
		tokenSecret       string
		trustConfig       TrustConfig
		agentConfig       agents.Config
		html2pdfConverter *html2pdf.Converter
		acmeService       *certmanager.ACMEService
		fileManager       *filemanager.Service
		auth              *auth.Service
		authz             *authz.Service
		logger            *log.Logger
	}

	TenantService struct {
		pg                                *pg.Client
		s3                                *s3.Client
		bucket                            string
		encryptionKey                     cipher.EncryptionKey
		scope                             coredata.Scoper
		baseURL                           string
		tokenSecret                       string
		trustConfig                       TrustConfig
		agent                             *agents.Agent
		fileManager                       *filemanager.Service
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
		TrustCenterReferences             *TrustCenterReferenceService
		TrustCenterFiles                  *TrustCenterFileService
		Nonconformities                   *NonconformityService
		Obligations                       *ObligationService
		Snapshots                         *SnapshotService
		ContinualImprovements             *ContinualImprovementService
		ProcessingActivities              *ProcessingActivityService
		Files                             *FileService
		CustomDomains                     *CustomDomainService
	}
)

func NewService(
	ctx context.Context,
	encryptionKey cipher.EncryptionKey,
	pgClient *pg.Client,
	s3Client *s3.Client,
	bucket string,
	baseURL string,
	tokenSecret string,
	trustConfig TrustConfig,
	agentConfig agents.Config,
	html2pdfConverter *html2pdf.Converter,
	acmeService *certmanager.ACMEService,
	fileManagerService *filemanager.Service,
	authService *auth.Service,
	authzService *authz.Service,
	logger *log.Logger,
) (*Service, error) {
	if bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	svc := &Service{
		pg:                pgClient,
		s3:                s3Client,
		bucket:            bucket,
		encryptionKey:     encryptionKey,
		baseURL:           baseURL,
		tokenSecret:       tokenSecret,
		trustConfig:       trustConfig,
		agentConfig:       agentConfig,
		html2pdfConverter: html2pdfConverter,
		acmeService:       acmeService,
		fileManager:       fileManagerService,
		auth:              authService,
		authz:             authzService,
		logger:            logger,
	}

	return svc, nil
}

func (s *Service) WithTenant(tenantID gid.TenantID) *TenantService {
	tenantService := &TenantService{
		pg:            s.pg,
		s3:            s.s3,
		bucket:        s.bucket,
		encryptionKey: s.encryptionKey,
		baseURL:       s.baseURL,
		scope:         coredata.NewScope(tenantID),
		tokenSecret:   s.tokenSecret,
		trustConfig:   s.trustConfig,
		agent:         agents.NewAgent(nil, s.agentConfig),
		fileManager:   s.fileManager,
	}

	tenantService.Frameworks = &FrameworkService{
		svc:               tenantService,
		html2pdfConverter: s.html2pdfConverter,
	}
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
	tenantService.VendorComplianceReports = &VendorComplianceReportService{svc: tenantService, fileValidator: filevalidation.NewValidator(filevalidation.CategoryDocument)}
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
	tenantService.TrustCenterAccesses = &TrustCenterAccessService{svc: tenantService}
	tenantService.TrustCenterReferences = &TrustCenterReferenceService{svc: tenantService}
	tenantService.TrustCenterFiles = &TrustCenterFileService{
		svc: tenantService,
		fileValidator: &filevalidation.FileValidator{
			MaxFileSize: 10 * 1024 * 1024, // 10MB
			AllowedMimeTypes: map[string]bool{
				"application/pdf": true,
			},
			AllowedExtensions: map[string][]string{
				".pdf": {"application/pdf"},
			},
		},
	}
	tenantService.Nonconformities = &NonconformityService{svc: tenantService}
	tenantService.Obligations = &ObligationService{svc: tenantService}
	tenantService.Snapshots = &SnapshotService{svc: tenantService}
	tenantService.ContinualImprovements = &ContinualImprovementService{svc: tenantService}
	tenantService.ProcessingActivities = &ProcessingActivityService{svc: tenantService}
	tenantService.Files = &FileService{svc: tenantService}
	tenantService.CustomDomains = &CustomDomainService{
		svc:           tenantService,
		encryptionKey: s.encryptionKey,
		acmeService:   s.acmeService,
		logger:        s.logger.Named("custom_domains"),
	}

	return tenantService
}

func (s *Service) ExportJob(ctx context.Context) error {
	exportJob, err := s.lockExportJob(ctx)
	if err != nil {
		return fmt.Errorf("cannot lock export job: %w", err)
	}

	tenantService := s.WithTenant(exportJob.ID.TenantID())

	var exportService ExportService

	switch exportJob.Type {
	case coredata.ExportJobTypeFramework:
		exportService = tenantService.Frameworks
	case coredata.ExportJobTypeDocument:
		exportService = tenantService.Documents
	default:
		unknownTypeErr := fmt.Errorf("unknown export job type: %q", exportJob.Type)
		if err := s.commitFailedExport(ctx, exportJob, unknownTypeErr); err != nil {
			return fmt.Errorf("unknown export job type %q, and cannot commit failed export: %w", exportJob.Type, err)
		}
		return unknownTypeErr
	}

	exportJob, buildErr := exportService.BuildAndUploadExport(ctx, exportJob.ID)
	if buildErr != nil {
		if err := s.commitFailedExport(ctx, exportJob, buildErr); err != nil {
			return fmt.Errorf(
				"cannot build and upload %s export: %w, and cannot commit failed export: %w",
				exportJob.Type,
				buildErr,
				err,
			)
		}
		return fmt.Errorf("cannot build and upload %s export: %w", exportJob.Type, buildErr)
	}

	if emailErr := exportService.SendExportEmail(ctx, *exportJob.FileID, exportJob.RecipientName, exportJob.RecipientEmail); emailErr != nil {
		if err := s.commitFailedExport(ctx, exportJob, emailErr); err != nil {
			return fmt.Errorf(
				"cannot send completion email: %w, and cannot commit failed export: %w",
				emailErr,
				err,
			)
		}
		return fmt.Errorf("cannot send completion email: %w", emailErr)
	}

	if err := s.commitSuccessfulExport(ctx, exportJob); err != nil {
		return fmt.Errorf("cannot commit successful %s export: %w", exportJob.Type, err)
	}

	return nil
}

func (s *Service) lockExportJob(ctx context.Context) (*coredata.ExportJob, error) {
	exportJob := &coredata.ExportJob{}
	var scope coredata.Scoper

	err := s.pg.WithTx(ctx,
		func(tx pg.Conn) error {
			if err := exportJob.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return fmt.Errorf("cannot load next pending export job: %w", err)
			}

			scope = coredata.NewScope(exportJob.ID.TenantID())

			exportJob.Status = coredata.ExportJobStatusProcessing
			exportJob.StartedAt = ref.Ref(time.Now())
			if err := exportJob.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update %s export job: %w", exportJob.Type, err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot lock export job: %w", err)
	}

	return exportJob, nil
}

func (s *Service) commitFailedExport(ctx context.Context, exportJob *coredata.ExportJob, failureErr error) error {
	exportJob.CompletedAt = ref.Ref(time.Now())
	exportJob.Status = coredata.ExportJobStatusFailed
	errorMsg := failureErr.Error()
	exportJob.Error = &errorMsg

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			scope := coredata.NewScope(exportJob.ID.TenantID())
			if err := exportJob.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update %s export job: %w", exportJob.Type, err)
			}

			return nil
		},
	)
}

func (s *Service) commitSuccessfulExport(ctx context.Context, exportJob *coredata.ExportJob) error {
	exportJob.CompletedAt = ref.Ref(time.Now())
	exportJob.Status = coredata.ExportJobStatusCompleted

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			scope := coredata.NewScope(exportJob.ID.TenantID())
			if err := exportJob.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update %s export job: %w", exportJob.Type, err)
			}

			return nil
		},
	)
}

func (s *Service) LoadOrganizationByDomain(ctx context.Context, domain string) (gid.GID, error) {
	var organizationID gid.GID

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var customDomain coredata.CustomDomain
			if err := customDomain.LoadByDomain(ctx, conn, coredata.NewNoScope(), s.encryptionKey, domain); err != nil {
				return fmt.Errorf("cannot load custom domain: %w", err)
			}

			var org coredata.Organization
			if err := org.LoadByCustomDomainID(ctx, conn, coredata.NewNoScope(), customDomain.ID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			organizationID = org.ID

			return nil
		},
	)

	return organizationID, err
}

type TrustCenterInfo struct {
	ID             gid.GID
	OrganizationID gid.GID
}

func (s *Service) LoadTrustCenterBySlug(ctx context.Context, slug string) (*TrustCenterInfo, error) {
	var info TrustCenterInfo

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var trustCenter coredata.TrustCenter
			if err := trustCenter.LoadBySlug(ctx, conn, slug); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			info.ID = trustCenter.ID
			info.OrganizationID = trustCenter.OrganizationID

			return nil
		},
	)

	return &info, err
}

func (s *Service) LoadTrustCenterByID(ctx context.Context, id gid.GID) (*TrustCenterInfo, error) {
	var info TrustCenterInfo

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			scope := coredata.NewScope(id.TenantID())
			var trustCenter coredata.TrustCenter
			if err := trustCenter.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			info.ID = trustCenter.ID
			info.OrganizationID = trustCenter.OrganizationID

			return nil
		},
	)

	return &info, err
}
