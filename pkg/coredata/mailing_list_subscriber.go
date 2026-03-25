// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type (
	MailingListSubscriber struct {
		ID             gid.GID                     `db:"id"`
		OrganizationID gid.GID                     `db:"organization_id"`
		MailingListID  gid.GID                     `db:"mailing_list_id"`
		FullName       string                      `db:"full_name"`
		Email          mail.Addr                   `db:"email"`
		Status         MailingListSubscriberStatus `db:"status"`
		CreatedAt      time.Time                   `db:"created_at"`
		UpdatedAt      time.Time                   `db:"updated_at"`
	}

	MailingListSubscribers []*MailingListSubscriber
)

func (cns *MailingListSubscriber) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM mailing_list_subscribers WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, cns.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query mailing list subscriber authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (cns *MailingListSubscriber) CursorKey(orderBy MailingListSubscriberOrderField) page.CursorKey {
	switch orderBy {
	case MailingListSubscriberOrderFieldCreatedAt:
		return page.NewCursorKey(cns.ID, cns.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (cns *MailingListSubscriber) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	mailing_list_id,
	full_name,
	email,
	status,
	created_at,
	updated_at
FROM
	mailing_list_subscribers
WHERE
	%s
	AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": id,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query mailing list subscriber: %w", err)
	}

	subscriber, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[MailingListSubscriber])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect mailing list subscriber: %w", err)
	}

	*cns = subscriber

	return nil
}

func (cns *MailingListSubscriber) LoadByMailingListIDAndEmail(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	mailingListID gid.GID,
	email mail.Addr,
) error {
	q := `
SELECT
	id,
	organization_id,
	mailing_list_id,
	full_name,
	email,
	status,
	created_at,
	updated_at
FROM
	mailing_list_subscribers
WHERE
	%s
	AND mailing_list_id = @mailing_list_id
	AND email = @email
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"mailing_list_id": mailingListID,
		"email":           email,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query mailing list subscriber: %w", err)
	}

	subscriber, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[MailingListSubscriber])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect mailing list subscriber: %w", err)
	}

	*cns = subscriber

	return nil
}

func (cns *MailingListSubscriber) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO mailing_list_subscribers (
	id,
	tenant_id,
	organization_id,
	mailing_list_id,
	full_name,
	email,
	status,
	created_at,
	updated_at
)
VALUES (
	@mailing_list_subscriber_id,
	@tenant_id,
	@organization_id,
	@mailing_list_id,
	@full_name,
	@email,
	@status,
	@created_at,
	@updated_at
);
`

	args := pgx.StrictNamedArgs{
		"mailing_list_subscriber_id": cns.ID,
		"tenant_id":                  scope.GetTenantID(),
		"organization_id":            cns.OrganizationID,
		"mailing_list_id":            cns.MailingListID,
		"full_name":                  cns.FullName,
		"email":                      cns.Email,
		"status":                     cns.Status,
		"created_at":                 cns.CreatedAt,
		"updated_at":                 cns.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert mailing list subscriber: %w", err)
	}

	return nil
}

func (cns *MailingListSubscriber) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE mailing_list_subscribers
SET
	status = @status,
	updated_at = @updated_at
WHERE
	%s
	AND id = @mailing_list_subscriber_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"mailing_list_subscriber_id": cns.ID,
		"status":                     cns.Status,
		"updated_at":                 cns.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	tag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update mailing list subscriber: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (cns *MailingListSubscriber) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM
	mailing_list_subscribers
WHERE
	%s
	AND id = @mailing_list_subscriber_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"mailing_list_subscriber_id": cns.ID}
	maps.Copy(args, scope.SQLArguments())

	tag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete mailing list subscriber: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (cnss *MailingListSubscribers) CountByMailingListID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	mailingListID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	mailing_list_subscribers
WHERE
	%s
	AND mailing_list_id = @mailing_list_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"mailing_list_id": mailingListID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count mailing list subscribers: %w", err)
	}

	return count, nil
}

func (cnss *MailingListSubscribers) LoadAllConfirmedByMailingListID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	mailingListID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	mailing_list_id,
	full_name,
	email,
	status,
	created_at,
	updated_at
FROM
	mailing_list_subscribers
WHERE
	%s
	AND mailing_list_id = @mailing_list_id
	AND status = 'CONFIRMED'
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"mailing_list_id": mailingListID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query confirmed mailing list subscribers: %w", err)
	}

	subscribers, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MailingListSubscriber])
	if err != nil {
		return fmt.Errorf("cannot collect confirmed mailing list subscribers: %w", err)
	}

	*cnss = subscribers

	return nil
}

func (cnss *MailingListSubscribers) LoadByMailingListID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	mailingListID gid.GID,
	cursor *page.Cursor[MailingListSubscriberOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	mailing_list_id,
	full_name,
	email,
	status,
	created_at,
	updated_at
FROM
	mailing_list_subscribers
WHERE
	%s
	AND mailing_list_id = @mailing_list_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"mailing_list_id": mailingListID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query mailing list subscribers: %w", err)
	}

	subscribers, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MailingListSubscriber])
	if err != nil {
		return fmt.Errorf("cannot collect mailing list subscribers: %w", err)
	}

	*cnss = subscribers

	return nil
}
