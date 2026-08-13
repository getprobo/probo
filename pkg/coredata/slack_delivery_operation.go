// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package coredata

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	SlackDeliveryOperationKind string

	SlackDeliveryOperation struct {
		ID                  gid.GID                    `db:"id"`
		OrganizationID      gid.GID                    `db:"organization_id"`
		OperationKey        string                     `db:"operation_key"`
		OperationKind       SlackDeliveryOperationKind `db:"operation_kind"`
		Payload             map[string]any             `db:"payload"`
		ClientMsgID         *string                    `db:"client_msg_id"`
		ProcessingStartedAt *time.Time                 `db:"processing_started_at"`
		CompletedAt         *time.Time                 `db:"completed_at"`
		AttemptCount        int                        `db:"attempt_count"`
		MaxAttempts         int                        `db:"max_attempts"`
		NextAttemptAt       *time.Time                 `db:"next_attempt_at"`
		LastError           *string                    `db:"last_error"`
		DeadLetteredAt      *time.Time                 `db:"dead_lettered_at"`
		CreatedAt           time.Time                  `db:"created_at"`
		UpdatedAt           time.Time                  `db:"updated_at"`
	}
)

const (
	SlackDeliveryOperationKindPostMessage    SlackDeliveryOperationKind = "POST_MESSAGE"
	SlackDeliveryOperationKindAddReaction    SlackDeliveryOperationKind = "ADD_REACTION"
	SlackDeliveryOperationDefaultMaxAttempts                            = 5
)

var (
	_ fmt.Stringer             = SlackDeliveryOperationKind("")
	_ encoding.TextMarshaler   = SlackDeliveryOperationKind("")
	_ encoding.TextUnmarshaler = (*SlackDeliveryOperationKind)(nil)
)

func (v SlackDeliveryOperationKind) String() string { return string(v) }

func (v SlackDeliveryOperationKind) IsValid() bool {
	switch v {
	case SlackDeliveryOperationKindPostMessage, SlackDeliveryOperationKindAddReaction:
		return true
	}

	return false
}

func (v SlackDeliveryOperationKind) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *SlackDeliveryOperationKind) UnmarshalText(text []byte) error {
	value := SlackDeliveryOperationKind(text)
	if !value.IsValid() {
		return fmt.Errorf("invalid SlackDeliveryOperationKind value: %q", text)
	}

	*v = value

	return nil
}

func NewSlackDeliveryOperation(
	scope Scoper,
	organizationID gid.GID,
	operationKey string,
	kind SlackDeliveryOperationKind,
	payload map[string]any,
) *SlackDeliveryOperation {
	now := time.Now()

	operation := &SlackDeliveryOperation{
		ID:             gid.New(scope.GetTenantID(), SlackDeliveryOperationEntityType),
		OrganizationID: organizationID,
		OperationKey:   operationKey,
		OperationKind:  kind,
		Payload:        payload,
		MaxAttempts:    SlackDeliveryOperationDefaultMaxAttempts,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if kind == SlackDeliveryOperationKindPostMessage {
		operation.ClientMsgID = new(uuid.MustNewV4().String())
	}

	return operation
}

func (o *SlackDeliveryOperation) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO slack_delivery_operations (
	id, tenant_id, organization_id, operation_key, operation_kind, payload,
	client_msg_id, processing_started_at, completed_at, attempt_count,
	max_attempts, next_attempt_at, last_error, dead_lettered_at, created_at,
	updated_at
) VALUES (
	@id, @tenant_id, @organization_id, @operation_key, @operation_kind, @payload,
	@client_msg_id, @processing_started_at, @completed_at, @attempt_count,
	@max_attempts, @next_attempt_at, @last_error, @dead_lettered_at, @created_at,
	@updated_at
)
ON CONFLICT (organization_id, operation_key) DO UPDATE
SET operation_key = slack_delivery_operations.operation_key
RETURNING
	id, organization_id, operation_key, operation_kind, payload, client_msg_id,
	processing_started_at, completed_at, attempt_count, max_attempts,
	next_attempt_at, last_error, dead_lettered_at, created_at, updated_at
`
	originalID := o.ID

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":                    o.ID,
			"tenant_id":             scope.GetTenantID(),
			"organization_id":       o.OrganizationID,
			"operation_key":         o.OperationKey,
			"operation_kind":        o.OperationKind,
			"payload":               o.Payload,
			"client_msg_id":         o.ClientMsgID,
			"processing_started_at": o.ProcessingStartedAt,
			"completed_at":          o.CompletedAt,
			"attempt_count":         o.AttemptCount,
			"max_attempts":          o.MaxAttempts,
			"next_attempt_at":       o.NextAttemptAt,
			"last_error":            o.LastError,
			"dead_lettered_at":      o.DeadLetteredAt,
			"created_at":            o.CreatedAt,
			"updated_at":            o.UpdatedAt,
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot upsert Slack delivery operation: %w", err)
	}
	defer rows.Close()

	operation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackDeliveryOperation])
	if err != nil {
		return false, fmt.Errorf("cannot collect Slack delivery operation upsert: %w", err)
	}

	*o = operation

	return originalID == o.ID, nil
}

func (o *SlackDeliveryOperation) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT id, organization_id, operation_key, operation_kind, payload,
	client_msg_id, processing_started_at, completed_at, attempt_count,
	max_attempts, next_attempt_at, last_error, dead_lettered_at, created_at,
	updated_at
FROM slack_delivery_operations
WHERE %s AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	return o.loadExactlyOne(ctx, conn, q, args)
}

func (o *SlackDeliveryOperation) ClaimNextForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
WITH candidate AS (
	SELECT id
	FROM slack_delivery_operations
	WHERE completed_at IS NULL
		AND dead_lettered_at IS NULL
		AND processing_started_at IS NULL
		AND attempt_count < max_attempts
		AND (next_attempt_at IS NULL OR next_attempt_at <= @now)
	ORDER BY created_at ASC, id ASC
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
UPDATE slack_delivery_operations operation
SET attempt_count = operation.attempt_count + 1,
	processing_started_at = @now,
	updated_at = @now
FROM candidate
WHERE operation.id = candidate.id
RETURNING operation.id, operation.organization_id, operation.operation_key,
	operation.operation_kind, operation.payload, operation.client_msg_id,
	operation.processing_started_at, operation.completed_at,
	operation.attempt_count, operation.max_attempts, operation.next_attempt_at,
	operation.last_error, operation.dead_lettered_at, operation.created_at,
	operation.updated_at
`

	return o.loadExactlyOne(ctx, conn, q, pgx.StrictNamedArgs{"now": now})
}

