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
	ComplianceBadgeService struct {
		svc *TenantService
	}

	CreateComplianceBadgeRequest struct {
		TrustCenterID gid.GID
		Name          string
		IconFile      File
	}

	UpdateComplianceBadgeRequest struct {
		ID       gid.GID
		Name     *string
		IconFile *File
		Rank     *int
	}
)

func (r *CreateComplianceBadgeRequest) Validate() error {
	v := validator.New()

	v.Check(r.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (r *UpdateComplianceBadgeRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.ComplianceBadgeEntityType))
	v.Check(r.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (s ComplianceBadgeService) ListForTrustCenterID(
	ctx context.Context,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.ComplianceBadgeOrderField],
) (*page.Page[*coredata.ComplianceBadge, coredata.ComplianceBadgeOrderField], error) {
	var badges coredata.ComplianceBadges

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := badges.LoadByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID, cursor); err != nil {
				return fmt.Errorf("cannot load compliance badges: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(badges, cursor), nil
}

func (s ComplianceBadgeService) CountForTrustCenterID(
	ctx context.Context,
	trustCenterID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			badges := coredata.ComplianceBadges{}
			count, err = badges.CountByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID)
			if err != nil {
				return fmt.Errorf("cannot count compliance badges: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s ComplianceBadgeService) Get(
	ctx context.Context,
	badgeID gid.GID,
) (*coredata.ComplianceBadge, error) {
	var badge coredata.ComplianceBadge

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := badge.LoadByID(ctx, conn, s.svc.scope, badgeID); err != nil {
				return fmt.Errorf("cannot load compliance badge: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return &badge, nil
}

func (s ComplianceBadgeService) Create(
	ctx context.Context,
	req *CreateComplianceBadgeRequest,
) (*coredata.ComplianceBadge, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	badgeID := gid.New(s.svc.scope.GetTenantID(), coredata.ComplianceBadgeEntityType)

	var badge *coredata.ComplianceBadge
	var iconKey string

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			fileID, s3Key, err := s.uploadIconFile(ctx, tx, req.IconFile, badgeID, req.TrustCenterID, now)
			if err != nil {
				return fmt.Errorf("cannot upload icon file: %w", err)
			}
			iconKey = s3Key

			badge = &coredata.ComplianceBadge{
				ID:             badgeID,
				OrganizationID: trustCenter.OrganizationID,
				TrustCenterID:  req.TrustCenterID,
				Name:           req.Name,
				IconFileID:     fileID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := badge.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert compliance badge: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		s.cleanupS3Object(ctx, iconKey)
		return nil, err
	}

	return badge, nil
}

func (s ComplianceBadgeService) Update(
	ctx context.Context,
	req *UpdateComplianceBadgeRequest,
) (*coredata.ComplianceBadge, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	var badge *coredata.ComplianceBadge
	var newFileID *gid.GID
	var iconKey string

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			badge = &coredata.ComplianceBadge{}

			if err := badge.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance badge: %w", err)
			}

			if req.IconFile != nil {
				fileID, s3Key, err := s.uploadIconFile(ctx, tx, *req.IconFile, req.ID, badge.TrustCenterID, now)
				if err != nil {
					return fmt.Errorf("cannot upload icon file: %w", err)
				}
				newFileID = &fileID
				iconKey = s3Key
			}

			if req.Name != nil {
				badge.Name = *req.Name
			}
			if newFileID != nil {
				badge.IconFileID = *newFileID
			}
			badge.UpdatedAt = now

			if req.Rank != nil {
				badge.Rank = *req.Rank
				if err := badge.UpdateRank(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update rank: %w", err)
				}
			}

			if err := badge.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update compliance badge: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		s.cleanupS3Object(ctx, iconKey)
		return nil, err
	}

	return badge, nil
}

func (s ComplianceBadgeService) Delete(
	ctx context.Context,
	badgeID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			badge := &coredata.ComplianceBadge{}

			if err := badge.LoadByID(ctx, tx, s.svc.scope, badgeID); err != nil {
				return fmt.Errorf("cannot load compliance badge: %w", err)
			}

			if err := badge.Delete(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete compliance badge: %w", err)
			}

			return nil
		},
	)
}

func (s ComplianceBadgeService) GenerateIconURL(
	ctx context.Context,
	badgeID gid.GID,
	duration time.Duration,
) (string, error) {
	badge := &coredata.ComplianceBadge{}
	file := &coredata.File{}

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := badge.LoadByID(ctx, tx, s.svc.scope, badgeID); err != nil {
				return fmt.Errorf("cannot load compliance badge: %w", err)
			}
			if err := file.LoadByID(ctx, tx, s.svc.scope, badge.IconFileID); err != nil {
				return fmt.Errorf("cannot load icon file: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return "", fmt.Errorf("cannot load compliance badge icon: %w", err)
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

func (s ComplianceBadgeService) uploadIconFile(
	ctx context.Context,
	tx pg.Conn,
	file File,
	badgeID gid.GID,
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
			if _, err = readSeeker.Seek(0, io.SeekStart); err != nil {
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
			"type":                "compliance-badge-icon",
			"compliance-badge-id": badgeID.String(),
			"organization-id":     trustCenter.OrganizationID.String(),
		},
	})
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot upload icon file to S3: %w", err)
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

func (s ComplianceBadgeService) cleanupS3Object(ctx context.Context, s3Key string) {
	if s3Key == "" {
		return
	}
	_, _ = s.svc.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(s3Key),
	})
}
