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
		Title                        *string
		Description                  **string
		WebsiteURL                   **string
		Email                        **string
		HeadquarterAddress           **string
	}

	UploadNDARequest struct {
		CompliancePortalID gid.GID
		File               io.Reader
		FileName           string
	}

	UpdateBrandRequest struct {
		CompliancePortalID gid.GID
		LogoFile           **FileUpload
		DarkLogoFile       **FileUpload
	}
)

const maxBrandFileSize = 5 * 1024 * 1024 // 5MB

func (utcr *UpdateRequest) Validate() error {
	v := validator.New()

	v.Check(utcr.ID, "id", validator.Required(), validator.GID(coredata.CompliancePortalEntityType))
	v.Check(utcr.Slug, "slug", validator.SafeText(NameMaxLength))
	v.Check(utcr.NonDisclosureAgreementFileID, "non_disclosure_agreement_file_id", validator.GID(coredata.FileEntityType))

	if utcr.Title != nil {
		v.Check(*utcr.Title, "title", validator.Required(), validator.SafeTextNoNewLine(NameMaxLength))
	}

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

	v.Check(utcndar.CompliancePortalID, "trust_center_id", validator.Required(), validator.GID(coredata.CompliancePortalEntityType))
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
	portalID gid.GID,
) (*coredata.CompliancePortal, error) {
	var portal *coredata.CompliancePortal

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			portal = &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, conn, scope, portalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load compliance portal: %w", err)
	}

	return portal, nil
}

func (s *Service) GetByOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.CompliancePortal, error) {
	var portal *coredata.CompliancePortal

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			portal = &coredata.CompliancePortal{}
			if err := portal.LoadByOrganizationID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return portal, nil
}

