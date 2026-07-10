// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package management

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/mail"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

type (
	UpdateRequest struct {
		ID                           gid.GID
		Active                       *bool
		Slug                         *string
		SearchEngineIndexing         *coredata.SearchEngineIndexing
		NonDisclosureAgreementFileID *gid.GID
		Description                  **string
		WebsiteURL                   **string
		Email                        **string
		HeadquarterAddress           **string
	}

	UploadNDARequest struct {
		TrustCenterID gid.GID
		File          io.Reader
		FileName      string
	}

	UpdateBrandRequest struct {
		TrustCenterID gid.GID
		LogoFile      **FileUpload
		DarkLogoFile  **FileUpload
	}
)

const maxBrandFileSize = 5 * 1024 * 1024 // 5MB

func (utcr *UpdateRequest) Validate() error {
	v := validator.New()

	v.Check(utcr.ID, "id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(utcr.Slug, "slug", validator.SafeText(NameMaxLength))
	v.Check(utcr.NonDisclosureAgreementFileID, "non_disclosure_agreement_file_id", validator.GID(coredata.FileEntityType))

	if utcr.Description != nil {
		v.Check(*utcr.Description, "description", validator.SafeText(ContentMaxLength))
	}

	if utcr.WebsiteURL != nil {
		v.Check(*utcr.WebsiteURL, "website_url", validator.SafeText(2048))
	}

	if utcr.Email != nil {
		v.Check(*utcr.Email, "email", validator.SafeText(255))
	}

	if utcr.HeadquarterAddress != nil {
		v.Check(*utcr.HeadquarterAddress, "headquarter_address", validator.SafeText(2048))
	}

	return v.Error()
}

func (utcndar *UploadNDARequest) Validate() error {
	v := validator.New()

	v.Check(utcndar.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(utcndar.FileName, "file_name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (req *UpdateBrandRequest) Validate() error {
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

func (s *Service) Get(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
) (*coredata.TrustCenter, error) {
	var trustCenter *coredata.TrustCenter

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load trust center: %w", err)
	}

	return trustCenter, nil
}

func (s *Service) GetByOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.TrustCenter, error) {
	var trustCenter *coredata.TrustCenter

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByOrganizationID(ctx, conn, scope, organizationID); err != nil {
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

func (s *Service) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateRequest,
) (*coredata.TrustCenter, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	var (
		trustCenter *coredata.TrustCenter
		file        *coredata.File
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if req.Active != nil {
				trustCenter.Active = *req.Active
			}

			if req.Slug != nil {
				trustCenter.Slug = *req.Slug
			}

			if req.SearchEngineIndexing != nil {
				trustCenter.SearchEngineIndexing = *req.SearchEngineIndexing
			}

			if req.Description != nil {
				trustCenter.Description = *req.Description
			}

			if req.WebsiteURL != nil {
				trustCenter.WebsiteURL = *req.WebsiteURL
			}

			if req.Email != nil {
				if *req.Email != nil {
					if _, err := mail.ParseAddress(**req.Email); err != nil {
						return fmt.Errorf("invalid email address: %w", err)
					}
				}

				trustCenter.Email = *req.Email
			}

			if req.HeadquarterAddress != nil {
				trustCenter.HeadquarterAddress = *req.HeadquarterAddress
			}

			trustCenter.UpdatedAt = time.Now()

			if err := trustCenter.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID != nil {
				file = &coredata.File{}
				if err := file.LoadByID(ctx, conn, scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
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

func (s *Service) UploadNDA(
	ctx context.Context,
	scope coredata.Scoper,
	req *UploadNDARequest,
) (*coredata.TrustCenter, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	var (
		trustCenter *coredata.TrustCenter
		file        *coredata.File
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.OrganizationID == gid.Nil {
				return fmt.Errorf("trust center %s has no organization", req.TrustCenterID)
			}

			objectKey, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("cannot generate object key: %w", err)
			}

			mimeType := mime.TypeByExtension(filepath.Ext(req.FileName))
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}

			now := time.Now()
			fileID := gid.New(scope.GetTenantID(), coredata.FileEntityType)

			file = &coredata.File{
				ID:             fileID,
				OrganizationID: trustCenter.OrganizationID,
				BucketName:     s.bucket,
				MimeType:       mimeType,
				FileName:       req.FileName,
				FileKey:        objectKey.String(),
				Visibility:     coredata.FileVisibilityPrivate,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			fileSize, err := s.fileManager.PutFile(
				ctx,
				file,
				req.File,
				map[string]string{
					"type":            "trust-center-nda",
					"trust-center-id": req.TrustCenterID.String(),
					"organization-id": trustCenter.OrganizationID.String(),
				},
			)
			if err != nil {
				return fmt.Errorf("cannot upload file to S3: %w", err)
			}

			file.FileSize = fileSize

			if err := file.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			trustCenter.NonDisclosureAgreementFileID = &fileID
			trustCenter.UpdatedAt = now

			if err := trustCenter.Update(ctx, conn, scope); err != nil {
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

func (s *Service) DeleteNDA(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
) (*coredata.TrustCenter, *coredata.File, error) {
	var trustCenter *coredata.TrustCenter

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			trustCenter.NonDisclosureAgreementFileID = nil
			trustCenter.UpdatedAt = time.Now()

			if err := trustCenter.Update(ctx, conn, scope); err != nil {
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

func (s *Service) UpdateBrand(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateBrandRequest,
) (*coredata.TrustCenter, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	var (
		trustCenter *coredata.TrustCenter
		ndaFile     *coredata.File
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			now := time.Now()

			if req.LogoFile != nil {
				if *req.LogoFile == nil {
					trustCenter.LogoFileID = nil
				} else {
					file, err := s.uploadBrandFile(ctx, scope, conn, *req.LogoFile, "trust-center-logo", trustCenter)
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
					file, err := s.uploadBrandFile(ctx, scope, conn, *req.DarkLogoFile, "trust-center-dark-logo", trustCenter)
					if err != nil {
						return fmt.Errorf("cannot upload dark logo file: %w", err)
					}

					trustCenter.DarkLogoFileID = &file.ID
				}
			}

			trustCenter.UpdatedAt = now

			if err := trustCenter.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID != nil {
				ndaFile = &coredata.File{}
				if err := ndaFile.LoadByID(ctx, conn, scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
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

func (s *Service) uploadBrandFile(
	ctx context.Context,
	scope coredata.Scoper,
	conn pg.Tx,
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

	_, err = s.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       &s.bucket,
		Key:          new(objectKey.String()),
		Body:         fileUpload.Content,
		ContentType:  &mimeType,
		CacheControl: new("max-age=3600, public"),
		Metadata: map[string]string{
			"type":            fileType,
			"trust-center-id": trustCenter.ID.String(),
			"organization-id": trustCenter.OrganizationID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("cannot upload file to S3: %w", err)
	}

	headOutput, err := s.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: new(s.bucket),
		Key:    new(objectKey.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get object metadata: %w", err)
	}

	now := time.Now()
	fileID := gid.New(scope.GetTenantID(), coredata.FileEntityType)

	file := &coredata.File{
		ID:             fileID,
		OrganizationID: trustCenter.OrganizationID,
		BucketName:     s.bucket,
		MimeType:       mimeType,
		FileName:       fileUpload.Filename,
		FileKey:        objectKey.String(),
		FileSize:       *headOutput.ContentLength,
		Visibility:     coredata.FileVisibilityPublic,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := file.Insert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot insert file: %w", err)
	}

	return file, nil
}

func (s *Service) GenerateNDAFileURL(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	var file *coredata.File

	trustCenter := &coredata.TrustCenter{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := trustCenter.LoadByID(ctx, conn, scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID == nil {
				return nil
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
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

	presignedURL, err := s.fileManager.GeneratePresignedURL(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s *Service) GenerateLogoURL(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}
	compliancePage := &coredata.TrustCenter{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.LogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, scope, *compliancePage.LogoFileID); err != nil {
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

	presignedURL, err := s.fileManager.GeneratePresignedURL(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s *Service) GenerateDarkLogoURL(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}
	compliancePage := &coredata.TrustCenter{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := compliancePage.LoadByID(ctx, conn, scope, compliancePageID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			if compliancePage.DarkLogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, scope, *compliancePage.DarkLogoFileID); err != nil {
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

	presignedURL, err := s.fileManager.GeneratePresignedURL(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s *Service) EmailPresenterConfig(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
) (emails.PresenterConfig, error) {
	var (
		compliancePage    = &coredata.TrustCenter{}
		organization      = &coredata.Organization{}
		logoFile          = &coredata.File{}
		compliancePageURL string
		emailPresenterCfg = emails.DefaultPresenterConfig(s.baseURL)
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
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

			publicURL, err := complianceportal.PublicURLForTrustCenter(
				ctx,
				conn,
				scope,
				compliancePage,
				s.baseDomain,
			)
			if err != nil {
				return fmt.Errorf("cannot resolve compliance page URL: %w", err)
			}

			compliancePageURL = publicURL

			return nil
		},
	)
	if err != nil {
		return emailPresenterCfg, err
	}

	emailPresenterCfg.BaseURL = compliancePageURL

	if compliancePage.LogoFileID != nil {
		if logoFile.FileKey == "" {
			return emailPresenterCfg, nil
		}

		emailPresenterCfg.SenderCompanyLogoPath = filepath.Join("/api/files/v1/public/", logoFile.ID.String())
		emailPresenterCfg.SenderCompanyName = organization.Name

		if compliancePage.WebsiteURL != nil {
			emailPresenterCfg.SenderCompanyWebsiteURL = *compliancePage.WebsiteURL
		}

		if compliancePage.HeadquarterAddress != nil {
			emailPresenterCfg.SenderCompanyHeadquarterAddress = *compliancePage.HeadquarterAddress
		}
	}

	return emailPresenterCfg, nil
}

func (s *Service) GetMailingList(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
) (*coredata.MailingList, error) {
	var mailingList *coredata.MailingList

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.MailingListID == nil {
				return nil
			}

			mailingList = &coredata.MailingList{}
			if err := mailingList.LoadByID(ctx, conn, scope, *trustCenter.MailingListID); err != nil {
				return fmt.Errorf("cannot load mailing list: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return mailingList, nil
}
