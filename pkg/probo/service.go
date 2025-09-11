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
	tenantService.TrustCenterAccesses = &TrustCenterAccessService{svc: tenantService, usrmgr: s.usrmgr}
	tenantService.Nonconformities = &NonconformityService{svc: tenantService}
	tenantService.Obligations = &ObligationService{svc: tenantService}
	tenantService.Snapshots = &SnapshotService{svc: tenantService}
	tenantService.ContinualImprovements = &ContinualImprovementService{svc: tenantService}
	tenantService.ProcessingActivities = &ProcessingActivityService{svc: tenantService}
	return tenantService
}

func (s *Service) ExportFrameworkJob(ctx context.Context) error {
	fe, scope, err := s.lockExport(ctx)
	if err != nil {
		return fmt.Errorf("cannot lock framework export: %w", err)
	}

	fe, buildErr := s.buildAndUploadExport(ctx, scope, fe)
	if buildErr != nil {
		if err := s.commitFailedExport(ctx, scope, fe); err != nil {
			return fmt.Errorf(
				"cannot build and upload framework export: %w, and cannot commit failed export: %w",
				buildErr,
				err,
			)
		}

		return fmt.Errorf("cannot build and upload framework export: %w", buildErr)
	}

	tenantService := s.WithTenant(scope.GetTenantID())
	if emailErr := tenantService.Frameworks.SendFrameworkExportEmail(ctx, *fe.FileID, fe.RecipientName, fe.RecipientEmail); emailErr != nil {
		if err := s.commitFailedExport(ctx, scope, fe); err != nil {
			return fmt.Errorf(
				"cannot send completion email: %w, and cannot commit failed export: %w",
				emailErr,
				err,
			)
		}

		return fmt.Errorf("cannot send completion email: %w", emailErr)
	}

	if err := s.commitSuccessfulExport(ctx, scope, fe); err != nil {
		return fmt.Errorf("cannot commit successful export: %w", err)
	}

	return nil
}

func (s *Service) lockExport(ctx context.Context) (*coredata.FrameworkExport, coredata.Scoper, error) {
	fe := &coredata.FrameworkExport{}
	var scope coredata.Scoper

	err := s.pg.WithTx(ctx,
		func(tx pg.Conn) error {
			if err := fe.LoadNextPendingForUpdateSkipLocked(ctx, tx); err != nil {
				return fmt.Errorf("cannot load next pending framework export: %w", err)
			}

			scope = coredata.NewScope(fe.ID.TenantID())

			fe.Status = coredata.FrameworkExportStatusProcessing
			fe.StartedAt = ref.Ref(time.Now())
			if err := fe.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update framework export: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot lock framework export: %w", err)
	}

	return fe, scope, nil
}

func (s *Service) buildAndUploadExport(ctx context.Context, scope coredata.Scoper, fe *coredata.FrameworkExport) (*coredata.FrameworkExport, error) {
	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, tx, scope, fe.FrameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			tempDir := os.TempDir()
			tempFile, err := os.CreateTemp(tempDir, "probo-framework-export-*.zip")
			if err != nil {
				return fmt.Errorf("cannot create temp file: %w", err)
			}
			defer tempFile.Close()
			defer os.Remove(tempFile.Name())

			tenantService := s.WithTenant(scope.GetTenantID())
			err = tenantService.Frameworks.Export(ctx, fe.FrameworkID, tempFile)
			if err != nil {
				return fmt.Errorf("cannot export framework: %w", err)
			}

			uuid, err := uuid.NewV4()
			if err != nil {
				return fmt.Errorf("cannot generate uuid: %w", err)
			}

			if _, err := tempFile.Seek(0, 0); err != nil {
				return fmt.Errorf("cannot seek temp file: %w", err)
			}

			fileInfo, err := tempFile.Stat()
			if err != nil {
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
				return fmt.Errorf("cannot upload file to S3: %w", err)
			}

			now := time.Now()

			file := coredata.File{
				ID:         gid.New(fe.ID.TenantID(), coredata.FileEntityType),
				BucketName: s.bucket,
				MimeType:   "application/zip",
				FileName:   fmt.Sprintf("%s Archive %s.zip", framework.Name, now.Format("2006-01-02")),
				FileKey:    uuid.String(),
				FileSize:   int(fileInfo.Size()),
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if err := file.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			fe.FileID = &file.ID
			if err := fe.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update framework export: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return fe, fmt.Errorf("cannot build and upload export: %w", err)
	}

	return fe, nil
}

func (s *Service) commitFailedExport(ctx context.Context, scope coredata.Scoper, fe *coredata.FrameworkExport) error {
	fe.CompletedAt = ref.Ref(time.Now())
	fe.Status = coredata.FrameworkExportStatusFailed

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := fe.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update framework export: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) commitSuccessfulExport(ctx context.Context, scope coredata.Scoper, fe *coredata.FrameworkExport) error {
	fe.CompletedAt = ref.Ref(time.Now())
	fe.Status = coredata.FrameworkExportStatusCompleted

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := fe.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update framework export: %w", err)
			}

			return nil
		},
	)
}
