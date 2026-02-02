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
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	TrustCenterFileService struct {
		svc           *TenantService
		fileValidator *filevalidation.FileValidator
	}

	CreateTrustCenterFileRequest struct {
		OrganizationID        gid.GID
		Name                  string
		Category              string
		File                  File
		TrustCenterVisibility coredata.TrustCenterVisibility
	}

	UpdateTrustCenterFileRequest struct {
		ID                    gid.GID
		Name                  *string
		Category              *string
		TrustCenterVisibility *coredata.TrustCenterVisibility
	}
)

func (ctcfr *CreateTrustCenterFileRequest) Validate() error {
	v := validator.New()

	v.Check(ctcfr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(ctcfr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ctcfr.Category, "category", validator.Required(), validator.SafeText(TitleMaxLength))
	v.Check(ctcfr.File, "file", validator.Required())
	v.Check(ctcfr.TrustCenterVisibility, "trust_center_visibility", validator.Required(), validator.OneOfSlice(coredata.TrustCenterVisibilities()))

	return v.Error()
}

func (utcfr *UpdateTrustCenterFileRequest) Validate() error {
	v := validator.New()

	v.Check(utcfr.ID, "id", validator.Required(), validator.GID(coredata.TrustCenterFileEntityType))
	v.Check(utcfr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(utcfr.Category, "category", validator.SafeText(TitleMaxLength))
	v.Check(utcfr.TrustCenterVisibility, "trust_center_visibility", validator.OneOfSlice(coredata.TrustCenterVisibilities()))

	return v.Error()
}

func (s TrustCenterFileService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterFileOrderField],
	filter *coredata.TrustCenterFileFilter,
) (*page.Page[*coredata.TrustCenterFile, coredata.TrustCenterFileOrderField], error) {
	var files coredata.TrustCenterFiles

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := files.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter); err != nil {
				return fmt.Errorf("cannot load trust center files: %w", err)
			}

			return nil
		})

	if err != nil {
		return nil, err
	}

	return page.NewPage(files, cursor), nil
}

func (s TrustCenterFileService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var err error
			count, err = (&coredata.TrustCenterFiles{}).CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count trust center files: %w", err)
			}

			return nil
		})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s TrustCenterFileService) Get(
	ctx context.Context,
	id gid.GID,
) (*coredata.TrustCenterFile, error) {
	var file *coredata.TrustCenterFile

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		file = &coredata.TrustCenterFile{}
		if err := file.LoadByID(ctx, conn, s.svc.scope, id); err != nil {
			return fmt.Errorf("cannot load trust center file: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s TrustCenterFileService) Create(
	ctx context.Context,
	req *CreateTrustCenterFileRequest,
) (*coredata.TrustCenterFile, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Validate file
	filename := req.File.Filename
	contentType := req.File.ContentType
	fileSize, err := s.svc.fileManager.GetFileSize(req.File.Content)
	if err != nil {
		return nil, fmt.Errorf("cannot get file size: %w", err)
	}

	if err := s.fileValidator.Validate(filename, contentType, fileSize); err != nil {
		return nil, err
	}

	now := time.Now()

	trustCenterFileID := gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterFileEntityType)

	var file *coredata.TrustCenterFile
	var s3Key string

	err = s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			fileID, objectKey, err := s.uploadFile(ctx, tx, req.File, trustCenterFileID, req.OrganizationID, now)
			if err != nil {
				return fmt.Errorf("cannot upload file: %w", err)
			}
			s3Key = objectKey

			file = &coredata.TrustCenterFile{
				ID:                    trustCenterFileID,
				OrganizationID:        req.OrganizationID,
				Name:                  req.Name,
				Category:              req.Category,
				FileID:                fileID,
				TrustCenterVisibility: req.TrustCenterVisibility,
				CreatedAt:             now,
				UpdatedAt:             now,
			}

			if err := file.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert trust center file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		s.cleanupS3Object(ctx, s3Key)
		return nil, err
	}

	return file, nil
}

func (s TrustCenterFileService) Update(
	ctx context.Context,
	req *UpdateTrustCenterFileRequest,
) (*coredata.TrustCenterFile, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	var file *coredata.TrustCenterFile

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			file = &coredata.TrustCenterFile{}

			if err := file.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load trust center file: %w", err)
			}

			if req.Name != nil {
				file.Name = *req.Name
			}
			if req.Category != nil {
				file.Category = *req.Category
			}
			if req.TrustCenterVisibility != nil {
				file.TrustCenterVisibility = *req.TrustCenterVisibility
			}
			file.UpdatedAt = now

			if err := file.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update trust center file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s TrustCenterFileService) Delete(
	ctx context.Context,
	trustCenterFileID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			file := &coredata.TrustCenterFile{}

			if err := file.LoadByID(ctx, tx, s.svc.scope, trustCenterFileID); err != nil {
				return fmt.Errorf("cannot load trust center file: %w", err)
			}

			if err := file.Delete(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete trust center file: %w", err)
			}

			return nil
		})

	return err
}

func (s TrustCenterFileService) GenerateFileURL(
	ctx context.Context,
	trustCenterFileID gid.GID,
	duration time.Duration,
) (string, error) {
	var storedFile *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			file := &coredata.TrustCenterFile{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, trustCenterFileID); err != nil {
				return fmt.Errorf("cannot load trust center file: %w", err)
			}

			storedFile = &coredata.File{}
			if err := storedFile.LoadByID(ctx, conn, s.svc.scope, file.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return "", err
	}

	fileURL, err := s.svc.fileManager.GenerateFileUrl(ctx, storedFile, duration)
	if err != nil {
		return "", fmt.Errorf("cannot generate file URL: %w", err)
	}

	return fileURL, nil
}

func (s TrustCenterFileService) uploadFile(
	ctx context.Context,
	tx pg.Conn,
	file File,
	trustCenterFileID gid.GID,
	organizationID gid.GID,
	now time.Time,
) (gid.GID, string, error) {
	fileID := gid.New(s.svc.scope.GetTenantID(), coredata.FileEntityType)

	objectKey, err := uuid.NewV7()
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot generate object key: %w", err)
	}

	var fileSize int64
	var fileContent io.ReadSeeker
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

	_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.svc.bucket),
		Key:         aws.String(objectKey.String()),
		Body:        fileContent,
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"type":                 "trust-center-file",
			"trust-center-file-id": trustCenterFileID.String(),
			"organization-id":      organizationID.String(),
		},
	})
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot upload file to S3: %w", err)
	}

	fileRecord := &coredata.File{
		ID:         fileID,
		BucketName: s.svc.bucket,
		MimeType:   contentType,
		FileName:   filename,
		FileKey:    objectKey.String(),
		FileSize:   fileSize,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := fileRecord.Insert(ctx, tx, s.svc.scope); err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot insert file: %w", err)
	}

	return fileID, objectKey.String(), nil
}

func (s TrustCenterFileService) cleanupS3Object(ctx context.Context, s3Key string) {
	if s3Key == "" {
		return
	}

	_, _ = s.svc.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(s3Key),
	})
}
