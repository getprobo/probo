// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/file"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/slack"
)

const (
	NameMaxLength    = 100
	TitleMaxLength   = 1000
	ContentMaxLength = 5000
)

type ExportService interface {
	BuildAndUploadExport(
		ctx context.Context,
		scope coredata.Scoper,
		exportJobID gid.GID,
	) (*coredata.ExportJob, error)
	SendExportEmail(
		ctx context.Context,
		scope coredata.Scoper,
		fileID gid.GID,
		recipientName string,
		recipientEmail mail.Addr,
	) error
}

type (
	// LLMConfig holds the model parameters used by the agents the probo
	// service runs.
	LLMConfig struct {
		Model       string
		Temperature float64
		MaxTokens   int
	}

	Service struct {
		pg                                    *pg.Client
		s3                                    *s3.Client
		bucket                                string
		encryptionKey                         cipher.EncryptionKey
		baseURL                               string
		tokenSecret                           string
		llmClient                             *llm.Client
		llmConfig                             LLMConfig
		html2pdfConverter                     *html2pdf.Converter
		acmeService                           *certmanager.ACMEService
		fileManager                           *filemanager.Service
		file                                  *file.Service
		logger                                *log.Logger
		slack                                 *slack.Service
		esign                                 *esign.Service
		connectorRegistry                     *connector.ConnectorRegistry
		invitationTokenValidity               time.Duration
		Frameworks                            *FrameworkService
		Measures                              *MeasureService
		Tasks                                 *TaskService
		Evidences                             *EvidenceService
		Organizations                         *OrganizationService
		ThirdParties                          *ThirdPartyService
		Documents                             *DocumentService
		DocumentApprovals                     *DocumentApprovalService
		Controls                              *ControlService
		Risks                                 *RiskService
		ThirdPartyComplianceReports           *ThirdPartyComplianceReportService
		ThirdPartyBusinessAssociateAgreements *ThirdPartyBusinessAssociateAgreementService
		ThirdPartyContacts                    *ThirdPartyContactService
		ThirdPartyDataPrivacyAgreements       *ThirdPartyDataPrivacyAgreementService
		ThirdPartyServices                    *ThirdPartyServiceService
		Connectors                            *ConnectorService
		Assets                                *AssetService
		Data                                  *DatumService
		Audits                                *AuditService
		WebhookSubscriptions                  *WebhookSubscriptionService
		TrustCenters                          *TrustCenterService
		TrustCenterAccesses                   *TrustCenterAccessService
		TrustCenterReferences                 *TrustCenterReferenceService
		TrustCenterFiles                      *TrustCenterFileService
		ComplianceFrameworks                  *ComplianceFrameworkService
		ComplianceExternalURLs                *ComplianceExternalURLService
		Findings                              *FindingService
		Obligations                           *ObligationService
		RightsRequests                        *RightsRequestService
		ProcessingActivities                  *ProcessingActivityService
		DataProtectionImpactAssessments       *DataProtectionImpactAssessmentService
		TransferImpactAssessments             *TransferImpactAssessmentService
		StatementsOfApplicability             *StatementOfApplicabilityService
		GeneratedDocuments                    *GeneratedDocumentService
		Files                                 *FileService
		CustomDomains                         *CustomDomainService
		SlackMessages                         *slack.Service
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
	llmClient *llm.Client,
	llmConfig LLMConfig,
	html2pdfConverter *html2pdf.Converter,
	acmeService *certmanager.ACMEService,
	fileManagerService *filemanager.Service,
	logger *log.Logger,
	slackService *slack.Service,
	iamService *iam.Service,
	esignService *esign.Service,
	connectorRegistry *connector.ConnectorRegistry,
	invitationTokenValidity time.Duration,
	fileService *file.Service,
) (*Service, error) {
	if bucket == "" {
		return nil, fmt.Errorf("bucket is required")
	}

	iamService.Authorizer.RegisterPolicySet(ProboPolicySet())

	svc := &Service{
		pg:                      pgClient,
		s3:                      s3Client,
		bucket:                  bucket,
		encryptionKey:           encryptionKey,
		baseURL:                 baseURL,
		tokenSecret:             tokenSecret,
		llmClient:               llmClient,
		llmConfig:               llmConfig,
		html2pdfConverter:       html2pdfConverter,
		acmeService:             acmeService,
		fileManager:             fileManagerService,
		file:                    fileService,
		logger:                  logger,
		slack:                   slackService,
		esign:                   esignService,
		connectorRegistry:       connectorRegistry,
		invitationTokenValidity: invitationTokenValidity,
	}

	svc.Frameworks = &FrameworkService{
		svc:               svc,
		html2pdfConverter: html2pdfConverter,
	}
	svc.Measures = &MeasureService{svc: svc}
	svc.Tasks = &TaskService{svc: svc}
	svc.Evidences = &EvidenceService{
		svc: svc,
		fileValidator: filevalidation.NewValidator(
			filevalidation.WithCategories(
				filevalidation.CategoryDocument,
				filevalidation.CategorySpreadsheet,
				filevalidation.CategoryPresentation,
				filevalidation.CategoryData,
				filevalidation.CategoryText,
				filevalidation.CategoryImage,
				filevalidation.CategoryVideo,
			),
		),
	}
	svc.ThirdParties = &ThirdPartyService{svc: svc}
	svc.Documents = &DocumentService{
		svc:                     svc,
		html2pdfConverter:       html2pdfConverter,
		invitationTokenValidity: invitationTokenValidity,
		tokenSecret:             tokenSecret,
	}
	svc.DocumentApprovals = &DocumentApprovalService{
		svc:                     svc,
		html2pdfConverter:       html2pdfConverter,
		invitationTokenValidity: invitationTokenValidity,
		tokenSecret:             tokenSecret,
	}
	svc.Organizations = &OrganizationService{
		svc: svc,
		fileValidator: filevalidation.NewValidator(
			filevalidation.WithCategories(filevalidation.CategoryImage),
		),
	}
	svc.Controls = &ControlService{svc: svc}
	svc.Risks = &RiskService{svc: svc}
	svc.ThirdPartyComplianceReports = &ThirdPartyComplianceReportService{
		svc: svc,
		fileValidator: filevalidation.NewValidator(
			filevalidation.WithCategories(filevalidation.CategoryDocument),
		),
	}
	svc.ThirdPartyBusinessAssociateAgreements = &ThirdPartyBusinessAssociateAgreementService{svc: svc}
	svc.ThirdPartyContacts = &ThirdPartyContactService{svc: svc}
	svc.ThirdPartyDataPrivacyAgreements = &ThirdPartyDataPrivacyAgreementService{svc: svc}
	svc.ThirdPartyServices = &ThirdPartyServiceService{svc: svc}
	svc.Connectors = &ConnectorService{svc: svc}
	svc.Assets = &AssetService{svc: svc}
	svc.Data = &DatumService{svc: svc}
	svc.Audits = &AuditService{svc: svc}
	svc.WebhookSubscriptions = &WebhookSubscriptionService{svc: svc}
	svc.TrustCenters = &TrustCenterService{svc: svc}
	svc.TrustCenterAccesses = &TrustCenterAccessService{svc: svc}
	svc.TrustCenterReferences = &TrustCenterReferenceService{svc: svc}
	svc.ComplianceFrameworks = &ComplianceFrameworkService{svc: svc}
	svc.ComplianceExternalURLs = &ComplianceExternalURLService{svc: svc}
	svc.TrustCenterFiles = &TrustCenterFileService{
		svc: svc,
		fileValidator: filevalidation.NewValidator(
			filevalidation.WithCategories(
				filevalidation.CategoryData,
				filevalidation.CategoryDocument,
				filevalidation.CategoryImage,
				filevalidation.CategoryPresentation,
				filevalidation.CategorySpreadsheet,
				filevalidation.CategoryText,
			),
			filevalidation.WithMaxFileSize(10*1024*1024), // 10MB
		),
	}
	svc.Findings = &FindingService{svc: svc}
	svc.Obligations = &ObligationService{svc: svc}
	svc.RightsRequests = &RightsRequestService{svc: svc}
	svc.ProcessingActivities = &ProcessingActivityService{svc: svc}
	svc.DataProtectionImpactAssessments = &DataProtectionImpactAssessmentService{svc: svc}
	svc.TransferImpactAssessments = &TransferImpactAssessmentService{svc: svc}
	svc.StatementsOfApplicability = &StatementOfApplicabilityService{svc: svc}
	svc.GeneratedDocuments = &GeneratedDocumentService{svc: svc}
	svc.Files = &FileService{svc: svc}
	svc.CustomDomains = &CustomDomainService{
		svc:           svc,
		encryptionKey: encryptionKey,
		acmeService:   acmeService,
		logger:        logger.Named("custom_domains"),
	}
	svc.SlackMessages = slackService

	return svc, nil
}

func (s *Service) ExportJob(ctx context.Context) error {
	exportJob, err := s.lockExportJob(ctx)
	if err != nil {
		return fmt.Errorf("cannot lock export job: %w", err)
	}

	scope := coredata.NewScope(exportJob.ID.TenantID())

	var exportService ExportService

	switch exportJob.Type {
	case coredata.ExportJobTypeFramework:
		exportService = s.Frameworks
	case coredata.ExportJobTypeDocument:
		exportService = s.Documents
	default:
		unknownTypeErr := fmt.Errorf("unknown export job type: %q", exportJob.Type)
		if err := s.commitFailedExport(ctx, exportJob, unknownTypeErr); err != nil {
			return fmt.Errorf("unknown export job type %q, and cannot commit failed export: %w", exportJob.Type, err)
		}

		return unknownTypeErr
	}

	updatedExportJob, buildErr := exportService.BuildAndUploadExport(ctx, scope, exportJob.ID)
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

	exportJob = updatedExportJob

	if emailErr := exportService.SendExportEmail(
		ctx,
		scope,
		*exportJob.FileID,
		exportJob.RecipientName,
		exportJob.RecipientEmail,
	); emailErr != nil {
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

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := exportJob.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return fmt.Errorf("cannot load next pending export job: %w", err)
			}

			scope = coredata.NewScope(exportJob.ID.TenantID())

			exportJob.Status = coredata.ExportJobStatusProcessing

			exportJob.StartedAt = new(time.Now())
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
	exportJob.CompletedAt = new(time.Now())
	exportJob.Status = coredata.ExportJobStatusFailed
	errorMsg := failureErr.Error()
	exportJob.Error = &errorMsg

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScope(exportJob.ID.TenantID())
			if err := exportJob.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update %s export job: %w", exportJob.Type, err)
			}

			return nil
		},
	)
}

func (s *Service) commitSuccessfulExport(ctx context.Context, exportJob *coredata.ExportJob) error {
	exportJob.CompletedAt = new(time.Now())
	exportJob.Status = coredata.ExportJobStatusCompleted

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScope(exportJob.ID.TenantID())
			if err := exportJob.Update(ctx, tx, scope); err != nil {
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
		func(ctx context.Context, conn pg.Querier) error {
			var customDomain coredata.CustomDomain
			if err := customDomain.LoadByDomain(ctx, conn, coredata.NewNoScope(), domain); err != nil {
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
