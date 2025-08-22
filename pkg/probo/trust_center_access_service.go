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
	"errors"
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
		Active        bool
	}

	UpdateTrustCenterAccessRequest struct {
		ID     gid.GID
		Name   *string
		Active *bool
	}

	DeleteTrustCenterAccessRequest struct {
		ID gid.GID
	}

	TrustCenterAccessData struct {
		TrustCenterID gid.GID `json:"trust_center_id"`
		Email         string  `json:"email"`
	}
)

const (
	trustCenterAccessEmailSubject  = "Trust Center Access Invitation - %s"
	trustCenterAccessEmailTemplate = `
	You have been granted access to %s's Trust Center!

	Click the link below to access it:

	[1] %s

	This link will expire in 7 days.

	If the link above doesn't work, copy and paste the entire URL into your browser.
	`
)

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
		s.svc.trustConfig.TokenSecret,
		s.svc.trustConfig.TokenType,
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
			Active:        req.Active,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := access.Insert(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert trust center access: %w", err)
		}

		if req.Active {
			if err := s.sendAccessEmail(ctx, tx, access); err != nil {
				return fmt.Errorf("failed to send access email: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return access, nil
}

func (s TrustCenterAccessService) Update(
	ctx context.Context,
	req *UpdateTrustCenterAccessRequest,
) (*coredata.TrustCenterAccess, error) {
	now := time.Now()

	var access *coredata.TrustCenterAccess

	if req.Name != nil && *req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		access = &coredata.TrustCenterAccess{}

		if err := access.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		shouldSendEmail := req.Active != nil && *req.Active && !access.Active
		if req.Name != nil {
			access.Name = *req.Name
		}
		if req.Active != nil {
			access.Active = *req.Active
		}
		access.UpdatedAt = now

		if err := access.Update(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot update trust center access: %w", err)
		}

		if shouldSendEmail {
			if err := s.sendAccessEmail(ctx, tx, access); err != nil {
				return fmt.Errorf("failed to send access email: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return access, nil
}

func (s TrustCenterAccessService) Delete(
	ctx context.Context,
	req *DeleteTrustCenterAccessRequest,
) error {
	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		access := &coredata.TrustCenterAccess{}

		if err := access.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
			return fmt.Errorf("cannot load trust center access: %w", err)
		}

		if err := access.Delete(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot delete trust center access: %w", err)
		}

		return nil
	})

	return err
}

func (s TrustCenterAccessService) sendAccessEmail(ctx context.Context, tx pg.Conn, access *coredata.TrustCenterAccess) error {
	accessToken, err := statelesstoken.NewToken(
		s.svc.trustConfig.TokenSecret,
		s.svc.trustConfig.TokenType,
		s.svc.trustConfig.TokenDuration,
		TrustCenterAccessData{
			TrustCenterID: access.TrustCenterID,
			Email:         access.Email,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot generate access token: %w", err)
	}

	trustCenter := &coredata.TrustCenter{}
	err = trustCenter.LoadByID(ctx, tx, s.svc.scope, access.TrustCenterID)
	if err != nil {
		return fmt.Errorf("cannot load trust center: %w", err)
	}

	organization := &coredata.Organization{}
	err = organization.LoadByID(ctx, tx, s.svc.scope, trustCenter.OrganizationID)
	if err != nil {
		return fmt.Errorf("cannot load organization: %w", err)
	}

	accessURL := url.URL{
		Scheme: "https",
		Host:   s.svc.hostname,
		Path:   "/trust/" + trustCenter.Slug + "/access",
		RawQuery: url.Values{
			"token": []string{accessToken},
		}.Encode(),
	}

	return s.sendTrustCenterAccessEmail(ctx, tx, access.Name, access.Email, organization.Name, accessURL.String())
}

func (s TrustCenterAccessService) sendTrustCenterAccessEmail(
	ctx context.Context,
	tx pg.Conn,
	name string,
	email string,
	companyName string,
	accessURL string,
) error {
	accessEmail := coredata.NewEmail(
		name,
		email,
		fmt.Sprintf(trustCenterAccessEmailSubject, companyName),
		fmt.Sprintf(trustCenterAccessEmailTemplate, companyName, accessURL),
	)

	if err := accessEmail.Insert(ctx, tx); err != nil {
		return fmt.Errorf("cannot insert access email: %w", err)
	}
	return nil
}
