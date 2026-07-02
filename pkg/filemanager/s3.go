// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package filemanager

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.probo.inc/probo/pkg/coredata"
)

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
