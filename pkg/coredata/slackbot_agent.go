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
	"go.gearno.de/kit/pg"
)

type (
	SlackbotAgent struct {
		AgentID     string              `db:"agent_id"`
		SessionID   string              `db:"session_id"`
		Channel     string              `db:"channel"`
		ThreadTS    string              `db:"thread_ts"`
		SlackUserID string              `db:"slack_user_id"`
		Status      SlackbotAgentStatus `db:"status"`
		Checkpoint  json.RawMessage     `db:"checkpoint"`
		Messages    json.RawMessage     `db:"messages"`
		CreatedAt   time.Time           `db:"created_at"`
		UpdatedAt   time.Time           `db:"updated_at"`
	}

	SlackbotAgents []*SlackbotAgent
)

func (a *SlackbotAgent) UpsertBySessionID(
	ctx context.Context,
	conn pg.Querier,
) (inserted bool, err error) {
	q := `
INSERT INTO slackbot_agents (
    agent_id,
    session_id,
    channel,
    thread_ts,
    slack_user_id,
    status,
    messages,
    created_at,
    updated_at
)
VALUES (
    @agent_id,
    @session_id,
    @channel,
    @thread_ts,
    @slack_user_id,
    @status,
    @messages,
    @created_at,
    @updated_at
)
ON CONFLICT (session_id) DO UPDATE
SET
    updated_at = EXCLUDED.updated_at
RETURNING
    agent_id,
    session_id,
    channel,
    thread_ts,
    slack_user_id,
    status,
    checkpoint,
    messages,
    created_at,
    updated_at
`

	originalAgentID := a.AgentID

	args := pgx.StrictNamedArgs{
		"agent_id":      a.AgentID,
		"session_id":    a.SessionID,
		"channel":       a.Channel,
		"thread_ts":     a.ThreadTS,
		"slack_user_id": a.SlackUserID,
		"status":        a.Status,
		"messages":      a.Messages,
		"created_at":    a.CreatedAt,
		"updated_at":    a.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert slackbot agent: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackbotAgent])
	if err != nil {
		return false, fmt.Errorf("cannot collect upserted slackbot agent: %w", err)
	}

	*a = row

	return originalAgentID == a.AgentID, nil
}

func (a *SlackbotAgent) LoadByAgentID(
	ctx context.Context,
	conn pg.Querier,
	agentID string,
) error {
	q := `
SELECT
    agent_id,
    session_id,
    channel,
    thread_ts,
    slack_user_id,
    status,
    checkpoint,
    messages,
    created_at,
    updated_at
FROM
    slackbot_agents
WHERE
    agent_id = @agent_id
LIMIT 1
`

	args := pgx.StrictNamedArgs{
		"agent_id": agentID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slackbot agent: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackbotAgent])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect slackbot agent: %w", err)
	}

	*a = row

	return nil
}

