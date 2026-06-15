// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/iam/policy"
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

func (d *DocumentVersionApprovalDecision) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM document_version_approval_decisions WHERE id = ANY(@resource_ids::text[])`

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

func (d *DocumentVersionApprovalDecision) LoadByID(
	ctx context.Context,
	conn pg.Querier,
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
	conn pg.Querier,
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
	conn pg.Querier,
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
	conn pg.Querier,
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
	conn pg.Tx,
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
	updated_at,
	notification_count
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
	@updated_at,
	@notification_count
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
		"notification_count":      0,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "document_version_approval_decisions_quorum_id_approver_id_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert document version approval decision: %w", err)
	}

	return nil
}

func (ds DocumentVersionApprovalDecisions) BulkInsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	if len(ds) == 0 {
		return nil
	}

	rows := make([][]any, 0, len(ds))
	for _, d := range ds {
		rows = append(
			rows,
			[]any{
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
				0,
			},
		)
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
			"notification_count",
		},
		pgx.CopyFromRows(rows),
	)

	return err
}

// LoadNextDueGroupForNotification loads every still-PENDING approval decision
// for the next (organization, approver) group that has at least one decision due
// for a notification, so the whole group can be emailed together. A decision is
// due once it is past its scheduled offset (see documentNotificationMaxCount)
// and has not reached the cap. The receiver is left empty when no group is due.
//
// A decision is only ever PENDING while its quorum is pending — resolving a
// quorum either approves all decisions or voids the remaining ones — so there is
// no need to join the quorum table here.
func (d *DocumentVersionApprovalDecisions) LoadNextDueGroupForNotification(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	debounceBefore time.Time,
	reminderInterval time.Duration,
) error {
	q := `
WITH next_group AS (
	SELECT
		organization_id,
		approver_id
	FROM
		document_version_approval_decisions
	WHERE
		state = @state
		AND notification_count < @max_notifications
		AND (
			(notification_count = 0 AND created_at < @debounce_before)
			OR (notification_count > 0 AND last_notified_at < @now::timestamptz - make_interval(secs => @reminder_interval_seconds * notification_count))
		)
		AND EXISTS (
			SELECT 1
			FROM document_version_approval_quorums q
			JOIN document_versions dv ON dv.id = q.version_id
			JOIN documents doc ON doc.id = dv.document_id
			WHERE q.id = document_version_approval_decisions.quorum_id
				AND doc.deleted_at IS NULL
				AND doc.archived_at IS NULL
		)
	GROUP BY
		organization_id,
		approver_id
	ORDER BY
		organization_id,
		approver_id
	LIMIT 1
)
SELECT
	d.id,
	d.organization_id,
	d.quorum_id,
	d.approver_id,
	d.state,
	d.comment,
	d.electronic_signature_id,
	d.decided_at,
	d.created_at,
	d.updated_at
FROM
	document_version_approval_decisions d
INNER JOIN next_group g
	ON g.organization_id = d.organization_id
	AND g.approver_id = d.approver_id
WHERE
	d.state = @state
	AND d.notification_count < @max_notifications
	AND EXISTS (
		SELECT 1
		FROM document_version_approval_quorums q
		JOIN document_versions dv ON dv.id = q.version_id
		JOIN documents doc ON doc.id = dv.document_id
		WHERE q.id = d.quorum_id
			AND doc.deleted_at IS NULL
			AND doc.archived_at IS NULL
	)
ORDER BY
	d.quorum_id
`

	args := pgx.StrictNamedArgs{
		"state":                     DocumentVersionApprovalDecisionStatePending,
		"max_notifications":         documentNotificationMaxCount,
		"now":                       now,
		"debounce_before":           debounceBefore,
		"reminder_interval_seconds": reminderInterval.Seconds(),
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query due approval decisions for notification: %w", err)
	}

	decisions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentVersionApprovalDecision])
	if err != nil {
		return fmt.Errorf("cannot collect due approval decisions for notification: %w", err)
	}

	*d = decisions

	return nil
}

// ClaimForNotification claims the receiver's decisions that are individually
// due for a notification and returns their ids. The conditional update doubles
// as the claim, so concurrent workers never email the same group twice. Callers
// advance the rest of the group with BumpRemainingForNotification in the same
// transaction.
func (d DocumentVersionApprovalDecisions) ClaimForNotification(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
	debounceBefore time.Time,
	reminderInterval time.Duration,
) ([]gid.GID, error) {
	ids := make([]gid.GID, len(d))
	for i, decision := range d {
		ids[i] = decision.ID
	}

	q := `
UPDATE document_version_approval_decisions
SET
	notification_count = notification_count + 1,
	last_notified_at = @now
WHERE
	id = ANY(@ids::text[])
	AND state = @state
	AND notification_count < @max_notifications
	AND (
		(notification_count = 0 AND created_at < @debounce_before)
		OR (notification_count > 0 AND last_notified_at < @now::timestamptz - make_interval(secs => @reminder_interval_seconds * notification_count))
	)
RETURNING id
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{
		"ids":                       ids,
		"state":                     DocumentVersionApprovalDecisionStatePending,
		"max_notifications":         documentNotificationMaxCount,
		"now":                       now,
		"debounce_before":           debounceBefore,
		"reminder_interval_seconds": reminderInterval.Seconds(),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot claim due approval decisions for notification: %w", err)
	}

	claimed, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect claimed approval decisions for notification: %w", err)
	}

	return claimed, nil
}

// BumpRemainingForNotification advances the notification schedule for the
// still-pending decisions in the group that were not individually claimed, so
// the whole emailed list moves forward together. It must run in the same
// transaction as ClaimForNotification.
func (d DocumentVersionApprovalDecisions) BumpRemainingForNotification(
	ctx context.Context,
	conn pg.Tx,
	claimed []gid.GID,
	now time.Time,
) ([]gid.GID, error) {
	ids := make([]gid.GID, len(d))
	for i, decision := range d {
		ids[i] = decision.ID
	}

	rest := remainingNotificationIDs(ids, claimed)
	if len(rest) == 0 {
		return nil, nil
	}

	q := `
UPDATE document_version_approval_decisions
SET
	notification_count = notification_count + 1,
	last_notified_at = @now
WHERE
	id = ANY(@ids::text[])
	AND state = @state
	AND notification_count < @max_notifications
RETURNING id
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{
		"ids":               rest,
		"state":             DocumentVersionApprovalDecisionStatePending,
		"max_notifications": documentNotificationMaxCount,
		"now":               now,
	})
	if err != nil {
		return nil, fmt.Errorf("cannot bump remaining approval decisions for notification: %w", err)
	}

	bumped, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect bumped approval decisions for notification: %w", err)
	}

	return bumped, nil
}

func (d *DocumentVersionApprovalDecision) Update(
	ctx context.Context,
	conn pg.Tx,
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
	conn pg.Tx,
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

func (d *DocumentVersionApprovalDecisions) VoidPendingByQuorumID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	quorumID gid.GID,
	now time.Time,
) error {
	q := `
UPDATE document_version_approval_decisions
SET
	state = 'VOIDED',
	updated_at = @updated_at
WHERE
	%s
	AND quorum_id = @quorum_id
	AND state = 'PENDING'
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"quorum_id":  quorumID,
		"updated_at": now,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot void pending approval decisions: %w", err)
	}

	return nil
}

func (d *DocumentVersionApprovalDecisions) CountByQuorumID(
	ctx context.Context,
	conn pg.Querier,
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
