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
	"go.probo.inc/probo/pkg/gid"
)

const SlackbotEventDefaultMaxAttempts = 5

// SlackbotEvent is the Slack event inbox. Rows are unscoped because HTTP
// ingress has not resolved an organization yet.
type SlackbotEvent struct {
	ID                   gid.GID         `db:"id"`
	EventID              string          `db:"event_id"`
	Envelope             json.RawMessage `db:"envelope"`
	ProcessingOwnerToken *string         `db:"processing_owner_token"`
	ProcessingStartedAt  *time.Time      `db:"processing_started_at"`
	ProcessedAt          *time.Time      `db:"processed_at"`
	AttemptCount         int             `db:"attempt_count"`
	MaxAttempts          int             `db:"max_attempts"`
	NextAttemptAt        *time.Time      `db:"next_attempt_at"`
	LastError            *string         `db:"last_error"`
	DeadLetteredAt       *time.Time      `db:"dead_lettered_at"`
	CreatedAt            time.Time       `db:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at"`
}

func NewSlackbotEvent(eventID string, envelope json.RawMessage) *SlackbotEvent {
	now := time.Now()

	return &SlackbotEvent{
		ID:          gid.New(gid.NilTenant, SlackbotEventEntityType),
		EventID:     eventID,
		Envelope:    envelope,
		MaxAttempts: SlackbotEventDefaultMaxAttempts,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (e *SlackbotEvent) Insert(ctx context.Context, conn pg.Querier) (bool, error) {
	q := `
INSERT INTO slackbot_events (
	id,
	event_id,
	envelope,
	processing_owner_token,
	processing_started_at,
	processed_at,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@event_id,
	@envelope,
	@processing_owner_token,
	@processing_started_at,
	@processed_at,
	@attempt_count,
	@max_attempts,
	@next_attempt_at,
	@last_error,
	@dead_lettered_at,
	@created_at,
	@updated_at
)
ON CONFLICT (event_id) DO NOTHING
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":                     e.ID,
			"event_id":               e.EventID,
			"envelope":               e.Envelope,
			"processing_owner_token": e.ProcessingOwnerToken,
			"processing_started_at":  e.ProcessingStartedAt,
			"processed_at":           e.ProcessedAt,
			"attempt_count":          e.AttemptCount,
			"max_attempts":           e.MaxAttempts,
			"next_attempt_at":        e.NextAttemptAt,
			"last_error":             e.LastError,
			"dead_lettered_at":       e.DeadLetteredAt,
			"created_at":             e.CreatedAt,
			"updated_at":             e.UpdatedAt,
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot insert Slackbot event: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (e *SlackbotEvent) LoadByEventID(
	ctx context.Context,
	conn pg.Querier,
	eventID string,
) error {
	q := `
SELECT
	id,
	event_id,
	envelope,
	processing_owner_token,
	processing_started_at,
	processed_at,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
FROM slackbot_events
WHERE event_id = @event_id
LIMIT 1
`

	return e.loadExactlyOne(ctx, conn, q, pgx.StrictNamedArgs{"event_id": eventID})
}

func (e *SlackbotEvent) ClaimNextForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
WITH candidate AS (
	SELECT id
	FROM slackbot_events
	WHERE processed_at IS NULL
		AND processing_started_at IS NULL
		AND processing_owner_token IS NULL
		AND dead_lettered_at IS NULL
		AND attempt_count < max_attempts
		AND (next_attempt_at IS NULL OR next_attempt_at <= @now)
	ORDER BY created_at ASC, id ASC
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
UPDATE slackbot_events e
SET
	attempt_count = e.attempt_count + 1,
	processing_owner_token = @owner_token,
	processing_started_at = @now,
	updated_at = @now
FROM candidate
WHERE e.id = candidate.id
RETURNING
	e.id,
	e.event_id,
	e.envelope,
	e.processing_owner_token,
	e.processing_started_at,
	e.processed_at,
	e.attempt_count,
	e.max_attempts,
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
			"now":         now,
			"owner_token": uuid.MustNewV4().String(),
		},
	)
}

func (e *SlackbotEvent) UpdateProcessingState(ctx context.Context, conn pg.Querier) error {
	if e.ProcessingOwnerToken == nil || *e.ProcessingOwnerToken == "" {
		panic("Slackbot event processing update requires an owner token")
	}

	q := `
UPDATE slackbot_events
SET
	processing_owner_token = NULL,
	processing_started_at = @processing_started_at,
	processed_at = @processed_at,
	attempt_count = @attempt_count,
	next_attempt_at = @next_attempt_at,
	last_error = @last_error,
	dead_lettered_at = @dead_lettered_at,
	updated_at = @updated_at
WHERE id = @id
	AND processing_owner_token = @processing_owner_token
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":                     e.ID,
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
		return fmt.Errorf("cannot update Slackbot event processing state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProcessingLeaseLost
	}

	return nil
}

func (e *SlackbotEvent) LoadDeadLetteredByEventID(
	ctx context.Context,
	conn pg.Querier,
	eventID string,
) error {
	q := `
SELECT
	id,
	event_id,
	envelope,
	processing_owner_token,
	processing_started_at,
	processed_at,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
FROM slackbot_events
WHERE
	event_id = @event_id
	AND dead_lettered_at IS NOT NULL
LIMIT 1
`

	return e.loadExactlyOne(ctx, conn, q, pgx.StrictNamedArgs{"event_id": eventID})
}

func (e *SlackbotEvent) Delete(ctx context.Context, conn pg.Querier) error {
	q := `DELETE FROM slackbot_events WHERE id = @id`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"id": e.ID})
	if err != nil {
		return fmt.Errorf("cannot delete Slackbot event: %w", err)
	}

	return nil
}

func DeleteProcessedSlackbotEventsBefore(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
) error {
	_, err := DeleteProcessedSlackbotEventsBeforeBatch(ctx, conn, before, 1000)

	return err
}

func DeleteProcessedSlackbotEventsBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM slackbot_events
	WHERE processed_at IS NOT NULL
		AND processed_at < @before
	ORDER BY processed_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM slackbot_events
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
		return 0, fmt.Errorf("cannot delete processed Slackbot events: %w", err)
	}

	return result.RowsAffected(), nil
}

func DeleteDeadLetteredSlackbotEventsBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM slackbot_events
	WHERE dead_lettered_at IS NOT NULL
		AND dead_lettered_at < @before
	ORDER BY dead_lettered_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM slackbot_events
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{"before": before, "limit": limit},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete dead-lettered Slackbot events: %w", err)
	}

	return result.RowsAffected(), nil
}

func ResetStaleProcessingSlackbotEvents(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	staleAfter time.Duration,
) error {
	q := `
UPDATE slackbot_events
SET
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
			"stale_error":     "Slackbot event processing lease expired",
		},
	)
	if err != nil {
		return fmt.Errorf("cannot reset stale processing Slackbot events: %w", err)
	}

	return nil
}

func (e *SlackbotEvent) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Slackbot event: %w", err)
	}

	event, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackbotEvent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect Slackbot event: %w", err)
	}

	*e = event

	return nil
}
