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
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/agents"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/filevalidation"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/html2pdf"
	"github.com/getprobo/probo/pkg/usrmgr"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
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
		NonconformityRegistries           *NonconformityRegistryService
		ComplianceRegistries              *ComplianceRegistryService
		Snapshots                         *SnapshotService
		ContinualImprovementRegistries    *ContinualImprovementRegistriesService
		ProcessingActivityRegistries      *ProcessingActivityRegistryService
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
		usrmgr:            usrmgrService,
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
	tenantService.NonconformityRegistries = &NonconformityRegistryService{svc: tenantService}
	tenantService.ComplianceRegistries = &ComplianceRegistryService{svc: tenantService}
	tenantService.Snapshots = &SnapshotService{svc: tenantService}
	tenantService.ContinualImprovementRegistries = &ContinualImprovementRegistriesService{svc: tenantService}
	tenantService.ProcessingActivityRegistries = &ProcessingActivityRegistryService{svc: tenantService}
	return tenantService
}

func (s *Service) ExportFrameworkJob(ctx context.Context) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			fe := &coredata.FrameworkExport{}
			if err := fe.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			scope := coredata.NewScope(fe.ID.TenantID())

			fe.Status = coredata.FrameworkExportStatusProcessing
			fe.StartedAt = ref.Ref(time.Now())
			if err := fe.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update framework export: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, tx, scope, fe.FrameworkID); err != nil {
				fe.Status = coredata.FrameworkExportStatusFailed
				fe.CompletedAt = ref.Ref(time.Now())
				if err := fe.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update framework export: %w", err)
				}

				return fmt.Errorf("cannot load framework: %w", err)
			}

			tenantService := s.WithTenant(fe.ID.TenantID())

			tempDir := os.TempDir()
			tempFile, err := os.CreateTemp(tempDir, "probo-framework-export-*.zip")
			if err != nil {
				fe.Status = coredata.FrameworkExportStatusFailed
				fe.CompletedAt = ref.Ref(time.Now())
				if err := fe.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update framework export: %w", err)
				}

				return fmt.Errorf("cannot create temp file: %w", err)
			}
			defer tempFile.Close()
			defer os.Remove(tempFile.Name())

			err = tenantService.Frameworks.Export(ctx, fe.FrameworkID, tempFile)
			if err != nil {
				fe.Status = coredata.FrameworkExportStatusFailed
				fe.CompletedAt = ref.Ref(time.Now())
				if err := fe.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update framework export: %w", err)
				}

				return fmt.Errorf("cannot export framework: %w", err)
			}

			uuid, err := uuid.NewV4()
			if err != nil {
				fe.Status = coredata.FrameworkExportStatusFailed
				fe.CompletedAt = ref.Ref(time.Now())
				if err := fe.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update framework export: %w", err)
				}

				return fmt.Errorf("cannot update framework export: %w", err)
			}

			if _, err := tempFile.Seek(0, 0); err != nil {
				fe.Status = coredata.FrameworkExportStatusFailed
				fe.CompletedAt = ref.Ref(time.Now())
				if err := fe.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update framework export: %w", err)
				}

				return fmt.Errorf("cannot seek temp file: %w", err)
			}

			fileInfo, err := tempFile.Stat()
			if err != nil {
				fe.Status = coredata.FrameworkExportStatusFailed
				fe.CompletedAt = ref.Ref(time.Now())
				if err := fe.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update framework export: %w", err)
				}

				return fmt.Errorf("cannot stat temp file: %w", err)
			}

			_, err = s.s3.PutObject(
				ctx,
				&s3.PutObjectInput{
					Bucket:        ref.Ref(s.bucket),
					Key:           ref.Ref(uuid.String()),
					Body:          tempFile,
					ContentLength: ref.Ref(fileInfo.Size()),
					ContentType:   ref.Ref("application/zip"),
					Metadata: map[string]string{
						"framework-id":        framework.ID.String(),
						"framework-export-id": fe.ID.String(),
					},
				},
			)
			if err != nil {
				fe.Status = coredata.FrameworkExportStatusFailed
				fe.CompletedAt = ref.Ref(time.Now())
				if err := fe.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update framework export: %w", err)
				}

				return fmt.Errorf("cannot upload file to S3: %w", err)
			}

			now := time.Now()

			file := coredata.File{
				ID:         gid.New(fe.ID.TenantID(), coredata.FileEntityType),
				BucketName: s.bucket,
				MimeType:   "application/zip",
				FileName:   fmt.Sprintf("%s Archive %s.zip", framework.Name, time.Now().Format("2006-01-02")),
				FileKey:    uuid.String(),
				FileSize:   int(fileInfo.Size()),
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if err := file.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			fe.FileID = &file.ID
			fe.CompletedAt = &now
			fe.Status = coredata.FrameworkExportStatusCompleted
			if err := fe.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update framework export: %w", err)
			}

			return nil
		},
	)
}
