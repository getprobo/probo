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
	"errors"
	"fmt"
	"io"
	"maps"
	"mime"
	"net/mail"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/slug"
	"go.probo.inc/probo/pkg/validator"
)

type (
	UpdateRequest struct {
		ID                           gid.GID
		Active                       *bool
		Slug                         *string
		SearchEngineIndexing         *coredata.SearchEngineIndexing
		NonDisclosureAgreementFileID *gid.GID
		EntityName                   *string
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

	if utcr.EntityName != nil {
		v.Check(*utcr.EntityName, "entity_name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
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

type CreateCompliancePortalRequest struct {
	OrganizationID gid.GID
	EntityName     string
}

func (r *CreateCompliancePortalRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.EntityName, "entity_name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (s *Service) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateCompliancePortalRequest,
) (*coredata.CompliancePortal, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var portal *coredata.CompliancePortal

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			now := time.Now()
			tenantID := organization.TenantID

			mailingList := &coredata.MailingList{
				ID:             gid.New(tenantID, coredata.MailingListEntityType),
				OrganizationID: organization.ID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := mailingList.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert mailing list: %w", err)
			}

			portal = &coredata.CompliancePortal{
				ID:                   gid.New(tenantID, coredata.CompliancePortalEntityType),
				OrganizationID:       organization.ID,
				TenantID:             tenantID,
				Active:               false,
				EntityName:           req.EntityName,
				SearchEngineIndexing: coredata.SearchEngineIndexingNotIndexable,
				MailingListID:        &mailingList.ID,
				CreatedAt:            now,
				UpdatedAt:            now,
			}

			const maxSlugAttempts = 5

			var insertErr error

			// Each attempt runs in its own savepoint: a slug collision
			// aborts only the failed INSERT and leaves the surrounding
			// transaction usable for the next attempt.
			for range maxSlugAttempts {
				portal.Slug = slug.MakeWithEntropy(req.EntityName)

				insertErr = tx.Savepoint(
					ctx,
					func(ctx context.Context, sp pg.Tx) error {
						return portal.Insert(ctx, sp, scope)
					},
				)
				if insertErr == nil {
					break
				}

				if !errors.Is(insertErr, coredata.ErrResourceAlreadyExists) {
					return fmt.Errorf("cannot insert compliance portal: %w", insertErr)
				}
			}

			if insertErr != nil {
				return ErrSlugAlreadyInUse
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	s.ensureDefaultManagedDomain(ctx, scope, portal)

	return portal, nil
}

func (s *Service) ensureDefaultManagedDomain(
	ctx context.Context,
	scope coredata.Scoper,
	portal *coredata.CompliancePortal,
) {
	if s.baseDomain == "" {
		return
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			defaultDomainHostname := portal.Slug + "." + s.baseDomain

			defaultDomain := coredata.NewCustomDomain(
				portal.TenantID,
				portal.OrganizationID,
				defaultDomainHostname,
				true,
			)

			certificate, err := s.certManager.EnsureCertificate(ctx, tx, scope, defaultDomainHostname)
			if err != nil {
				return fmt.Errorf("cannot ensure certificate for default custom domain: %w", err)
			}

			defaultDomain.CertificateID = &certificate.ID

			if err := defaultDomain.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert default custom domain: %w", err)
			}

			portal.DefaultDomainID = &defaultDomain.ID
			portal.UpdatedAt = time.Now()

			if err := portal.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update compliance portal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		s.logger.ErrorCtx(
			ctx,
			"cannot provision default managed domain for compliance portal",
			log.Error(err),
			log.String("compliance_portal_id", portal.ID.String()),
		)
	}
}

func (s *Service) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.CompliancePortalOrderField],
) (*page.Page[*coredata.CompliancePortal, coredata.CompliancePortalOrderField], error) {
	var portals coredata.CompliancePortals

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := portals.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
				return fmt.Errorf("cannot list compliance portals: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(portals, cursor), nil
}

func (s *Service) CountForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			count, err = (&coredata.CompliancePortals{}).CountByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
			)
			if err != nil {
				return fmt.Errorf("cannot count compliance portals: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
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

			if req.EntityName != nil {
				portal.EntityName = *req.EntityName
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

func (s *Service) Delete(
	ctx context.Context,
	scope coredata.Scoper,
	portalID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			portal := &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, tx, scope, portalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			var files coredata.CompliancePortalFiles

			var err error

			files, err = page.LoadAll(
				ctx,
				page.OrderBy[coredata.CompliancePortalFileOrderField]{
					Field:     coredata.CompliancePortalFileOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
				func(ctx context.Context, cursor *page.Cursor[coredata.CompliancePortalFileOrderField]) ([]*coredata.CompliancePortalFile, error) {
					var batch coredata.CompliancePortalFiles
					if err := batch.LoadByCompliancePortalID(ctx, tx, scope, portalID, cursor, coredata.NewCompliancePortalFileFilter()); err != nil {
						return nil, err
					}

					return batch, nil
				},
			)
			if err != nil {
				return fmt.Errorf("cannot list portal files: %w", err)
			}

			for _, file := range files {
				if err := file.Delete(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot delete portal file: %w", err)
				}
			}

			if portal.DefaultDomainID != nil {
				if err := s.removeDomain(ctx, tx, scope, *portal.DefaultDomainID); err != nil {
					return err
				}
			}

			if portal.CustomDomainID != nil {
				if err := s.removeDomain(ctx, tx, scope, *portal.CustomDomainID); err != nil {
					return err
				}
			}

			mailingListID := portal.MailingListID

			if err := portal.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete compliance portal: %w", err)
			}

			if mailingListID != nil {
				q := fmt.Sprintf(`DELETE FROM mailing_lists WHERE %s AND id = @id`, scope.SQLFragment())
				args := pgx.StrictNamedArgs{"id": *mailingListID}
				maps.Copy(args, scope.SQLArguments())

				if _, err := tx.Exec(ctx, q, args); err != nil {
					return fmt.Errorf("cannot delete mailing list: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) removeDomain(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	domainID gid.GID,
) error {
	domain := &coredata.CustomDomain{}
	if err := domain.LoadByID(ctx, tx, scope, domainID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load custom domain: %w", err)
	}

	if err := domain.Delete(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot delete custom domain: %w", err)
	}

	if domain.CertificateID != nil {
		if err := s.certManager.Delete(ctx, tx, scope, *domain.CertificateID); err != nil {
			return fmt.Errorf("cannot delete certificate: %w", err)
		}
	}

	return nil
}
