// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package filesign

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
)

type Service struct {
	pg          *pg.Client
	fileManager *filemanager.Service
}

func NewService(pgClient *pg.Client, fileManager *filemanager.Service) *Service {
	return &Service{
		pg:          pgClient,
		fileManager: fileManager,
	}
}

// GeneratePresignedFileURL returns a short-lived S3 presigned URL for a PUBLIC file.
// Prefer file.Service.GenerateFileURL in most cases — it returns a stable, cacheable
// application URL. Only use this when a direct S3 URL with a controlled TTL is required
// (e.g. the /api/files/v1 HTTP handler that issues the presign-on-redirect).
func (s *Service) GeneratePresignedFileURL(
	ctx context.Context,
	fileID gid.GID,
	expiresIn time.Duration,
) (string, error) {
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

	return s.fileManager.GenerateFileUrl(ctx, file, expiresIn)
}

// LoadAnyActiveFile loads a non-deleted file by ID regardless of visibility.
// Use only at the HTTP layer where visibility is checked explicitly by the caller.
func (s *Service) LoadAnyActiveFile(
	ctx context.Context,
	fileID gid.GID,
) (*coredata.File, error) {
	file := &coredata.File{}

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := file.LoadNonDeletedByID(ctx, conn, fileID); err != nil {
			return fmt.Errorf("cannot load file: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return file, nil
}

// GeneratePresignedURLForFile returns a short-lived S3 presigned URL for an
// already-loaded file. The caller is responsible for any visibility or
// authorization checks before calling this method.
func (s *Service) GeneratePresignedURLForFile(
	ctx context.Context,
	file *coredata.File,
	expiresIn time.Duration,
) (string, error) {
	return s.fileManager.GenerateFileUrl(ctx, file, expiresIn)
}
