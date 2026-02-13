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

type WebhookConfigurationService struct {
	svc *TenantService
}

type (
	CreateWebhookConfigurationRequest struct {
		OrganizationID gid.GID
		EndpointURL    string
		SelectedEvents []coredata.WebhookEventType
	}

	UpdateWebhookConfigurationRequest struct {
		WebhookConfigurationID gid.GID
		EndpointURL            *string
		SelectedEvents         []coredata.WebhookEventType
	}
)

func (r *CreateWebhookConfigurationRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.EndpointURL, "endpoint_url", validator.Required(), validator.URL())

	return v.Error()
}

func (r *UpdateWebhookConfigurationRequest) Validate() error {
	v := validator.New()

	v.Check(r.WebhookConfigurationID, "webhook_configuration_id", validator.Required(), validator.GID(coredata.WebhookConfigurationEntityType))
	v.Check(r.EndpointURL, "endpoint_url", validator.NotEmpty(), validator.URL())

	return v.Error()
}

func (s WebhookConfigurationService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.WebhookConfigurationOrderField],
) (*page.Page[*coredata.WebhookConfiguration, coredata.WebhookConfigurationOrderField], error) {
	var configurations coredata.WebhookConfigurations
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err := configurations.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organization.ID,
				cursor,
			)
			if err != nil {
				return fmt.Errorf("cannot load webhook configurations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(configurations, cursor), nil
}

func (s WebhookConfigurationService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			configurations := &coredata.WebhookConfigurations{}
			count, err = configurations.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count webhook configurations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s WebhookConfigurationService) Get(
	ctx context.Context,
	webhookConfigurationID gid.GID,
) (*coredata.WebhookConfiguration, error) {
	wc := &coredata.WebhookConfiguration{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := wc.LoadByID(ctx, conn, s.svc.scope, webhookConfigurationID); err != nil {
				return fmt.Errorf("cannot load webhook configuration: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return wc, nil
}

func (s WebhookConfigurationService) Create(
	ctx context.Context,
	req CreateWebhookConfigurationRequest,
) (*coredata.WebhookConfiguration, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	var wc *coredata.WebhookConfiguration
	organization := &coredata.Organization{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			wc = &coredata.WebhookConfiguration{
				ID:             gid.New(organization.ID.TenantID(), coredata.WebhookConfigurationEntityType),
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
				return fmt.Errorf("cannot insert webhook configuration: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return wc, nil
}

func (s WebhookConfigurationService) Update(
	ctx context.Context,
	req UpdateWebhookConfigurationRequest,
) (*coredata.WebhookConfiguration, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	wc := &coredata.WebhookConfiguration{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := wc.LoadByID(ctx, conn, s.svc.scope, req.WebhookConfigurationID); err != nil {
				return fmt.Errorf("cannot load webhook configuration: %w", err)
			}

			if req.EndpointURL != nil {
				wc.EndpointURL = *req.EndpointURL
			}
			if req.SelectedEvents != nil {
				wc.SelectedEvents = req.SelectedEvents
			}

			wc.UpdatedAt = time.Now()

			if err := wc.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update webhook configuration: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return wc, nil
}

func (s WebhookConfigurationService) GetSigningSecret(
	ctx context.Context,
	webhookConfigurationID gid.GID,
) (string, error) {
	wc := &coredata.WebhookConfiguration{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := wc.LoadByID(ctx, conn, s.svc.scope, webhookConfigurationID); err != nil {
				return fmt.Errorf("cannot load webhook configuration: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return "", err
	}

	return wc.DecryptSigningSecret(s.svc.encryptionKey)
}

func (s WebhookConfigurationService) ListEventsForConfigurationID(
	ctx context.Context,
	webhookConfigurationID gid.GID,
	cursor *page.Cursor[coredata.WebhookEventOrderField],
) (*page.Page[*coredata.WebhookEvent, coredata.WebhookEventOrderField], error) {
	var events coredata.WebhookEvents

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := events.LoadByConfigurationID(ctx, conn, s.svc.scope, webhookConfigurationID, cursor); err != nil {
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

func (s WebhookConfigurationService) CountEventsForConfigurationID(
	ctx context.Context,
	webhookConfigurationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			events := &coredata.WebhookEvents{}
			count, err = events.CountByConfigurationID(ctx, conn, s.svc.scope, webhookConfigurationID)

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

func (s WebhookConfigurationService) Delete(
	ctx context.Context,
	webhookConfigurationID gid.GID,
) error {
	wc := &coredata.WebhookConfiguration{ID: webhookConfigurationID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := wc.LoadByID(ctx, conn, s.svc.scope, webhookConfigurationID); err != nil {
				return fmt.Errorf("cannot load webhook configuration: %w", err)
			}

			if err := wc.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete webhook configuration: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return err
	}

	return nil
}
