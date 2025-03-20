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
		ExpectedVersion          int
		Kind                     *coredata.PeopleKind
		Position                 *string
		FullName                 *string
		PrimaryEmailAddress      *string
		AdditionalEmailAddresses *[]string
	}

	CreatePeopleRequest struct {
		OrganizationID           gid.GID
		FullName                 string
		PrimaryEmailAddress      string
		AdditionalEmailAddresses []string
		Kind                     coredata.PeopleKind
		Position                 string
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

func (s PeopleService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.PeopleOrderField],
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
	params := coredata.UpdatePeopleParams{
		ExpectedVersion:          req.ExpectedVersion,
		Kind:                     req.Kind,
		Position:                 req.Position,
		FullName:                 req.FullName,
		PrimaryEmailAddress:      req.PrimaryEmailAddress,
		AdditionalEmailAddresses: req.AdditionalEmailAddresses,
	}

	people := &coredata.People{ID: req.ID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			return people.Update(ctx, conn, s.svc.scope, params)
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
	now := time.Now()
	peopleID, err := gid.NewGID(s.svc.scope.GetTenantID(), coredata.PeopleEntityType)
	if err != nil {
		return nil, fmt.Errorf("cannot create people global id: %w", err)
	}

	organization := &coredata.Organization{}
	people := &coredata.People{
		ID:                       peopleID,
		OrganizationID:           req.OrganizationID,
		Kind:                     req.Kind,
		Position:                 req.Position,
		FullName:                 req.FullName,
		PrimaryEmailAddress:      req.PrimaryEmailAddress,
		AdditionalEmailAddresses: req.AdditionalEmailAddresses,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	err = s.svc.pg.WithTx(
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
