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
	ComplianceNewsletterSubscriber struct {
		ID             gid.GID      `db:"id"`
		TenantID       gid.TenantID `db:"tenant_id"`
		OrganizationID gid.GID      `db:"organization_id"`
		TrustCenterID  gid.GID      `db:"trust_center_id"`
		Email          mail.Addr    `db:"email"`
		CreatedAt      time.Time    `db:"created_at"`
		UpdatedAt      time.Time    `db:"updated_at"`
	}

	ComplianceNewsletterSubscribers []*ComplianceNewsletterSubscriber
)

func (cns *ComplianceNewsletterSubscriber) CursorKey(orderBy ComplianceNewsletterSubscriberOrderField) page.CursorKey {
	switch orderBy {
	case ComplianceNewsletterSubscriberOrderFieldCreatedAt:
		return page.NewCursorKey(cns.ID, cns.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (cns *ComplianceNewsletterSubscriber) LoadByTrustCenterIDAndEmail(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
	email mail.Addr,
) error {
	q := `
SELECT
	id,
	tenant_id,
	organization_id,
	trust_center_id,
	email,
	created_at,
	updated_at
FROM
	compliance_newsletter_subscribers
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND email = @email
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": trustCenterID,
		"email":           email,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance newsletter subscriber: %w", err)
	}

	subscriber, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ComplianceNewsletterSubscriber])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect compliance newsletter subscriber: %w", err)
	}

	*cns = subscriber

	return nil
}

func (cns *ComplianceNewsletterSubscriber) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO compliance_newsletter_subscribers (
	id,
	tenant_id,
	organization_id,
	trust_center_id,
	email,
	created_at,
	updated_at
)
VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@trust_center_id,
	@email,
	@created_at,
	@updated_at
)
RETURNING
	id,
	tenant_id,
	organization_id,
	trust_center_id,
	email,
	created_at,
	updated_at;
`

	args := pgx.StrictNamedArgs{
		"id":              cns.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": cns.OrganizationID,
		"trust_center_id": cns.TrustCenterID,
		"email":           cns.Email,
		"created_at":      cns.CreatedAt,
		"updated_at":      cns.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert compliance newsletter subscriber: %w", err)
	}

	subscriber, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ComplianceNewsletterSubscriber])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot collect compliance newsletter subscriber: %w", err)
	}

	*cns = subscriber

	return nil
}

func (cns *ComplianceNewsletterSubscriber) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM
	compliance_newsletter_subscribers
WHERE
	%s
	AND id = @id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": cns.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance newsletter subscriber: %w", err)
	}

	return nil
}

func (cnss *ComplianceNewsletterSubscribers) LoadByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[ComplianceNewsletterSubscriberOrderField],
) error {
	q := `
SELECT
	id,
	tenant_id,
	organization_id,
	trust_center_id,
	email,
	created_at,
	updated_at
FROM
	compliance_newsletter_subscribers
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": trustCenterID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query compliance newsletter subscribers: %w", err)
	}

	subscribers, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ComplianceNewsletterSubscriber])
	if err != nil {
		return fmt.Errorf("cannot collect compliance newsletter subscribers: %w", err)
	}

	*cnss = subscribers

	return nil
}
