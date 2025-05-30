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
	"github.com/getprobo/probo/pkg/page"
	"go.gearno.de/kit/pg"
)

type DatumService struct {
	svc *TenantService
}

type CreateDatumRequest struct {
	OrganizationID  gid.GID
	Name            string
	DataSensitivity coredata.DataSensitivity
	OwnerID         gid.GID
	VendorIDs       []gid.GID
}

type UpdateDatumRequest struct {
	ID              gid.GID
	Name            *string
	DataSensitivity *coredata.DataSensitivity
	OwnerID         *gid.GID
	VendorIDs       []gid.GID
}

func (s DatumService) Get(
	ctx context.Context,
	datumID gid.GID,
) (*coredata.Data, error) {
	datum := &coredata.Data{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return datum.LoadByID(ctx, conn, s.svc.scope, datumID)
		},
	)

	if err != nil {
		return nil, err
	}

	return datum, nil
}

func (s DatumService) GetByOwnerID(
	ctx context.Context,
	ownerID gid.GID,
) (*coredata.Data, error) {
	datum := &coredata.Data{OwnerID: ownerID}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return datum.LoadByOwnerID(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, err
	}

	return datum, nil
}

func (s DatumService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DatumOrderField],
) (*page.Page[*coredata.Data, coredata.DatumOrderField], error) {
	var data coredata.DataList

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return data.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
				cursor,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(data, cursor), nil
}

func (s DatumService) Update(
	ctx context.Context,
	req UpdateDatumRequest,
) (*coredata.Data, error) {
	now := time.Now()

	existing := &coredata.Data{}
	if err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return existing.LoadByID(ctx, conn, s.svc.scope, req.ID)
	}); err != nil {
		return nil, fmt.Errorf("cannot load data: %w", err)
	}

	datum := &coredata.Data{
		ID:              req.ID,
		OrganizationID:  existing.OrganizationID,
		Name:            existing.Name,
		DataSensitivity: existing.DataSensitivity,
		OwnerID:         existing.OwnerID,
		CreatedAt:       existing.CreatedAt,
		UpdatedAt:       now,
	}

	// Update fields from request
	if req.Name != nil {
		datum.Name = *req.Name
	}
	if req.DataSensitivity != nil {
		datum.DataSensitivity = *req.DataSensitivity
	}
	if req.OwnerID != nil {
		datum.OwnerID = *req.OwnerID
	}

	if err := datum.UpdateWithVendorsTx(ctx, s.svc.pg, s.svc.scope, req.VendorIDs, now); err != nil {
		return nil, err
	}

	return datum, nil
}

func (s DatumService) Create(
	ctx context.Context,
	req CreateDatumRequest,
) (*coredata.Data, error) {
	now := time.Now()
	datumID := gid.New(s.svc.scope.GetTenantID(), coredata.DatumEntityType)

	datum := &coredata.Data{
		ID:              datumID,
		OrganizationID:  req.OrganizationID,
		Name:            req.Name,
		DataSensitivity: req.DataSensitivity,
		OwnerID:         req.OwnerID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			return datum.CreateWithVendors(ctx, conn, s.svc.scope, req.VendorIDs, now)
		},
	)

	if err != nil {
		return nil, err
	}

	return datum, nil
}

func (s DatumService) Delete(
	ctx context.Context,
	datumID gid.GID,
) error {
	datum := &coredata.Data{ID: datumID}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return datum.Delete(ctx, conn, s.svc.scope)
		},
	)
}

func (s DatumService) ListVendors(
	ctx context.Context,
	datumID gid.GID,
	cursor *page.Cursor[coredata.VendorOrderField],
) (*page.Page[*coredata.Vendor, coredata.VendorOrderField], error) {
	var vendors coredata.Vendors

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendors.LoadByDatumID(ctx, conn, s.svc.scope, datumID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(vendors, cursor), nil
}

func (s VendorService) ListForDatumID(
	ctx context.Context,
	datumID gid.GID,
	cursor *page.Cursor[coredata.VendorOrderField],
) (*page.Page[*coredata.Vendor, coredata.VendorOrderField], error) {
	var vendors coredata.Vendors

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendors.LoadByDatumID(
				ctx,
				conn,
				s.svc.scope,
				datumID,
				cursor,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(vendors, cursor), nil
}
