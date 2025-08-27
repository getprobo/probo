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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/getprobo/probo/pkg/usrmgr"
	"go.gearno.de/kit/pg"
)

type (
	TrustCenterAccessService struct {
		svc    *TenantService
		usrmgr *usrmgr.Service
	}

	CreateTrustCenterAccessRequest struct {
		TrustCenterID gid.GID
		Email         string
		Name          string
	}
)

const (
	TokenTypeTrustCenterAccess = "trust_center_access"
)

func (s TrustCenterAccessService) ValidateToken(
	ctx context.Context,
	tokenString string,
) (*probo.TrustCenterAccessData, error) {
	token, err := statelesstoken.ValidateToken[probo.TrustCenterAccessData](
		s.svc.tokenSecret,
		TokenTypeTrustCenterAccess,
		tokenString,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot validate trust center access token: %w", err)
	}

	access := &coredata.TrustCenterAccess{}
	err = s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		err := access.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, token.Data.TrustCenterID, token.Data.Email)
		if err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		if !access.Active {
			return fmt.Errorf("trust center access is not active")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &token.Data, nil
}

func (s TrustCenterAccessService) Create(
	ctx context.Context,
	req *CreateTrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	if !strings.Contains(req.Email, "@") {
		return nil, fmt.Errorf("invalid email address")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	now := time.Now()

	var access *coredata.TrustCenterAccess

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		existingAccess := &coredata.TrustCenterAccess{}
		err := existingAccess.LoadByTrustCenterIDAndEmail(ctx, tx, s.svc.scope, req.TrustCenterID, req.Email)

		if err == nil {
			if existingAccess.Active {
				return fmt.Errorf("active trust center access already exists for this email")
			}
			if err := existingAccess.Delete(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete existing trust center access: %w", err)
			}
		} else {
			var notFoundErr *coredata.ErrTrustCenterAccessNotFound
			if !errors.As(err, &notFoundErr) {
				return fmt.Errorf("cannot load trust center access: %w", err)
			}
		}

		access = &coredata.TrustCenterAccess{
			ID:            gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterAccessEntityType),
			TenantID:      s.svc.scope.GetTenantID(),
			TrustCenterID: req.TrustCenterID,
			Email:         req.Email,
			Name:          req.Name,
			Active:        false,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := access.Insert(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert trust center access: %w", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return access, nil
}
