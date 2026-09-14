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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
)

const LinearWebhookEventDefaultMaxAttempts = 5

// LinearWebhookEvent is the Linear webhook inbox. Rows are unscoped because
// HTTP ingress has not resolved an organization yet.
type LinearWebhookEvent struct {
	DeliveryID           string          `db:"delivery_id"`
	Envelope             json.RawMessage `db:"envelope"`
	ProcessingOwnerToken *string         `db:"processing_owner_token"`
	ProcessingStartedAt  *time.Time      `db:"processing_started_at"`
	ProcessedAt          *time.Time      `db:"processed_at"`
	AttemptCount         int             `db:"attempt_count"`
	NextAttemptAt        *time.Time      `db:"next_attempt_at"`
	LastError            *string         `db:"last_error"`
	DeadLetteredAt       *time.Time      `db:"dead_lettered_at"`
	CreatedAt            time.Time       `db:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at"`
}

func NewLinearWebhookEvent(deliveryID string, envelope json.RawMessage) *LinearWebhookEvent {
	now := time.Now()

	return &LinearWebhookEvent{
		DeliveryID:   deliveryID,
		Envelope:     envelope,
		AttemptCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (e *LinearWebhookEvent) Insert(ctx context.Context, conn pg.Querier) (bool, error) {
	q := `
INSERT INTO linear_webhook_events (
    delivery_id,
    envelope,
    processing_owner_token,
    processing_started_at,
    processed_at,
    attempt_count,
    next_attempt_at,
    last_error,
    dead_lettered_at,
    created_at,
    updated_at
) VALUES (
    @delivery_id,
    @envelope,
    @processing_owner_token,
    @processing_started_at,
    @processed_at,
    @attempt_count,
    @next_attempt_at,
    @last_error,
    @dead_lettered_at,
    @created_at,
    @updated_at
)
ON CONFLICT (delivery_id) DO NOTHING
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"delivery_id":            e.DeliveryID,
			"envelope":               e.Envelope,
			"processing_owner_token": e.ProcessingOwnerToken,
			"processing_started_at":  e.ProcessingStartedAt,
			"processed_at":           e.ProcessedAt,
			"attempt_count":          e.AttemptCount,
			"next_attempt_at":        e.NextAttemptAt,
			"last_error":             e.LastError,
			"dead_lettered_at":       e.DeadLetteredAt,
			"created_at":             e.CreatedAt,
			"updated_at":             e.UpdatedAt,
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot insert Linear webhook event: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (e *LinearWebhookEvent) ClaimNextForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
WITH candidate AS (
    SELECT delivery_id
    FROM linear_webhook_events
    WHERE processed_at IS NULL
        AND processing_started_at IS NULL
        AND processing_owner_token IS NULL
        AND dead_lettered_at IS NULL
        AND attempt_count < @max_attempts
        AND (next_attempt_at IS NULL OR next_attempt_at <= @now)
    ORDER BY created_at ASC, delivery_id ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
UPDATE linear_webhook_events e
SET
    attempt_count = e.attempt_count + 1,
    processing_owner_token = @owner_token,
    processing_started_at = @now,
    updated_at = @now
FROM candidate
WHERE e.delivery_id = candidate.delivery_id
RETURNING
    e.delivery_id,
    e.envelope,
    e.processing_owner_token,
    e.processing_started_at,
    e.processed_at,
    e.attempt_count,
    e.next_attempt_at,
    e.last_error,
    e.dead_lettered_at,
    e.created_at,
    e.updated_at
`

	return e.loadExactlyOne(
		ctx,
		conn,
		q,
		pgx.StrictNamedArgs{
			"now":          now,
			"owner_token":  uuid.MustNewV4().String(),
			"max_attempts": LinearWebhookEventDefaultMaxAttempts,
		},
	)
}

func (e *LinearWebhookEvent) UpdateProcessingState(ctx context.Context, conn pg.Querier) error {
	if e.ProcessingOwnerToken == nil || *e.ProcessingOwnerToken == "" {
		panic("Linear webhook event processing update requires an owner token")
	}

	q := `
UPDATE linear_webhook_events
SET
    processing_owner_token = NULL,
    processing_started_at = @processing_started_at,
    processed_at = @processed_at,
    attempt_count = @attempt_count,
    next_attempt_at = @next_attempt_at,
    last_error = @last_error,
    dead_lettered_at = @dead_lettered_at,
    updated_at = @updated_at
WHERE delivery_id = @delivery_id
    AND processing_owner_token = @processing_owner_token
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"delivery_id":            e.DeliveryID,
			"processing_owner_token": e.ProcessingOwnerToken,
			"processing_started_at":  e.ProcessingStartedAt,
			"processed_at":           e.ProcessedAt,
			"attempt_count":          e.AttemptCount,
			"next_attempt_at":        e.NextAttemptAt,
			"last_error":             e.LastError,
			"dead_lettered_at":       e.DeadLetteredAt,
			"updated_at":             e.UpdatedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot update Linear webhook event processing state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProcessingLeaseLost
	}

	return nil
}

func ResetStaleProcessingLinearWebhookEvents(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	staleAfter time.Duration,
) error {
	q := `
UPDATE linear_webhook_events
SET
    processing_owner_token = NULL,
    processing_started_at = NULL,
    next_attempt_at = CASE
        WHEN attempt_count >= @max_attempts THEN NULL
        ELSE @now::timestamptz
    END,
    dead_lettered_at = CASE
        WHEN attempt_count >= @max_attempts THEN @now::timestamptz
        ELSE NULL
    END,
    last_error = @stale_error,
    updated_at = @now
WHERE processed_at IS NULL
    AND processing_started_at IS NOT NULL
    AND processing_started_at < @stale_threshold
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"now":             now,
			"stale_threshold": now.Add(-staleAfter),
			"stale_error":     "Linear webhook event processing lease expired",
			"max_attempts":    LinearWebhookEventDefaultMaxAttempts,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot reset stale Linear webhook events: %w", err)
	}

	return nil
}

func (e *LinearWebhookEvent) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Linear webhook event: %w", err)
	}

	event, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[LinearWebhookEvent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect Linear webhook event: %w", err)
	}

	*e = event

	return nil
}
