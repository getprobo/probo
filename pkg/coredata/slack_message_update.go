// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	SlackMessageUpdate struct {
		ID             gid.GID        `db:"id"`
		SlackMessageID gid.GID        `db:"slack_message_id"`
		Body           map[string]any `db:"body"`
		CreatedAt      time.Time      `db:"created_at"`
		UpdatedAt      time.Time      `db:"updated_at"`
		SentAt         *time.Time     `db:"sent_at"`
		Error          *string        `db:"error"`
	}

	ErrNoUnsentSlackMessageUpdate struct{}
)

func (e ErrNoUnsentSlackMessageUpdate) Error() string {
	return "no unsent slack message update found"
}

func NewSlackMessageUpdate(
	scope Scoper,
	slackMessageID gid.GID,
	body map[string]any,
) *SlackMessageUpdate {
	now := time.Now()
	return &SlackMessageUpdate{
		ID:             gid.New(scope.GetTenantID(), SlackMessageUpdateEntityType),
		SlackMessageID: slackMessageID,
		Body:           body,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (s *SlackMessageUpdate) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO slack_message_updates (id, tenant_id, slack_message_id, body, created_at, updated_at, sent_at, error)
VALUES (@id, @tenant_id, @slack_message_id, @body, @created_at, @updated_at, @sent_at, @error)
	`

	args := pgx.StrictNamedArgs{
		"id":               s.ID,
		"tenant_id":        scope.GetTenantID(),
		"slack_message_id": s.SlackMessageID,
		"body":             s.Body,
		"created_at":       s.CreatedAt,
		"updated_at":       s.UpdatedAt,
		"sent_at":          s.SentAt,
		"error":            s.Error,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert slack message update: %w", err)
	}

	return nil
}

func (s *SlackMessageUpdate) LoadLatestBySlackMessageID(
	ctx context.Context,
	conn pg.Conn,
	slackMessageID gid.GID,
) error {
	q := `
SELECT id, slack_message_id, body, created_at, updated_at, sent_at, error
FROM slack_message_updates
WHERE slack_message_id = @slack_message_id
ORDER BY created_at DESC
LIMIT 1
	`

	args := pgx.StrictNamedArgs{
		"slack_message_id": slackMessageID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slack message updates: %w", err)
	}

	update, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackMessageUpdate])
	if err != nil {
		return err
	}

	*s = update

	return nil
}

func (s *SlackMessageUpdate) LoadNextUnsentForUpdate(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT smu.id, smu.slack_message_id, smu.body, smu.created_at, smu.updated_at, smu.sent_at, smu.error
FROM slack_message_updates smu
INNER JOIN slack_messages sm ON smu.slack_message_id = sm.id
WHERE smu.sent_at IS NULL
	AND smu.error IS NULL
	AND sm.sent_at IS NOT NULL
	AND sm.error IS NULL
ORDER BY smu.created_at ASC
LIMIT 1
FOR UPDATE OF smu
	`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query slack message updates: %w", err)
	}

	update, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackMessageUpdate])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoUnsentSlackMessageUpdate{}
		}

		return fmt.Errorf("cannot collect slack message update: %w", err)
	}

	*s = update

	return nil
}

func (s *SlackMessageUpdate) Update(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
UPDATE slack_message_updates
SET body = @body, updated_at = @updated_at, sent_at = @sent_at, error = @error
WHERE id = @id
	`

	args := pgx.StrictNamedArgs{
		"id":         s.ID,
		"body":       s.Body,
		"updated_at": s.UpdatedAt,
		"sent_at":    s.SentAt,
		"error":      s.Error,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update slack message update: %w", err)
	}

	return nil
}
