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
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	StatementOfApplicabilityDefaultApprover struct {
		StatementOfApplicabilityID gid.GID   `db:"statement_of_applicability_id"`
		ApproverProfileID          gid.GID   `db:"approver_profile_id"`
		OrganizationID             gid.GID   `db:"organization_id"`
		CreatedAt                  time.Time `db:"created_at"`
		UpdatedAt                  time.Time `db:"updated_at"`
	}

	StatementOfApplicabilityDefaultApprovers []*StatementOfApplicabilityDefaultApprover
)

// LoadByStatementOfApplicabilityID loads all default approvers for a statement of applicability.
func (das *StatementOfApplicabilityDefaultApprovers) LoadByStatementOfApplicabilityID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	statementOfApplicabilityID gid.GID,
) error {
	q := `
SELECT
	statement_of_applicability_id,
	approver_profile_id,
	organization_id,
	created_at,
	updated_at
FROM statement_of_applicability_default_approvers
WHERE
	%s
	AND statement_of_applicability_id = @statement_of_applicability_id
ORDER BY created_at ASC;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"statement_of_applicability_id": statementOfApplicabilityID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query statement of applicability default approvers: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[StatementOfApplicabilityDefaultApprover])
	if err != nil {
		return fmt.Errorf("cannot collect statement of applicability default approvers: %w", err)
	}

	*das = result
	return nil
}

// MergeByStatementOfApplicabilityID merges the given approver profile IDs for a statement of applicability,
// inserting new ones, keeping existing ones, and deleting removed ones.
func (das *StatementOfApplicabilityDefaultApprovers) MergeByStatementOfApplicabilityID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	statementOfApplicabilityID gid.GID,
	organizationID gid.GID,
	approverProfileIDs []gid.GID,
) error {
	q := `
MERGE INTO statement_of_applicability_default_approvers AS target
USING (
	SELECT unnest(@approver_profile_ids::text[]) AS approver_profile_id
) AS source
ON
	%s
	AND target.statement_of_applicability_id = @statement_of_applicability_id
	AND target.approver_profile_id = source.approver_profile_id
WHEN NOT MATCHED THEN
	INSERT (statement_of_applicability_id, approver_profile_id, tenant_id, organization_id, created_at, updated_at)
	VALUES (@statement_of_applicability_id, source.approver_profile_id, @tenant_id, @organization_id, @now, @now)
WHEN NOT MATCHED BY SOURCE
	AND %s
	AND target.statement_of_applicability_id = @statement_of_applicability_id THEN
	DELETE;
`

	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())

	now := time.Now()

	ids := make([]string, len(approverProfileIDs))
	for i, id := range approverProfileIDs {
		ids[i] = id.String()
	}

	args := pgx.StrictNamedArgs{
		"statement_of_applicability_id": statementOfApplicabilityID,
		"approver_profile_ids":          ids,
		"tenant_id":                     scope.GetTenantID(),
		"organization_id":               organizationID,
		"now":                           now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot merge statement of applicability default approvers: %w", err)
	}

	result := make(StatementOfApplicabilityDefaultApprovers, 0, len(approverProfileIDs))
	for _, profileID := range approverProfileIDs {
		result = append(result, &StatementOfApplicabilityDefaultApprover{
			StatementOfApplicabilityID: statementOfApplicabilityID,
			ApproverProfileID:          profileID,
			OrganizationID:             organizationID,
			CreatedAt:                  now,
			UpdatedAt:                  now,
		})
	}

	*das = result
	return nil
}
