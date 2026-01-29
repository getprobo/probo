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
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

type (
	TrustCenterService struct {
		svc *TenantService
	}

	UpdateTrustCenterRequest struct {
		ID                           gid.GID
		Active                       *bool
		Slug                         *string
		NonDisclosureAgreementFileID *gid.GID
	}

	UploadTrustCenterNDARequest struct {
		TrustCenterID gid.GID
		File          io.Reader
		FileName      string
	}
)

func (utcr *UpdateTrustCenterRequest) Validate() error {
	v := validator.New()

	v.Check(utcr.ID, "id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(utcr.Slug, "slug", validator.SafeText(NameMaxLength))
	v.Check(utcr.NonDisclosureAgreementFileID, "non_disclosure_agreement_file_id", validator.GID(coredata.FileEntityType))

	return v.Error()
}

func (utcndar *UploadTrustCenterNDARequest) Validate() error {
	v := validator.New()

	v.Check(utcndar.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(utcndar.FileName, "file_name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (s TrustCenterService) Get(
	ctx context.Context,
	trustCenterID gid.GID,
) (*coredata.TrustCenter, *coredata.File, error) {
	var trustCenter *coredata.TrustCenter
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID != nil {
				file = &coredata.File{}
				if err := file.LoadByID(ctx, conn, s.svc.scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
					return fmt.Errorf("cannot load file: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot load trust center: %w", err)
	}

	return trustCenter, file, nil
}

func (s TrustCenterService) GetByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.TrustCenter, *coredata.File, error) {
	var trustCenter *coredata.TrustCenter
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID != nil {
				file = &coredata.File{}
				if err := file.LoadByID(ctx, conn, s.svc.scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
					return fmt.Errorf("cannot load file: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return trustCenter, file, nil
}

func (s TrustCenterService) Update(
	ctx context.Context,
	req *UpdateTrustCenterRequest,
) (*coredata.TrustCenter, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	var trustCenter *coredata.TrustCenter
	var file *coredata.File

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if req.Active != nil {
				trustCenter.Active = *req.Active
			}
			if req.Slug != nil {
				trustCenter.Slug = *req.Slug
			}

			trustCenter.UpdatedAt = time.Now()

			if err := trustCenter.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID != nil {
				file = &coredata.File{}
				if err := file.LoadByID(ctx, conn, s.svc.scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
					return fmt.Errorf("cannot load file: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return trustCenter, file, nil
}

func (s TrustCenterService) UploadNDA(
	ctx context.Context,
	req *UploadTrustCenterNDARequest,
) (*coredata.TrustCenter, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	objectKey, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate object key: %w", err)
	}

	var trustCenter *coredata.TrustCenter
	var file *coredata.File

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			mimeType := mime.TypeByExtension(filepath.Ext(req.FileName))

			_, err := s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      &s.svc.bucket,
				Key:         aws.String(objectKey.String()),
				Body:        req.File,
				ContentType: &mimeType,
				Metadata: map[string]string{
					"type":            "trust-center-nda",
					"trust-center-id": req.TrustCenterID.String(),
					"organization-id": trustCenter.OrganizationID.String(),
				},
			})
			if err != nil {
				return fmt.Errorf("cannot upload file to S3: %w", err)
			}

			headOutput, err := s.svc.s3.HeadObject(ctx, &s3.HeadObjectInput{
				Bucket: aws.String(s.svc.bucket),
				Key:    aws.String(objectKey.String()),
			})
			if err != nil {
				return fmt.Errorf("cannot get object metadata: %w", err)
			}

			now := time.Now()
			fileID := gid.New(s.svc.scope.GetTenantID(), coredata.FileEntityType)

			file = &coredata.File{
				ID:         fileID,
				BucketName: s.svc.bucket,
				MimeType:   mimeType,
				FileName:   req.FileName,
				FileKey:    objectKey.String(),
				FileSize:   *headOutput.ContentLength,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if err := file.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			trustCenter.NonDisclosureAgreementFileID = &fileID
			trustCenter.UpdatedAt = now

			if err := trustCenter.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update trust center: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return trustCenter, file, nil
}

func (s TrustCenterService) DeleteNDA(
	ctx context.Context,
	trustCenterID gid.GID,
) (*coredata.TrustCenter, *coredata.File, error) {
	var trustCenter *coredata.TrustCenter

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			trustCenter.NonDisclosureAgreementFileID = nil
			trustCenter.UpdatedAt = time.Now()

			if err := trustCenter.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update trust center: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return trustCenter, nil, nil
}

func (s TrustCenterService) GenerateNDAFileURL(
	ctx context.Context,
	trustCenterID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	var file *coredata.File
	trustCenter := &coredata.TrustCenter{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID == nil {
				return nil
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if trustCenter.NonDisclosureAgreementFileID == nil {
		return nil, nil
	}

	presignedURL, err := s.svc.fileManager.GenerateFileUrl(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s TrustCenterService) GenerateLogoURL(
	ctx context.Context,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}
	compliancePage := &coredata.TrustCenter{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := compliancePage.LoadByID(ctx, conn, s.svc.scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.LogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, s.svc.scope, *compliancePage.LogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if compliancePage.LogoFileID == nil {
		return nil, nil
	}

	if file.FileKey == "" {
		return nil, nil
	}

	presignedURL, err := s.svc.fileManager.GenerateFileUrl(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s TrustCenterService) GenerateDarkLogoURL(
	ctx context.Context,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}
	compliancePage := &coredata.TrustCenter{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := compliancePage.LoadByID(ctx, conn, s.svc.scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.DarkLogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, s.svc.scope, *compliancePage.DarkLogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if compliancePage.LogoFileID == nil {
		return nil, nil
	}

	if file.FileKey == "" {
		return nil, nil
	}

	presignedURL, err := s.svc.fileManager.GenerateFileUrl(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}
