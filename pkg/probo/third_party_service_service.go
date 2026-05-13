// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	ThirdPartyServiceService struct {
		svc *TenantService
	}

	CreateThirdPartyServiceRequest struct {
		ThirdPartyID gid.GID
		Name         string
		Description  *string
	}

	UpdateThirdPartyServiceRequest struct {
		ID          gid.GID
		Name        *string
		Description **string
	}
)

func (cvsr *CreateThirdPartyServiceRequest) Validate() error {
	v := validator.New()

	v.Check(cvsr.ThirdPartyID, "third_party_id", validator.Required(), validator.GID(coredata.ThirdPartyEntityType))
	v.Check(cvsr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cvsr.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (uvsr *UpdateThirdPartyServiceRequest) Validate() error {
	v := validator.New()

	v.Check(uvsr.ID, "id", validator.Required(), validator.GID(coredata.ThirdPartyServiceEntityType))
	v.Check(uvsr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uvsr.Description, "description", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s ThirdPartyServiceService) Get(
	ctx context.Context,
	thirdPartyServiceID gid.GID,
) (*coredata.ThirdPartyService, error) {
	thirdPartyService := &coredata.ThirdPartyService{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := thirdPartyService.LoadByID(ctx, conn, s.svc.scope, thirdPartyServiceID)
			if err != nil {
				return fmt.Errorf("cannot load thirdParty service: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdPartyService, nil
}

func (s ThirdPartyServiceService) List(
	ctx context.Context,
	thirdPartyID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyServiceOrderField],
) (*page.Page[*coredata.ThirdPartyService, coredata.ThirdPartyServiceOrderField], error) {
	var thirdPartyServices coredata.ThirdPartyServices

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := thirdPartyServices.LoadByThirdPartyID(ctx, conn, s.svc.scope, thirdPartyID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load thirdParty services: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdPartyServices, cursor), nil
}

func (s ThirdPartyServiceService) Create(
	ctx context.Context,
	req CreateThirdPartyServiceRequest,
) (*coredata.ThirdPartyService, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	thirdPartyService := &coredata.ThirdPartyService{
		ID:           gid.New(s.svc.scope.GetTenantID(), coredata.ThirdPartyServiceEntityType),
		ThirdPartyID: req.ThirdPartyID,
		Name:         req.Name,
		Description:  req.Description,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			thirdParty := &coredata.ThirdParty{}
			if err := thirdParty.LoadByID(ctx, conn, s.svc.scope, req.ThirdPartyID); err != nil {
				return fmt.Errorf("cannot load thirdParty: %w", err)
			}

			thirdPartyService.OrganizationID = thirdParty.OrganizationID

			if err := thirdPartyService.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert thirdParty service: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdPartyService, nil
}

func (s ThirdPartyServiceService) Update(
	ctx context.Context,
	req UpdateThirdPartyServiceRequest,
) (*coredata.ThirdPartyService, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	thirdPartyService := &coredata.ThirdPartyService{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			err := thirdPartyService.LoadByID(ctx, conn, s.svc.scope, req.ID)
			if err != nil {
				return fmt.Errorf("cannot load thirdParty service: %w", err)
			}

			if req.Name != nil {
				thirdPartyService.Name = *req.Name
			}
			if req.Description != nil {
				thirdPartyService.Description = *req.Description
			}
			thirdPartyService.UpdatedAt = time.Now()

			if err := thirdPartyService.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update thirdParty service: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdPartyService, nil
}

func (s ThirdPartyServiceService) Delete(
	ctx context.Context,
	thirdPartyServiceID gid.GID,
) error {
	thirdPartyService := coredata.ThirdPartyService{ID: thirdPartyServiceID}
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := thirdPartyService.LoadByID(ctx, conn, s.svc.scope, thirdPartyServiceID); err != nil {
				return fmt.Errorf("cannot load thirdParty service: %w", err)
			}

			if err := thirdPartyService.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete thirdParty service: %w", err)
			}

			return nil
		},
	)
}
