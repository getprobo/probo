// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	DocumentDefaultApprover struct {
		DocumentID        gid.GID   `db:"document_id"`
		ApproverProfileID gid.GID   `db:"approver_profile_id"`
		OrganizationID    gid.GID   `db:"organization_id"`
		CreatedAt         time.Time `db:"created_at"`
		UpdatedAt         time.Time `db:"updated_at"`
	}

	DocumentDefaultApprovers []*DocumentDefaultApprover
)

// LoadByDocumentID loads all default approvers for a document.
func (das *DocumentDefaultApprovers) LoadByDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
) error {
	q := `
SELECT
	document_id,
	approver_profile_id,
	organization_id,
	created_at,
	updated_at
FROM document_default_approvers
WHERE
	%s
	AND document_id = @document_id
ORDER BY created_at ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document default approvers: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentDefaultApprover])
	if err != nil {
		return fmt.Errorf("cannot collect document default approvers: %w", err)
	}

	*das = result
	return nil
}

// MergeByDocumentID merges the given approver profile IDs for a document,
// inserting new ones, keeping existing ones, and deleting removed ones.
func (das *DocumentDefaultApprovers) MergeByDocumentID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	approverProfileIDs []gid.GID,
) error {
	q := `
MERGE INTO document_default_approvers AS target
USING (
	SELECT unnest(@approver_profile_ids::text[]) AS approver_profile_id
) AS source
ON
	%s
	AND target.document_id = @document_id
	AND target.approver_profile_id = source.approver_profile_id
WHEN NOT MATCHED THEN
	INSERT (document_id, approver_profile_id, tenant_id, organization_id, created_at, updated_at)
	VALUES (@document_id, source.approver_profile_id, @tenant_id, @organization_id, @now, @now)
WHEN NOT MATCHED BY SOURCE
	AND %s
	AND target.document_id = @document_id THEN
	DELETE;
`

	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())

	now := time.Now()

	ids := make([]string, len(approverProfileIDs))
	for i, id := range approverProfileIDs {
		ids[i] = id.String()
	}

	args := pgx.StrictNamedArgs{
		"document_id":          documentID,
		"approver_profile_ids": ids,
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      organizationID,
		"now":                  now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot merge document default approvers: %w", err)
	}

	result := make(DocumentDefaultApprovers, 0, len(approverProfileIDs))
	for _, profileID := range approverProfileIDs {
		result = append(result, &DocumentDefaultApprover{
			DocumentID:        documentID,
			ApproverProfileID: profileID,
			OrganizationID:    organizationID,
			CreatedAt:         now,
			UpdatedAt:         now,
		})
	}

	*das = result
	return nil
}
