// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

type (
	ThirdPartyDataPrivacyAgreementService struct {
		svc *TenantService
	}

	ThirdPartyDataPrivacyAgreementCreateRequest struct {
		File       io.Reader
		ValidFrom  *time.Time
		ValidUntil *time.Time
		FileName   string
	}

	ThirdPartyDataPrivacyAgreementUpdateRequest struct {
		ValidFrom  **time.Time
		ValidUntil **time.Time
	}
)

func (vdpacr *ThirdPartyDataPrivacyAgreementCreateRequest) Validate() error {
	v := validator.New()

	v.Check(vdpacr.FileName, "file_name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(vdpacr.ValidUntil, "valid_until", validator.After(vdpacr.ValidFrom))

	return v.Error()
}

func (vdpaur *ThirdPartyDataPrivacyAgreementUpdateRequest) Validate() error {
	v := validator.New()

	v.Check(vdpaur.ValidUntil, "valid_until", validator.After(vdpaur.ValidFrom))

	return v.Error()
}

func (s ThirdPartyDataPrivacyAgreementService) GetByThirdPartyID(
	ctx context.Context,
	thirdPartyID gid.GID,
) (*coredata.ThirdPartyDataPrivacyAgreement, *coredata.File, error) {
	var thirdPartyDataPrivacyAgreement *coredata.ThirdPartyDataPrivacyAgreement
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			thirdPartyDataPrivacyAgreement = &coredata.ThirdPartyDataPrivacyAgreement{}
			if err := thirdPartyDataPrivacyAgreement.LoadByThirdPartyID(ctx, conn, s.svc.scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot load thirdParty data privacy agreement: %w", err)
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, thirdPartyDataPrivacyAgreement.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return thirdPartyDataPrivacyAgreement, file, nil
}

func (s ThirdPartyDataPrivacyAgreementService) Upload(
	ctx context.Context,
	thirdPartyID gid.GID,
	req *ThirdPartyDataPrivacyAgreementCreateRequest,
) (*coredata.ThirdPartyDataPrivacyAgreement, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	objectKey, err := uuid.NewV7()
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate object key: %w", err)
	}

	var thirdPartyDataPrivacyAgreement *coredata.ThirdPartyDataPrivacyAgreement
	var file *coredata.File
	var thirdParty *coredata.ThirdParty

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			thirdParty = &coredata.ThirdParty{}
			if err := thirdParty.LoadByID(ctx, conn, s.svc.scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot load thirdParty: %w", err)
			}

			mimeType := mime.TypeByExtension(filepath.Ext(req.FileName))
			_, err := s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
				Bucket:       &s.svc.bucket,
				Key:          new(objectKey.String()),
				Body:         req.File,
				ContentType:  &mimeType,
				CacheControl: new("private, max-age=3600"),
				Metadata: map[string]string{
					"type":            "thirdParty-data-privacy-agreement",
					"thirdParty-id":   thirdPartyID.String(),
					"organization-id": thirdParty.OrganizationID.String(),
				},
			})
			if err != nil {
				return fmt.Errorf("cannot upload file to S3: %w", err)
			}
			headOutput, err := s.svc.s3.HeadObject(ctx, &s3.HeadObjectInput{
				Bucket: new(s.svc.bucket),
				Key:    new(objectKey.String()),
			})
			if err != nil {
				return fmt.Errorf("cannot get object metadata: %w", err)
			}

			now := time.Now()
			fileID := gid.New(s.svc.scope.GetTenantID(), coredata.FileEntityType)
			thirdPartyDataPrivacyAgreementID := gid.New(s.svc.scope.GetTenantID(), coredata.ThirdPartyDataPrivacyAgreementEntityType)
			file = &coredata.File{
				ID:         fileID,
				BucketName: s.svc.bucket,
				MimeType:   mimeType,
				FileName:   req.FileName,
				FileKey:    objectKey.String(),
				FileSize:   *headOutput.ContentLength,
				Visibility: coredata.FileVisibilityPrivate,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			thirdPartyDataPrivacyAgreement = &coredata.ThirdPartyDataPrivacyAgreement{
				ID:             thirdPartyDataPrivacyAgreementID,
				OrganizationID: thirdParty.OrganizationID,
				ThirdPartyID:   thirdPartyID,
				ValidFrom:      req.ValidFrom,
				ValidUntil:     req.ValidUntil,
				FileID:         fileID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := file.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			if err := thirdPartyDataPrivacyAgreement.Upsert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert thirdParty data privacy agreement: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return thirdPartyDataPrivacyAgreement, file, nil
}

func (s ThirdPartyDataPrivacyAgreementService) Get(
	ctx context.Context,
	thirdPartyDataPrivacyAgreementID gid.GID,
) (*coredata.ThirdPartyDataPrivacyAgreement, *coredata.File, error) {
	var thirdPartyDataPrivacyAgreement *coredata.ThirdPartyDataPrivacyAgreement
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			thirdPartyDataPrivacyAgreement = &coredata.ThirdPartyDataPrivacyAgreement{}
			if err := thirdPartyDataPrivacyAgreement.LoadByID(ctx, conn, s.svc.scope, thirdPartyDataPrivacyAgreementID); err != nil {
				return fmt.Errorf("cannot load thirdParty data privacy agreement: %w", err)
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, thirdPartyDataPrivacyAgreement.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot load thirdParty data privacy agreement: %w", err)
	}

	return thirdPartyDataPrivacyAgreement, file, nil
}

func (s ThirdPartyDataPrivacyAgreementService) GenerateFileURL(
	ctx context.Context,
	thirdPartyDataPrivacyAgreementID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			thirdPartyDataPrivacyAgreement := &coredata.ThirdPartyDataPrivacyAgreement{}
			if err := thirdPartyDataPrivacyAgreement.LoadByID(ctx, conn, s.svc.scope, thirdPartyDataPrivacyAgreementID); err != nil {
				return fmt.Errorf("cannot load thirdParty data privacy agreement: %w", err)
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, thirdPartyDataPrivacyAgreement.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	encodedFilename := url.QueryEscape(file.FileName)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     new(s.svc.bucket),
		Key:                        new(file.FileKey),
		ResponseCacheControl:       new("max-age=3600, public"),
		ResponseContentDisposition: new(contentDisposition),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}

func (s ThirdPartyDataPrivacyAgreementService) Update(
	ctx context.Context,
	thirdPartyID gid.GID,
	req *ThirdPartyDataPrivacyAgreementUpdateRequest,
) (*coredata.ThirdPartyDataPrivacyAgreement, *coredata.File, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	existingAgreement := &coredata.ThirdPartyDataPrivacyAgreement{}
	file := &coredata.File{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := existingAgreement.LoadByThirdPartyID(ctx, conn, s.svc.scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot load existing thirdParty data privacy agreement: %w", err)
			}

			now := time.Now()
			if req.ValidFrom != nil {
				existingAgreement.ValidFrom = *req.ValidFrom
			}
			if req.ValidUntil != nil {
				existingAgreement.ValidUntil = *req.ValidUntil
			}

			existingAgreement.UpdatedAt = now

			if err := existingAgreement.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update thirdParty data privacy agreement: %w", err)
			}

			if err := file.LoadByID(ctx, conn, s.svc.scope, existingAgreement.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return existingAgreement, file, nil
}

func (s ThirdPartyDataPrivacyAgreementService) Delete(
	ctx context.Context,
	thirdPartyDataPrivacyAgreementID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			thirdPartyDataPrivacyAgreement := &coredata.ThirdPartyDataPrivacyAgreement{}
			if err := thirdPartyDataPrivacyAgreement.LoadByID(ctx, conn, s.svc.scope, thirdPartyDataPrivacyAgreementID); err != nil {
				return fmt.Errorf("cannot load thirdParty data privacy agreement: %w", err)
			}

			if err := thirdPartyDataPrivacyAgreement.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete thirdParty data privacy agreement: %w", err)
			}

			return nil
		},
	)
}

func (s ThirdPartyDataPrivacyAgreementService) DeleteByThirdPartyID(
	ctx context.Context,
	thirdPartyID gid.GID,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			thirdPartyDataPrivacyAgreement := &coredata.ThirdPartyDataPrivacyAgreement{}
			if err := thirdPartyDataPrivacyAgreement.LoadByThirdPartyID(ctx, conn, s.svc.scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot load thirdParty data privacy agreement: %w", err)
			}

			if err := thirdPartyDataPrivacyAgreement.DeleteByThirdPartyID(ctx, conn, s.svc.scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot delete thirdParty data privacy agreement: %w", err)
			}

			return nil
		},
	)
}
