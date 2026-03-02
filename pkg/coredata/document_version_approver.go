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
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	DocumentVersionApprover struct {
		DocumentVersionID gid.GID      `db:"document_version_id"`
		ApproverProfileID gid.GID      `db:"approver_profile_id"`
		OrganizationID    gid.GID      `db:"organization_id"`
		TenantID          gid.TenantID `db:"tenant_id"`
		CreatedAt         time.Time    `db:"created_at"`
	}

	DocumentVersionApprovers []*DocumentVersionApprover
)

func (dva DocumentVersionApprover) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    document_version_approvers (
        document_version_id,
        approver_profile_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @document_version_id,
    @approver_profile_id,
    @organization_id,
    @tenant_id,
    @created_at
)
ON CONFLICT (document_version_id, approver_profile_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"document_version_id": dva.DocumentVersionID,
		"approver_profile_id": dva.ApproverProfileID,
		"organization_id":     dva.OrganizationID,
		"tenant_id":           scope.GetTenantID(),
		"created_at":          dva.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		return fmt.Errorf("cannot insert document version approver: %w", err)
	}

	return nil
}

func (dva *DocumentVersionApprovers) LoadByDocumentVersionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
) error {
	q := `
SELECT
    document_version_id,
    approver_profile_id,
    organization_id,
    tenant_id,
    created_at
FROM
    document_version_approvers
WHERE
    %s
    AND document_version_id = @document_version_id
ORDER BY created_at ASC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version approvers: %w", err)
	}

	approvers, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionApprover])
	if err != nil {
		return fmt.Errorf("cannot collect document version approvers: %w", err)
	}

	*dva = approvers

	return nil
}

func (dva *DocumentVersionApprovers) DeleteByDocumentVersionID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentVersionID gid.GID,
) error {
	q := `
DELETE FROM
    document_version_approvers
WHERE
    %s
    AND document_version_id = @document_version_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_version_id": documentVersionID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete document version approvers: %w", err)
	}

	return nil
}

func (dva *DocumentVersionApprovers) ApproverProfileIDs() []gid.GID {
	ids := make([]gid.GID, len(*dva))
	for i, a := range *dva {
		ids[i] = a.ApproverProfileID
	}
	return ids
}
