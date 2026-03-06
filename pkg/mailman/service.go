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

package mailman

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type Service struct {
	pg *pg.Client
}

func NewService(pgClient *pg.Client) *Service {
	return &Service{pg: pgClient}
}

func (s *Service) UpdateMailingList(
	ctx context.Context,
	id gid.GID,
	replyTo *mail.Addr,
) (*coredata.MailingList, error) {
	scope := coredata.NewScopeFromObjectID(id)
	ml := coredata.MailingList{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := ml.LoadByID(ctx, conn, scope, id); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrMailingListNotFound
				}
				return fmt.Errorf("cannot load mailing list: %w", err)
			}

			ml.ReplyTo = replyTo
			ml.UpdatedAt = time.Now()

			if err := ml.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update mailing list: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &ml, nil
}

func (s *Service) GetSubscriber(
	ctx context.Context,
	mailingListID gid.GID,
	email mail.Addr,
) (*coredata.MailingListSubscriber, error) {
	scope := coredata.NewScopeFromObjectID(mailingListID)
	subscriber := coredata.MailingListSubscriber{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := subscriber.LoadByMailingListIDAndEmail(ctx, conn, scope, mailingListID, email); err != nil {
				return fmt.Errorf("cannot load mailing list subscriber: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &subscriber, nil
}

func (s *Service) CreateSubscriber(
	ctx context.Context,
	mailingListID gid.GID,
	email mail.Addr,
	fullName string,
) (*coredata.MailingListSubscriber, error) {
	scope := coredata.NewScopeFromObjectID(mailingListID)
	now := time.Now()

	subscriber := &coredata.MailingListSubscriber{
		ID:            gid.New(scope.GetTenantID(), coredata.MailingListSubscriberEntityType),
		MailingListID: mailingListID,
		FullName:      fullName,
		Email:         email,
		Status:        coredata.MailingListSubscriberStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			ml := coredata.MailingList{}
			if err := ml.LoadByID(ctx, conn, scope, mailingListID); err != nil {
				return fmt.Errorf("cannot load mailing list: %w", err)
			}
			subscriber.OrganizationID = ml.OrganizationID

			if err := subscriber.Insert(ctx, conn, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrSubscriberAlreadyExist
				}
				return fmt.Errorf("cannot insert mailing list subscriber: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return subscriber, nil
}

func (s *Service) DeleteSubscriber(
	ctx context.Context,
	id gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(id)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			subscriber := coredata.MailingListSubscriber{ID: id}
			if err := subscriber.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete mailing list subscriber: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CountSubscribers(
	ctx context.Context,
	mailingListID gid.GID,
) (int, error) {
	scope := coredata.NewScopeFromObjectID(mailingListID)
	count := 0

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			subscribers := coredata.MailingListSubscribers{}
			count, err = subscribers.CountByMailingListID(ctx, conn, scope, mailingListID)
			if err != nil {
				return fmt.Errorf("cannot count mailing list subscribers: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) ListSubscribers(
	ctx context.Context,
	mailingListID gid.GID,
	cursor *page.Cursor[coredata.MailingListSubscriberOrderField],
) (*page.Page[*coredata.MailingListSubscriber, coredata.MailingListSubscriberOrderField], error) {
	scope := coredata.NewScopeFromObjectID(mailingListID)
	subscribers := coredata.MailingListSubscribers{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := subscribers.LoadByMailingListID(ctx, conn, scope, mailingListID, cursor); err != nil {
				return fmt.Errorf("cannot load mailing list subscribers: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(subscribers, cursor), nil
}
