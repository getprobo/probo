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

package notification

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type Service struct {
	pg *pg.Client
}

func NewService(pgClient *pg.Client) *Service {
	return &Service{pg: pgClient}
}

func (s *Service) ListNotifications(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	identityID gid.GID,
	cursor *page.Cursor[coredata.NotificationOrderField],
	filter *coredata.NotificationFilter,
) (*page.Page[*coredata.Notification, coredata.NotificationOrderField], error) {
	var notifications coredata.Notifications

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			profile, err := loadRecipient(ctx, conn, scope, organizationID, identityID)
			if err != nil {
				return fmt.Errorf("cannot load notification recipient: %w", err)
			}

			if err := notifications.LoadByRecipientProfileID(
				ctx,
				conn,
				scope,
				organizationID,
				profile.ID,
				cursor,
				filter,
			); err != nil {
				return fmt.Errorf("cannot load notifications: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(notifications, cursor), nil
}

func (s *Service) CountNotifications(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	identityID gid.GID,
	filter *coredata.NotificationFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			profile, err := loadRecipient(ctx, conn, scope, organizationID, identityID)
			if err != nil {
				return fmt.Errorf("cannot load notification recipient: %w", err)
			}

			count, err = (&coredata.Notifications{}).CountByRecipientProfileID(
				ctx,
				conn,
				scope,
				organizationID,
				profile.ID,
				filter,
			)
			if err != nil {
				return fmt.Errorf("cannot count notifications: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) GetNotification(
	ctx context.Context,
	scope coredata.Scoper,
	notificationID gid.GID,
	identityID gid.GID,
) (*coredata.Notification, error) {
	notification := &coredata.Notification{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := notification.LoadByID(ctx, conn, scope, notificationID); err != nil {
				return fmt.Errorf("cannot load notification: %w", err)
			}

			profile, err := loadRecipient(
				ctx,
				conn,
				scope,
				notification.OrganizationID,
				identityID,
			)
			if err != nil {
				return fmt.Errorf("cannot load notification recipient: %w", err)
			}

			if notification.RecipientProfileID != profile.ID {
				return coredata.ErrResourceNotFound
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return notification, nil
}

func (s *Service) MarkAllNotificationsRead(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	identityID gid.GID,
	types []coredata.NotificationType,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			profile, err := loadRecipient(ctx, tx, scope, organizationID, identityID)
			if err != nil {
				return fmt.Errorf("cannot load notification recipient: %w", err)
			}

			if err := (&coredata.Notifications{}).MarkAllRead(
				ctx,
				tx,
				scope,
				organizationID,
				profile.ID,
				types,
				time.Now(),
			); err != nil {
				return err
			}

			return nil
		},
	)
}

func loadRecipient(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	identityID gid.GID,
) (*coredata.MembershipProfile, error) {
	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByIdentityIDAndOrganizationID(
		ctx,
		conn,
		scope,
		identityID,
		organizationID,
	); err != nil {
		return nil, fmt.Errorf("cannot load membership profile: %w", err)
	}

	return profile, nil
}

type EmailPreference struct {
	Type  coredata.NotificationType
	Email coredata.NotificationEmailDelivery
}

func (s *Service) ListPreferences(
	ctx context.Context,
	organizationID gid.GID,
	identityID gid.GID,
) ([]EmailPreference, error) {
	var stored coredata.NotificationPreferences

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			scope := coredata.NewScopeFromObjectID(organizationID)
			profile, err := loadRecipient(ctx, conn, scope, organizationID, identityID)
			if err != nil {
				return err
			}

			if err := stored.LoadByProfileID(ctx, conn, profile.ID); err != nil {
				return fmt.Errorf("cannot load notification preferences: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	byType := make(map[coredata.NotificationType]coredata.NotificationEmailDelivery, len(stored))
	for _, preference := range stored {
		byType[preference.Type] = preference.EmailDelivery
	}

	preferences := make([]EmailPreference, 0, len(coredata.NotificationTypes()))
	for _, notificationType := range coredata.NotificationTypes() {
		email, ok := byType[notificationType]
		if !ok {
			email = coredata.DefaultNotificationEmailDelivery(notificationType)
		}

		preferences = append(preferences, EmailPreference{
			Type:  notificationType,
			Email: email,
		})
	}

	return preferences, nil
}

func (s *Service) UpdatePreference(
	ctx context.Context,
	organizationID gid.GID,
	identityID gid.GID,
	notificationType coredata.NotificationType,
	email coredata.NotificationEmailDelivery,
) (EmailPreference, error) {
	if !notificationType.IsValid() || !email.IsValid() {
		return EmailPreference{}, fmt.Errorf("cannot update notification preference: invalid value")
	}

	now := time.Now()
	preference := &coredata.NotificationPreference{
		OrganizationID: organizationID,
		Type:           notificationType,
		EmailDelivery:  email,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(organizationID)
			profile, err := loadRecipient(ctx, tx, scope, organizationID, identityID)
			if err != nil {
				return err
			}

			preference.RecipientProfileID = profile.ID
			if err := preference.Upsert(ctx, tx); err != nil {
				return fmt.Errorf("cannot update notification preference: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return EmailPreference{}, err
	}

	return EmailPreference{
		Type:  preference.Type,
		Email: preference.EmailDelivery,
	}, nil
}
