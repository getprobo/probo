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
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

func (e *SlackbotProcessedEvent) Exists(ctx context.Context, conn pg.Querier) (bool, error) {
	q := `SELECT event_id, created_at FROM slackbot_processed_events WHERE event_id = @event_id`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"event_id": e.EventID})
	if err != nil {
		return false, fmt.Errorf("cannot query slackbot processed event: %w", err)
	}

	event, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackbotProcessedEvent])
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("cannot collect slackbot processed event: %w", err)
	}

	*e = event

	return true, nil
}

// SlackbotProcessedEvent is the inner event-effect dedupe ledger. It is
// unscoped because Slack event IDs are unique per app, not per tenant.
type SlackbotProcessedEvent struct {
	EventID   string    `db:"event_id"`
	CreatedAt time.Time `db:"created_at"`
}

// Claim inserts the event id. Returns false when it was already claimed
// (Slack retry) so callers can skip side effects.
func (e *SlackbotProcessedEvent) Claim(
	ctx context.Context,
	conn pg.Querier,
) (bool, error) {
	q := `
INSERT INTO slackbot_processed_events (
    event_id,
    created_at
)
VALUES (
    @event_id,
    @created_at
)
ON CONFLICT (event_id) DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"event_id":   e.EventID,
		"created_at": e.CreatedAt,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot claim slackbot processed event: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func DeleteSlackbotProcessedEventsBefore(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
) error {
	_, err := DeleteSlackbotProcessedEventsBeforeBatch(ctx, conn, before, 1000)

	return err
}

func DeleteSlackbotProcessedEventsBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT event_id
	FROM slackbot_processed_events
	WHERE created_at < @before
	ORDER BY created_at ASC, event_id ASC
	LIMIT @limit
)
DELETE FROM slackbot_processed_events
WHERE event_id IN (SELECT event_id FROM doomed)
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
		return 0, fmt.Errorf("cannot delete Slackbot processed events: %w", err)
	}

	return result.RowsAffected(), nil
}
