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

package management

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	CreateFileRequest struct {
		OrganizationID             gid.GID
		Name                       string
		Category                   string
		File                       File
		CompliancePortalVisibility coredata.CompliancePortalVisibility
	}

	UpdateFileRequest struct {
		ID                         gid.GID
		Name                       *string
		Category                   *string
		CompliancePortalVisibility *coredata.CompliancePortalVisibility
	}
)

func (ctcfr *CreateFileRequest) Validate() error {
	v := validator.New()

	v.Check(ctcfr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(ctcfr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ctcfr.Category, "category", validator.Required(), validator.SafeText(TitleMaxLength))
	v.Check(ctcfr.File, "file", validator.Required())
	v.Check(ctcfr.CompliancePortalVisibility, "trust_center_visibility", validator.Required(), validator.OneOfSlice(coredata.CompliancePortalVisibilities()))

	return v.Error()
}

func (utcfr *UpdateFileRequest) Validate() error {
	v := validator.New()

	v.Check(utcfr.ID, "id", validator.Required(), validator.GID(coredata.CompliancePortalFileEntityType))
	v.Check(utcfr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(utcfr.Category, "category", validator.SafeText(TitleMaxLength))
	v.Check(utcfr.CompliancePortalVisibility, "trust_center_visibility", validator.OneOfSlice(coredata.CompliancePortalVisibilities()))

	return v.Error()
}

func (s *Service) ListFilesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.CompliancePortalFileOrderField],
	filter *coredata.CompliancePortalFileFilter,
) (*page.Page[*coredata.CompliancePortalFile, coredata.CompliancePortalFileOrderField], error) {
	var files coredata.CompliancePortalFiles

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := files.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter); err != nil {
				return fmt.Errorf("cannot load compliance page files: %w", err)
			}

			return nil
		})
	if err != nil {
		return nil, err
	}

	return page.NewPage(files, cursor), nil
}

func (s *Service) CountFilesForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			count, err = (&coredata.CompliancePortalFiles{}).CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count compliance page files: %w", err)
			}

			return nil
		})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) GetFile(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.CompliancePortalFile, error) {
	var file *coredata.CompliancePortalFile

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			file = &coredata.CompliancePortalFile{}
			if err := file.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load compliance page file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *Service) CreateFile(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateFileRequest,
) (*coredata.CompliancePortalFile, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Validate file
	filename := req.File.Filename
	contentType := req.File.ContentType

	fileSize, err := filemanager.GetFileSize(req.File.Content)
	if err != nil {
		return nil, fmt.Errorf("cannot get file size: %w", err)
	}

	if err := s.fileValidator.Validate(filename, contentType, fileSize); err != nil {
		return nil, err
	}

	now := time.Now()

	compliancePortalFileID := gid.New(scope.GetTenantID(), coredata.CompliancePortalFileEntityType)

	var (
		file  *coredata.CompliancePortalFile
		s3Key string
	)

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			fileID, objectKey, err := s.uploadFile(ctx, scope, tx, req.File, compliancePortalFileID, req.OrganizationID, now)
			if err != nil {
				return fmt.Errorf("cannot upload file: %w", err)
			}

			s3Key = objectKey

			file = &coredata.CompliancePortalFile{
				ID:                         compliancePortalFileID,
				OrganizationID:             req.OrganizationID,
				Name:                       req.Name,
				Category:                   req.Category,
				FileID:                     fileID,
				CompliancePortalVisibility: req.CompliancePortalVisibility,
				CreatedAt:                  now,
				UpdatedAt:                  now,
			}

			if err := file.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert compliance page file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		s.cleanupFileS3Object(ctx, scope, s3Key)
		return nil, err
	}

	return file, nil
}

