// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
)

type (
	FileService struct {
		svc *Service
	}

	File struct {
		Content     io.Reader
		Filename    string
		Size        int64
		ContentType string
	}

	FileUpload struct {
		Content     io.Reader
		Filename    string
		Size        int64
		ContentType string
	}
)

func (s FileService) Get(
	ctx context.Context, scope coredata.Scoper,
	fileID gid.GID,
) (*coredata.File, error) {
	file := &coredata.File{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadByID(ctx, conn, scope, fileID); err != nil {
				return fmt.Errorf("cannot load file %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load file: %w", err)
	}

	return file, nil
}

func (s FileService) GetByIDs(
	ctx context.Context, scope coredata.Scoper,
	fileIDs ...gid.GID,
) (coredata.Files, error) {
	var files coredata.Files

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := files.LoadByIDs(
				ctx,
				conn,
				scope,
				fileIDs,
			); err != nil {
				return fmt.Errorf("cannot load files by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (s FileService) UploadAndSaveFile(
	ctx context.Context, scope coredata.Scoper,
	fileValidator *filevalidation.FileValidator,
	s3Metadata map[string]string,
	req *FileUpload,
) (*coredata.File, error) {
	objectKey, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("cannot generate object key: %w", err)
	}

	mimeType := req.ContentType
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	if err := fileValidator.Validate(req.Filename, mimeType, req.Size); err != nil {
		return nil, fmt.Errorf("cannot validate file: %w", err)
	}

	_, err = s.svc.s3.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:       &s.svc.bucket,
			Key:          new(objectKey.String()),
			Body:         req.Content,
			Metadata:     s3Metadata,
			ContentType:  new(mimeType),
			CacheControl: new("private, max-age=3600"),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot upload file to S3: %w", err)
	}

	headOutput, err := s.svc.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: new(s.svc.bucket),
		Key:    new(objectKey.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get object metadata: %w", err)
	}

	now := time.Now()

	fileID := gid.New(scope.GetTenantID(), coredata.FileEntityType)

	var file *coredata.File

	// Extract organization ID from S3 metadata
	organizationIDStr, hasOrgID := s3Metadata["organization-id"]

	var organizationID gid.GID

	if hasOrgID {
		var err error

		organizationID, err = gid.ParseGID(organizationIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid organization-id in metadata: %w", err)
		}
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			file = &coredata.File{
				ID:             fileID,
				OrganizationID: organizationID,
				BucketName:     s.svc.bucket,
				MimeType:       mimeType,
				FileName:       req.Filename,
				FileKey:        objectKey.String(),
				FileSize:       *headOutput.ContentLength,
				Visibility:     coredata.FileVisibilityPrivate,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := file.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s FileService) GenerateFileURL(
	ctx context.Context, scope coredata.Scoper,
	fileID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	file, err := s.Get(ctx, scope, fileID)
	if err != nil {
		return "", fmt.Errorf("cannot get file: %w", err)
	}

	presignedURL, err := s.svc.fileManager.GenerateFileUrl(ctx, file, expiresIn)
	if err != nil {
		return "", fmt.Errorf("cannot generate file URL: %w", err)
	}

	return presignedURL, nil
}
