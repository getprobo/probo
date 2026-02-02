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

package filemanager

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.probo.inc/probo/pkg/coredata"
)

type Service struct {
	s3Client *s3.Client
}

func NewService(s3Client *s3.Client) *Service {
	return &Service{
		s3Client: s3Client,
	}
}

func (s *Service) GetFileBase64(
	ctx context.Context,
	file *coredata.File,
) (base64Data string, mimeType string, err error) {
	result, err := s.s3Client.GetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: &file.BucketName,
			Key:    &file.FileKey,
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

	if result.ContentType == nil || *result.ContentType == "" {
		return "", "", fmt.Errorf("no MIME type available for file %s", file.FileKey)
	}

	base64Data = base64.StdEncoding.EncodeToString(fileData)
	mimeType = *result.ContentType

	return base64Data, mimeType, nil
}

func (s *Service) GetFileSize(content io.Reader) (int64, error) {
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

func (s *Service) PutFile(
	ctx context.Context,
	file *coredata.File,
	content io.Reader,
	metadata map[string]string,
) (int64, error) {
	_, err := s.s3Client.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      &file.BucketName,
			Key:         &file.FileKey,
			Body:        content,
			ContentType: &file.MimeType,
			Metadata:    metadata,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot upload file to S3: %w", err)
	}

	headOutput, err := s.s3Client.HeadObject(
		ctx,
		&s3.HeadObjectInput{
			Bucket: &file.BucketName,
			Key:    &file.FileKey,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot get object metadata: %w", err)
	}

	return *headOutput.ContentLength, nil
}

func (s *Service) GenerateFileUrl(
	ctx context.Context,
	file *coredata.File,
	expiresIn time.Duration,
) (string, error) {
	presignClient := s3.NewPresignClient(s.s3Client)

	encodedFilename := url.QueryEscape(file.FileName)
	contentDisposition := fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket:                     &file.BucketName,
			Key:                        &file.FileKey,
			ResponseCacheControl:       aws.String("max-age=3600, public"),
			ResponseContentDisposition: &contentDisposition,
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = expiresIn
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}
