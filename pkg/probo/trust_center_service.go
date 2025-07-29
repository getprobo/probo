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
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"go.gearno.de/kit/pg"
)

type (
	TrustCenterService struct {
		svc *TenantService
	}

	UpdateTrustCenterRequest struct {
		ID     gid.GID
		Active *bool
		Slug   *string
	}
)

func (s TrustCenterService) Get(
	ctx context.Context,
	trustCenterID gid.GID,
) (*coredata.TrustCenter, error) {
	trustCenter := &coredata.TrustCenter{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return trustCenter.LoadByID(ctx, conn, s.svc.scope, trustCenterID)
		},
	)

	if err != nil {
		return nil, err
	}

	return trustCenter, nil
}

func (s TrustCenterService) GetByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (*coredata.TrustCenter, error) {
	trustCenter := &coredata.TrustCenter{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return trustCenter.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID)
		},
	)

	if err != nil {
		return nil, err
	}

	return trustCenter, nil
}

func (s *TrustCenterService) Update(
	ctx context.Context,
	req *UpdateTrustCenterRequest,
) (*coredata.TrustCenter, error) {
	trustCenter := &coredata.TrustCenter{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := trustCenter.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			if req.Active != nil {
				trustCenter.Active = *req.Active
			}
			if req.Slug != nil {
				trustCenter.Slug = *req.Slug
			}

			trustCenter.UpdatedAt = time.Now()

			if err := trustCenter.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update trust center: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return trustCenter, nil
}
