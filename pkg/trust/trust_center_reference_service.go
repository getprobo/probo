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

	file, err := s.svc.fileManager.GetPublicFile(ctx, reference.LogoFileID)
	if err != nil {
		return "", err
	}

	return s.svc.fileManager.GenerateFileURL(file), nil
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
