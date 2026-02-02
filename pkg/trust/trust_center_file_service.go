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
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/watermarkpdf"
)

type TrustCenterFileService struct {
	svc *TenantService
}

func (s *TrustCenterFileService) Get(
	ctx context.Context,
	trustCenterFileID gid.GID,
) (*coredata.TrustCenterFile, error) {
	trustCenterFile := &coredata.TrustCenterFile{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := trustCenterFile.LoadByID(ctx, conn, s.svc.scope, trustCenterFileID)
			if err != nil {
				return fmt.Errorf("cannot load trust center file: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return trustCenterFile, nil
}

func (s *TrustCenterFileService) ListForOrganizationId(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterFileOrderField],
	filter *coredata.TrustCenterFileFilter,
) (*page.Page[*coredata.TrustCenterFile, coredata.TrustCenterFileOrderField], error) {
	var trustCenterFiles coredata.TrustCenterFiles

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := trustCenterFiles.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load trust center files: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(trustCenterFiles, cursor), nil
}

func (s *TrustCenterFileService) ExportFile(
	ctx context.Context,
	trustCenterFileID gid.GID,
	email mail.Addr,
) ([]byte, error) {
	pdfData, err := s.exportFileData(ctx, trustCenterFileID)
	if err != nil {
		return nil, fmt.Errorf("cannot export trust center file: %w", err)
	}

	watermarkedPDF, err := watermarkpdf.AddConfidentialWithTimestamp(pdfData, email)
	if err != nil {
		return nil, fmt.Errorf("cannot add watermark to PDF: %w", err)
	}

	return watermarkedPDF, nil
}

func (s *TrustCenterFileService) ExportFileWithoutWatermark(
	ctx context.Context,
	trustCenterFileID gid.GID,
) ([]byte, error) {
	return s.exportFileData(ctx, trustCenterFileID)
}

func (s *TrustCenterFileService) exportFileData(
	ctx context.Context,
	trustCenterFileID gid.GID,
) ([]byte, error) {
	var trustCenterFile *coredata.TrustCenterFile
	var file *coredata.File

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		trustCenterFile = &coredata.TrustCenterFile{}
		if err := trustCenterFile.LoadByID(ctx, conn, s.svc.scope, trustCenterFileID); err != nil {
			return fmt.Errorf("cannot load trust center file: %w", err)
		}

		file = &coredata.File{}
		if err := file.LoadByID(ctx, conn, s.svc.scope, trustCenterFile.FileID); err != nil {
			return fmt.Errorf("cannot load file: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	result, err := s.svc.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(file.FileKey),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot download file from S3: %w", err)
	}
	defer func() { _ = result.Body.Close() }()

	fileData, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read file data: %w", err)
	}

	return fileData, nil
}
