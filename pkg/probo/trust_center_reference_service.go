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
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	TrustCenterReferenceService struct {
		svc *TenantService
	}

	CreateTrustCenterReferenceRequest struct {
		TrustCenterID gid.GID
		Name          string
		Description   *string
		WebsiteURL    string
		LogoFile      File
	}

	UpdateTrustCenterReferenceRequest struct {
		ID          gid.GID
		Name        *string
		Description **string
		WebsiteURL  *string
		LogoFile    *File
		Rank        *int
	}
)

func (ctcrr *CreateTrustCenterReferenceRequest) Validate() error {
	v := validator.New()

	v.Check(ctcrr.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(ctcrr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ctcrr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(ctcrr.WebsiteURL, "website_url", validator.Required(), validator.SafeText(2048))

	return v.Error()
}

func (utcrr *UpdateTrustCenterReferenceRequest) Validate() error {
	v := validator.New()

	v.Check(utcrr.ID, "id", validator.Required(), validator.GID(coredata.TrustCenterReferenceEntityType))
	v.Check(utcrr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(utcrr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(utcrr.WebsiteURL, "website_url", validator.SafeText(2048))

	return v.Error()
}

func (s TrustCenterReferenceService) ListForTrustCenterID(
	ctx context.Context,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterReferenceOrderField],
) (*page.Page[*coredata.TrustCenterReference, coredata.TrustCenterReferenceOrderField], error) {
	var references coredata.TrustCenterReferences

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		err := references.LoadByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID, cursor)
		if err != nil {
			return fmt.Errorf("cannot load trust center references: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return page.NewPage(references, cursor), nil
}

func (s TrustCenterReferenceService) CountForTrustCenterID(
	ctx context.Context,
	trustCenterID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) (err error) {
		references := coredata.TrustCenterReferences{}
		count, err = references.CountByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID)
		if err != nil {
			return fmt.Errorf("cannot count trust center references: %w", err)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s TrustCenterReferenceService) Get(
	ctx context.Context,
	referenceID gid.GID,
) (*coredata.TrustCenterReference, error) {
	var reference coredata.TrustCenterReference

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		err := reference.LoadByID(ctx, conn, s.svc.scope, referenceID)
		if err != nil {
			return fmt.Errorf("cannot load trust center reference: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &reference, nil
}

func (s TrustCenterReferenceService) Create(
	ctx context.Context,
	req *CreateTrustCenterReferenceRequest,
) (*coredata.TrustCenterReference, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	referenceID := gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterReferenceEntityType)

	var reference *coredata.TrustCenterReference

	var logoKey string

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		trustCenter := &coredata.TrustCenter{}
		if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
			return fmt.Errorf("cannot load trust center: %w", err)
		}

		fileID, s3Key, err := s.uploadLogoFile(ctx, tx, req.LogoFile, referenceID, req.TrustCenterID, now)
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

		if err := reference.Insert(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert trust center reference: %w", err)
		}

		return nil
	})

	if err != nil {
		s.cleanupS3Object(ctx, logoKey)
		return nil, err
	}

	return reference, nil
}

func (s TrustCenterReferenceService) Update(
	ctx context.Context,
	req *UpdateTrustCenterReferenceRequest,
) (*coredata.TrustCenterReference, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	var reference *coredata.TrustCenterReference
	var newFileID *gid.GID
	var logoKey string

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		reference = &coredata.TrustCenterReference{}

		if err := reference.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
			return fmt.Errorf("cannot load trust center reference: %w", err)
		}

		if req.LogoFile != nil {
			fileID, s3Key, err := s.uploadLogoFile(ctx, tx, *req.LogoFile, req.ID, reference.TrustCenterID, now)
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
			if err := reference.UpdateRank(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update rank: %w", err)
			}
		}

		if err := reference.Update(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot update trust center reference: %w", err)
		}

		return nil
	})

	if err != nil {
		s.cleanupS3Object(ctx, logoKey)
		return nil, err
	}

	return reference, nil
}

func (s TrustCenterReferenceService) Delete(
	ctx context.Context,
	trustCenterReferenceID gid.GID,
) error {
	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		reference := &coredata.TrustCenterReference{}

		if err := reference.LoadByID(ctx, tx, s.svc.scope, trustCenterReferenceID); err != nil {
			return fmt.Errorf("cannot load trust center reference: %w", err)
		}

		if err := reference.Delete(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot delete trust center reference: %w", err)
		}

		return nil
	})

	return err
}

func (s TrustCenterReferenceService) GenerateLogoURL(
	ctx context.Context,
	referenceID gid.GID,
	duration time.Duration,
) (string, error) {
	reference := &coredata.TrustCenterReference{}
	file := &coredata.File{}
	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		err := reference.LoadByID(ctx, tx, s.svc.scope, referenceID)
		if err != nil {
			return fmt.Errorf("cannot load trust center reference: %w", err)
		}

		err = file.LoadByID(ctx, tx, s.svc.scope, reference.LogoFileID)
		if err != nil {
			return fmt.Errorf("cannot load logo file: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", nil
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	encodedFilename := url.PathEscape(file.FileName)
	contentDisposition := fmt.Sprintf("inline; filename=\"%s\"; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(file.FileKey),
		ResponseCacheControl:       aws.String("max-age=3600, public"),
		ResponseContentDisposition: aws.String(contentDisposition),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}

func (s TrustCenterReferenceService) uploadLogoFile(
	ctx context.Context,
	tx pg.Conn,
	file File,
	referenceID gid.GID,
	trustCenterID gid.GID,
	now time.Time,
) (gid.GID, string, error) {
	fileID := gid.New(s.svc.scope.GetTenantID(), coredata.FileEntityType)

	objectKey, err := uuid.NewV7()
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot generate object key: %w", err)
	}

	trustCenter := &coredata.TrustCenter{}
	if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, trustCenterID); err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot load trust center: %w", err)
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
			"type":                      "trust-center-reference-logo",
			"trust-center-reference-id": referenceID.String(),
			"organization-id":           trustCenter.OrganizationID.String(),
		},
	})
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot upload logo file to S3: %w", err)
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

func (s TrustCenterReferenceService) cleanupS3Object(ctx context.Context, s3Key string) {
	if s3Key == "" {
		return
	}

	_, _ = s.svc.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(s3Key),
	})
}
