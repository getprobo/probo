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
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/filevalidation"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/html2pdf"
	"github.com/getprobo/probo/pkg/auth"
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
		hostname          string
		tokenSecret       string
		trustConfig       TrustConfig
		agentConfig       agents.Config
		html2pdfConverter *html2pdf.Converter
		auth              *auth.Service
		authz             *authz.Service
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
		TrustCenterReferences             *TrustCenterReferenceService
		Nonconformities                   *NonconformityService
		Obligations                       *ObligationService
		Snapshots                         *SnapshotService
		ContinualImprovements             *ContinualImprovementService
		ProcessingActivities              *ProcessingActivityService
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
	authService *auth.Service,
	authzService *authz.Service,
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
		auth:              authService,
		authz:             authzService,
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
	tenantService.TrustCenterAccesses = &TrustCenterAccessService{svc: tenantService}
	tenantService.TrustCenterReferences = &TrustCenterReferenceService{svc: tenantService}
	tenantService.Nonconformities = &NonconformityService{svc: tenantService}
	tenantService.Obligations = &ObligationService{svc: tenantService}
	tenantService.Snapshots = &SnapshotService{svc: tenantService}
	tenantService.ContinualImprovements = &ContinualImprovementService{svc: tenantService}
	tenantService.ProcessingActivities = &ProcessingActivityService{svc: tenantService}
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
