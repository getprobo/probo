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
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	WebhookEvent struct {
		ID                    gid.GID            `db:"id"`
		WebhookDataID         gid.GID            `db:"webhook_data_id"`
		WebhookSubscriptionID gid.GID            `db:"webhook_subscription_id"`
		Status                WebhookEventStatus `db:"status"`
		Response              json.RawMessage    `db:"response"`
		ProcessingOwnerToken  *string            `db:"processing_owner_token"`
		ProcessingStartedAt   *time.Time         `db:"processing_started_at"`
		AttemptCount          int                `db:"attempt_count"`
		MaxAttempts           int                `db:"max_attempts"`
		NextAttemptAt         *time.Time         `db:"next_attempt_at"`
		LastError             *string            `db:"last_error"`
		CompletedAt           *time.Time         `db:"completed_at"`
		DeadLetteredAt        *time.Time         `db:"dead_lettered_at"`
		CreatedAt             time.Time          `db:"created_at"`
		UpdatedAt             time.Time          `db:"updated_at"`
	}

	WebhookEvents []*WebhookEvent
)

const WebhookEventDefaultMaxAttempts = 12

func (w WebhookEvent) CursorKey(orderBy WebhookEventOrderField) page.CursorKey {
	switch orderBy {
	case WebhookEventOrderFieldCreatedAt:
		return page.NewCursorKey(w.ID, w.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (w *WebhookEvents) LoadBySubscriptionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	webhookSubscriptionID gid.GID,
	cursor *page.Cursor[WebhookEventOrderField],
) error {
	q := `
SELECT
    id,
    webhook_data_id,
    webhook_subscription_id,
    status,
    response,
    processing_owner_token,
    processing_started_at,
    attempt_count,
    max_attempts,
    next_attempt_at,
    last_error,
    completed_at,
    dead_lettered_at,
    created_at,
    updated_at
FROM
    webhook_events
WHERE
    %s
    AND webhook_subscription_id = @webhook_subscription_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"webhook_subscription_id": webhookSubscriptionID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query webhook events: %w", err)
	}

	events, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[WebhookEvent])
	if err != nil {
		return fmt.Errorf("cannot collect webhook events: %w", err)
	}

	*w = events

	return nil
}

func (w *WebhookEvents) CountBySubscriptionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	webhookSubscriptionID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(*)
FROM webhook_events
WHERE %s
    AND webhook_subscription_id = @webhook_subscription_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"webhook_subscription_id": webhookSubscriptionID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count webhook events: %w", err)
	}

	return count, nil
}

func (w *WebhookEvent) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    webhook_data_id,
    webhook_subscription_id,
    status,
    response,
    processing_owner_token,
    processing_started_at,
    attempt_count,
    max_attempts,
    next_attempt_at,
    last_error,
    completed_at,
    dead_lettered_at,
    created_at,
    updated_at
FROM webhook_events
WHERE %s
    AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query webhook event: %w", err)
	}

	event, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[WebhookEvent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect webhook event: %w", err)
	}

	*w = event

	return nil
}

func (w *WebhookEvent) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO webhook_events (
    id,
    tenant_id,
    webhook_data_id,
    webhook_subscription_id,
    status,
    response,
    processing_owner_token,
    processing_started_at,
    attempt_count,
    max_attempts,
    next_attempt_at,
    last_error,
    completed_at,
    dead_lettered_at,
    created_at,
    updated_at
)
VALUES (
    @id,
    @tenant_id,
    @webhook_data_id,
    @webhook_subscription_id,
    @status,
    @response,
    @processing_owner_token,
    @processing_started_at,
    @attempt_count,
    @max_attempts,
    @next_attempt_at,
    @last_error,
    @completed_at,
    @dead_lettered_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                      w.ID,
		"tenant_id":               scope.GetTenantID(),
		"webhook_data_id":         w.WebhookDataID,
		"webhook_subscription_id": w.WebhookSubscriptionID,
		"status":                  w.Status,
		"response":                w.Response,
		"processing_owner_token":  w.ProcessingOwnerToken,
		"processing_started_at":   w.ProcessingStartedAt,
		"attempt_count":           w.AttemptCount,
		"max_attempts":            w.MaxAttempts,
		"next_attempt_at":         w.NextAttemptAt,
		"last_error":              w.LastError,
		"completed_at":            w.CompletedAt,
		"dead_lettered_at":        w.DeadLetteredAt,
		"created_at":              w.CreatedAt,
		"updated_at":              w.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert webhook event: %w", err)
	}

	return nil
}

func (w *WebhookEvent) ClaimNextForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
WITH candidate AS (
    SELECT id
    FROM webhook_events
    WHERE status = @pending_status
        AND completed_at IS NULL
        AND dead_lettered_at IS NULL
        AND processing_started_at IS NULL
        AND processing_owner_token IS NULL
        AND attempt_count < max_attempts
        AND (next_attempt_at IS NULL OR next_attempt_at <= @now)
    ORDER BY created_at ASC, id ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
UPDATE webhook_events event
SET
    attempt_count = event.attempt_count + 1,
    processing_owner_token = @owner_token,
    processing_started_at = @now,
    updated_at = @now
FROM candidate
WHERE event.id = candidate.id
RETURNING
    event.id,
    event.webhook_data_id,
    event.webhook_subscription_id,
    event.status,
    event.response,
    event.processing_owner_token,
    event.processing_started_at,
    event.attempt_count,
    event.max_attempts,
    event.next_attempt_at,
    event.last_error,
    event.completed_at,
    event.dead_lettered_at,
    event.created_at,
    event.updated_at
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"pending_status": WebhookEventStatusPending,
			"now":            now,
			"owner_token":    uuid.MustNewV4().String(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot claim webhook event: %w", err)
	}

	event, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[WebhookEvent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect claimed webhook event: %w", err)
	}

	*w = event

	return nil
}

func (w *WebhookEvent) UpdateDeliveryState(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	if w.ProcessingOwnerToken == nil || *w.ProcessingOwnerToken == "" {
		panic("webhook event update requires an owner token")
	}

	q := `
UPDATE webhook_events
SET
    status = @status,
    response = @response,
    processing_owner_token = NULL,
    processing_started_at = @processing_started_at,
    next_attempt_at = @next_attempt_at,
    last_error = @last_error,
    completed_at = @completed_at,
    dead_lettered_at = @dead_lettered_at,
    updated_at = @updated_at
WHERE %s
    AND id = @id
    AND processing_owner_token = @processing_owner_token
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                     w.ID,
		"status":                 w.Status,
		"response":               w.Response,
		"processing_owner_token": w.ProcessingOwnerToken,
		"processing_started_at":  w.ProcessingStartedAt,
		"next_attempt_at":        w.NextAttemptAt,
		"last_error":             w.LastError,
		"completed_at":           w.CompletedAt,
		"dead_lettered_at":       w.DeadLetteredAt,
		"updated_at":             w.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update webhook event delivery state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProcessingLeaseLost
	}

	return nil
}

func ResetStaleWebhookEvents(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	staleAfter time.Duration,
) error {
	q := `
UPDATE webhook_events
SET
    status = CASE
        WHEN attempt_count >= max_attempts THEN @failed_status::webhook_event_status
        ELSE @pending_status::webhook_event_status
    END,
    processing_owner_token = NULL,
    processing_started_at = NULL,
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
			"pending_status":  WebhookEventStatusPending,
			"failed_status":   WebhookEventStatusFailed,
			"now":             now,
			"stale_threshold": now.Add(-staleAfter),
			"last_error":      "Webhook delivery lease expired",
		},
	)
	if err != nil {
		return fmt.Errorf("cannot reset stale webhook events: %w", err)
	}

	return nil
}
