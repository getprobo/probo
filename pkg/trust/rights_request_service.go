// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package trust

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/validator"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

// RightsRequestDeadlineDays is the number of days a portal-submitted data
// subject request is given before its response deadline. Thirty days matches
// the GDPR Article 12(3) one-month standard (the shorter of GDPR / CCPA), and
// the console can adjust it afterwards.
const RightsRequestDeadlineDays = 30

type (
	RightsRequestService struct {
		svc *Service
	}

	// CreateRightsRequest is a data subject request submitted from the trust
	// portal. The organization comes from the current compliance page and the
	// contact from the verified viewer's identity, so neither is client-supplied.
	CreateRightsRequest struct {
		OrganizationID gid.GID
		RequestType    coredata.RightsRequestType
		DataSubject    *string
		Contact        string
		Details        *string
	}
)

// Validate bounds the free-text fields with the same rules the console applies,
// so this public portal mutation can't persist oversized or unsafe input.
func (r *CreateRightsRequest) Validate() error {
	v := validator.New()

	v.Check(r.DataSubject, "data_subject", validator.SafeText(probo.ContentMaxLength))
	v.Check(r.Details, "details", validator.SafeText(probo.ContentMaxLength))

	return v.Error()
}

func (s *RightsRequestService) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateRightsRequest,
) (*coredata.RightsRequest, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	deadline := now.AddDate(0, 0, RightsRequestDeadlineDays)

	request := &coredata.RightsRequest{
		ID:             gid.New(scope.GetTenantID(), coredata.RightsRequestEntityType),
		OrganizationID: req.OrganizationID,
		RequestType:    req.RequestType,
		RequestState:   coredata.RightsRequestStateTodo,
		DataSubject:    req.DataSubject,
		Contact:        &req.Contact,
		Details:        req.Details,
		Deadline:       &deadline,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := request.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert rights request: %w", err)
			}

			if err := webhook.InsertData(
				ctx,
				tx,
				scope,
				request.OrganizationID,
				coredata.WebhookEventTypeRightRequestCreated,
				webhooktypes.NewRightsRequest(request),
			); err != nil {
				return fmt.Errorf("cannot insert webhook event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return request, nil
}

func (s RightsRequestService) ListForOrganizationIDAndContact(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	contact string,
	cursor *page.Cursor[coredata.RightsRequestOrderField],
) (*page.Page[*coredata.RightsRequest, coredata.RightsRequestOrderField], error) {
	var requests coredata.RightsRequests

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := requests.LoadByOrganizationIDAndContact(ctx, conn, scope, organizationID, contact, cursor)
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
