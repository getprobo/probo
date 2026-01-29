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
	"mime"
	"net/mail"
	"path/filepath"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

type (
	OrganizationService struct {
		svc           *TenantService
		fileValidator *filevalidation.FileValidator
	}

	CreateOrganizationRequest struct {
		Name string
	}

	UpdateOrganizationRequest struct {
		ID                 gid.GID
		Name               *string
		File               *File
		HorizontalLogoFile *File
		Description        **string
		WebsiteURL         **string
		Email              **string
		HeadquarterAddress **string
	}

	UpdateOrganizationContextRequest struct {
		OrganizationID gid.GID
		Summary        **string
	}
)

func (cor *CreateOrganizationRequest) Validate() error {
	v := validator.New()

	v.Check(cor.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (uor *UpdateOrganizationRequest) Validate() error {
	v := validator.New()

	v.Check(uor.ID, "id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(uor.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uor.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(uor.WebsiteURL, "website_url", validator.SafeText(2048))
	v.Check(uor.Email, "email", validator.SafeText(255))
	v.Check(uor.HeadquarterAddress, "headquarter_address", validator.SafeText(2048))
	v.Check(uor.File, "file", validator.NotEmpty())
	v.Check(uor.HorizontalLogoFile, "horizontal_logo_file", validator.NotEmpty())

	return v.Error()
}

func (uocr *UpdateOrganizationContextRequest) Validate() error {
	v := validator.New()

	v.Check(uocr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(uocr.Summary, "summary", validator.SafeText(30_000))

	return v.Error()
}

func (s OrganizationService) Get(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.Organization, error) {
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return organization.LoadByID(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s OrganizationService) GetContextSummary(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.OrganizationContext, error) {
	organizationContext := &coredata.OrganizationContext{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := organizationContext.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
			)
			if err != nil {
				return fmt.Errorf("cannot load organization context: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organizationContext, nil
}

func (s OrganizationService) UpdateContext(
	ctx context.Context,
	req UpdateOrganizationContextRequest,
) (*coredata.OrganizationContext, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	organization := &coredata.Organization{}
	organizationContext := &coredata.OrganizationContext{}

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := organization.LoadByID(ctx, tx, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := organizationContext.LoadByOrganizationID(ctx, tx, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization context: %w", err)
			}

			if req.Summary != nil {
				organizationContext.Summary = *req.Summary
				organizationContext.UpdatedAt = time.Now()

				if err := organizationContext.Update(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update organization context: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organizationContext, nil
}

func (s OrganizationService) Update(
	ctx context.Context,
	req UpdateOrganizationRequest,
) (*coredata.Organization, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	organization := &coredata.Organization{}

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := organization.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			now := time.Now()
			organization.UpdatedAt = now

			if req.Name != nil {
				organization.Name = *req.Name
			}

			if req.Description != nil {
				organization.Description = *req.Description
			}

			if req.WebsiteURL != nil {
				organization.WebsiteURL = *req.WebsiteURL
			}

			if req.Email != nil {
				if *req.Email != nil {
					if _, err := mail.ParseAddress(**req.Email); err != nil {
						return fmt.Errorf("invalid email address: %w", err)
					}
				}
				organization.Email = *req.Email
			}

			if req.HeadquarterAddress != nil {
				organization.HeadquarterAddress = *req.HeadquarterAddress
			}

			if err := organization.Update(ctx, s.svc.scope, tx); err != nil {
				return fmt.Errorf("cannot update organization: %w", err)
			}

			if req.File != nil {
				fileID := gid.New(s.svc.scope.GetTenantID(), coredata.FileEntityType)
				objectKey, err := uuid.NewV7()
				if err != nil {
					return fmt.Errorf("cannot generate object key: %w", err)
				}

				filename := req.File.Filename
				contentType := req.File.ContentType

				if contentType == "" {
					contentType = "application/octet-stream"
					if filename != "" {
						if detectedType := mime.TypeByExtension(filepath.Ext(filename)); detectedType != "" {
							contentType = detectedType
						}
					}
				}

				fileSize, err := s.svc.fileManager.GetFileSize(req.File.Content)
				if err != nil {
					return fmt.Errorf("cannot get file size: %w", err)
				}

				if err := s.fileValidator.Validate(filename, contentType, fileSize); err != nil {
					return err
				}

				fileRecord := &coredata.File{
					ID:         fileID,
					BucketName: s.svc.bucket,
					MimeType:   contentType,
					FileName:   filename,
					FileKey:    objectKey.String(),
					CreatedAt:  now,
					UpdatedAt:  now,
				}

				fileSize, err = s.svc.fileManager.PutFile(ctx, fileRecord, req.File.Content, map[string]string{
					"type":            "organization-logo",
					"organization-id": organization.ID.String(),
				})
				if err != nil {
					return fmt.Errorf("cannot upload logo file: %w", err)
				}

				fileRecord.FileSize = fileSize

				if err := fileRecord.Insert(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot insert file: %w", err)
				}

				organization.LogoFileID = &fileID
			}

			if req.HorizontalLogoFile != nil {
				fileID := gid.New(s.svc.scope.GetTenantID(), coredata.FileEntityType)
				objectKey, err := uuid.NewV7()
				if err != nil {
					return fmt.Errorf("cannot generate object key: %w", err)
				}

				filename := req.HorizontalLogoFile.Filename
				contentType := req.HorizontalLogoFile.ContentType

				if contentType == "" {
					contentType = "application/octet-stream"
					if filename != "" {
						if detectedType := mime.TypeByExtension(filepath.Ext(filename)); detectedType != "" {
							contentType = detectedType
						}
					}
				}

				fileSize, err := s.svc.fileManager.GetFileSize(req.HorizontalLogoFile.Content)
				if err != nil {
					return fmt.Errorf("cannot get file size: %w", err)
				}

				if err := s.fileValidator.Validate(filename, contentType, fileSize); err != nil {
					return err
				}

				fileRecord := &coredata.File{
					ID:         fileID,
					BucketName: s.svc.bucket,
					MimeType:   contentType,
					FileName:   filename,
					FileKey:    objectKey.String(),
					CreatedAt:  now,
					UpdatedAt:  now,
				}

				fileSize, err = s.svc.fileManager.PutFile(ctx, fileRecord, req.HorizontalLogoFile.Content, map[string]string{
					"type":            "organization-horizontal-logo",
					"organization-id": organization.ID.String(),
				})
				if err != nil {
					return fmt.Errorf("cannot upload horizontal logo file: %w", err)
				}

				fileRecord.FileSize = fileSize

				if err := fileRecord.Insert(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot insert file: %w", err)
				}

				organization.HorizontalLogoFileID = &fileID
			}

			if err := organization.Update(ctx, s.svc.scope, tx); err != nil {
				return fmt.Errorf("cannot update organization: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s OrganizationService) GenerateLogoURL(
	ctx context.Context,
	organizationID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if organization.LogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, s.svc.scope, *organization.LogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
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

func (s OrganizationService) GenerateHorizontalLogoURL(
	ctx context.Context,
	organizationID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if organization.HorizontalLogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, s.svc.scope, *organization.HorizontalLogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
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

func (s OrganizationService) DeleteHorizontalLogo(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.Organization, error) {
	organization := &coredata.Organization{}

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := organization.LoadByID(ctx, tx, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			organization.HorizontalLogoFileID = nil
			organization.UpdatedAt = time.Now()

			if err := organization.Update(ctx, s.svc.scope, tx); err != nil {
				return fmt.Errorf("cannot update organization: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organization, nil
}
