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
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

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
