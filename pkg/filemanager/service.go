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

package filemanager

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"time"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type Service struct {
	pg       *pg.Client
	baseURL  *baseurl.BaseURL
	s3Client *awss3.Client
}

func NewService(
	pgClient *pg.Client,
	baseURL *baseurl.BaseURL,
	s3Client *awss3.Client,
) *Service {
	return &Service{
		pg:       pgClient,
		baseURL:  baseURL,
		s3Client: s3Client,
	}
}

func (s *Service) GetFileBase64(
	ctx context.Context,
	file *coredata.File,
) (base64Data string, mimeType string, err error) {
	result, err := s.s3Client.GetObject(
		ctx,
		&awss3.GetObjectInput{
			Bucket: new(file.BucketName),
			Key:    new(file.FileKey),
		},
	)
	if err != nil {
		return "", "", fmt.Errorf("cannot get file from S3: %w", err)
	}

	defer func() { _ = result.Body.Close() }()

	fileData, err := io.ReadAll(result.Body)
	if err != nil {
		return "", "", fmt.Errorf("cannot read file data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(fileData), file.MimeType, nil
}

func (s *Service) GetFileBytes(
	ctx context.Context,
	file *coredata.File,
) ([]byte, error) {
	result, err := s.s3Client.GetObject(
		ctx,
		&awss3.GetObjectInput{
			Bucket: new(file.BucketName),
			Key:    new(file.FileKey),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get file from S3: %w", err)
	}

	defer func() { _ = result.Body.Close() }()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read file data: %w", err)
	}

	return data, nil
}

func (s *Service) PutFile(
	ctx context.Context,
	file *coredata.File,
	content io.Reader,
	metadata map[string]string,
) (int64, error) {
	_, err := s.s3Client.PutObject(
		ctx,
		&awss3.PutObjectInput{
			Bucket:       new(file.BucketName),
			Key:          new(file.FileKey),
			Body:         content,
			ContentType:  new(file.MimeType),
			CacheControl: new("private, max-age=3600"),
			Metadata:     metadata,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot upload file to S3: %w", err)
	}

	headOutput, err := s.s3Client.HeadObject(
		ctx,
		&awss3.HeadObjectInput{
			Bucket: new(file.BucketName),
			Key:    new(file.FileKey),
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot get object metadata: %w", err)
	}

	return *headOutput.ContentLength, nil
}

func (s *Service) GeneratePresignedFileURL(
	ctx context.Context,
	file *coredata.File,
	expiresIn time.Duration,
) (string, error) {
	presignClient := awss3.NewPresignClient(s.s3Client)

	encodedFilename := url.QueryEscape(file.FileName)
	contentDisposition := fmt.Sprintf(
		"attachment; filename=%q; filename*=UTF-8''%s",
		encodedFilename,
		encodedFilename,
	)

	presignedReq, err := presignClient.PresignGetObject(
		ctx,
		&awss3.GetObjectInput{
			Bucket:                     new(file.BucketName),
			Key:                        new(file.FileKey),
			ResponseCacheControl:       new("max-age=3600, public"),
			ResponseContentType:        new(file.MimeType),
			ResponseContentDisposition: &contentDisposition,
		},
		func(opts *awss3.PresignOptions) {
			opts.Expires = expiresIn
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}

// DownloadAPIPath returns the stable files API path for a stored file.
func DownloadAPIPath(file *coredata.File) string {
	if file.Visibility == coredata.FileVisibilityPublic {
		return "/api/files/v1/public/" + file.ID.String()
	}

	return "/api/files/v1/" + file.ID.String()
}

// BuildDownloadURL returns the absolute app URL that routes through the files API.
func (s *Service) BuildDownloadURL(file *coredata.File) (string, error) {
	url, err := s.baseURL.AppendPath(DownloadAPIPath(file)).String()
	if err != nil {
		return "", fmt.Errorf("cannot build file URL: %w", err)
	}

	return url, nil
}

// GenerateFileURL loads a public file from DB and returns the stable app URL
// /api/files/v1/public/{id}. Used when a long-lived embeddable URL is needed
// (e.g. trust center logos).
func (s *Service) GenerateFileURL(
	ctx context.Context,
	fileID gid.GID,
) (string, error) {
	file := &coredata.File{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadPublicByID(ctx, conn, fileID); err != nil {
				return fmt.Errorf("cannot load public file: %w", err)
			}

			return nil
		})
	if err != nil {
		return "", err
	}

	return s.BuildDownloadURL(file)
}

// GeneratePublicPresignedFileURL loads a public file from DB and returns a
// short-lived S3 presigned URL. Used by the public HTTP handler.
func (s *Service) GeneratePublicPresignedFileURL(
	ctx context.Context,
	fileID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	file := &coredata.File{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadPublicByID(ctx, conn, fileID); err != nil {
				return fmt.Errorf("cannot load public file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return s.GeneratePresignedFileURL(ctx, file, expiresIn)
}

// GetFileSize determines the byte size of a seekable io.Reader by seeking to
// the end and back. Returns an error if content is not seekable.
func GetFileSize(content io.Reader) (int64, error) {
	seeker, ok := content.(io.Seeker)
	if !ok {
		return 0, fmt.Errorf("cannot determine file size: content is not seekable")
	}

	size, err := seeker.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("cannot determine file size: %w", err)
	}

	_, err = seeker.Seek(0, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("cannot reset file position: %w", err)
	}

	return size, nil
}
