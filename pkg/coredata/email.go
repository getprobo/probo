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

package coredata

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type (
	Email struct {
		ID                  gid.GID    `db:"id"`
		RecipientEmail      string     `db:"recipient_email"`
		RecipientName       string     `db:"recipient_name"`
		ReplyTo             *mail.Addr `db:"reply_to"`
		UnsubscribeURL      *string    `db:"unsubscribe_url"`
		MailingListUpdateID *gid.GID   `db:"mailing_list_update_id"`
		Subject             string     `db:"subject"`
		TextBody            string     `db:"text_body"`
		HtmlBody            *string    `db:"html_body"`
		CreatedAt           time.Time  `db:"created_at"`
		UpdatedAt           time.Time  `db:"updated_at"`
		SentAt              *time.Time `db:"sent_at"`
	}

	Emails []*Email

	EmailOptions struct {
		ReplyTo             *mail.Addr
		UnsubscribeURL      *string
		MailingListUpdateID *gid.GID
	}
)

var (
	ErrNoUnsentEmail = errors.New("no unsent email found")
)

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
// Email is identity-scoped (not org-scoped), so it returns an empty map.
func (e *Email) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	return map[string]string{}, nil
}

func NewEmail(
	recipientName string,
	recipientEmail mail.Addr,
	subject string,
	textBody string,
	htmlBody *string,
	opts *EmailOptions,
) *Email {
	now := time.Now()
	e := &Email{
		ID:             gid.New(gid.NilTenant, EmailEntityType),
		RecipientName:  recipientName,
		RecipientEmail: recipientEmail.String(),
		Subject:        subject,
		TextBody:       textBody,
		HtmlBody:       htmlBody,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if opts != nil {
		e.ReplyTo = opts.ReplyTo
		e.UnsubscribeURL = opts.UnsubscribeURL
		e.MailingListUpdateID = opts.MailingListUpdateID
	}

	return e
}

func (e *Email) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO emails (id, recipient_email, recipient_name, reply_to, unsubscribe_url, mailing_list_update_id, subject, text_body, html_body, created_at, updated_at)
VALUES (@id, @recipient_email, @recipient_name, @reply_to, @unsubscribe_url, @mailing_list_update_id, @subject, @text_body, @html_body, @created_at, @updated_at)
	`

	args := pgx.StrictNamedArgs{
		"id":                     e.ID,
		"recipient_email":        e.RecipientEmail,
		"recipient_name":         e.RecipientName,
		"reply_to":               e.ReplyTo,
		"unsubscribe_url":        e.UnsubscribeURL,
		"mailing_list_update_id": e.MailingListUpdateID,
		"subject":                e.Subject,
		"text_body":              e.TextBody,
		"html_body":              e.HtmlBody,
		"created_at":             e.CreatedAt,
		"updated_at":             e.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (emails Emails) BulkInsert(
	ctx context.Context,
	conn pg.Conn,
) error {
	if len(emails) == 0 {
		return nil
	}

	rows := make([][]any, 0, len(emails))
	for _, e := range emails {
		rows = append(rows, []any{
			e.ID,
			e.RecipientEmail,
			e.RecipientName,
			e.ReplyTo,
			e.UnsubscribeURL,
			e.MailingListUpdateID,
			e.Subject,
			e.TextBody,
			e.HtmlBody,
			e.CreatedAt,
			e.UpdatedAt,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"emails"},
		[]string{"id", "recipient_email", "recipient_name", "reply_to", "unsubscribe_url", "mailing_list_update_id", "subject", "text_body", "html_body", "created_at", "updated_at"},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (e *Email) LoadNextUnsentForUpdate(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT id, recipient_email, recipient_name, reply_to, unsubscribe_url, subject, text_body, html_body, created_at, updated_at, sent_at
FROM emails
WHERE sent_at IS NULL
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE
	`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return err
	}

	email, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Email])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoUnsentEmail
		}

		return fmt.Errorf("cannot collect email: %w", err)
	}

	*e = email

	return nil
}

func (e *Email) Update(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
UPDATE emails
SET sent_at = @sent_at, updated_at = @updated_at
WHERE id = @id
	`

	args := pgx.StrictNamedArgs{
		"id":         e.ID,
		"sent_at":    e.SentAt,
		"updated_at": e.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	return err
}
