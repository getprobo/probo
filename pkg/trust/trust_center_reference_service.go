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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.gearno.de/kit/pg"
)

type TrustCenterReferenceService struct {
	svc *TenantService
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

func (s TrustCenterReferenceService) Get(
	ctx context.Context,
	referenceID gid.GID,
) (*coredata.TrustCenterReference, error) {
	reference := &coredata.TrustCenterReference{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := reference.LoadByID(ctx, conn, s.svc.scope, referenceID)
			if err != nil {
				return fmt.Errorf("cannot load trust center reference: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return reference, nil
}