func (s *Service) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateRequest,
) (*coredata.CompliancePortal, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	var (
		portal *coredata.CompliancePortal
		file   *coredata.File
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			portal = &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			if req.Active != nil {
				portal.Active = *req.Active
			}

			if req.Slug != nil {
				portal.Slug = *req.Slug
			}

			if req.SearchEngineIndexing != nil {
				portal.SearchEngineIndexing = *req.SearchEngineIndexing
			}

			if req.Title != nil {
				portal.Title = *req.Title
			}

			if req.Description != nil {
				portal.Description = *req.Description
			}

			if req.WebsiteURL != nil {
				portal.WebsiteURL = *req.WebsiteURL
			}

			if req.Email != nil {
				if *req.Email != nil {
					if _, err := mail.ParseAddress(**req.Email); err != nil {
						return fmt.Errorf("invalid email address: %w", err)
					}
				}

				portal.Email = *req.Email
			}

			if req.HeadquarterAddress != nil {
				portal.HeadquarterAddress = *req.HeadquarterAddress
			}

			portal.UpdatedAt = time.Now()

			if err := portal.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update compliance portal: %w", err)
			}

			if portal.NonDisclosureAgreementFileID != nil {
				file = &coredata.File{}
				if err := file.LoadByID(ctx, conn, scope, *portal.NonDisclosureAgreementFileID); err != nil {
					return fmt.Errorf("cannot load file: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return portal, file, nil
}

func (s *Service) UploadNDA(
	ctx context.Context,
	scope coredata.Scoper,
	req *UploadNDARequest,
) (*coredata.CompliancePortal, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	var (
		portal *coredata.CompliancePortal
		file   *coredata.File
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			portal = &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, conn, scope, req.CompliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			if portal.OrganizationID == gid.Nil {
				return fmt.Errorf("compliance portal %s has no organization", req.CompliancePortalID)
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
				OrganizationID: portal.OrganizationID,
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
					"type":               "compliance-page-nda",
					"compliance-page-id": req.CompliancePortalID.String(),
					"organization-id":    portal.OrganizationID.String(),
				},
			)
			if err != nil {
				return fmt.Errorf("cannot upload file to S3: %w", err)
			}

			file.FileSize = fileSize

			if err := file.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			portal.NonDisclosureAgreementFileID = &fileID
			portal.UpdatedAt = now

			if err := portal.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update compliance portal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return portal, file, nil
}

func (s *Service) DeleteNDA(
	ctx context.Context,
	scope coredata.Scoper,
	portalID gid.GID,
) (*coredata.CompliancePortal, *coredata.File, error) {
	var portal *coredata.CompliancePortal

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			portal = &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, conn, scope, portalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			portal.NonDisclosureAgreementFileID = nil
			portal.UpdatedAt = time.Now()

			if err := portal.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update compliance portal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return portal, nil, nil
}

func (s *Service) UpdateBrand(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateBrandRequest,
) (*coredata.CompliancePortal, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	var (
		portal  *coredata.CompliancePortal
		ndaFile *coredata.File
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			portal = &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, conn, scope, req.CompliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			now := time.Now()

			if req.LogoFile != nil {
				if *req.LogoFile == nil {
					portal.LogoFileID = nil
				} else {
					file, err := s.uploadBrandFile(ctx, scope, conn, *req.LogoFile, "compliance-page-logo", portal)
					if err != nil {
						return fmt.Errorf("cannot upload logo file: %w", err)
					}

					portal.LogoFileID = &file.ID
				}
			}

			if req.DarkLogoFile != nil {
				if *req.DarkLogoFile == nil {
					portal.DarkLogoFileID = nil
				} else {
					file, err := s.uploadBrandFile(ctx, scope, conn, *req.DarkLogoFile, "compliance-page-dark-logo", portal)
					if err != nil {
						return fmt.Errorf("cannot upload dark logo file: %w", err)
					}

					portal.DarkLogoFileID = &file.ID
				}
			}

			portal.UpdatedAt = now

			if err := portal.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update compliance portal: %w", err)
			}

			if portal.NonDisclosureAgreementFileID != nil {
				ndaFile = &coredata.File{}
				if err := ndaFile.LoadByID(ctx, conn, scope, *portal.NonDisclosureAgreementFileID); err != nil {
					return fmt.Errorf("cannot load nda file: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return portal, ndaFile, nil
}

func (s *Service) uploadBrandFile(
	ctx context.Context,
	scope coredata.Scoper,
	conn pg.Tx,
	fileUpload *FileUpload,
	fileType string,
	portal *coredata.CompliancePortal,
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
			"type":               fileType,
			"compliance-page-id": portal.ID.String(),
			"organization-id":    portal.OrganizationID.String(),
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
		OrganizationID: portal.OrganizationID,
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
	portalID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	var file *coredata.File

	portal := &coredata.CompliancePortal{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := portal.LoadByID(ctx, conn, scope, portalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			if portal.NonDisclosureAgreementFileID == nil {
				return nil
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, scope, *portal.NonDisclosureAgreementFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if portal.NonDisclosureAgreementFileID == nil {
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
	portalID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}
	portal := &coredata.CompliancePortal{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := portal.LoadByID(ctx, conn, scope, portalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			if portal.LogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, scope, *portal.LogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if portal.LogoFileID == nil {
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
	portalID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}
	portal := &coredata.CompliancePortal{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := portal.LoadByID(ctx, conn, scope, portalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			if portal.DarkLogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, scope, *portal.DarkLogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if portal.DarkLogoFileID == nil {
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
	portalID gid.GID,
) (emails.PresenterConfig, error) {
	var (
		portal            = &coredata.CompliancePortal{}
		organization      = &coredata.Organization{}
		logoFile          = &coredata.File{}
		portalURL         string
		emailPresenterCfg = emails.DefaultPresenterConfig(s.baseURL)
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := portal.LoadByID(ctx, conn, scope, portalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			if portal.LogoFileID != nil {
				if err := logoFile.LoadByID(ctx, conn, scope, *portal.LogoFileID); err != nil {
					return fmt.Errorf("cannot load logoFile: %w", err)
				}
			}

			if err := organization.LoadByID(ctx, conn, scope, portal.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			publicURL, err := s.PublicURLForCompliancePortal(ctx, conn, scope, portal)
			if err != nil {
				return fmt.Errorf("cannot resolve compliance portal URL: %w", err)
			}

			portalURL = publicURL

			return nil
		},
	)
	if err != nil {
		return emailPresenterCfg, err
	}

	emailPresenterCfg.BaseURL = portalURL

	if portal.LogoFileID != nil {
		if logoFile.FileKey == "" {
			return emailPresenterCfg, nil
		}

		emailPresenterCfg.SenderCompanyLogoPath = filepath.Join("/api/files/v1/public/", logoFile.ID.String())
		emailPresenterCfg.SenderCompanyName = organization.Name

		if portal.WebsiteURL != nil {
			emailPresenterCfg.SenderCompanyWebsiteURL = *portal.WebsiteURL
		}

		if portal.HeadquarterAddress != nil {
			emailPresenterCfg.SenderCompanyHeadquarterAddress = *portal.HeadquarterAddress
		}
	}

	return emailPresenterCfg, nil
}

func (s *Service) GetMailingList(
	ctx context.Context,
	scope coredata.Scoper,
	portalID gid.GID,
) (*coredata.MailingList, error) {
	var mailingList *coredata.MailingList

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			portal := &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, conn, scope, portalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			if portal.MailingListID == nil {
				return nil
			}

			mailingList = &coredata.MailingList{}
			if err := mailingList.LoadByID(ctx, conn, scope, *portal.MailingListID); err != nil {
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
