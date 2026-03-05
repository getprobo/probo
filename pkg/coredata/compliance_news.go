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
	ComplianceNews struct {
		ID             gid.GID              `db:"id"`
		OrganizationID gid.GID              `db:"organization_id"`
		TrustCenterID  gid.GID              `db:"trust_center_id"`
		Title          string               `db:"title"`
		Body           string               `db:"body"`
		Status         ComplianceNewsStatus `db:"status"`
		CreatedAt      time.Time            `db:"created_at"`
		UpdatedAt      time.Time            `db:"updated_at"`
	}

	ComplianceNewsItems []*ComplianceNews
)

func (cn *ComplianceNews) CursorKey(orderBy ComplianceNewsOrderField) page.CursorKey {
	switch orderBy {
	case ComplianceNewsOrderFieldCreatedAt:
		return page.NewCursorKey(cn.ID, cn.CreatedAt)
	case ComplianceNewsOrderFieldUpdatedAt:
		return page.NewCursorKey(cn.ID, cn.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (cn *ComplianceNews) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM compliance_news WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, cn.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query compliance news authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (cn *ComplianceNews) Insert(ctx context.Context, conn pg.Conn, scope Scoper) error {
	q := `
INSERT INTO compliance_news (
	id,
	tenant_id,
	organization_id,
	trust_center_id,
	title,
	body,
	status,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@trust_center_id,
	@title,
	@body,
	@status,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":              cn.ID,
		"organization_id": cn.OrganizationID,
		"trust_center_id": cn.TrustCenterID,
		"title":           cn.Title,
		"body":            cn.Body,
		"status":          cn.Status,
		"created_at":      cn.CreatedAt,
		"updated_at":      cn.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (cn *ComplianceNews) Update(ctx context.Context, conn pg.Conn, scope Scoper) error {
	q := `
UPDATE compliance_news
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
		"id":         cn.ID,
		"title":      cn.Title,
		"body":       cn.Body,
		"status":     cn.Status,
		"updated_at": cn.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	tag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update compliance news: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrResourceNotFound
	}
	return nil
}

func (cn *ComplianceNews) Delete(ctx context.Context, conn pg.Conn, scope Scoper) error {
	q := `
DELETE FROM compliance_news
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id": cn.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	tag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete compliance news: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrResourceNotFound
	}
	return nil
}

func (cn *ComplianceNews) LoadByID(ctx context.Context, conn pg.Conn, scope Scoper, id gid.GID) error {
	q := `
SELECT
	id,
	organization_id,
	trust_center_id,
	title,
	body,
	status,
	created_at,
	updated_at
FROM compliance_news
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
		return fmt.Errorf("cannot query compliance news: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ComplianceNews])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect compliance news: %w", err)
	}

	*cn = result
	return nil
}

func (cnl *ComplianceNewsItems) LoadSentByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[ComplianceNewsOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	trust_center_id,
	title,
	body,
	status,
	created_at,
	updated_at
FROM compliance_news
WHERE
	%s
	AND trust_center_id = @trust_center_id
	AND status = 'SENT'
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
		return fmt.Errorf("cannot query sent compliance news: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ComplianceNews])
	if err != nil {
		return fmt.Errorf("cannot collect sent compliance news: %w", err)
	}

	*cnl = results
	return nil
}

func (cnl *ComplianceNewsItems) LoadByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
	cursor *page.Cursor[ComplianceNewsOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	trust_center_id,
	title,
	body,
	status,
	created_at,
	updated_at
FROM compliance_news
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
		return fmt.Errorf("cannot query compliance news: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ComplianceNews])
	if err != nil {
		return fmt.Errorf("cannot collect compliance news: %w", err)
	}

	*cnl = results
	return nil
}

func (cnl *ComplianceNewsItems) CountByTrustCenterID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	trustCenterID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(*)
FROM compliance_news
WHERE
	%s
	AND trust_center_id = @trust_center_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"trust_center_id": trustCenterID,
	}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count compliance news: %w", err)
	}

	return count, nil
}
