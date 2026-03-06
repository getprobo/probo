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
	"go.probo.inc/probo/pkg/page"
)

type (
	MailingListUpdate struct {
		ID             gid.GID                 `db:"id"`
		OrganizationID gid.GID                 `db:"organization_id"`
		MailingListID  gid.GID                 `db:"mailing_list_id"`
		Title          string                  `db:"title"`
		Body           string                  `db:"body"`
		Status         MailingListUpdateStatus `db:"status"`
		CreatedAt      time.Time               `db:"created_at"`
		UpdatedAt      time.Time               `db:"updated_at"`
	}

	MailingListUpdateItems []*MailingListUpdate
)

func (mlu *MailingListUpdate) CursorKey(orderBy MailingListUpdateOrderField) page.CursorKey {
	switch orderBy {
	case MailingListUpdateOrderFieldCreatedAt:
		return page.NewCursorKey(mlu.ID, mlu.CreatedAt)
	case MailingListUpdateOrderFieldUpdatedAt:
		return page.NewCursorKey(mlu.ID, mlu.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (mlu *MailingListUpdate) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM mailing_list_updates WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, mlu.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query mailing list update authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (mlu *MailingListUpdate) Insert(ctx context.Context, conn pg.Conn, scope Scoper) error {
	q := `
INSERT INTO mailing_list_updates (
	id,
	tenant_id,
	organization_id,
	mailing_list_id,
	title,
	body,
	status,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@mailing_list_id,
	@title,
	@body,
	@status,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":              mlu.ID,
		"organization_id": mlu.OrganizationID,
		"mailing_list_id": mlu.MailingListID,
		"title":           mlu.Title,
		"body":            mlu.Body,
		"status":          mlu.Status,
		"created_at":      mlu.CreatedAt,
		"updated_at":      mlu.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (mlu *MailingListUpdate) Update(ctx context.Context, conn pg.Conn, scope Scoper) error {
	q := `
UPDATE mailing_list_updates
SET
	title      = @title,
	body       = @body,
	status     = @status,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         mlu.ID,
		"title":      mlu.Title,
		"body":       mlu.Body,
		"status":     mlu.Status,
		"updated_at": mlu.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	tag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update mailing list update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrResourceNotFound
	}
	return nil
}

func (mlu *MailingListUpdate) Delete(ctx context.Context, conn pg.Conn, scope Scoper) error {
	q := `
DELETE FROM mailing_list_updates
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": mlu.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	tag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete mailing list update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrResourceNotFound
	}
	return nil
}

func (mlu *MailingListUpdate) LoadByID(ctx context.Context, conn pg.Conn, scope Scoper, id gid.GID) error {
	q := `
SELECT
	id,
	organization_id,
	mailing_list_id,
	title,
	body,
	status,
	created_at,
	updated_at
FROM mailing_list_updates
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
		return fmt.Errorf("cannot query mailing list update: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[MailingListUpdate])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect mailing list update: %w", err)
	}

	*mlu = result
	return nil
}

func (mlul *MailingListUpdateItems) LoadSentByMailingListID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	mailingListID gid.GID,
	cursor *page.Cursor[MailingListUpdateOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	mailing_list_id,
	title,
	body,
	status,
	created_at,
	updated_at
FROM mailing_list_updates
WHERE
	%s
	AND mailing_list_id = @mailing_list_id
	AND status = 'SENT'
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
		return fmt.Errorf("cannot query sent mailing list updates: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MailingListUpdate])
	if err != nil {
		return fmt.Errorf("cannot collect sent mailing list updates: %w", err)
	}

	*mlul = results
	return nil
}

func (mlul *MailingListUpdateItems) LoadByMailingListID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	mailingListID gid.GID,
	cursor *page.Cursor[MailingListUpdateOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	mailing_list_id,
	title,
	body,
	status,
	created_at,
	updated_at
FROM mailing_list_updates
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
		return fmt.Errorf("cannot query mailing list updates: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MailingListUpdate])
	if err != nil {
		return fmt.Errorf("cannot collect mailing list updates: %w", err)
	}

	*mlul = results
	return nil
}

func (mlul *MailingListUpdateItems) CountByMailingListID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	mailingListID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(*)
FROM mailing_list_updates
WHERE
	%s
	AND mailing_list_id = @mailing_list_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"mailing_list_id": mailingListID,
	}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count mailing list updates: %w", err)
	}

	return count, nil
}
