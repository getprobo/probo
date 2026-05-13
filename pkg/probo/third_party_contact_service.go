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
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	ThirdPartyContactService struct {
		svc *TenantService
	}

	CreateThirdPartyContactRequest struct {
		ThirdPartyID gid.GID
		FullName     *string
		Email        *mail.Addr
		Phone        *string
		Role         *string
	}

	UpdateThirdPartyContactRequest struct {
		ID       gid.GID
		FullName **string
		Email    **mail.Addr
		Phone    **string
		Role     **string
	}
)

func (cvcr *CreateThirdPartyContactRequest) Validate() error {
	v := validator.New()

	v.Check(cvcr.ThirdPartyID, "third_party_id", validator.Required(), validator.GID(coredata.ThirdPartyEntityType))
	v.Check(cvcr.FullName, "fullName", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cvcr.Phone, "phone", validator.SafeText(NameMaxLength))
	v.Check(cvcr.Role, "role", validator.SafeText(TitleMaxLength))

	return v.Error()
}

func (uvcr *UpdateThirdPartyContactRequest) Validate() error {
	v := validator.New()

	v.Check(uvcr.ID, "id", validator.Required(), validator.GID(coredata.ThirdPartyContactEntityType))
	v.Check(uvcr.FullName, "fullName", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(uvcr.Phone, "phone", validator.SafeText(NameMaxLength))
	v.Check(uvcr.Role, "role", validator.SafeText(TitleMaxLength))

	return v.Error()
}

func (s ThirdPartyContactService) Get(
	ctx context.Context,
	thirdPartyContactID gid.GID,
) (*coredata.ThirdPartyContact, error) {
	thirdPartyContact := &coredata.ThirdPartyContact{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := thirdPartyContact.LoadByID(ctx, conn, s.svc.scope, thirdPartyContactID)
			if err != nil {
				return fmt.Errorf("cannot load thirdParty contact: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdPartyContact, nil
}

func (s ThirdPartyContactService) List(
	ctx context.Context,
	thirdPartyID gid.GID,
	cursor *page.Cursor[coredata.ThirdPartyContactOrderField],
) (*page.Page[*coredata.ThirdPartyContact, coredata.ThirdPartyContactOrderField], error) {
	var thirdPartyContacts coredata.ThirdPartyContacts

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := thirdPartyContacts.LoadByThirdPartyID(ctx, conn, s.svc.scope, thirdPartyID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load thirdParty contacts: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(thirdPartyContacts, cursor), nil
}

func (s ThirdPartyContactService) Create(
	ctx context.Context,
	req CreateThirdPartyContactRequest,
) (*coredata.ThirdPartyContact, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	thirdPartyContact := &coredata.ThirdPartyContact{
		ID:           gid.New(s.svc.scope.GetTenantID(), coredata.ThirdPartyContactEntityType),
		ThirdPartyID: req.ThirdPartyID,
		FullName:     req.FullName,
		Email:        req.Email,
		Phone:        req.Phone,
		Role:         req.Role,
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

			thirdPartyContact.OrganizationID = thirdParty.OrganizationID

			if err := thirdPartyContact.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert thirdParty contact: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdPartyContact, nil
}

func (s ThirdPartyContactService) Update(
	ctx context.Context,
	req UpdateThirdPartyContactRequest,
) (*coredata.ThirdPartyContact, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	thirdPartyContact := &coredata.ThirdPartyContact{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			err := thirdPartyContact.LoadByID(ctx, conn, s.svc.scope, req.ID)
			if err != nil {
				return fmt.Errorf("cannot load thirdParty contact: %w", err)
			}

			if req.FullName != nil {
				thirdPartyContact.FullName = *req.FullName
			}
			if req.Email != nil {
				thirdPartyContact.Email = *req.Email
			}
			if req.Phone != nil {
				thirdPartyContact.Phone = *req.Phone
			}
			if req.Role != nil {
				thirdPartyContact.Role = *req.Role
			}
			thirdPartyContact.UpdatedAt = time.Now()

			return thirdPartyContact.Update(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, err
	}

	return thirdPartyContact, nil
}

func (s ThirdPartyContactService) Delete(
	ctx context.Context,
	thirdPartyContactID gid.GID,
) error {
	thirdPartyContact := coredata.ThirdPartyContact{ID: thirdPartyContactID}
	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := thirdPartyContact.LoadByID(ctx, conn, s.svc.scope, thirdPartyContactID); err != nil {
				return fmt.Errorf("cannot load thirdParty contact: %w", err)
			}

			if err := thirdPartyContact.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete thirdParty contact: %w", err)
			}

			return nil
		},
	)
}
