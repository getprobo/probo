// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package task

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
	tasksync "go.probo.inc/probo/pkg/task/sync"
)

type attachmentReader struct {
	pg     *pg.Client
	files  *filemanager.Service
	bucket string
}

func (r *attachmentReader) ReadAttachment(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	fileID gid.GID,
) (tasksync.Attachment, error) {
	if r == nil || r.files == nil {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: storage is not configured", fileID)
	}

	file := &coredata.File{}

	err := r.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadByID(ctx, conn, scope, fileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: %w", fileID, err)
	}

	if file.OrganizationID != organizationID {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: file belongs to another organization", fileID)
	}

	if file.DeletedAt != nil {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: file is deleted", fileID)
	}

	if file.Visibility != coredata.FileVisibilityPrivate {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: file is not private", fileID)
	}

	body, err := r.files.GetFileBytes(ctx, file)
	if err != nil {
		return tasksync.Attachment{}, fmt.Errorf("cannot read attachment %q: %w", fileID, err)
	}

	name := strings.TrimSpace(file.FileName)
	if name == "" {
		name = "file"
	}

	mimeType := strings.TrimSpace(file.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return tasksync.Attachment{
		Name:     name,
		MimeType: mimeType,
		Body:     body,
	}, nil
}

func (r *attachmentReader) StoreAttachment(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	file tasksync.Attachment,
) (tasksync.StoredAttachment, error) {
	if r == nil || r.files == nil || r.pg == nil || r.bucket == "" {
		return tasksync.StoredAttachment{}, fmt.Errorf("cannot store attachment: storage is not configured")
	}

	name, mimeType, err := normalizeStoredAttachment(file.Name, file.MimeType, int64(len(file.Body)))
	if err != nil {
		return tasksync.StoredAttachment{}, err
	}

	objectKey, err := uuid.NewV7()
	if err != nil {
		return tasksync.StoredAttachment{}, fmt.Errorf("cannot generate object key: %w", err)
	}

	now := time.Now()
	record := &coredata.File{
		ID:             gid.New(scope.GetTenantID(), coredata.FileEntityType),
		OrganizationID: organizationID,
		BucketName:     r.bucket,
		MimeType:       mimeType,
		FileName:       name,
		FileKey:        objectKey.String(),
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	size, err := r.files.PutFile(
		ctx,
		record,
		bytes.NewReader(file.Body),
		map[string]string{
			"type":            "file",
			"organization-id": organizationID.String(),
		},
	)
	if err != nil {
		return tasksync.StoredAttachment{}, fmt.Errorf("cannot upload attachment: %w", err)
	}

	record.FileSize = size

	err = r.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := record.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return tasksync.StoredAttachment{}, err
	}

	return tasksync.StoredAttachment{
		ID:       record.ID,
		Name:     record.FileName,
		MimeType: record.MimeType,
		Size:     record.FileSize,
	}, nil
}

func normalizeStoredAttachment(name, contentType string, size int64) (string, string, error) {
	name = pathBase(name)
	media := filevalidation.ResolveContentType(contentType, name)

	// A supported extension must agree with the downloaded media type.
	// A PDF named diagram.png is rejected instead of stored as an image.
	if _, ok := filevalidation.ContentTypeForAttachment(name); ok {
		if err := filevalidation.NewAttachmentValidator().Validate(name, media, size); err != nil {
			return "", "", err
		}

		return name, media, nil
	}

	ext := extensionForContentType(media)

	if ext == "" {
		if err := filevalidation.NewAttachmentValidator().Validate("file", media, size); err != nil {
			return "", "", err
		}

		return "", "", fmt.Errorf("cannot store attachment: content type %q is not allowed", media)
	}

	base := strings.TrimSuffix(name, filepath.Ext(name))
	if base == "" || base == "." || base == "file" {
		name = "file" + ext
	} else {
		name = base + ext
	}

	if len(name) > prosemirror.MaxFileNameBytes {
		name = trimFileName(name)
	}

	mimeType, ok := filevalidation.ContentTypeForAttachment(name)
	if !ok {
		return "", "", fmt.Errorf("cannot store attachment: content type %q is not allowed", media)
	}

	if err := filevalidation.NewAttachmentValidator().Validate(name, mimeType, size); err != nil {
		return "", "", err
	}

	return name, mimeType, nil
}

func pathBase(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, `\`, "/")
	name = filepath.Base(name)

	if name == "." || name == "/" || name == "" {
		return "file"
	}

	if len(name) > prosemirror.MaxFileNameBytes {
		return trimFileName(name)
	}

	return name
}

func trimFileName(name string) string {
	if len(name) <= prosemirror.MaxFileNameBytes {
		return name
	}

	ext := filepath.Ext(name)
	if len(ext) >= prosemirror.MaxFileNameBytes || !utf8.ValidString(ext) {
		return trimAtRune(name, prosemirror.MaxFileNameBytes)
	}

	base := trimAtRune(strings.TrimSuffix(name, ext), prosemirror.MaxFileNameBytes-len(ext))

	return base + ext
}

func trimAtRune(name string, limit int) string {
	if len(name) <= limit {
		return name
	}

	end := limit
	for end > 0 && !utf8.RuneStart(name[end]) {
		end--
	}

	if end == 0 {
		return ""
	}

	return name[:end]
}

func extensionForContentType(contentType string) string {
	for _, fileType := range filevalidation.FileTypes {
		if fileType.MimeType == contentType && len(fileType.Extensions) > 0 {
			return fileType.Extensions[0]
		}
	}

	return ""
}
