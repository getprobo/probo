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
	"net/url"
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
	VendorBusinessAssociateAgreementService struct {
		svc *TenantService
	}

	VendorBusinessAssociateAgreementCreateRequest struct {
		File       io.Reader
		ValidFrom  *time.Time
		ValidUntil *time.Time
		FileName   string
	}

	VendorBusinessAssociateAgreementUpdateRequest struct {
		ValidFrom  **time.Time
		ValidUntil **time.Time
	}
)

func (vbaacr *VendorBusinessAssociateAgreementCreateRequest) Validate() error {
	v := validator.New()

	v.Check(vbaacr.FileName, "file_name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(vbaacr.ValidUntil, "valid_until", validator.After(vbaacr.ValidFrom))

	return v.Error()
}

func (vbaaur *VendorBusinessAssociateAgreementUpdateRequest) Validate() error {
	v := validator.New()

	v.Check(vbaaur.ValidUntil, "valid_until", validator.After(vbaaur.ValidFrom))

	return v.Error()
}

func (s VendorBusinessAssociateAgreementService) GetByVendorID(
	ctx context.Context,
	vendorID gid.GID,
) (*coredata.VendorBusinessAssociateAgreement, *coredata.File, error) {
	var vendorBusinessAssociateAgreement *coredata.VendorBusinessAssociateAgreement
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			vendorBusinessAssociateAgreement = &coredata.VendorBusinessAssociateAgreement{}
			if err := vendorBusinessAssociateAgreement.LoadByVendorID(ctx, conn, s.svc.scope, vendorID); err != nil {
				return fmt.Errorf("cannot load vendor business associate agreement: %w", err)
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, vendorBusinessAssociateAgreement.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return vendorBusinessAssociateAgreement, file, nil
}

func (s VendorBusinessAssociateAgreementService) Upload(
	ctx context.Context,
	vendorID gid.GID,
	req *VendorBusinessAssociateAgreementCreateRequest,
) (*coredata.VendorBusinessAssociateAgreement, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	objectKey, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate object key: %w", err)
	}

	var vendorBusinessAssociateAgreement *coredata.VendorBusinessAssociateAgreement
	var file *coredata.File

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			vendor := &coredata.Vendor{}
			if err := vendor.LoadByID(ctx, conn, s.svc.scope, vendorID); err != nil {
				return fmt.Errorf("cannot load vendor: %w", err)
			}

			mimeType := mime.TypeByExtension(filepath.Ext(req.FileName))

			_, err := s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      &s.svc.bucket,
				Key:         aws.String(objectKey.String()),
				Body:        req.File,
				ContentType: &mimeType,
				Metadata: map[string]string{
					"type":            "vendor-business-associate-agreement",
					"vendor-id":       vendorID.String(),
					"organization-id": vendor.OrganizationID.String(),
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
			vendorBusinessAssociateAgreementID := gid.New(s.svc.scope.GetTenantID(), coredata.VendorBusinessAssociateAgreementEntityType)

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

			vendorBusinessAssociateAgreement = &coredata.VendorBusinessAssociateAgreement{
				ID:             vendorBusinessAssociateAgreementID,
				OrganizationID: vendor.OrganizationID,
				VendorID:       vendorID,
				ValidFrom:      req.ValidFrom,
				ValidUntil:     req.ValidUntil,
				FileID:         fileID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := file.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			if err := vendorBusinessAssociateAgreement.Upsert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert vendor business associate agreement: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return vendorBusinessAssociateAgreement, file, nil
}

func (s VendorBusinessAssociateAgreementService) Get(
	ctx context.Context,
	vendorBusinessAssociateAgreementID gid.GID,
) (*coredata.VendorBusinessAssociateAgreement, *coredata.File, error) {
	var vendorBusinessAssociateAgreement *coredata.VendorBusinessAssociateAgreement
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			vendorBusinessAssociateAgreement = &coredata.VendorBusinessAssociateAgreement{}
			if err := vendorBusinessAssociateAgreement.LoadByID(ctx, conn, s.svc.scope, vendorBusinessAssociateAgreementID); err != nil {
				return fmt.Errorf("cannot load vendor business associate agreement: %w", err)
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, vendorBusinessAssociateAgreement.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot load vendor business associate agreement: %w", err)
	}

	return vendorBusinessAssociateAgreement, file, nil
}

func (s VendorBusinessAssociateAgreementService) GenerateFileURL(
	ctx context.Context,
	vendorBusinessAssociateAgreementID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			vendorBusinessAssociateAgreement := &coredata.VendorBusinessAssociateAgreement{}
			if err := vendorBusinessAssociateAgreement.LoadByID(ctx, conn, s.svc.scope, vendorBusinessAssociateAgreementID); err != nil {
				return fmt.Errorf("cannot load vendor business associate agreement: %w", err)
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, vendorBusinessAssociateAgreement.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	encodedFilename := url.QueryEscape(file.FileName)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(file.FileKey),
		ResponseCacheControl:       aws.String("max-age=3600, public"),
		ResponseContentDisposition: aws.String(contentDisposition),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}

func (s VendorBusinessAssociateAgreementService) Update(
	ctx context.Context,
	vendorID gid.GID,
	req *VendorBusinessAssociateAgreementUpdateRequest,
) (*coredata.VendorBusinessAssociateAgreement, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	existingAgreement := &coredata.VendorBusinessAssociateAgreement{}
	file := &coredata.File{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := existingAgreement.LoadByVendorID(ctx, conn, s.svc.scope, vendorID); err != nil {
				return fmt.Errorf("cannot load existing vendor business associate agreement: %w", err)
			}

			now := time.Now()
			if req.ValidFrom != nil {
				existingAgreement.ValidFrom = *req.ValidFrom
			}
			if req.ValidUntil != nil {
				existingAgreement.ValidUntil = *req.ValidUntil
			}

			existingAgreement.UpdatedAt = now

			if err := existingAgreement.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update vendor business associate agreement: %w", err)
			}

			if err := file.LoadByID(ctx, conn, s.svc.scope, existingAgreement.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return existingAgreement, file, nil
}

func (s VendorBusinessAssociateAgreementService) Delete(
	ctx context.Context,
	vendorBusinessAssociateAgreementID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			vendorBusinessAssociateAgreement := &coredata.VendorBusinessAssociateAgreement{}
			if err := vendorBusinessAssociateAgreement.LoadByID(ctx, conn, s.svc.scope, vendorBusinessAssociateAgreementID); err != nil {
				return fmt.Errorf("cannot load vendor business associate agreement: %w", err)
			}

			if err := vendorBusinessAssociateAgreement.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete vendor business associate agreement: %w", err)
			}

			return nil
		},
	)
}

func (s VendorBusinessAssociateAgreementService) DeleteByVendorID(
	ctx context.Context,
	vendorID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			vendorBusinessAssociateAgreement := &coredata.VendorBusinessAssociateAgreement{}
			if err := vendorBusinessAssociateAgreement.LoadByVendorID(ctx, conn, s.svc.scope, vendorID); err != nil {
				return fmt.Errorf("cannot load vendor business associate agreement: %w", err)
			}

			if err := vendorBusinessAssociateAgreement.DeleteByVendorID(ctx, conn, s.svc.scope, vendorID); err != nil {
				return fmt.Errorf("cannot delete vendor business associate agreement: %w", err)
			}

			return nil
		},
	)
}
