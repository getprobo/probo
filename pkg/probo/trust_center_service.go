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
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
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

	UpdateTrustCenterBrandRequest struct {
		TrustCenterID gid.GID
		LogoFile      **FileUpload
		DarkLogoFile  **FileUpload
	}
)

const maxBrandFileSize = 5 * 1024 * 1024 // 5MB

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

func (req *UpdateTrustCenterBrandRequest) Validate() error {
	fv := filevalidation.NewValidator(
		filevalidation.WithCategories(filevalidation.CategoryImage),
		filevalidation.WithMaxFileSize(maxBrandFileSize),
	)

	if req.LogoFile != nil && *req.LogoFile != nil {
		logoFile := *req.LogoFile
		if err := fv.Validate(logoFile.Filename, logoFile.ContentType, logoFile.Size); err != nil {
			return fmt.Errorf("invalid logo file: %w", err)
		}
	}

	if req.DarkLogoFile != nil && *req.DarkLogoFile != nil {
		darkLogoFile := *req.DarkLogoFile
		if err := fv.Validate(darkLogoFile.Filename, darkLogoFile.ContentType, darkLogoFile.Size); err != nil {
			return fmt.Errorf("invalid dark logo file: %w", err)
		}
	}

	return nil
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

func (s TrustCenterService) UpdateTrustCenterBrand(
	ctx context.Context,
	req *UpdateTrustCenterBrandRequest,
) (*coredata.TrustCenter, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	var trustCenter *coredata.TrustCenter
	var ndaFile *coredata.File

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			now := time.Now()

			if req.LogoFile != nil {
				if *req.LogoFile == nil {
					trustCenter.LogoFileID = nil
				} else {
					file, err := s.uploadFile(ctx, conn, *req.LogoFile, "trust-center-logo", trustCenter)
					if err != nil {
						return fmt.Errorf("cannot upload logo file: %w", err)
					}
					trustCenter.LogoFileID = &file.ID
				}
			}

			if req.DarkLogoFile != nil {
				if *req.DarkLogoFile == nil {
					trustCenter.DarkLogoFileID = nil
				} else {
					file, err := s.uploadFile(ctx, conn, *req.DarkLogoFile, "trust-center-dark-logo", trustCenter)
					if err != nil {
						return fmt.Errorf("cannot upload dark logo file: %w", err)
					}
					trustCenter.DarkLogoFileID = &file.ID
				}
			}

			trustCenter.UpdatedAt = now

			if err := trustCenter.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID != nil {
				ndaFile = &coredata.File{}
				if err := ndaFile.LoadByID(ctx, conn, s.svc.scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
					return fmt.Errorf("cannot load nda file: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return trustCenter, ndaFile, nil
}

func (s TrustCenterService) uploadFile(
	ctx context.Context,
	conn pg.Conn,
	fileUpload *FileUpload,
	fileType string,
	trustCenter *coredata.TrustCenter,
) (*coredata.File, error) {
	objectKey, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("cannot generate object key: %w", err)
	}

	mimeType := fileUpload.ContentType
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fileUpload.Filename))
	}

	_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.svc.bucket,
		Key:         aws.String(objectKey.String()),
		Body:        fileUpload.Content,
		ContentType: &mimeType,
		Metadata: map[string]string{
			"type":            fileType,
			"trust-center-id": trustCenter.ID.String(),
			"organization-id": trustCenter.OrganizationID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("cannot upload file to S3: %w", err)
	}

	headOutput, err := s.svc.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(objectKey.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get object metadata: %w", err)
	}

	now := time.Now()
	fileID := gid.New(s.svc.scope.GetTenantID(), coredata.FileEntityType)

	file := &coredata.File{
		ID:             fileID,
		OrganizationID: trustCenter.OrganizationID,
		BucketName:     s.svc.bucket,
		MimeType:       mimeType,
		FileName:       fileUpload.Filename,
		FileKey:        objectKey.String(),
		FileSize:       *headOutput.ContentLength,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := file.Insert(ctx, conn, s.svc.scope); err != nil {
		return nil, fmt.Errorf("cannot insert file: %w", err)
	}

	return file, nil
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

	if compliancePage.DarkLogoFileID == nil {
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

func (s *TrustCenterService) EmailPresenterConfig(ctx context.Context, compliancePageID gid.GID) (emails.PresenterConfig, error) {
	var (
		compliancePage    = &coredata.TrustCenter{}
		organization      = &coredata.Organization{}
		customDomain      *coredata.CustomDomain
		logoFile          = &coredata.File{}
		emailPresenterCfg = emails.DefaultPresenterConfig(s.svc.bucket, s.svc.baseURL)
	)

	scope := coredata.NewScopeFromObjectID(compliancePageID)

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.LogoFileID != nil {
				if err := logoFile.LoadByID(ctx, conn, scope, *compliancePage.LogoFileID); err != nil {
					return fmt.Errorf("cannot load logoFile: %w", err)
				}
			}

			if err := organization.LoadByID(ctx, conn, scope, compliancePage.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			customDomain = &coredata.CustomDomain{}
			if err := customDomain.LoadByOrganizationID(ctx, conn, scope, s.svc.encryptionKey, organization.ID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load custom domain: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return emailPresenterCfg, err
	}

	parsedBaseURL, err := url.Parse(s.svc.baseURL)
	if err != nil {
		return emailPresenterCfg, fmt.Errorf("cannot parse base URL: %w", err)
	}

	baseURL := url.URL{
		Scheme: parsedBaseURL.Scheme,
		Host:   parsedBaseURL.Host,
		Path:   "/trust/" + compliancePage.Slug,
	}

	if customDomain != nil && customDomain.SSLStatus == coredata.CustomDomainSSLStatusActive {
		baseURL.Host = customDomain.Domain
		baseURL.Scheme = "https"
		baseURL.Path = ""
	}

	emailPresenterCfg.BaseURL = baseURL.String()

	if compliancePage.LogoFileID != nil {
		if logoFile.FileKey == "" {
			return emailPresenterCfg, nil
		}

		emailPresenterCfg.SenderCompanyLogo = emails.Asset{
			Name:       logoFile.FileName,
			ObjectKey:  logoFile.FileKey,
			BucketName: logoFile.BucketName,
			MimeType:   logoFile.MimeType,
		}

		emailPresenterCfg.SenderCompanyName = organization.Name

		if organization.WebsiteURL != nil {
			emailPresenterCfg.SenderCompanyWebsiteURL = *organization.WebsiteURL
		}

		if organization.HeadquarterAddress != nil {
			emailPresenterCfg.SenderCompanyHeadquarterAddress = *organization.HeadquarterAddress
		}
	}

	return emailPresenterCfg, nil
}
