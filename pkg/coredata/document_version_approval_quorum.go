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
	"go.probo.inc/probo/pkg/page"
)

type (
	DocumentVersionApprovalQuorum struct {
		ID             gid.GID                             `db:"id"`
		OrganizationID gid.GID                             `db:"organization_id"`
		VersionID      gid.GID                             `db:"version_id"`
		Status         DocumentVersionApprovalQuorumStatus `db:"status"`
		CreatedAt      time.Time                           `db:"created_at"`
		UpdatedAt      time.Time                           `db:"updated_at"`
	}

	DocumentVersionApprovalQuorums []*DocumentVersionApprovalQuorum
)

func (q DocumentVersionApprovalQuorum) CursorKey(orderBy DocumentVersionApprovalQuorumOrderField) page.CursorKey {
	switch orderBy {
	case DocumentVersionApprovalQuorumOrderFieldCreatedAt:
		return page.NewCursorKey(q.ID, q.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (q *DocumentVersionApprovalQuorum) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	query := `SELECT organization_id FROM document_version_approval_quorums WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, query, q.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query approval quorum authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (q *DocumentVersionApprovalQuorum) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	query := `
SELECT
	id,
	organization_id,
	version_id,
	status,
	created_at,
	updated_at
FROM
	document_version_approval_quorums
WHERE
	id = @id
	AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query approval quorum: %w", err)
	}

	quorum, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersionApprovalQuorum])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect approval quorum: %w", err)
	}

	*q = quorum

	return nil
}

func (q *DocumentVersionApprovalQuorum) LoadLastByDocumentVersionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
) error {
	query := `
SELECT
	id,
	organization_id,
	version_id,
	status,
	created_at,
	updated_at
FROM
	document_version_approval_quorums
WHERE
	%s
	AND version_id = @version_id
ORDER BY created_at DESC
LIMIT 1
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query last approval quorum: %w", err)
	}

	quorum, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersionApprovalQuorum])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect last approval quorum: %w", err)
	}

	*q = quorum

	return nil
}

func (q *DocumentVersionApprovalQuorums) LoadAllByDocumentVersionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
	cursor *page.Cursor[DocumentVersionApprovalQuorumOrderField],
) error {
	query := `
SELECT
	id,
	organization_id,
	version_id,
	status,
	created_at,
	updated_at
FROM
	document_version_approval_quorums
WHERE
	%s
	AND version_id = @version_id
	AND %s
`

	query = fmt.Sprintf(query, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query approval quorums: %w", err)
	}

	quorums, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionApprovalQuorum])
	if err != nil {
		return fmt.Errorf("cannot collect approval quorums: %w", err)
	}

	*q = quorums

	return nil
}

func (q *DocumentVersionApprovalQuorums) CountByDocumentVersionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
) (int, error) {
	query := `
SELECT
	COUNT(id)
FROM
	document_version_approval_quorums
WHERE
	%s
	AND version_id = @version_id
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, query, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (q *DocumentVersionApprovalQuorum) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	query := `
INSERT INTO document_version_approval_quorums (
	id,
	tenant_id,
	organization_id,
	version_id,
	status,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@version_id,
	@status,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              q.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": q.OrganizationID,
		"version_id":      q.VersionID,
		"status":          q.Status,
		"created_at":      q.CreatedAt,
		"updated_at":      q.UpdatedAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert approval quorum: %w", err)
	}

	return nil
}

func (q *DocumentVersionApprovalQuorum) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	query := `
DELETE FROM document_version_approval_quorums
WHERE
	%s
	AND id = @id
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": q.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot delete approval quorum: %w", err)
	}

	return nil
}

func (q *DocumentVersionApprovalQuorum) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	query := `
UPDATE document_version_approval_quorums
SET
	status = @status,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         q.ID,
		"status":     q.Status,
		"updated_at": q.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot update approval quorum: %w", err)
	}

	return nil
}
