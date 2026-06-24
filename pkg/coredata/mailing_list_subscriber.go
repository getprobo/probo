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
	"go.probo.inc/probo/pkg/iam/policy"
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

func (cns *MailingListSubscriber) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM mailing_list_subscribers WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
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
	conn pg.Querier,
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
	conn pg.Querier,
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
	conn pg.Tx,
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
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "mailing_list_subscribers_mailing_list_id_email_key" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert mailing list subscriber: %w", err)
	}

	return nil
}

func (cns *MailingListSubscriber) Update(
	ctx context.Context,
	conn pg.Tx,
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
	conn pg.Tx,
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

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete mailing list subscriber: %w", err)
	}

	return nil
}

func (cnss *MailingListSubscribers) CountByMailingListID(
	ctx context.Context,
	conn pg.Querier,
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

func (cnss *MailingListSubscribers) LoadConfirmedByMailingListID(
	ctx context.Context,
	conn pg.Querier,
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
	AND status = 'CONFIRMED'
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"mailing_list_id": mailingListID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

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
	conn pg.Querier,
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
