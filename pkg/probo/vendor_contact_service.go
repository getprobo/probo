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
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"go.gearno.de/kit/pg"
)

type (
	VendorContactService struct {
		svc *TenantService
	}

	CreateVendorContactRequest struct {
		VendorID gid.GID
		Name     *string
		Email    *string
		Phone    *string
		Role     *string
	}

	UpdateVendorContactRequest struct {
		ID    gid.GID
		Name  *string
		Email *string
		Phone *string
		Role  *string
	}
)

func (s VendorContactService) Get(
	ctx context.Context,
	vendorContactID gid.GID,
) (*coredata.VendorContact, error) {
	vendorContact := &coredata.VendorContact{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendorContact.LoadByID(ctx, conn, s.svc.scope, vendorContactID)
		},
	)

	if err != nil {
		return nil, err
	}

	return vendorContact, nil
}

func (s VendorContactService) List(
	ctx context.Context,
	vendorID gid.GID,
	cursor *page.Cursor[coredata.VendorContactOrderField],
) (*page.Page[*coredata.VendorContact, coredata.VendorContactOrderField], error) {
	var vendorContacts coredata.VendorContacts

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendorContacts.LoadByVendorID(ctx, conn, s.svc.scope, vendorID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(vendorContacts, cursor), nil
}

func (s VendorContactService) Create(
	ctx context.Context,
	req CreateVendorContactRequest,
) (*coredata.VendorContact, error) {
	now := time.Now()
	vendorContact := &coredata.VendorContact{
		ID:        gid.New(s.svc.scope.GetTenantID(), coredata.VendorContactEntityType),
		VendorID:  req.VendorID,
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			return vendorContact.Insert(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, err
	}

	return vendorContact, nil
}

func (s VendorContactService) Update(
	ctx context.Context,
	req UpdateVendorContactRequest,
) (*coredata.VendorContact, error) {
	vendorContact := &coredata.VendorContact{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			err := vendorContact.LoadByID(ctx, conn, s.svc.scope, req.ID)
			if err != nil {
				return err
			}

			if req.Name != nil {
				vendorContact.Name = req.Name
			}
			if req.Email != nil {
				vendorContact.Email = req.Email
			}
			if req.Phone != nil {
				vendorContact.Phone = req.Phone
			}
			if req.Role != nil {
				vendorContact.Role = req.Role
			}
			vendorContact.UpdatedAt = time.Now()

			return vendorContact.Update(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, err
	}

	return vendorContact, nil
}

func (s VendorContactService) Delete(
	ctx context.Context,
	vendorContactID gid.GID,
) error {
	vendorContact := coredata.VendorContact{ID: vendorContactID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return vendorContact.Delete(ctx, conn, s.svc.scope)
		},
	)
}
