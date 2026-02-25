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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type MailingList struct {
	ID             gid.GID    `db:"id"`
	OrganizationID gid.GID    `db:"organization_id"`
	ReplyTo        *mail.Addr `db:"reply_to"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}

func (ml *MailingList) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM mailing_lists WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, ml.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query mailing list authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (ml *MailingList) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	reply_to,
	created_at,
	updated_at
FROM
	mailing_lists
WHERE
	%s
	AND id = @mailing_list_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"mailing_list_id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query mailing list: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[MailingList])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect mailing list: %w", err)
	}

	*ml = result

	return nil
}

func (ml *MailingList) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE mailing_lists
SET
	reply_to = @reply_to,
	updated_at = @updated_at
WHERE
	%s
	AND id = @mailing_list_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"mailing_list_id": ml.ID,
		"reply_to":        ml.ReplyTo,
		"updated_at":      ml.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	tag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update mailing list: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (ml *MailingList) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO mailing_lists (
	id,
	tenant_id,
	organization_id,
	reply_to,
	created_at,
	updated_at
)
VALUES (
	@mailing_list_id,
	@tenant_id,
	@organization_id,
	@reply_to,
	@created_at,
	@updated_at
);
`

	args := pgx.StrictNamedArgs{
		"mailing_list_id": ml.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": ml.OrganizationID,
		"reply_to":        ml.ReplyTo,
		"created_at":      ml.CreatedAt,
		"updated_at":      ml.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert mailing list: %w", err)
	}

	return nil
}
