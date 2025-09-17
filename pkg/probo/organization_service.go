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
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/mail"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			organization.UpdatedAt = time.Now()

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
						return &coredata.ErrInvalidValue{Field: "email"}
					}
				}
				organization.Email = *req.Email
			}

			if req.HeadquarterAddress != nil {
				organization.HeadquarterAddress = *req.HeadquarterAddress
			}

			if req.File != nil {
				objectKey, err := uuid.NewV7()
				if err != nil {
					return fmt.Errorf("cannot generate object key: %w", err)
				}

				var fileSize int64
				var fileContent io.ReadSeeker
				filename := req.File.Filename
				contentType := req.File.ContentType

				if seeker, ok := req.File.Content.(io.Seeker); ok {
					if req.File.Size <= 0 {
						size, err := seeker.Seek(0, io.SeekEnd)
						if err != nil {
							return fmt.Errorf("cannot determine file size: %w", err)
						}
						fileSize = size

						_, err = seeker.Seek(0, io.SeekStart)
						if err != nil {
							return fmt.Errorf("cannot reset file position: %w", err)
						}
					} else {
						fileSize = req.File.Size
					}
					fileContent = req.File.Content.(io.ReadSeeker)
				} else {
					buf, err := io.ReadAll(req.File.Content)
					if err != nil {
						return fmt.Errorf("cannot read file: %w", err)
					}
					fileSize = int64(len(buf))
					fileContent = bytes.NewReader(buf)
				}

				if contentType == "" {
					contentType = "application/octet-stream"
					if filename != "" {
						if detectedType := mime.TypeByExtension(filepath.Ext(filename)); detectedType != "" {
							contentType = detectedType
						}
					}
				}

				if err := s.fileValidator.Validate(filename, contentType, fileSize); err != nil {
					return err
				}

				_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
					Bucket:      aws.String(s.svc.bucket),
					Key:         aws.String(objectKey.String()),
					Body:        fileContent,
					ContentType: aws.String(contentType),
					Metadata: map[string]string{
						"type":            "organization-logo",
						"organization-id": organization.ID.String(),
					},
				})

				if err != nil {
					return fmt.Errorf("cannot upload file to S3: %w", err)
				}

				organization.LogoObjectKey = objectKey.String()
			}

			if err := organization.Update(ctx, s.svc.scope, conn); err != nil {
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
	organization, err := s.Get(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot get organization: %w", err)
	}

	if organization.LogoObjectKey == "" {
		return nil, nil
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	encodedFilename := url.QueryEscape(organization.Name)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(organization.LogoObjectKey),
		ResponseCacheControl:       aws.String("max-age=3600, public"),
		ResponseContentDisposition: aws.String(contentDisposition),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return nil, fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return &presignedReq.URL, nil
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

	if organization.LogoObjectKey != "" {
		_, err := s.svc.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.svc.bucket),
			Key:    aws.String(organization.LogoObjectKey),
		})
		if err != nil {
			return fmt.Errorf("organization deleted but failed to delete logo from S3: %w", err)
		}
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
