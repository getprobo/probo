// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package management

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
	CreateCustomLinkRequest struct {
		CompliancePortalID gid.GID
		Name               string
		URL                string
	}

	UpdateCustomLinkRequest struct {
		ID   gid.GID
		Name string
		URL  string
		Rank *int
	}

	DeleteCustomLinkRequest struct {
		ID gid.GID
	}
)

func (r *CreateCustomLinkRequest) Validate() error {
	v := validator.New()
	v.Check(r.CompliancePortalID, "trust_center_id", validator.Required(), validator.GID(coredata.CompliancePortalEntityType))
	v.Check(r.URL, "url", validator.Required(), validator.URL())

	return v.Error()
}

func (r *UpdateCustomLinkRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.ComplianceCustomLinkEntityType))
	v.Check(r.URL, "url", validator.Required(), validator.URL())
	v.Check(r.Rank, "rank", validator.Min(1))

	return v.Error()
}

func (r *DeleteCustomLinkRequest) Validate() error {
	v := validator.New()
	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.ComplianceCustomLinkEntityType))

	return v.Error()
}

func (s *Service) ListCustomLinks(
	ctx context.Context,
	scope coredata.Scoper,
	compliancePageID gid.GID,
	cursor *page.Cursor[coredata.ComplianceCustomLinkOrderField],
) (*page.Page[*coredata.ComplianceCustomLink, coredata.ComplianceCustomLinkOrderField], error) {
	var items coredata.ComplianceCustomLinks

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := items.LoadByCompliancePortalID(ctx, conn, scope, compliancePageID, cursor); err != nil {
				return fmt.Errorf("cannot load custom links: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(items, cursor), nil
}

func (s *Service) CreateCustomLink(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateCustomLinkRequest,
) (*coredata.ComplianceCustomLink, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	id := gid.New(scope.GetTenantID(), coredata.ComplianceCustomLinkEntityType)

	var item *coredata.ComplianceCustomLink

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			compliancePage := &coredata.CompliancePortal{}
			if err := compliancePage.LoadByID(ctx, tx, scope, req.CompliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance page: %w", err)
			}

			item = &coredata.ComplianceCustomLink{
				ID:                 id,
				OrganizationID:     compliancePage.OrganizationID,
				CompliancePortalID: req.CompliancePortalID,
				Name:               req.Name,
				URL:                req.URL,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if err := item.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert custom link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) UpdateCustomLink(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateCustomLinkRequest,
) (*coredata.ComplianceCustomLink, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var item *coredata.ComplianceCustomLink

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			item = &coredata.ComplianceCustomLink{}

			if err := item.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load custom link: %w", err)
			}

			item.Name = req.Name
			item.URL = req.URL
			item.UpdatedAt = time.Now()

			if req.Rank != nil {
				item.Rank = *req.Rank
				if err := item.UpdateRank(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update custom link rank: %w", err)
				}
			}

			if err := item.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update custom link: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *Service) DeleteCustomLink(
	ctx context.Context,
	scope coredata.Scoper,
	req *DeleteCustomLinkRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			item := &coredata.ComplianceCustomLink{}

			if err := item.LoadByID(ctx, tx, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load custom link: %w", err)
			}

			if err := item.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete custom link: %w", err)
			}

			return nil
		},
	)
}
