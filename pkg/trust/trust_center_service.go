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

package trust

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type TrustCenterService struct {
	svc *TenantService
}

func (s TrustCenterService) Get(
	ctx context.Context,
	trustCenterID gid.GID,
) (*coredata.TrustCenter, *coredata.File, error) {
	var trustCenter *coredata.TrustCenter
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			trustCenter = &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID != nil {
				file = &coredata.File{}
				if err := file.LoadByID(ctx, conn, s.svc.scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
					return fmt.Errorf("cannot load file: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot load trust center: %w", err)
	}

	return trustCenter, file, nil
}

func (s TrustCenterService) GetByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.TrustCenter, error) {
	trustCenter := &coredata.TrustCenter{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := trustCenter.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return trustCenter, nil
}

func (s TrustCenterService) GenerateNDAFileURL(
	ctx context.Context,
	trustCenterID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	var file *coredata.File

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, trustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if trustCenter.NonDisclosureAgreementFileID == nil {
				return fmt.Errorf("no NDA file found")
			}

			file = &coredata.File{}
			if err := file.LoadByID(ctx, conn, s.svc.scope, *trustCenter.NonDisclosureAgreementFileID); err != nil {
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
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(file.FileKey),
		ResponseCacheControl:       aws.String("max-age=3600, public"),
		ResponseContentDisposition: aws.String(contentDisposition),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}

func (s TrustCenterService) GenerateLogoURL(
	ctx context.Context,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	compliancePage, _, err := s.Get(ctx, compliancePageID)
	if err != nil {
		return nil, fmt.Errorf("cannot get compliance page: %w", err)
	}

	if compliancePage.LogoFileID == nil {
		return nil, nil
	}

	url, err := s.generateFileURL(ctx, *compliancePage.LogoFileID, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return url, nil
}

func (s TrustCenterService) GenerateDarkLogoURL(
	ctx context.Context,
	compliancePageID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	compliancePage, _, err := s.Get(ctx, compliancePageID)
	if err != nil {
		return nil, fmt.Errorf("cannot get compliance page: %w", err)
	}

	if compliancePage.DarkLogoFileID == nil {
		return nil, nil
	}

	url, err := s.generateFileURL(ctx, *compliancePage.DarkLogoFileID, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return url, nil
}

func (s TrustCenterService) generateFileURL(ctx context.Context, fileID gid.GID, expiresIn time.Duration) (*string, error) {
	file := &coredata.File{}
	if err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return file.LoadByID(ctx, conn, s.svc.scope, fileID)
		},
	); err != nil {
		return nil, fmt.Errorf("cannot load file: %w", err)
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	encodedFilename := url.QueryEscape(file.FileName)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(file.FileKey),
		ResponseCacheControl:       aws.String("max-age=3600, public"),
		ResponseContentDisposition: aws.String(contentDisposition),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return nil, fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return &presignedReq.URL, nil
}