func (s *Service) UpdateFile(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateFileRequest,
) (*coredata.CompliancePortalFile, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	var file *coredata.CompliancePortalFile

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			file = &coredata.CompliancePortalFile{}

			if err := file.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance page file: %w", err)
			}

			if req.Name != nil {
				file.Name = *req.Name
			}

			if req.Category != nil {
				file.Category = *req.Category
			}

			if req.CompliancePortalVisibility != nil {
				file.CompliancePortalVisibility = *req.CompliancePortalVisibility
			}

			file.UpdatedAt = now

			if err := file.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update compliance page file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *Service) DeleteFile(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalFileID gid.GID,
) error {
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			file := &coredata.CompliancePortalFile{}

			if err := file.LoadByID(ctx, tx, scope, compliancePortalFileID); err != nil {
				return fmt.Errorf("cannot load compliance page file: %w", err)
			}

			if err := file.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete compliance page file: %w", err)
			}

			return nil
		})

	return err
}

func (s *Service) GenerateFileURL(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePortalFileID gid.GID,
	duration time.Duration,
) (string, error) {
	var storedFile *coredata.File

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			file := &coredata.CompliancePortalFile{}
			if err := file.LoadByID(ctx, conn, scope, compliancePortalFileID); err != nil {
				return fmt.Errorf("cannot load compliance page file: %w", err)
			}

			storedFile = &coredata.File{}
			if err := storedFile.LoadByID(ctx, conn, scope, file.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	fileURL, err := s.fileManager.GeneratePresignedURL(ctx, storedFile, duration)
	if err != nil {
		return "", fmt.Errorf("cannot generate file URL: %w", err)
	}

	return fileURL, nil
}

func (s *Service) uploadFile(
	ctx context.Context,
	scope coredata.Scoper,
	tx pg.Tx,
	file File,
	compliancePortalFileID gid.GID,
	organizationID gid.GID,
	now time.Time,
) (gid.GID, string, error) {
	fileID := gid.New(scope.GetTenantID(), coredata.FileEntityType)

	objectKey, err := uuid.NewV7()
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot generate object key: %w", err)
	}

	var (
		fileSize    int64
		fileContent io.ReadSeeker
	)

	filename := file.Filename
	contentType := file.ContentType

	if readSeeker, ok := file.Content.(io.ReadSeeker); ok {
		if file.Size <= 0 {
			size, err := readSeeker.Seek(0, io.SeekEnd)
			if err != nil {
				return gid.GID{}, "", fmt.Errorf("cannot determine file size: %w", err)
			}

			fileSize = size

			_, err = readSeeker.Seek(0, io.SeekStart)
			if err != nil {
				return gid.GID{}, "", fmt.Errorf("cannot reset file position: %w", err)
			}
		} else {
			fileSize = file.Size
		}

		fileContent = readSeeker
	} else {
		buf, err := io.ReadAll(file.Content)
		if err != nil {
			return gid.GID{}, "", fmt.Errorf("cannot read file: %w", err)
		}

		fileSize = int64(len(buf))
		fileContent = bytes.NewReader(buf)
	}

	if contentType == "" {
		contentType = "application/octet-stream"

		if filename != "" {
			if detectedType := mime.TypeByExtension(filepath.Ext(filename)); detectedType != "" {
				contentType = detectedType
			}
		}
	}

	_, err = s.s3.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:       new(s.bucket),
			Key:          new(objectKey.String()),
			Body:         fileContent,
			ContentType:  new(contentType),
			CacheControl: new("private, max-age=3600"),
			Metadata: map[string]string{
				"type":                    "compliance-page-file",
				"compliance-page-file-id": compliancePortalFileID.String(),
				"organization-id":         organizationID.String(),
			},
		},
	)
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot upload file to S3: %w", err)
	}

	fileRecord := &coredata.File{
		ID:             fileID,
		OrganizationID: organizationID,
		BucketName:     s.bucket,
		MimeType:       contentType,
		FileName:       filename,
		FileKey:        objectKey.String(),
		FileSize:       fileSize,
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := fileRecord.Insert(ctx, tx, scope); err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot insert file: %w", err)
	}

	return fileID, objectKey.String(), nil
}

func (s *Service) cleanupFileS3Object(
	ctx context.Context,
	scope coredata.Scoper,
	s3Key string,
) {
	if s3Key == "" {
		return
	}

	_, _ = s.s3.DeleteObject(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: new(s.bucket),
			Key:    new(s3Key),
		},
	)
}
