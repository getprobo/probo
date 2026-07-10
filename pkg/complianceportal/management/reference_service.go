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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	CreateReferenceRequest struct {
		TrustCenterID gid.GID
		Name          string
		Description   *string
		WebsiteURL    string
		LogoFile      File
	}

	UpdateReferenceRequest struct {
		ID          gid.GID
		Name        *string
		Description **string
		WebsiteURL  *string
		LogoFile    *File
		Rank        *int
	}
)

func (ctcrr *CreateReferenceRequest) Validate() error {
	v := validator.New()

	v.Check(ctcrr.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(ctcrr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ctcrr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(ctcrr.WebsiteURL, "website_url", validator.Required(), validator.SafeText(2048))

	return v.Error()
}

func (utcrr *UpdateReferenceRequest) Validate() error {
	v := validator.New()

	v.Check(utcrr.ID, "id", validator.Required(), validator.GID(coredata.TrustCenterReferenceEntityType))
	v.Check(utcrr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(utcrr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(utcrr.WebsiteURL, "website_url", validator.SafeText(2048))

	return v.Error()
}

func (s *Service) ListReferences(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterReferenceOrderField],
) (*page.Page[*coredata.TrustCenterReference, coredata.TrustCenterReferenceOrderField], error) {
	var references coredata.TrustCenterReferences

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := references.LoadByTrustCenterID(ctx, conn, scope, trustCenterID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load trust center references: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(references, cursor), nil
}

func (s *Service) CountReferences(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			references := coredata.TrustCenterReferences{}

			count, err = references.CountByTrustCenterID(ctx, conn, scope, trustCenterID)
			if err != nil {
				return fmt.Errorf("cannot count trust center references: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) GetReference(
	ctx context.Context,
	scope coredata.Scoper,
	referenceID gid.GID,
) (*coredata.TrustCenterReference, error) {
	var reference coredata.TrustCenterReference

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := reference.LoadByID(ctx, conn, scope, referenceID)
			if err != nil {
				return fmt.Errorf("cannot load trust center reference: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &reference, nil
}

func (s *Service) CreateReference(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateReferenceRequest,
) (*coredata.TrustCenterReference, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	referenceID := gid.New(scope.GetTenantID(), coredata.TrustCenterReferenceEntityType)

	var reference *coredata.TrustCenterReference

	var logoKey string

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, tx, scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			fileID, s3Key, err := s.uploadReferenceLogoFile(ctx, scope, tx, req.LogoFile, referenceID, req.TrustCenterID, now)
			if err != nil {
				return fmt.Errorf("cannot upload logo file: %w", err)
			}

			logoKey = s3Key

			reference = &coredata.TrustCenterReference{
				ID:             referenceID,
				OrganizationID: trustCenter.OrganizationID,
				TrustCenterID:  req.TrustCenterID,
				Name:           req.Name,
				Description:    req.Description,
				WebsiteURL:     req.WebsiteURL,
				LogoFileID:     fileID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := reference.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert trust center reference: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		s.cleanupReferenceS3Object(ctx, scope, logoKey)
		return nil, err
	}

	return reference, nil
}

func (s *Service) UpdateReference(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateReferenceRequest,
) (*coredata.TrustCenterReference, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	var (
		reference *coredata.TrustCenterReference
		newFileID *gid.GID
		logoKey   string
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			reference = &coredata.TrustCenterReference{}

			if err := reference.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load trust center reference: %w", err)
			}

			if req.LogoFile != nil {
				fileID, s3Key, err := s.uploadReferenceLogoFile(ctx, scope, tx, *req.LogoFile, req.ID, reference.TrustCenterID, now)
				if err != nil {
					return fmt.Errorf("cannot upload logo file: %w", err)
				}

				newFileID = &fileID
				logoKey = s3Key
			}

			if req.Name != nil {
				reference.Name = *req.Name
			}

			if req.Description != nil {
				reference.Description = *req.Description
			}

			if req.WebsiteURL != nil {
				reference.WebsiteURL = *req.WebsiteURL
			}

			if newFileID != nil {
				reference.LogoFileID = *newFileID
			}

			reference.UpdatedAt = now

			if req.Rank != nil {
				reference.Rank = *req.Rank
				if err := reference.UpdateRank(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update rank: %w", err)
				}
			}

			if err := reference.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update trust center reference: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		s.cleanupReferenceS3Object(ctx, scope, logoKey)
		return nil, err
	}

	return reference, nil
}

func (s *Service) DeleteReference(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterReferenceID gid.GID,
) error {
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			reference := &coredata.TrustCenterReference{}

			if err := reference.LoadByID(ctx, tx, scope, trustCenterReferenceID); err != nil {
				return fmt.Errorf("cannot load trust center reference: %w", err)
			}

			if err := reference.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete trust center reference: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s *Service) GenerateReferenceLogoURL(
	ctx context.Context,
	scope coredata.Scoper,
	referenceID gid.GID,
) (string, error) {
	reference := &coredata.TrustCenterReference{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return reference.LoadByID(ctx, tx, scope, referenceID)
		},
	)
	if err != nil {
		return "", fmt.Errorf("cannot load trust center reference: %w", err)
	}

	file, err := s.fileManager.GetPublicFile(ctx, reference.LogoFileID)
	if err != nil {
		return "", err
	}

	return s.fileManager.GenerateFileURL(file), nil
}

func (s *Service) uploadReferenceLogoFile(
	ctx context.Context,
	scope coredata.Scoper,
	tx pg.Tx,
	file File,
	referenceID gid.GID,
	trustCenterID gid.GID,
	now time.Time,
) (gid.GID, string, error) {
	fileID := gid.New(scope.GetTenantID(), coredata.FileEntityType)

	objectKey, err := uuid.NewV7()
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot generate object key: %w", err)
	}

	trustCenter := &coredata.TrustCenter{}
	if err := trustCenter.LoadByID(ctx, tx, scope, trustCenterID); err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot load trust center: %w", err)
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
			CacheControl: new("max-age=3600, public"),
			Metadata: map[string]string{
				"type":                      "trust-center-reference-logo",
				"trust-center-reference-id": referenceID.String(),
				"organization-id":           trustCenter.OrganizationID.String(),
			},
		},
	)
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot upload logo file to S3: %w", err)
	}

	fileRecord := &coredata.File{
		ID:             fileID,
		OrganizationID: trustCenter.OrganizationID,
		BucketName:     s.bucket,
		MimeType:       contentType,
		FileName:       filename,
		FileKey:        objectKey.String(),
		FileSize:       fileSize,
		Visibility:     coredata.FileVisibilityPublic,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := fileRecord.Insert(ctx, tx, scope); err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot insert file: %w", err)
	}

	return fileID, objectKey.String(), nil
}

func (s *Service) cleanupReferenceS3Object(
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
