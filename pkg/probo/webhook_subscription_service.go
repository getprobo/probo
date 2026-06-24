// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

type WebhookSubscriptionService struct {
	svc *Service
}

type (
	CreateWebhookSubscriptionRequest struct {
		OrganizationID gid.GID
		EndpointURL    string
		SelectedEvents []coredata.WebhookEventType
	}

	UpdateWebhookSubscriptionRequest struct {
		WebhookSubscriptionID gid.GID
		EndpointURL           *string
		SelectedEvents        []coredata.WebhookEventType
	}
)

func (r *CreateWebhookSubscriptionRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.EndpointURL, "endpoint_url", validator.Required(), validator.HTTPSUrl())

	return v.Error()
}

func (r *UpdateWebhookSubscriptionRequest) Validate() error {
	v := validator.New()

	v.Check(r.WebhookSubscriptionID, "webhook_subscription_id", validator.Required(), validator.GID(coredata.WebhookSubscriptionEntityType))
	v.Check(r.EndpointURL, "endpoint_url", validator.NotEmpty(), validator.HTTPSUrl())

	return v.Error()
}

func (s WebhookSubscriptionService) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.WebhookSubscriptionOrderField],
) (*page.Page[*coredata.WebhookSubscription, coredata.WebhookSubscriptionOrderField], error) {
	var subscriptions coredata.WebhookSubscriptions

	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err := subscriptions.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organization.ID,
				cursor,
			)
			if err != nil {
				return fmt.Errorf("cannot load webhook subscriptions: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(subscriptions, cursor), nil
}

func (s WebhookSubscriptionService) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			subscriptions := &coredata.WebhookSubscriptions{}

			count, err = subscriptions.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count webhook subscriptions: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s WebhookSubscriptionService) Get(
	ctx context.Context, scope coredata.Scoper,
	webhookSubscriptionID gid.GID,
) (*coredata.WebhookSubscription, error) {
	wc := &coredata.WebhookSubscription{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := wc.LoadByID(ctx, conn, scope, webhookSubscriptionID); err != nil {
				return fmt.Errorf("cannot load webhook subscription: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return wc, nil
}

func (s WebhookSubscriptionService) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateWebhookSubscriptionRequest,
) (*coredata.WebhookSubscription, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	var wc *coredata.WebhookSubscription

	organization := &coredata.Organization{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			wc = &coredata.WebhookSubscription{
				ID:             gid.New(organization.ID.TenantID(), coredata.WebhookSubscriptionEntityType),
				OrganizationID: organization.ID,
				EndpointURL:    req.EndpointURL,
				SelectedEvents: req.SelectedEvents,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if _, err := wc.GenerateSigningSecret(s.svc.encryptionKey); err != nil {
				return fmt.Errorf("cannot generate signing secret: %w", err)
			}

			if err := wc.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert webhook subscription: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return wc, nil
}

func (s WebhookSubscriptionService) Update(
	ctx context.Context, scope coredata.Scoper,
	req UpdateWebhookSubscriptionRequest,
) (*coredata.WebhookSubscription, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	wc := &coredata.WebhookSubscription{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := wc.LoadByID(ctx, conn, scope, req.WebhookSubscriptionID); err != nil {
				return fmt.Errorf("cannot load webhook subscription: %w", err)
			}

			if req.EndpointURL != nil {
				wc.EndpointURL = *req.EndpointURL
			}

			if req.SelectedEvents != nil {
				wc.SelectedEvents = req.SelectedEvents
			}

			wc.UpdatedAt = time.Now()

			if err := wc.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update webhook subscription: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return wc, nil
}

func (s WebhookSubscriptionService) GetSigningSecret(
	ctx context.Context, scope coredata.Scoper,
	webhookSubscriptionID gid.GID,
) (string, error) {
	wc := &coredata.WebhookSubscription{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := wc.LoadByID(ctx, conn, scope, webhookSubscriptionID); err != nil {
				return fmt.Errorf("cannot load webhook subscription: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return wc.DecryptSigningSecret(s.svc.encryptionKey)
}

func (s WebhookSubscriptionService) ListEventsForSubscriptionID(
	ctx context.Context, scope coredata.Scoper,
	webhookSubscriptionID gid.GID,
	cursor *page.Cursor[coredata.WebhookEventOrderField],
) (*page.Page[*coredata.WebhookEvent, coredata.WebhookEventOrderField], error) {
	var events coredata.WebhookEvents

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := events.LoadBySubscriptionID(ctx, conn, scope, webhookSubscriptionID, cursor); err != nil {
				return fmt.Errorf("cannot load webhook events: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(events, cursor), nil
}

func (s WebhookSubscriptionService) CountEventsForSubscriptionID(
	ctx context.Context, scope coredata.Scoper,
	webhookSubscriptionID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			events := &coredata.WebhookEvents{}

			count, err = events.CountBySubscriptionID(ctx, conn, scope, webhookSubscriptionID)
			if err != nil {
				return fmt.Errorf("cannot count webhook events: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s WebhookSubscriptionService) Delete(
	ctx context.Context, scope coredata.Scoper,
	webhookSubscriptionID gid.GID,
) error {
	wc := &coredata.WebhookSubscription{ID: webhookSubscriptionID}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := wc.LoadByID(ctx, conn, scope, webhookSubscriptionID); err != nil {
				return fmt.Errorf("cannot load webhook subscription: %w", err)
			}

			if err := wc.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete webhook subscription: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}
