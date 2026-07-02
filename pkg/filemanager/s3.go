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
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"go.probo.inc/probo/pkg/coredata"
)

type FileObject struct {
	Body                io.ReadCloser
	ContentType         string
	ContentLength       int64
	ContentRange        string
	ETag                string
	LastModified        time.Time
	NotModified         bool
	PartialContent      bool
	RangeNotSatisfiable bool
}

type FileConditions struct {
	IfNoneMatch     string
	IfModifiedSince time.Time
	Range           string
}

func (s *Service) GetFileBase64(
	ctx context.Context,
	file *coredata.File,
) (base64Data string, mimeType string, err error) {
	result, err := s.s3Client.GetObject(
		ctx,
		&s3.GetObjectInput{
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
		&s3.GetObjectInput{
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

func (s *Service) OpenFile(
	ctx context.Context,
	file *coredata.File,
	conds FileConditions,
) (*FileObject, error) {
	input := &s3.GetObjectInput{
		Bucket: new(file.BucketName),
		Key:    new(file.FileKey),
	}
	if conds.IfNoneMatch != "" {
		input.IfNoneMatch = &conds.IfNoneMatch
	}

	if !conds.IfModifiedSince.IsZero() {
		input.IfModifiedSince = &conds.IfModifiedSince
	}

	if conds.Range != "" {
		input.Range = &conds.Range
	}

	result, err := s.s3Client.GetObject(ctx, input)
	if err != nil {
		if respErr, ok := errors.AsType[*smithyhttp.ResponseError](err); ok {
			switch respErr.HTTPStatusCode() {
			case http.StatusNotModified:
				return &FileObject{NotModified: true}, nil
			case http.StatusRequestedRangeNotSatisfiable:
				return &FileObject{
					RangeNotSatisfiable: true,
					ContentLength:       file.FileSize,
				}, nil
			}
		}

		return nil, fmt.Errorf("cannot get file from S3: %w", err)
	}

	obj := &FileObject{
		Body:          result.Body,
		ContentType:   file.MimeType,
		ContentLength: file.FileSize,
		LastModified:  file.UpdatedAt,
	}
	if result.ETag != nil {
		obj.ETag = *result.ETag
	}

	if result.LastModified != nil {
		obj.LastModified = *result.LastModified
	}

	// A Range request that S3 honors comes back as 206 Partial Content with a
	// Content-Range header and a ContentLength scoped to the returned slice. An
	// If-Range mismatch (or no Range) yields a normal 200 with the full object,
	// so we only override the length/status when Content-Range is present.
	if result.ContentRange != nil {
		obj.ContentRange = *result.ContentRange
		obj.PartialContent = true
	}

	if result.ContentLength != nil {
		obj.ContentLength = *result.ContentLength
	}

	return obj, nil
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
		&s3.HeadObjectInput{
			Bucket: new(file.BucketName),
			Key:    new(file.FileKey),
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot get object metadata: %w", err)
	}

	return *headOutput.ContentLength, nil
}

func (s *Service) GeneratePresignedURL(
	ctx context.Context,
	file *coredata.File,
	expiresIn time.Duration,
) (string, error) {
	presignClient := s3.NewPresignClient(s.s3Client)

	contentDisposition := fmt.Sprintf(
		"attachment; filename=%q; filename*=UTF-8''%s",
		asciiFilename(file.FileName),
		url.PathEscape(file.FileName),
	)

	presignedReq, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket:                     new(file.BucketName),
			Key:                        new(file.FileKey),
			ResponseCacheControl:       new("max-age=3600, public"),
			ResponseContentType:        new(file.MimeType),
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

func asciiFilename(filename string) string {
	var b strings.Builder
	b.Grow(len(filename))

	for _, r := range filename {
		if r < 0x20 || r > 0x7e {
			b.WriteByte('_')
			continue
		}

		b.WriteRune(r)
	}

	return b.String()
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