func (o *SlackDeliveryOperation) UpdateDeliveryState(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
UPDATE slack_delivery_operations
SET processing_started_at = @processing_started_at,
	completed_at = @completed_at,
	next_attempt_at = @next_attempt_at,
	last_error = @last_error,
	dead_lettered_at = @dead_lettered_at,
	updated_at = @updated_at
WHERE %s AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":                    o.ID,
		"processing_started_at": o.ProcessingStartedAt,
		"completed_at":          o.CompletedAt,
		"next_attempt_at":       o.NextAttemptAt,
		"last_error":            o.LastError,
		"dead_lettered_at":      o.DeadLetteredAt,
		"updated_at":            o.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update Slack delivery operation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func ResetStaleSlackDeliveryOperations(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	staleAfter time.Duration,
) error {
	q := `
UPDATE slack_delivery_operations
SET processing_started_at = NULL,
	next_attempt_at = CASE
		WHEN attempt_count >= max_attempts THEN NULL
		ELSE @now::timestamptz
	END,
	dead_lettered_at = CASE
		WHEN attempt_count >= max_attempts THEN @now::timestamptz
		ELSE NULL
	END,
	last_error = @last_error,
	updated_at = @now
WHERE completed_at IS NULL
	AND dead_lettered_at IS NULL
	AND processing_started_at < @stale_threshold
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"now":             now,
			"stale_threshold": now.Add(-staleAfter),
			"last_error":      "Slack delivery operation lease expired",
		},
	)
	if err != nil {
		return fmt.Errorf("cannot reset stale Slack delivery operations: %w", err)
	}

	return nil
}

func DeleteCompletedSlackDeliveryOperationsBefore(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM slack_delivery_operations
	WHERE completed_at IS NOT NULL
		AND completed_at < @before
	ORDER BY completed_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM slack_delivery_operations
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"before": before,
			"limit":  limit,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete completed Slack delivery operations: %w", err)
	}

	return result.RowsAffected(), nil
}

func DeleteDeadLetteredSlackDeliveryOperationsBefore(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM slack_delivery_operations
	WHERE dead_lettered_at IS NOT NULL
		AND dead_lettered_at < @before
	ORDER BY dead_lettered_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM slack_delivery_operations
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{"before": before, "limit": limit},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete dead-lettered Slack delivery operations: %w", err)
	}

	return result.RowsAffected(), nil
}

func (o *SlackDeliveryOperation) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Slack delivery operation: %w", err)
	}

	operation, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackDeliveryOperation])
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrResourceNotFound
	}

	if err != nil {
		return fmt.Errorf("cannot collect Slack delivery operation: %w", err)
	}

	*o = operation

	return nil
}
