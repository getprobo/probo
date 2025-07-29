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
	"net/url"
	"strings"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
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
		SendEmail     bool
	}

	UpdateTrustCenterAccessRequest struct {
		AccessID  gid.GID
		Email     *string
		Name      *string
		Active    *bool
		SendEmail bool
	}

	DeleteTrustCenterAccessRequest struct {
		AccessID gid.GID
	}

	RevokeTrustCenterAccessRequest struct {
		AccessID gid.GID
	}

	TrustCenterAccessData struct {
		TrustCenterID gid.GID `json:"trust_center_id"`
		Email         string  `json:"email"`
	}
)

const (
	TokenTypeTrustCenterAccess = "trust_center_access"
)

func (s *TenantService) getURLScheme() string {
	if s.https {
		return "https"
	}
	return "http"
}

func (s TrustCenterAccessService) RevokeAccess(
	ctx context.Context,
	req *RevokeTrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	access := &coredata.TrustCenterAccess{}

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		if err := access.LoadByID(ctx, tx, s.svc.scope, req.AccessID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		access.Active = false
		access.UpdatedAt = time.Now()

		if err := access.Update(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot update trust center access: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return access, nil
}

func (s TrustCenterAccessService) ListForTrustCenterID(
	ctx context.Context,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.TrustCenterAccessOrderField],
) (*page.Page[*coredata.TrustCenterAccess, coredata.TrustCenterAccessOrderField], error) {
	var accesses coredata.TrustCenterAccesses

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return accesses.LoadByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID, cursor)
	})

	if err != nil {
		return nil, err
	}

	return page.NewPage(accesses, cursor), nil
}

func (s TrustCenterAccessService) ValidateToken(
	ctx context.Context,
	tokenString string,
) (*TrustCenterAccessData, error) {
	token, err := statelesstoken.ValidateToken[TrustCenterAccessData](
		s.svc.tokenSecret,
		TokenTypeTrustCenterAccess,
		tokenString,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot validate trust center access token: %w", err)
	}

	access := &coredata.TrustCenterAccess{}
	err = s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return access.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, token.Data.TrustCenterID, token.Data.Email)
	})

	if err != nil {
		return nil, fmt.Errorf("access not found or revoked: %w", err)
	}

	if !access.Active {
		return nil, fmt.Errorf("access has been revoked")
	}

	return &token.Data, nil
}

func (s TrustCenterAccessService) IsAccessActive(
	ctx context.Context,
	trustCenterID gid.GID,
	email string,
) (bool, error) {
	access := &coredata.TrustCenterAccess{}
	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return access.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, trustCenterID, email)
	})

	if err != nil {
		return false, fmt.Errorf("cannot load trust center access: %w", err)
	}

	return access.Active, nil
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

	existingAccess := &coredata.TrustCenterAccess{}
	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return existingAccess.LoadByTrustCenterIDAndEmail(ctx, conn, s.svc.scope, req.TrustCenterID, req.Email)
	})

	var access *coredata.TrustCenterAccess

	if err == nil {
		access = existingAccess
		access.Name = req.Name
		access.Active = true
		access.UpdatedAt = now

		err = s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
			if err := access.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update trust center access: %w", err)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		access = &coredata.TrustCenterAccess{
			ID:            gid.New(s.svc.scope.GetTenantID(), coredata.TrustCenterAccessEntityType),
			TenantID:      s.svc.scope.GetTenantID(),
			TrustCenterID: req.TrustCenterID,
			Email:         req.Email,
			Name:          req.Name,
			Active:        true,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		err = s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
			if err := access.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert trust center access: %w", err)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	if req.SendEmail {
		if err := s.sendAccessEmail(ctx, access); err != nil {
			fmt.Printf("Failed to send access email\n")
		}
	}

	return access, nil
}

func (s TrustCenterAccessService) Update(
	ctx context.Context,
	req *UpdateTrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	access := &coredata.TrustCenterAccess{}

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		if err := access.LoadByID(ctx, tx, s.svc.scope, req.AccessID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		if req.Email != nil {
			if !strings.Contains(*req.Email, "@") {
				return fmt.Errorf("invalid email address")
			}
			access.Email = *req.Email
		}

		if req.Name != nil {
			if *req.Name == "" {
				return fmt.Errorf("name cannot be empty")
			}
			access.Name = *req.Name
		}

		if req.Active != nil {
			access.Active = *req.Active
		}

		access.UpdatedAt = time.Now()

		if err := access.Update(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot update trust center access: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if req.SendEmail && access.Active {
		if err := s.sendAccessEmail(ctx, access); err != nil {
			fmt.Printf("Failed to send access email\n")
		}
	}

	return access, nil
}

func (s TrustCenterAccessService) Delete(
	ctx context.Context,
	req *DeleteTrustCenterAccessRequest,
) error {
	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		access := &coredata.TrustCenterAccess{}

		if err := access.LoadByID(ctx, tx, s.svc.scope, req.AccessID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		if err := access.Delete(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot delete trust center access: %w", err)
		}

		return nil
	})

	return err
}

func (s TrustCenterAccessService) sendAccessEmail(ctx context.Context, access *coredata.TrustCenterAccess) error {
	accessToken, err := statelesstoken.NewToken(
		s.svc.tokenSecret,
		TokenTypeTrustCenterAccess,
		7*24*time.Hour,
		TrustCenterAccessData{
			TrustCenterID: access.TrustCenterID,
			Email:         access.Email,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot generate access token: %w", err)
	}

	trustCenter := &coredata.TrustCenter{}
	err = s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return trustCenter.LoadByID(ctx, conn, s.svc.scope, access.TrustCenterID)
	})
	if err != nil {
		return fmt.Errorf("cannot load trust center: %w", err)
	}

	organization := &coredata.Organization{}
	err = s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return organization.LoadByID(ctx, conn, s.svc.scope, trustCenter.OrganizationID)
	})
	if err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	accessURL := url.URL{
		Scheme:   s.svc.getURLScheme(),
		Host:     s.svc.hostname,
		Path:     "/trust/" + trustCenter.Slug + "/access",
		RawQuery: "token=" + url.QueryEscape(accessToken),
	}

	return s.usrmgr.SendTrustCenterAccessEmail(ctx, access.Name, access.Email, organization.Name, accessURL.String())
}
