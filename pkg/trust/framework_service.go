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
	"time"

	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type FrameworkService struct {
	svc *TenantService
}

func (s FrameworkService) Get(
	ctx context.Context,
	frameworkID gid.GID,
) (*coredata.Framework, error) {
	framework := &coredata.Framework{}

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		err := framework.LoadByID(ctx, conn, s.svc.scope, frameworkID)
		if err != nil {
			return fmt.Errorf("cannot load framework: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return framework, nil
}

func (s FrameworkService) GenerateLightLogoURL(
	ctx context.Context,
	frameworkID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, s.svc.scope, frameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if framework.LightLogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, s.svc.scope, *framework.LightLogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if file.FileKey == "" {
		return nil, nil
	}

	presignedURL, err := s.svc.fileManager.GenerateFileUrl(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}

func (s FrameworkService) GenerateDarkLogoURL(
	ctx context.Context,
	frameworkID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	file := &coredata.File{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, s.svc.scope, frameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			if framework.DarkLogoFileID == nil {
				return nil
			}

			if err := file.LoadByID(ctx, conn, s.svc.scope, *framework.DarkLogoFileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if file.FileKey == "" {
		return nil, nil
	}

	presignedURL, err := s.svc.fileManager.GenerateFileUrl(ctx, file, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate file URL: %w", err)
	}

	return &presignedURL, nil
}
