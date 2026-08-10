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
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	SlackbotInteraction struct {
		InteractionID string          `db:"interaction_id"`
		AgentID       string          `db:"agent_id"`
		EventID       *string         `db:"event_id"`
		EventType     string          `db:"event_type"`
		Payload       json.RawMessage `db:"payload"`
		ProcessedAt   *time.Time      `db:"processed_at"`
		CreatedAt     time.Time       `db:"created_at"`
	}

	SlackbotInteractions []*SlackbotInteraction
)

func (i *SlackbotInteraction) Insert(
	ctx context.Context,
	conn pg.Querier,
) error {
	q := `
INSERT INTO slackbot_interactions (
    interaction_id,
    agent_id,
    event_id,
    event_type,
    payload,
    created_at
)
VALUES (
    @interaction_id,
    @agent_id,
    @event_id,
    @event_type,
    @payload,
    @created_at
)
ON CONFLICT (event_id) DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"interaction_id": i.InteractionID,
		"agent_id":       i.AgentID,
		"event_id":       i.EventID,
		"event_type":     i.EventType,
		"payload":        i.Payload,
		"created_at":     i.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert slackbot interaction: %w", err)
	}

	return nil
}

func (i *SlackbotInteractions) LoadPendingByAgentID(
	ctx context.Context,
	conn pg.Querier,
	agentID string,
	limit int,
) error {
	q := `
SELECT
    interaction_id,
    agent_id,
    event_id,
    event_type,
    payload,
    processed_at,
    created_at
FROM
    slackbot_interactions
WHERE
    agent_id = @agent_id
    AND processed_at IS NULL
ORDER BY
    created_at ASC
LIMIT @limit
`

	args := pgx.StrictNamedArgs{
		"agent_id": agentID,
		"limit":    limit,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query pending slackbot interactions: %w", err)
	}

	interactions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SlackbotInteraction])
	if err != nil {
		return fmt.Errorf("cannot collect pending slackbot interactions: %w", err)
	}

	*i = interactions

	return nil
}

func MarkSlackbotInteractionsProcessed(
	ctx context.Context,
	conn pg.Querier,
	interactionIDs []string,
	now time.Time,
) error {
	if len(interactionIDs) == 0 {
		return nil
	}

	q := `
UPDATE slackbot_interactions
SET
    processed_at = @processed_at
WHERE
    interaction_id = ANY(@interaction_ids)
`

	args := pgx.StrictNamedArgs{
		"interaction_ids": interactionIDs,
		"processed_at":    now,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark slackbot interactions processed: %w", err)
	}

	return nil
}

func DeleteSlackbotInteractionsByAgentIDs(
	ctx context.Context,
	conn pg.Querier,
	agentIDs []string,
) error {
	if len(agentIDs) == 0 {
		return nil
	}

	q := `
DELETE FROM slackbot_interactions
WHERE
    agent_id = ANY(@agent_ids)
`

	args := pgx.StrictNamedArgs{
		"agent_ids": agentIDs,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete slackbot interactions by agent ids: %w", err)
	}

	return nil
}
