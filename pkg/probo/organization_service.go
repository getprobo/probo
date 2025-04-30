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
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
)

type (
	OrganizationService struct {
		svc *TenantService
	}

	CreateOrganizationRequest struct {
		Name string
	}

	UpdateOrganizationRequest struct {
		ID   gid.GID
		Name *string
		File io.Reader
	}
)

func (s OrganizationService) Create(
	ctx context.Context,
	req CreateOrganizationRequest,
) (*coredata.Organization, error) {
	now := time.Now()
	organizationID, err := gid.NewGID(s.svc.scope.GetTenantID(), coredata.OrganizationEntityType)
	if err != nil {
		return nil, fmt.Errorf("cannot create organization global id: %w", err)
	}

	organization := &coredata.Organization{
		ID:        organizationID,
		TenantID:  organizationID.TenantID(),
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.svc.pg.WithConn(
		ctx,
		func(tx pg.Conn) error {
			if err := organization.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert organization: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s OrganizationService) Get(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.Organization, error) {
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return organization.LoadByID(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s OrganizationService) Update(
	ctx context.Context,
	req UpdateOrganizationRequest,
) (*coredata.Organization, error) {
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			organization.UpdatedAt = time.Now()

			if req.Name != nil {
				organization.Name = *req.Name
			}

			if req.File != nil {
				objectKey, err := uuid.NewV7()
				if err != nil {
					return fmt.Errorf("cannot generate object key: %w", err)
				}

				_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
					Bucket: aws.String(s.svc.bucket),
					Key:    aws.String(objectKey.String()),
					Body:   req.File,
				})

				if err != nil {
					return fmt.Errorf("cannot upload file to S3: %w", err)
				}

				organization.LogoObjectKey = objectKey.String()
			}

			if err := organization.Update(ctx, s.svc.scope, conn); err != nil {
				return fmt.Errorf("cannot update organization: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s OrganizationService) GenerateLogoURL(
	ctx context.Context,
	organizationID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	organization, err := s.Get(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot get organization: %w", err)
	}

	if organization.LogoObjectKey == "" {
		return nil, nil
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	encodedFilename := url.QueryEscape(organization.Name)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(organization.LogoObjectKey),
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
