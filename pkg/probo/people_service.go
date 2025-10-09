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

type (
	PeopleService struct {
		svc *TenantService
	}

	UpdatePeopleRequest struct {
		ID                       gid.GID
		Kind                     *coredata.PeopleKind
		FullName                 *string
		PrimaryEmailAddress      *string
		AdditionalEmailAddresses *[]string
		Position                 **string
		ContractStartDate        **time.Time
		ContractEndDate          **time.Time
	}

	CreatePeopleRequest struct {
		OrganizationID           gid.GID
		FullName                 string
		PrimaryEmailAddress      string
		AdditionalEmailAddresses []string
		Kind                     coredata.PeopleKind
		Position                 *string
		ContractStartDate        *time.Time
		ContractEndDate          *time.Time
	}
)

func (s PeopleService) Get(
	ctx context.Context,
	peopleID gid.GID,
) (*coredata.People, error) {
	people := &coredata.People{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return people.LoadByID(ctx, conn, s.svc.scope, peopleID)
		},
	)

	if err != nil {
		return nil, err
	}

	return people, nil
}

func (s PeopleService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.PeopleFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			peoples := coredata.Peoples{}
			count, err = peoples.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count peoples: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s PeopleService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.PeopleOrderField],
	filter *coredata.PeopleFilter,
) (*page.Page[*coredata.People, coredata.PeopleOrderField], error) {
	var peoples coredata.Peoples

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return peoples.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
				cursor,
				filter,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(peoples, cursor), nil
}

func (s PeopleService) Update(
	ctx context.Context,
	req UpdatePeopleRequest,
) (*coredata.People, error) {
	people := &coredata.People{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := people.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load people: %w", err)
			}

			if req.Kind != nil {
				people.Kind = *req.Kind
			}

			if req.FullName != nil {
				people.FullName = *req.FullName
			}

			if req.PrimaryEmailAddress != nil {
				people.PrimaryEmailAddress = *req.PrimaryEmailAddress
			}

			if req.AdditionalEmailAddresses != nil {
				people.AdditionalEmailAddresses = *req.AdditionalEmailAddresses
			}

			if req.Position != nil {
				people.Position = *req.Position
			}

			if req.ContractStartDate != nil {
				people.ContractStartDate = *req.ContractStartDate
			}

			if req.ContractEndDate != nil {
				people.ContractEndDate = *req.ContractEndDate
			}

			if people.ContractStartDate != nil && people.ContractEndDate != nil {
				if people.ContractEndDate.Before(*people.ContractStartDate) {
					return fmt.Errorf("contract end date must be after or equal to start date")
				}
			}

			people.UpdatedAt = time.Now()

			return people.Update(ctx, conn, s.svc.scope)
		})
	if err != nil {
		return nil, err
	}

	return people, nil
}

func (s PeopleService) Create(
	ctx context.Context,
	req CreatePeopleRequest,
) (*coredata.People, error) {
	if req.ContractStartDate != nil && req.ContractEndDate != nil {
		if req.ContractEndDate.Before(*req.ContractStartDate) {
			return nil, fmt.Errorf("contract end date must be after or equal to start date")
		}
	}

	now := time.Now()
	peopleID := gid.New(s.svc.scope.GetTenantID(), coredata.PeopleEntityType)

	organization := &coredata.Organization{}
	people := &coredata.People{
		ID:                       peopleID,
		OrganizationID:           req.OrganizationID,
		Kind:                     req.Kind,
		FullName:                 req.FullName,
		PrimaryEmailAddress:      req.PrimaryEmailAddress,
		AdditionalEmailAddresses: req.AdditionalEmailAddresses,
		Position:                 req.Position,
		ContractStartDate:        req.ContractStartDate,
		ContractEndDate:          req.ContractEndDate,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization %q: %w", req.OrganizationID, err)
			}

			if err := people.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert people: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return people, nil
}

func (s PeopleService) Delete(
	ctx context.Context,
	peopleID gid.GID,
) error {
	people := coredata.People{ID: peopleID}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return people.Delete(ctx, conn, s.svc.scope)
		},
	)
}
