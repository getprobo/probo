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

type WebhookSubscriptionService struct {
	svc *TenantService
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
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.WebhookSubscriptionOrderField],
) (*page.Page[*coredata.WebhookSubscription, coredata.WebhookSubscriptionOrderField], error) {
	var subscriptions coredata.WebhookSubscriptions
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err := subscriptions.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
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
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			subscriptions := &coredata.WebhookSubscriptions{}
			count, err = subscriptions.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
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
	ctx context.Context,
	webhookSubscriptionID gid.GID,
) (*coredata.WebhookSubscription, error) {
	wc := &coredata.WebhookSubscription{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := wc.LoadByID(ctx, conn, s.svc.scope, webhookSubscriptionID); err != nil {
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
	ctx context.Context,
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
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
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

			if err := wc.Insert(ctx, conn, s.svc.scope); err != nil {
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
	ctx context.Context,
	req UpdateWebhookSubscriptionRequest,
) (*coredata.WebhookSubscription, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	wc := &coredata.WebhookSubscription{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := wc.LoadByID(ctx, conn, s.svc.scope, req.WebhookSubscriptionID); err != nil {
				return fmt.Errorf("cannot load webhook subscription: %w", err)
			}

			if req.EndpointURL != nil {
				wc.EndpointURL = *req.EndpointURL
			}
			if req.SelectedEvents != nil {
				wc.SelectedEvents = req.SelectedEvents
			}

			wc.UpdatedAt = time.Now()

			if err := wc.Update(ctx, conn, s.svc.scope); err != nil {
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
	ctx context.Context,
	webhookSubscriptionID gid.GID,
) (string, error) {
	wc := &coredata.WebhookSubscription{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := wc.LoadByID(ctx, conn, s.svc.scope, webhookSubscriptionID); err != nil {
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
	ctx context.Context,
	webhookSubscriptionID gid.GID,
	cursor *page.Cursor[coredata.WebhookEventOrderField],
) (*page.Page[*coredata.WebhookEvent, coredata.WebhookEventOrderField], error) {
	var events coredata.WebhookEvents

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := events.LoadBySubscriptionID(ctx, conn, s.svc.scope, webhookSubscriptionID, cursor); err != nil {
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
	ctx context.Context,
	webhookSubscriptionID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			events := &coredata.WebhookEvents{}
			count, err = events.CountBySubscriptionID(ctx, conn, s.svc.scope, webhookSubscriptionID)

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
	ctx context.Context,
	webhookSubscriptionID gid.GID,
) error {
	wc := &coredata.WebhookSubscription{ID: webhookSubscriptionID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := wc.LoadByID(ctx, conn, s.svc.scope, webhookSubscriptionID); err != nil {
				return fmt.Errorf("cannot load webhook subscription: %w", err)
			}

			if err := wc.Delete(ctx, conn, s.svc.scope); err != nil {
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
