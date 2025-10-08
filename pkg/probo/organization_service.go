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

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/filevalidation"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/slug"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
)

var (
	proboVendor = struct {
		Name                 string
		Description          string
		LegalName            string
		HeadquarterAddress   string
		WebsiteURL           string
		PrivacyPolicyURL     string
		TermsOfServiceURL    string
		SubprocessorsListURL string
	}{
		Name:                 "Probo",
		Description:          "Probo is an open-source compliance platform that helps startups achieve SOC 2 and ISO 27001 certifications quickly and affordably, with expert guidance and no vendor lock-in.",
		LegalName:            "Probo Inc.",
		HeadquarterAddress:   "490 Post St, Suite 640,San Francisco, CA 94102, United States",
		WebsiteURL:           "https://www.getprobo.com/",
		PrivacyPolicyURL:     "https://www.getprobo.com/privacy",
		TermsOfServiceURL:    "https://www.getprobo.com/terms",
		SubprocessorsListURL: "https://www.getprobo.com/subprocessors",
	}
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
)

func (s OrganizationService) Create(
	ctx context.Context,
	req CreateOrganizationRequest,
) (*coredata.Organization, error) {
	now := time.Now()
	organizationID := gid.New(s.svc.scope.GetTenantID(), coredata.OrganizationEntityType)

	organization := &coredata.Organization{
		ID:        organizationID,
		TenantID:  organizationID.TenantID(),
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := organization.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert organization: %w", err)
			}

			trustCenter := &coredata.TrustCenter{
				ID:             gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterEntityType),
				OrganizationID: organization.ID,
				TenantID:       organization.TenantID,
				Active:         false,
				Slug:           slug.Make(organization.Name),
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := trustCenter.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert trust center: %w", err)
			}

			if err := s.createProboVendor(ctx, tx, organization, now); err != nil {
				return fmt.Errorf("cannot create Probo vendor: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organization, nil
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

func (s OrganizationService) Update(
	ctx context.Context,
	req UpdateOrganizationRequest,
) (*coredata.Organization, error) {
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

func (s OrganizationService) Delete(
	ctx context.Context,
	organizationID gid.GID,
) error {
	organization := &coredata.Organization{}

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := organization.LoadByID(ctx, tx, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			// Delete documents first because versions and signatures have a delete restriction on people
			// that must be resolved before deleting the organization
			document := &coredata.Document{}
			if err := document.DeleteByOrganizationID(ctx, tx, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot delete documents: %w", err)
			}

			if err := organization.Delete(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete organization: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (s OrganizationService) createProboVendor(ctx context.Context, tx pg.Conn, organization *coredata.Organization, now time.Time) error {
	proboData := &coredata.Vendor{
		ID:                   gid.New(s.svc.scope.GetTenantID(), coredata.VendorEntityType),
		TenantID:             organization.TenantID,
		OrganizationID:       organization.ID,
		Name:                 proboVendor.Name,
		Description:          &proboVendor.Description,
		Category:             coredata.VendorCategorySecurity,
		HeadquarterAddress:   &proboVendor.HeadquarterAddress,
		LegalName:            &proboVendor.LegalName,
		WebsiteURL:           &proboVendor.WebsiteURL,
		PrivacyPolicyURL:     &proboVendor.PrivacyPolicyURL,
		TermsOfServiceURL:    &proboVendor.TermsOfServiceURL,
		SubprocessorsListURL: &proboVendor.SubprocessorsListURL,
		ShowOnTrustCenter:    false,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := proboData.Insert(ctx, tx, s.svc.scope); err != nil {
		return fmt.Errorf("cannot insert trust center: %w", err)
	}

	return nil
}
