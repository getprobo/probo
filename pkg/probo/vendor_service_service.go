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
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	VendorServiceService struct {
		svc *TenantService
	}

	CreateVendorServiceRequest struct {
		VendorID    gid.GID
		Name        string
		Description *string
	}

	UpdateVendorServiceRequest struct {
		ID          gid.GID
		Name        *string
		Description **string
	}
)

func (cvsr *CreateVendorServiceRequest) Validate() error {
	v := validator.New()

	v.Check(cvsr.VendorID, "vendor_id", validator.Required(), validator.GID(coredata.VendorEntityType))
	v.Check(cvsr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cvsr.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (uvsr *UpdateVendorServiceRequest) Validate() error {
	v := validator.New()

	v.Check(uvsr.ID, "id", validator.Required(), validator.GID(coredata.VendorServiceEntityType))
	v.Check(uvsr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uvsr.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s VendorServiceService) Get(
	ctx context.Context,
	vendorServiceID gid.GID,
) (*coredata.VendorService, error) {
	vendorService := &coredata.VendorService{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := vendorService.LoadByID(ctx, conn, s.svc.scope, vendorServiceID)
			if err != nil {
				return fmt.Errorf("cannot load vendor service: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return vendorService, nil
}

func (s VendorServiceService) List(
	ctx context.Context,
	vendorID gid.GID,
	cursor *page.Cursor[coredata.VendorServiceOrderField],
) (*page.Page[*coredata.VendorService, coredata.VendorServiceOrderField], error) {
	var vendorServices coredata.VendorServices

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := vendorServices.LoadByVendorID(ctx, conn, s.svc.scope, vendorID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load vendor services: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(vendorServices, cursor), nil
}

func (s VendorServiceService) Create(
	ctx context.Context,
	req CreateVendorServiceRequest,
) (*coredata.VendorService, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	vendorService := &coredata.VendorService{
		ID:          gid.New(s.svc.scope.GetTenantID(), coredata.VendorServiceEntityType),
		VendorID:    req.VendorID,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			vendor := &coredata.Vendor{}
			if err := vendor.LoadByID(ctx, conn, s.svc.scope, req.VendorID); err != nil {
				return fmt.Errorf("cannot load vendor: %w", err)
			}

			vendorService.OrganizationID = vendor.OrganizationID

			if err := vendorService.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert vendor service: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return vendorService, nil
}

func (s VendorServiceService) Update(
	ctx context.Context,
	req UpdateVendorServiceRequest,
) (*coredata.VendorService, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	vendorService := &coredata.VendorService{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			err := vendorService.LoadByID(ctx, conn, s.svc.scope, req.ID)
			if err != nil {
				return fmt.Errorf("cannot load vendor service: %w", err)
			}

			if req.Name != nil {
				vendorService.Name = *req.Name
			}
			if req.Description != nil {
				vendorService.Description = *req.Description
			}
			vendorService.UpdatedAt = time.Now()

			if err := vendorService.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update vendor service: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return vendorService, nil
}

func (s VendorServiceService) Delete(
	ctx context.Context,
	vendorServiceID gid.GID,
) error {
	vendorService := coredata.VendorService{ID: vendorServiceID}
	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := vendorService.LoadByID(ctx, conn, s.svc.scope, vendorServiceID); err != nil {
				return fmt.Errorf("cannot load vendor service: %w", err)
			}

			if err := vendorService.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete vendor service: %w", err)
			}

			return nil
		},
	)
}
