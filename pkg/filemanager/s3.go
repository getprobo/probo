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
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/objectstorage"
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
	IfRange         string
	Range           string
}

func (s *Service) GetFileBase64(
	ctx context.Context,
	file *coredata.File,
) (base64Data string, mimeType string, err error) {
	result, err := s.store.GetObject(ctx, file.BucketName, file.FileKey, objectstorage.GetOptions{})
	if err != nil {
		return "", "", fmt.Errorf("cannot get file from object storage: %w", err)
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
	result, err := s.store.GetObject(ctx, file.BucketName, file.FileKey, objectstorage.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get file from object storage: %w", err)
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
	options := objectstorage.GetOptions{
		IfNoneMatch:     conds.IfNoneMatch,
		IfModifiedSince: conds.IfModifiedSince,
		Range:           conds.Range,
		FileSize:        file.FileSize,
	}

	if conds.Range != "" {
		honorRange := true

		if conds.IfRange != "" {
			head, err := s.store.HeadObject(ctx, file.BucketName, file.FileKey)
			if err != nil {
				return nil, fmt.Errorf("cannot head file in object storage: %w", err)
			}

			etag := head.ETag

			lastModified := file.UpdatedAt
			if !head.LastModified.IsZero() {
				lastModified = head.LastModified
			}

			honorRange = ifRangeMatches(conds.IfRange, etag, lastModified)
		}

		if !honorRange {
			options.Range = ""
		}
	}

	result, err := s.store.GetObject(ctx, file.BucketName, file.FileKey, options)
	if err != nil {
		return nil, fmt.Errorf("cannot get file from object storage: %w", err)
	}

	lastModified := result.LastModified
	if lastModified.IsZero() {
		lastModified = file.UpdatedAt
	}

	obj := &FileObject{
		Body:                result.Body,
		ContentType:         file.MimeType,
		ContentLength:       result.ContentLength,
		ContentRange:        result.ContentRange,
		ETag:                result.ETag,
		LastModified:        lastModified,
		NotModified:         result.NotModified,
		PartialContent:      result.PartialContent,
		RangeNotSatisfiable: result.RangeNotSatisfiable,
	}

	return obj, nil
}

type putFileConfig struct {
	attachmentDisposition bool
}

// PutFileOption configures optional PutFile behavior.
type PutFileOption func(*putFileConfig)

// WithAttachmentContentDisposition stores Content-Disposition: attachment on the
// uploaded object so browsers download it when served from object storage.
func WithAttachmentContentDisposition() PutFileOption {
	return func(cfg *putFileConfig) {
		cfg.attachmentDisposition = true
	}
}

func (s *Service) PutFile(
	ctx context.Context,
	file *coredata.File,
	content io.Reader,
	metadata map[string]string,
	opts ...PutFileOption,
) (int64, error) {
	cfg := putFileConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	options := objectstorage.PutOptions{
		ContentType:  file.MimeType,
		CacheControl: "private, max-age=3600",
		Metadata:     metadata,
	}
	if cfg.attachmentDisposition && file.FileName != "" {
		options.ContentDisposition = attachmentContentDisposition(file.FileName)
	}

	if err := s.store.UploadObject(ctx, file.BucketName, file.FileKey, content, options); err != nil {
		return 0, fmt.Errorf("cannot upload file to object storage: %w", err)
	}

	headOutput, err := s.store.HeadObject(ctx, file.BucketName, file.FileKey)
	if err != nil {
		return 0, fmt.Errorf("cannot get object metadata: %w", err)
	}

	return headOutput.ContentLength, nil
}

func (s *Service) GeneratePresignedURL(
	ctx context.Context,
	file *coredata.File,
	expiresIn time.Duration,
) (string, error) {
	contentDisposition := attachmentContentDisposition(file.FileName)

	presignedURL, err := s.store.PresignGetObject(
		ctx,
		file.BucketName,
		file.FileKey,
		objectstorage.PresignOptions{
			Expires:                    expiresIn,
			ResponseCacheControl:       "max-age=3600, public",
			ResponseContentType:        file.MimeType,
			ResponseContentDisposition: contentDisposition,
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot generate file download URL: %w", err)
	}

	return presignedURL, nil
}

// ifRangeMatches reports whether an If-Range validator still matches the
// current representation, following the strong-comparison rules of RFC 9110
// section 13.1.3. Object storage has no native If-Range support, so callers
// evaluate it before deciding whether to forward a Range header. An entity-tag
// validator must strong-match the current ETag (weak tags on either side never
// satisfy If-Range); otherwise the validator is an HTTP-date compared for
// equality against Last-Modified.
func ifRangeMatches(ifRange, etag string, lastModified time.Time) bool {
	ifRange = strings.TrimSpace(ifRange)
	if ifRange == "" {
		return true
	}

	if strings.HasPrefix(ifRange, `"`) {
		return etag != "" && !strings.HasPrefix(etag, `W/`) && ifRange == etag
	}

	if strings.HasPrefix(ifRange, `W/`) {
		return false
	}

	t, err := http.ParseTime(ifRange)
	if err != nil {
		return false
	}

	return !lastModified.IsZero() && lastModified.Truncate(time.Second).Equal(t)
}

func attachmentContentDisposition(filename string) string {
	return fmt.Sprintf(
		"attachment; filename=%q; filename*=UTF-8''%s",
		asciiFilename(filename),
		url.PathEscape(filename),
	)
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
