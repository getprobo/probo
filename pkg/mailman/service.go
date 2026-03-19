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

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

const (
	updateTitleMaxLength        = 200
	updateBodyMaxLength         = 50000
	subscriberFullNameMaxLength = 200
)

const (
	pathUnsubscribe = "/mail-actions/unsubscribe"
	pathConfirm     = "/mail-actions/confirm"
)

type Service struct {
	pg            *pg.Client
	fm            *filemanager.Service
	tokenSecret   string
	apiBaseURL    *baseurl.BaseURL
	bucket        string
	encryptionKey cipher.EncryptionKey
	logger        *log.Logger
}

func NewService(pgClient *pg.Client, fm *filemanager.Service, tokenSecret string, apiBaseURL *baseurl.BaseURL, bucket string, encryptionKey cipher.EncryptionKey, logger *log.Logger) *Service {
	return &Service{pg: pgClient, fm: fm, tokenSecret: tokenSecret, apiBaseURL: apiBaseURL, bucket: bucket, encryptionKey: encryptionKey, logger: logger}
}

type (
	CreateMailingListUpdateRequest struct {
		MailingListID gid.GID
		Title         string
		Body          string
	}

	UpdateMailingListUpdateRequest struct {
		ID    gid.GID
		Title *string
		Body  *string
	}

	CreateSubscriberRequest struct {
		MailingListID gid.GID
		Email         mail.Addr
		FullName      string
		Confirmed     bool
	}
)

func (r *CreateMailingListUpdateRequest) Validate() error {
	v := validator.New()

	v.Check(r.MailingListID, "mailing_list_id", validator.Required(), validator.GID(coredata.MailingListEntityType))
	v.Check(r.Title, "title", validator.Required(), validator.SafeTextNoNewLine(updateTitleMaxLength))
	v.Check(r.Body, "body", validator.Required(), validator.SafeText(updateBodyMaxLength))

	return v.Error()
}

func (r *UpdateMailingListUpdateRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.MailingListUpdateEntityType))
	v.Check(r.Title, "title", validator.SafeTextNoNewLine(updateTitleMaxLength))
	v.Check(r.Body, "body", validator.SafeText(updateBodyMaxLength))

	return v.Error()
}

func (r *CreateSubscriberRequest) Validate() error {
	v := validator.New()

	v.Check(r.MailingListID, "mailing_list_id", validator.Required(), validator.GID(coredata.MailingListEntityType))
	v.Check(r.Email, "email", validator.Required(), validator.NotEmpty())
	v.Check(r.FullName, "full_name", validator.Required(), validator.SafeTextNoNewLine(subscriberFullNameMaxLength))

	return v.Error()
}

func (s *Service) UpdateMailingList(
	ctx context.Context,
	id gid.GID,
	replyTo *mail.Addr,
) (*coredata.MailingList, error) {
	var ml coredata.MailingList
	scope := coredata.NewScopeFromObjectID(id)

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
	req *CreateSubscriberRequest,
) (*coredata.MailingListSubscriber, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	mailingListID := req.MailingListID
	email := req.Email
	fullName := req.FullName
	scope := coredata.NewScopeFromObjectID(mailingListID)

	status := coredata.MailingListSubscriberStatusPending
	var emailRecord *coredata.Email
	if req.Confirmed {
		status = coredata.MailingListSubscriberStatusConfirmed
	} else {
		var err error
		emailRecord, err = s.buildConfirmationMail(ctx, mailingListID, email, fullName)
		if err != nil {
			return nil, fmt.Errorf("cannot build confirmation mail: %w", err)
		}
	}

	now := time.Now()
	subscriber := &coredata.MailingListSubscriber{
		ID:            gid.New(scope.GetTenantID(), coredata.MailingListSubscriberEntityType),
		MailingListID: mailingListID,
		FullName:      fullName,
		Email:         email,
		Status:        status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var ml coredata.MailingList
			if err := ml.LoadByID(ctx, tx, scope, mailingListID); err != nil {
				return fmt.Errorf("cannot load mailing list: %w", err)
			}
			subscriber.OrganizationID = ml.OrganizationID

			if err := subscriber.Insert(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceAlreadyExists) {
					return ErrSubscriberAlreadyExist
				}
				return fmt.Errorf("cannot insert mailing list subscriber: %w", err)
			}

			if emailRecord != nil {
				if err := emailRecord.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert subscription confirmation email: %w", err)
				}
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return subscriber, nil
}