func (a *SlackbotAgent) ClaimAvailable(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
) (bool, error) {
	q := `
UPDATE slackbot_agents
SET
    status = @busy_status,
    updated_at = @updated_at
WHERE agent_id IN (
    SELECT agent_id
    FROM slackbot_agents
    WHERE
        agent_id = @agent_id
        AND status = @available_status
    FOR UPDATE SKIP LOCKED
)
`

	args := pgx.StrictNamedArgs{
		"agent_id":         a.AgentID,
		"busy_status":      SlackbotAgentStatusBusy,
		"available_status": SlackbotAgentStatusAvailable,
		"updated_at":       now,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot claim slackbot agent: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (a *SlackbotAgent) MarkAvailable(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
) error {
	q := `
UPDATE slackbot_agents
SET
    status = @status,
    checkpoint = NULL,
    updated_at = @updated_at
WHERE
    agent_id = @agent_id
`

	args := pgx.StrictNamedArgs{
		"agent_id":   a.AgentID,
		"status":     SlackbotAgentStatusAvailable,
		"updated_at": now,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark slackbot agent available: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	a.Status = SlackbotAgentStatusAvailable
	a.Checkpoint = nil
	a.UpdatedAt = now

	return nil
}

func (a *SlackbotAgent) UpdateMessages(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
) error {
	q := `
UPDATE slackbot_agents
SET
    messages = @messages,
    updated_at = @updated_at
WHERE
    agent_id = @agent_id
`

	args := pgx.StrictNamedArgs{
		"agent_id":   a.AgentID,
		"messages":   a.Messages,
		"updated_at": now,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update slackbot agent messages: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	a.UpdatedAt = now

	return nil
}

func (a *SlackbotAgent) UpdateCheckpoint(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
) error {
	q := `
UPDATE slackbot_agents
SET
    checkpoint = @checkpoint,
    updated_at = @updated_at
WHERE
    agent_id = @agent_id
`

	args := pgx.StrictNamedArgs{
		"agent_id":   a.AgentID,
		"checkpoint": a.Checkpoint,
		"updated_at": now,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update slackbot agent checkpoint: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	a.UpdatedAt = now

	return nil
}

func (a *SlackbotAgents) ClaimResumable(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
) error {
	q := `
UPDATE slackbot_agents
SET
    updated_at = @updated_at
WHERE agent_id IN (
    SELECT agent_id
    FROM slackbot_agents
    WHERE
        status = @busy_status
        AND checkpoint IS NOT NULL
    FOR UPDATE SKIP LOCKED
)
RETURNING
    agent_id,
    session_id,
    channel,
    thread_ts,
    slack_user_id,
    status,
    checkpoint,
    messages,
    created_at,
    updated_at
`

	args := pgx.StrictNamedArgs{
		"busy_status": SlackbotAgentStatusBusy,
		"updated_at":  now,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot claim resumable slackbot agents: %w", err)
	}

	agents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SlackbotAgent])
	if err != nil {
		return fmt.Errorf("cannot collect resumable slackbot agents: %w", err)
	}

	*a = agents

	return nil
}

func (a *SlackbotAgents) LoadBySessionIDPrefix(
	ctx context.Context,
	conn pg.Querier,
	sessionIDPrefix string,
	limit int,
) error {
	q := `
SELECT
    agent_id,
    session_id,
    channel,
    thread_ts,
    slack_user_id,
    status,
    checkpoint,
    messages,
    created_at,
    updated_at
FROM
    slackbot_agents
WHERE
    session_id LIKE @session_id_prefix || '%'
ORDER BY
    created_at ASC
LIMIT @limit
`

	args := pgx.StrictNamedArgs{
		"session_id_prefix": sessionIDPrefix,
		"limit":             limit,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slackbot agents by session id prefix: %w", err)
	}

	agents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[SlackbotAgent])
	if err != nil {
		return fmt.Errorf("cannot collect slackbot agents by session id prefix: %w", err)
	}

	*a = agents

	return nil
}

func CountSlackbotAgentsBySessionIDPrefix(
	ctx context.Context,
	conn pg.Querier,
	sessionIDPrefix string,
) (int, error) {
	q := `
SELECT COUNT(*)
FROM slackbot_agents
WHERE session_id LIKE @session_id_prefix || '%'
`

	args := pgx.StrictNamedArgs{
		"session_id_prefix": sessionIDPrefix,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot count slackbot agents by session id prefix: %w", err)
	}

	count, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[int])
	if err != nil {
		return 0, fmt.Errorf("cannot collect slackbot agent count: %w", err)
	}

	return count, nil
}

func (a *SlackbotAgent) Delete(
	ctx context.Context,
	conn pg.Querier,
) error {
	q := `
DELETE FROM slackbot_agents
WHERE
    agent_id = @agent_id
`

	args := pgx.StrictNamedArgs{
		"agent_id": a.AgentID,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete slackbot agent: %w", err)
	}

	return nil
}
