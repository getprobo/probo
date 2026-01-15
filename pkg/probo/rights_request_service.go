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

type RightsRequestService struct {
	svc *TenantService
}

type (
	CreateRightsRequestRequest struct {
		OrganizationID gid.GID
		RequestType    *coredata.RightsRequestType
		RequestState   *coredata.RightsRequestState
		DataSubject    *string
		Contact        *string
		Details        *string
		Deadline       *time.Time
		ActionTaken    *string
	}

	UpdateRightsRequestRequest struct {
		ID           gid.GID
		RequestType  *coredata.RightsRequestType
		RequestState *coredata.RightsRequestState
		DataSubject  **string
		Contact      **string
		Details      **string
		Deadline     **time.Time
		ActionTaken  **string
	}
)

func (crrr *CreateRightsRequestRequest) Validate() error {
	v := validator.New()

	v.Check(crrr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(crrr.RequestType, "request_type", validator.Required(), validator.OneOfSlice(coredata.RightsRequestTypes()))
	v.Check(crrr.RequestState, "request_state", validator.Required(), validator.OneOfSlice(coredata.RightsRequestStates()))
	v.Check(crrr.DataSubject, "data_subject", validator.SafeText(ContentMaxLength))
	v.Check(crrr.Contact, "contact", validator.SafeText(ContentMaxLength))
	v.Check(crrr.Details, "details", validator.SafeText(ContentMaxLength))
	v.Check(crrr.ActionTaken, "action_taken", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (urrr *UpdateRightsRequestRequest) Validate() error {
	v := validator.New()

	v.Check(urrr.ID, "id", validator.Required(), validator.GID(coredata.RightsRequestEntityType))
	v.Check(urrr.RequestType, "request_type", validator.OneOfSlice(coredata.RightsRequestTypes()))
	v.Check(urrr.RequestState, "request_state", validator.OneOfSlice(coredata.RightsRequestStates()))
	v.Check(urrr.DataSubject, "data_subject", validator.SafeText(ContentMaxLength))
	v.Check(urrr.Contact, "contact", validator.SafeText(ContentMaxLength))
	v.Check(urrr.Details, "details", validator.SafeText(ContentMaxLength))
	v.Check(urrr.ActionTaken, "action_taken", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s RightsRequestService) Get(
	ctx context.Context,
	rightsRequestID gid.GID,
) (*coredata.RightsRequest, error) {
	request := &coredata.RightsRequest{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := request.LoadByID(ctx, conn, s.svc.scope, rightsRequestID); err != nil {
				return fmt.Errorf("cannot load rights request: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return request, nil
}

func (s *RightsRequestService) Create(
	ctx context.Context,
	req *CreateRightsRequestRequest,
) (*coredata.RightsRequest, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	request := &coredata.RightsRequest{
		ID:             gid.New(s.svc.scope.GetTenantID(), coredata.RightsRequestEntityType),
		OrganizationID: req.OrganizationID,
		RequestType:    *req.RequestType,
		RequestState:   *req.RequestState,
		DataSubject:    req.DataSubject,
		Contact:        req.Contact,
		Details:        req.Details,
		Deadline:       req.Deadline,
		ActionTaken:    req.ActionTaken,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := request.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert rights request: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return request, nil
}

func (s *RightsRequestService) Update(
	ctx context.Context,
	req *UpdateRightsRequestRequest,
) (*coredata.RightsRequest, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	request := &coredata.RightsRequest{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := request.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load rights request: %w", err)
			}

			if req.RequestType != nil {
				request.RequestType = *req.RequestType
			}

			if req.RequestState != nil {
				request.RequestState = *req.RequestState
			}

			if req.DataSubject != nil {
				request.DataSubject = *req.DataSubject
			}

			if req.Contact != nil {
				request.Contact = *req.Contact
			}

			if req.Details != nil {
				request.Details = *req.Details
			}

			if req.Deadline != nil {
				request.Deadline = *req.Deadline
			}

			if req.ActionTaken != nil {
				request.ActionTaken = *req.ActionTaken
			}

			request.UpdatedAt = time.Now()

			if err := request.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update rights request: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return request, nil
}

func (s *RightsRequestService) Delete(
	ctx context.Context,
	rightsRequestID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			request := &coredata.RightsRequest{}
			if err := request.LoadByID(ctx, conn, s.svc.scope, rightsRequestID); err != nil {
				return fmt.Errorf("cannot load rights request: %w", err)
			}

			if err := request.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete rights request: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s RightsRequestService) CountByOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			requests := coredata.RightsRequests{}
			count, err = requests.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count rights requests: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s RightsRequestService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.RightsRequestOrderField],
) (*page.Page[*coredata.RightsRequest, coredata.RightsRequestOrderField], error) {
	var requests coredata.RightsRequests

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := requests.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load rights requests: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(requests, cursor), nil
}