func (s *Service) UnsubscribeByEmail(
	ctx context.Context,
	mailingListID gid.GID,
	email mail.Addr,
) error {
	scope := coredata.NewScopeFromObjectID(mailingListID)
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var subscriber coredata.MailingListSubscriber
			if err := subscriber.LoadByMailingListIDAndEmail(ctx, tx, scope, mailingListID, email); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrSubscriberNotFound
				}
				return fmt.Errorf("cannot load mailing list subscriber: %w", err)
			}

			wasConfirmed := subscriber.Status == coredata.MailingListSubscriberStatusConfirmed

			if err := subscriber.Delete(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrSubscriberNotFound
				}
				return fmt.Errorf("cannot delete mailing list subscriber: %w", err)
			}

			if wasConfirmed {
				emailRecord, err := s.buildUnsubscriptionMail(ctx, mailingListID, subscriber.Email, subscriber.FullName)
				if err != nil {
					return fmt.Errorf("cannot build unsubscription email: %w", err)
				}

				if err := emailRecord.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert unsubscription email: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) ConfirmSubscriberByEmail(
	ctx context.Context,
	mailingListID gid.GID,
	email mail.Addr,
) error {
	scope := coredata.NewScopeFromObjectID(mailingListID)

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var subscriber coredata.MailingListSubscriber
			if err := subscriber.LoadByMailingListIDAndEmail(ctx, conn, scope, mailingListID, email); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrSubscriberNotFound
				}
				return fmt.Errorf("cannot load mailing list subscriber: %w", err)
			}

			subscriber.Status = coredata.MailingListSubscriberStatusConfirmed
			subscriber.UpdatedAt = time.Now()

			if err := subscriber.Update(ctx, conn, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrSubscriberNotFound
				}
				return fmt.Errorf("cannot update mailing list subscriber: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) DeleteSubscriber(
	ctx context.Context,
	id gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(id)

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var subscriber coredata.MailingListSubscriber
			if err := subscriber.LoadByID(ctx, tx, scope, id); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrSubscriberNotFound
				}
				return fmt.Errorf("cannot load mailing list subscriber: %w", err)
			}

			wasConfirmed := subscriber.Status == coredata.MailingListSubscriberStatusConfirmed

			if err := subscriber.Delete(ctx, tx, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrSubscriberNotFound
				}
				return fmt.Errorf("cannot delete mailing list subscriber: %w", err)
			}

			if wasConfirmed {
				emailRecord, err := s.buildUnsubscriptionMail(ctx, subscriber.MailingListID, subscriber.Email, subscriber.FullName)
				if err != nil {
					return fmt.Errorf("cannot build unsubscription email: %w", err)
				}

				if err := emailRecord.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert unsubscription email: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) CountSubscribers(
	ctx context.Context,
	mailingListID gid.GID,
) (int, error) {
	scope := coredata.NewScopeFromObjectID(mailingListID)
	var count int

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
	var subscribers coredata.MailingListSubscribers

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

func (s *Service) CreateMailingListUpdate(
	ctx context.Context,
	req *CreateMailingListUpdateRequest,
) (*coredata.MailingListUpdate, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	mailingListID := req.MailingListID
	scope := coredata.NewScopeFromObjectID(mailingListID)
	now := time.Now()

	mlu := &coredata.MailingListUpdate{
		ID:            gid.New(scope.GetTenantID(), coredata.MailingListUpdateEntityType),
		MailingListID: mailingListID,
		Title:         req.Title,
		Body:          req.Body,
		Status:        coredata.MailingListUpdateStatusDraft,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var ml coredata.MailingList
			if err := ml.LoadByID(ctx, conn, scope, mailingListID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrMailingListNotFound
				}
				return fmt.Errorf("cannot load mailing list: %w", err)
			}

			mlu.OrganizationID = ml.OrganizationID

			if err := mlu.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert mailing list update: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return mlu, nil
}

func (s *Service) GetMailingListUpdate(
	ctx context.Context,
	id gid.GID,
) (*coredata.MailingListUpdate, error) {
	scope := coredata.NewScopeFromObjectID(id)
	var mlu coredata.MailingListUpdate

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := mlu.LoadByID(ctx, conn, scope, id); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrMailingListUpdateNotFound
				}
				return fmt.Errorf("cannot load mailing list update: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &mlu, nil
}

func (s *Service) UpdateMailingListUpdate(
	ctx context.Context,
	req *UpdateMailingListUpdateRequest,
) (*coredata.MailingListUpdate, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	scope := coredata.NewScopeFromObjectID(req.ID)
	var mlu coredata.MailingListUpdate

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := mlu.LoadByID(ctx, conn, scope, req.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrMailingListUpdateNotFound
				}
				return fmt.Errorf("cannot load mailing list update: %w", err)
			}

			if mlu.Status != coredata.MailingListUpdateStatusDraft {
				return ErrMailingListUpdateAlreadySent
			}

			if req.Title != nil {
				mlu.Title = *req.Title
			}
			if req.Body != nil {
				mlu.Body = *req.Body
			}
			mlu.UpdatedAt = time.Now()

			if err := mlu.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update mailing list update: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &mlu, nil
}

