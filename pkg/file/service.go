// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package file

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
)

type Service struct {
	pg          *pg.Client
	baseURL     *baseurl.BaseURL
	fileManager *filemanager.Service
}

func NewService(pgClient *pg.Client, baseURL *baseurl.BaseURL, fileManager *filemanager.Service) *Service {
	return &Service{
		pg:          pgClient,
		baseURL:     baseURL,
		fileManager: fileManager,
	}
}

// GenerateFileURL returns the stable application URL for a public file.
// The URL points to the public endpoint which redirects to a presigned S3 URL.
func (s *Service) GenerateFileURL(ctx context.Context, fileID gid.GID) (string, error) {
	file := &coredata.File{}

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := file.LoadPublicByID(ctx, conn, fileID); err != nil {
			return fmt.Errorf("cannot load public file: %w", err)
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	url, err := s.baseURL.AppendPath("/api/files/v1/public/" + fileID.String()).String()
	if err != nil {
		return "", fmt.Errorf("cannot build file URL: %w", err)
	}

	return url, nil
}

// GeneratePublicPresignedURL loads a public file and returns a short-lived S3 presigned URL.
// Used by the public /api/files/v1/public/{id} HTTP handler.
func (s *Service) GeneratePublicPresignedURL(ctx context.Context, fileID gid.GID, expiresIn time.Duration) (string, error) {
	file := &coredata.File{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadPublicByID(ctx, conn, fileID); err != nil {
				return fmt.Errorf("cannot load public file: %w", err)
			}

			return nil
		})
	if err != nil {
		return "", err
	}

	return s.fileManager.GenerateFileUrl(ctx, file, expiresIn)
}

// GeneratePresignedURL returns a short-lived S3 presigned URL for an already-loaded file.
// Used by the authenticated /api/files/v1/{id} HTTP handler to avoid a second DB round-trip.
func (s *Service) GeneratePresignedURL(ctx context.Context, file *coredata.File, expiresIn time.Duration) (string, error) {
	return s.fileManager.GenerateFileUrl(ctx, file, expiresIn)
}
