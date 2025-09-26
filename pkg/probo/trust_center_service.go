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
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
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

	DeleteTrustCenterNDARequest struct {
		TrustCenterID gid.GID
	}
)

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
	objectKey, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate object key: %w", err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(req.FileName))

	_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.svc.bucket,
		Key:         aws.String(objectKey.String()),
		Body:        req.File,
		ContentType: &mimeType,
		Metadata: map[string]string{
			"type":            "trust-center-nda",
			"trust-center-id": req.TrustCenterID.String(),
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot upload file to S3: %w", err)
	}

	headOutput, err := s.svc.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(objectKey.String()),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get object metadata: %w", err)
	}

	now := time.Now()

	fileID := gid.New(s.svc.scope.GetTenantID(), coredata.FileEntityType)

	var trustCenter *coredata.TrustCenter
	var file *coredata.File

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

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
	req *DeleteTrustCenterNDARequest,
) (*coredata.TrustCenter, *coredata.File, error) {
	var trustCenter *coredata.TrustCenter

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, req.TrustCenterID); err != nil {
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

	return trustCenter, nil, nil // File is nil after deletion
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
		return nil, fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return &presignedReq.URL, nil
}