func (s *Service) SendMailingListUpdate(
	ctx context.Context,
	id gid.GID,
) (*coredata.MailingListUpdate, error) {
	scope := coredata.NewScopeFromObjectID(id)
	var mlu coredata.MailingListUpdate

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := mlu.LoadByID(ctx, conn, scope, id); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrMailingListUpdateNotFound
				}
				return fmt.Errorf("cannot load mailing list update: %w", err)
			}

			if mlu.Status != coredata.MailingListUpdateStatusDraft {
				return ErrMailingListUpdateAlreadySent
			}

			mlu.Status = coredata.MailingListUpdateStatusEnqueued
			mlu.UpdatedAt = time.Now()

			if err := mlu.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot queue mailing list update for sending: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &mlu, nil
}

func (s *Service) DeleteMailingListUpdate(
	ctx context.Context,
	id gid.GID,
) error {
	scope := coredata.NewScopeFromObjectID(id)

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			mlu := coredata.MailingListUpdate{ID: id}
			if err := mlu.Delete(ctx, conn, scope); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrMailingListUpdateNotFound
				}
				return fmt.Errorf("cannot delete mailing list update: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) ListMailingListUpdates(
	ctx context.Context,
	mailingListID gid.GID,
	cursor *page.Cursor[coredata.MailingListUpdateOrderField],
) (*page.Page[*coredata.MailingListUpdate, coredata.MailingListUpdateOrderField], error) {
	scope := coredata.NewScopeFromObjectID(mailingListID)
	var items coredata.MailingListUpdateItems

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := items.LoadByMailingListID(ctx, conn, scope, mailingListID, cursor); err != nil {
				return fmt.Errorf("cannot load mailing list updates: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(items, cursor), nil
}

func (s *Service) ListSentMailingListUpdates(
	ctx context.Context,
	mailingListID gid.GID,
	cursor *page.Cursor[coredata.MailingListUpdateOrderField],
) (*page.Page[*coredata.MailingListUpdate, coredata.MailingListUpdateOrderField], error) {
	scope := coredata.NewScopeFromObjectID(mailingListID)
	var items coredata.MailingListUpdateItems

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := items.LoadSentByMailingListID(ctx, conn, scope, mailingListID, cursor); err != nil {
				return fmt.Errorf("cannot load sent mailing list updates: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(items, cursor), nil
}

func (s *Service) CountMailingListUpdates(
	ctx context.Context,
	mailingListID gid.GID,
) (int, error) {
	scope := coredata.NewScopeFromObjectID(mailingListID)
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var items coredata.MailingListUpdateItems
			var err error
			count, err = items.CountByMailingListID(ctx, conn, scope, mailingListID)
			if err != nil {
				return fmt.Errorf("cannot count mailing list updates: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) CreateUpdateEmails(
	ctx context.Context,
	mailingListID gid.GID,
	mailingListUpdateID gid.GID,
	updateTitle string,
	updateBody string,
) error {
	scope := coredata.NewScopeFromObjectID(mailingListID)

	presenterCfg, orgName, compliancePageURL, replyTo, err := s.UpdateEmailConfig(ctx, mailingListID)
	if err != nil {
		return fmt.Errorf("cannot get update email config: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var subscribers coredata.MailingListSubscribers
			if err := subscribers.LoadAllConfirmedByMailingListID(ctx, tx, scope, mailingListID); err != nil {
				return fmt.Errorf("cannot load confirmed subscribers: %w", err)
			}

			if len(subscribers) == 0 {
				return nil
			}

			emailRecords := make(coredata.Emails, 0, len(subscribers))
			for _, sub := range subscribers {
				unsubscribeURL, err := s.buildUnsubscribeURL(mailingListID, sub.Email)
				if err != nil {
					return fmt.Errorf("cannot generate unsubscribe URL: %w", err)
				}

				subject, textBody, htmlBody, err := emails.NewPresenterFromConfig(s.fm, presenterCfg, sub.FullName).
					RenderMailingListNews(ctx, orgName, updateTitle, updateBody, compliancePageURL, unsubscribeURL)
				if err != nil {
					return fmt.Errorf("cannot render mailing list update email: %w", err)
				}

				emailRecords = append(
					emailRecords,
					coredata.NewEmail(
						sub.FullName,
						sub.Email,
						subject,
						textBody,
						htmlBody,
						&coredata.EmailOptions{
							ReplyTo:             replyTo,
							UnsubscribeURL:      &unsubscribeURL,
							MailingListUpdateID: &mailingListUpdateID,
						},
					),
				)
			}

			if err := emailRecords.BulkInsert(ctx, tx); err != nil {
				return fmt.Errorf("cannot bulk insert update emails: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) buildConfirmationMail(
	ctx context.Context,
	mailingListID gid.GID,
	email mail.Addr,
	fullName string,
) (*coredata.Email, error) {
	unsubscribeURL, err := s.buildUnsubscribeURL(mailingListID, email)
	if err != nil {
		return nil, fmt.Errorf("cannot generate unsubscribe URL: %w", err)
	}

	confirmURL, err := s.buildConfirmURL(mailingListID, email)
	if err != nil {
		return nil, fmt.Errorf("cannot generate confirm URL: %w", err)
	}

	presenterCfg, orgName, replyTo, err := s.SubscriptionConfirmationEmailConfig(ctx, mailingListID)
	if err != nil {
		return nil, fmt.Errorf("cannot get subscription confirmation email config: %w", err)
	}

	subject, textBody, htmlBody, err := emails.NewPresenterFromConfig(s.fm, presenterCfg, fullName).
		RenderMailingListSubscription(ctx, orgName, confirmURL, unsubscribeURL)
	if err != nil {
		return nil, fmt.Errorf("cannot render subscription confirmation email: %w", err)
	}

	return coredata.NewEmail(fullName, email, subject, textBody, htmlBody, &coredata.EmailOptions{ReplyTo: replyTo, UnsubscribeURL: &unsubscribeURL}), nil
}

func (s *Service) buildUnsubscriptionMail(
	ctx context.Context,
	mailingListID gid.GID,
	email mail.Addr,
	fullName string,
) (*coredata.Email, error) {
	presenterCfg, orgName, replyTo, err := s.UnsubscribeEmailConfig(ctx, mailingListID)
	if err != nil {
		return nil, fmt.Errorf("cannot get unsubscription email config: %w", err)
	}

	subject, textBody, htmlBody, err := emails.NewPresenterFromConfig(s.fm, presenterCfg, fullName).
		RenderMailingListUnsubscription(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("cannot render unsubscription email: %w", err)
	}

	return coredata.NewEmail(fullName, email, subject, textBody, htmlBody, &coredata.EmailOptions{ReplyTo: replyTo}), nil
}

func (s *Service) buildUnsubscribeURL(mailingListID gid.GID, email mail.Addr) (string, error) {
	if s.tokenSecret == "" {
		return "", nil
	}
	token, err := newUnsubscribeToken(s.tokenSecret, mailingListID, email)
	if err != nil {
		return "", err
	}
	return s.apiBaseURL.WithPath(pathUnsubscribe).WithQuery("token", token).String()
}

func (s *Service) buildConfirmURL(mailingListID gid.GID, email mail.Addr) (string, error) {
	if s.tokenSecret == "" {
		return "", nil
	}
	token, err := newConfirmToken(s.tokenSecret, mailingListID, email)
	if err != nil {
		return "", err
	}
	return s.apiBaseURL.WithPath(pathConfirm).WithQuery("token", token).String()
}
