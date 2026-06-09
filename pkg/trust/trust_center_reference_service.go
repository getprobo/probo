// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type TrustCenterReferenceService struct {
	svc *Service
}

func (s TrustCenterReferenceService) ListForTrustCenterID(
	ctx context.Context,
	scope coredata.Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterReferenceOrderField],
) (*page.Page[*coredata.TrustCenterReference, coredata.TrustCenterReferenceOrderField], error) {
	var references coredata.TrustCenterReferences

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		err := references.LoadByTrustCenterID(ctx, conn, scope, trustCenterID, cursor)
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
	scope coredata.Scoper,
	referenceID gid.GID,
) (string, error) {
	reference := &coredata.TrustCenterReference{}

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return reference.LoadByID(ctx, tx, scope, referenceID)
	})
	if err != nil {
		return "", fmt.Errorf("cannot load trust center reference: %w", err)
	}

	return s.svc.file.GenerateFileURL(ctx, reference.LogoFileID)
}

func (s TrustCenterReferenceService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	referenceID gid.GID,
) (*coredata.TrustCenterReference, error) {
	reference := &coredata.TrustCenterReference{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := reference.LoadByID(ctx, conn, scope, referenceID)
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
