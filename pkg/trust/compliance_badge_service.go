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
	"go.probo.inc/probo/pkg/page"
)

type ComplianceBadgeService struct {
	svc *TenantService
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

func (s ComplianceBadgeService) Get(
	ctx context.Context,
	badgeID gid.GID,
) (*coredata.ComplianceBadge, error) {
	badge := &coredata.ComplianceBadge{}

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

	return badge, nil
}
