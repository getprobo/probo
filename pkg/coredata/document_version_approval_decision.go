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
	DocumentVersionApprovalDecision struct {
		ID                    gid.GID                              `db:"id"`
		OrganizationID        gid.GID                              `db:"organization_id"`
		QuorumID              gid.GID                              `db:"quorum_id"`
		ApproverID            gid.GID                              `db:"approver_id"`
		State                 DocumentVersionApprovalDecisionState `db:"state"`
		Comment               *string                              `db:"comment"`
		ElectronicSignatureID *gid.GID                             `db:"electronic_signature_id"`
		DecidedAt             *time.Time                           `db:"decided_at"`
		CreatedAt             time.Time                            `db:"created_at"`
		UpdatedAt             time.Time                            `db:"updated_at"`
	}

	DocumentVersionApprovalDecisions []*DocumentVersionApprovalDecision
)

func (d DocumentVersionApprovalDecision) CursorKey(orderBy DocumentVersionApprovalDecisionOrderField) page.CursorKey {
	switch orderBy {
	case DocumentVersionApprovalDecisionOrderFieldCreatedAt:
		return page.NewCursorKey(d.ID, d.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (d *DocumentVersionApprovalDecision) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM document_version_approval_decisions WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, d.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query document version approval decision authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (d *DocumentVersionApprovalDecision) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	quorum_id,
	approver_id,
	state,
	comment,
	electronic_signature_id,
	decided_at,
	created_at,
	updated_at
FROM
	document_version_approval_decisions
WHERE
	id = @id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version approval decision: %w", err)
	}

	decision, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersionApprovalDecision])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect document version approval decision: %w", err)
	}

	*d = decision

	return nil
}

func (d *DocumentVersionApprovalDecision) LoadByQuorumIDAndApproverID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	quorumID gid.GID,
	approverID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	quorum_id,
	approver_id,
	state,
	comment,
	electronic_signature_id,
	decided_at,
	created_at,
	updated_at
FROM
	document_version_approval_decisions
WHERE
	%s
	AND quorum_id = @quorum_id
	AND approver_id = @approver_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"quorum_id":   quorumID,
		"approver_id": approverID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version approval decision: %w", err)
	}

	decision, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentVersionApprovalDecision])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect document version approval decision: %w", err)
	}

	*d = decision

	return nil
}

func (d *DocumentVersionApprovalDecisions) CountApprovedByQuorumID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	quorumID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	document_version_approval_decisions
WHERE
	%s
	AND quorum_id = @quorum_id
	AND state = 'APPROVED'
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"quorum_id": quorumID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (d *DocumentVersionApprovalDecisions) LoadByQuorumID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	quorumID gid.GID,
	cursor *page.Cursor[DocumentVersionApprovalDecisionOrderField],
	filter *DocumentVersionApprovalDecisionFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	quorum_id,
	approver_id,
	state,
	comment,
	electronic_signature_id,
	decided_at,
	created_at,
	updated_at
FROM
	document_version_approval_decisions
WHERE
	%s
	AND quorum_id = @quorum_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"quorum_id": quorumID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document version approval decisions: %w", err)
	}

	decisions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionApprovalDecision])
	if err != nil {
		return fmt.Errorf("cannot collect document version approval decisions: %w", err)
	}

	*d = decisions

	return nil
}

func (d *DocumentVersionApprovalDecision) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO document_version_approval_decisions (
	id,
	tenant_id,
	organization_id,
	quorum_id,
	approver_id,
	state,
	comment,
	electronic_signature_id,
	decided_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@quorum_id,
	@approver_id,
	@state,
	@comment,
	@electronic_signature_id,
	@decided_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                      d.ID,
		"tenant_id":               scope.GetTenantID(),
		"organization_id":         d.OrganizationID,
		"quorum_id":               d.QuorumID,
		"approver_id":             d.ApproverID,
		"state":                   d.State,
		"comment":                 d.Comment,
		"electronic_signature_id": d.ElectronicSignatureID,
		"decided_at":              d.DecidedAt,
		"created_at":              d.CreatedAt,
		"updated_at":              d.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert document version approval decision: %w", err)
	}

	return nil
}

func (ds DocumentVersionApprovalDecisions) BulkInsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	if len(ds) == 0 {
		return nil
	}

	rows := make([][]any, 0, len(ds))
	for _, d := range ds {
		rows = append(rows, []any{
			d.ID,
			scope.GetTenantID(),
			d.OrganizationID,
			d.QuorumID,
			d.ApproverID,
			d.State,
			d.Comment,
			d.ElectronicSignatureID,
			d.DecidedAt,
			d.CreatedAt,
			d.UpdatedAt,
		})
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"document_version_approval_decisions"},
		[]string{
			"id",
			"tenant_id",
			"organization_id",
			"quorum_id",
			"approver_id",
			"state",
			"comment",
			"electronic_signature_id",
			"decided_at",
			"created_at",
			"updated_at",
		},
		pgx.CopyFromRows(rows),
	)
	return err
}

func (d *DocumentVersionApprovalDecision) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE document_version_approval_decisions
SET
	state = @state,
	comment = @comment,
	electronic_signature_id = @electronic_signature_id,
	decided_at = @decided_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                      d.ID,
		"state":                   d.State,
		"comment":                 d.Comment,
		"electronic_signature_id": d.ElectronicSignatureID,
		"decided_at":              d.DecidedAt,
		"updated_at":              d.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document version approval decision: %w", err)
	}

	return nil
}

func (d *DocumentVersionApprovalDecision) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM document_version_approval_decisions
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": d.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete document version approval decision: %w", err)
	}

	return nil
}

func (d *DocumentVersionApprovalDecisions) CountByQuorumID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	quorumID gid.GID,
	filter *DocumentVersionApprovalDecisionFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	document_version_approval_decisions
WHERE
	%s
	AND quorum_id = @quorum_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"quorum_id": quorumID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}
