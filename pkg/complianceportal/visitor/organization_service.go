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

package visitor

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *Service) GetOrganization(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.Organization, error) {
	organization := &coredata.Organization{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := organization.LoadByID(
				ctx,
				conn,
				scope,
				organizationID,
			)
			if err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (s *Service) GenerateOrganizationLogoURL(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	organization, err := s.GetOrganization(ctx, scope, organizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot get organization: %w", err)
	}

	if organization.LogoFileID == nil {
		return nil, nil
	}

	file := &coredata.File{}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return file.LoadByID(ctx, conn, scope, *organization.LogoFileID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load file: %w", err)
	}

	presignClient := s3.NewPresignClient(s.s3)

	encodedFilename := url.QueryEscape(file.FileName)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket:                     new(s.bucket),
			Key:                        new(file.FileKey),
			ResponseCacheControl:       new("max-age=3600, public"),
			ResponseContentDisposition: new(contentDisposition),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = expiresIn
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return &presignedReq.URL, nil
}
