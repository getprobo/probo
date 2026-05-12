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

package probo

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	ComplianceExternalURLService struct {
		svc *TenantService
	}

	CreateComplianceExternalURLRequest struct {
		TrustCenterID gid.GID
		Name          string
		URL           string
	}

	UpdateComplianceExternalURLRequest struct {
		ID   gid.GID
		Name string
		URL  string
		Rank *int
	}

	DeleteComplianceExternalURLRequest struct {
		ID gid.GID
	}
)

func (r *CreateComplianceExternalURLRequest) Validate() error {
	v := validator.New()
	v.Check(r.TrustCenterID, "trust_center_id", validator.Required(), validator.GID(coredata.TrustCenterEntityType))
	v.Check(r.URL, "url", validator.Required(), validator.URL())
	return v.Error()
}

func (r *UpdateComplianceExternalURLRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.ComplianceExternalURLEntityType))
	v.Check(r.URL, "url", validator.Required(), validator.URL())
	v.Check(r.Rank, "rank", validator.Min(1))
	return v.Error()
}

func (r *DeleteComplianceExternalURLRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.ComplianceExternalURLEntityType))
	return v.Error()
}

func (s ComplianceExternalURLService) List(
	ctx context.Context,
	trustCenterID gid.GID,
	cursor *page.Cursor[coredata.ComplianceExternalURLOrderField],
) (*page.Page[*coredata.ComplianceExternalURL, coredata.ComplianceExternalURLOrderField], error) {
	var items coredata.ComplianceExternalURLs

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := items.LoadByTrustCenterID(ctx, conn, s.svc.scope, trustCenterID, cursor); err != nil {
				return fmt.Errorf("cannot load compliance external URLs: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(items, cursor), nil
}

func (s ComplianceExternalURLService) Create(
	ctx context.Context,
	req *CreateComplianceExternalURLRequest,
) (*coredata.ComplianceExternalURL, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	id := gid.New(s.svc.scope.GetTenantID(), coredata.ComplianceExternalURLEntityType)

	var item *coredata.ComplianceExternalURL

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			trustCenter := &coredata.TrustCenter{}
			if err := trustCenter.LoadByID(ctx, tx, s.svc.scope, req.TrustCenterID); err != nil {
				return fmt.Errorf("cannot load trust center: %w", err)
			}

			item = &coredata.ComplianceExternalURL{
				ID:             id,
				OrganizationID: trustCenter.OrganizationID,
				TrustCenterID:  req.TrustCenterID,
				Name:           req.Name,
				URL:            req.URL,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := item.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert compliance external URL: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s ComplianceExternalURLService) Update(
	ctx context.Context,
	req *UpdateComplianceExternalURLRequest,
) (*coredata.ComplianceExternalURL, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item *coredata.ComplianceExternalURL

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			item = &coredata.ComplianceExternalURL{}

			if err := item.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance external URL: %w", err)
			}

			item.Name = req.Name
			item.URL = req.URL
			item.UpdatedAt = time.Now()

			if req.Rank != nil {
				item.Rank = *req.Rank
				if err := item.UpdateRank(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot update compliance external URL rank: %w", err)
				}
			}

			if err := item.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update compliance external URL: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s ComplianceExternalURLService) Delete(
	ctx context.Context,
	req *DeleteComplianceExternalURLRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			item := &coredata.ComplianceExternalURL{}

			if err := item.LoadByID(ctx, tx, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load compliance external URL: %w", err)
			}

			if err := item.Delete(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete compliance external URL: %w", err)
			}

			return nil
		},
	)
}
