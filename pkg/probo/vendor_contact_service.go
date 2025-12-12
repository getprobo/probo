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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	VendorContactService struct {
		svc *TenantService
	}

	CreateVendorContactRequest struct {
		VendorID gid.GID
		FullName *string
		Email    *mail.Addr
		Phone    *string
		Role     *string
	}

	UpdateVendorContactRequest struct {
		ID       gid.GID
		FullName **string
		Email    **mail.Addr
		Phone    **string
		Role     **string
	}
)

func (cvcr *CreateVendorContactRequest) Validate() error {
	v := validator.New()

	v.Check(cvcr.VendorID, "vendor_id", validator.Required(), validator.GID(coredata.VendorEntityType))
	v.Check(cvcr.FullName, "fullName", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cvcr.Phone, "phone", validator.SafeText(NameMaxLength))
	v.Check(cvcr.Role, "role", validator.SafeText(TitleMaxLength))

	return v.Error()
}

func (uvcr *UpdateVendorContactRequest) Validate() error {
	v := validator.New()

	v.Check(uvcr.ID, "id", validator.Required(), validator.GID(coredata.VendorContactEntityType))
	v.Check(uvcr.FullName, "fullName", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uvcr.Phone, "phone", validator.SafeText(NameMaxLength))
	v.Check(uvcr.Role, "role", validator.SafeText(TitleMaxLength))

	return v.Error()
}

func (s VendorContactService) Get(
	ctx context.Context,
	vendorContactID gid.GID,
) (*coredata.VendorContact, error) {
	vendorContact := &coredata.VendorContact{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := vendorContact.LoadByID(ctx, conn, s.svc.scope, vendorContactID)
			if err != nil {
				return fmt.Errorf("cannot load vendor contact: %w", err)
			}

			return nil
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
			err := vendorContacts.LoadByVendorID(ctx, conn, s.svc.scope, vendorID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load vendor contacts: %w", err)
			}

			return nil
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
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	vendorContact := &coredata.VendorContact{
		ID:        gid.New(s.svc.scope.GetTenantID(), coredata.VendorContactEntityType),
		VendorID:  req.VendorID,
		FullName:  req.FullName,
		Email:     req.Email,
		Phone:     req.Phone,
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			vendor := &coredata.Vendor{}
			if err := vendor.LoadByID(ctx, conn, s.svc.scope, req.VendorID); err != nil {
				return fmt.Errorf("cannot load vendor: %w", err)
			}

			vendorContact.OrganizationID = vendor.OrganizationID

			if err := vendorContact.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert vendor contact: %w", err)
			}

			return nil
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
	if err := req.Validate(); err != nil {
		return nil, err
	}

	vendorContact := &coredata.VendorContact{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			err := vendorContact.LoadByID(ctx, conn, s.svc.scope, req.ID)
			if err != nil {
				return fmt.Errorf("cannot load vendor contact: %w", err)
			}

			if req.FullName != nil {
				vendorContact.FullName = *req.FullName
			}
			if req.Email != nil {
				vendorContact.Email = *req.Email
			}
			if req.Phone != nil {
				vendorContact.Phone = *req.Phone
			}
			if req.Role != nil {
				vendorContact.Role = *req.Role
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
	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := vendorContact.LoadByID(ctx, conn, s.svc.scope, vendorContactID); err != nil {
				return fmt.Errorf("cannot load vendor contact: %w", err)
			}

			if err := vendorContact.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete vendor contact: %w", err)
			}

			return nil
		},
	)
}
