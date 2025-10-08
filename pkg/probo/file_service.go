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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/filevalidation"
	"github.com/getprobo/probo/pkg/gid"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
)

type (
	FileService struct {
		svc *TenantService
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
	ctx context.Context,
	fileID gid.GID,
) (*coredata.File, error) {
	file := &coredata.File{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := file.LoadByID(ctx, conn, s.svc.scope, fileID); err != nil {
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

func (s FileService) UploadAndSaveFile(
	ctx context.Context,
	fileValidator *filevalidation.FileValidator,
	s3Metadata map[string]string,
	req *FileUpload) (*coredata.File, error) {
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

	_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.svc.bucket,
		Key:         aws.String(objectKey.String()),
		Body:        req.Content,
		Metadata:    s3Metadata,
		ContentType: aws.String(mimeType),
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
	var file *coredata.File

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {

			file = &coredata.File{
				ID:         fileID,
				BucketName: s.svc.bucket,
				MimeType:   mimeType,
				FileName:   req.Filename,
				FileKey:    objectKey.String(),
				FileSize:   *headOutput.ContentLength,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if err := file.Insert(ctx, conn, s.svc.scope); err != nil {
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

func (s FileService) DeleteFileFromS3(ctx context.Context, fileKey string) error {
	_, err := s.svc.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(fileKey),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}

func (s FileService) GenerateFileTempURL(
	ctx context.Context,
	fileID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	file, err := s.Get(ctx, fileID)
	if err != nil {
		return "", fmt.Errorf("cannot get file: %w", err)
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(file.FileKey),
		ResponseCacheControl:       aws.String("max-age=3600, public"),
		ResponseContentType:        aws.String(file.MimeType),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", file.FileName)),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}
