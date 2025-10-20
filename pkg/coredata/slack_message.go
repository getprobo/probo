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
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	SlackMessage struct {
		ID             gid.GID          `db:"id"`
		OrganizationID gid.GID          `db:"organization_id"`
		Type           SlackMessageType `db:"type"`
		Body           map[string]any   `db:"body"`
		MessageTS      *string          `db:"message_ts"`
		ChannelID      *string          `db:"channel_id"`
		RequesterEmail *string          `db:"requester_email"`
		CreatedAt      time.Time        `db:"created_at"`
		UpdatedAt      time.Time        `db:"updated_at"`
		SentAt         *time.Time       `db:"sent_at"`
		Error          *string          `db:"error"`
	}

	ErrNoUnsentSlackMessage struct{}
)

func (e ErrNoUnsentSlackMessage) Error() string {
	return "no unsent slack message found"
}

func NewSlackMessage(
	scope Scoper,
	organizationID gid.GID,
	messageType SlackMessageType,
	body map[string]any,
	requesterEmail *string,
) *SlackMessage {
	now := time.Now()
	return &SlackMessage{
		ID:             gid.New(scope.GetTenantID(), SlackMessageEntityType),
		OrganizationID: organizationID,
		Type:           messageType,
		Body:           body,
		RequesterEmail: requesterEmail,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (s *SlackMessage) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO slack_messages (id, tenant_id, organization_id, type, body, requester_email, created_at, updated_at)
VALUES (@id, @tenant_id, @organization_id, @type, @body, @requester_email, @created_at, @updated_at)
	`

	args := pgx.StrictNamedArgs{
		"id":              s.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": s.OrganizationID,
		"type":            s.Type,
		"body":            s.Body,
		"requester_email": s.RequesterEmail,
		"created_at":      s.CreatedAt,
		"updated_at":      s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert slack message: %w", err)
	}

	return nil
}

func (s *SlackMessage) LoadNextUnsentForUpdate(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, created_at, updated_at, sent_at, error
FROM slack_messages
WHERE sent_at IS NULL AND error IS NULL
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE
	`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query slack messages: %w", err)
	}

	message, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackMessage])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoUnsentSlackMessage{}
		}

		return fmt.Errorf("cannot collect slack message: %w", err)
	}

	*s = message

	return nil
}

// This is used for Slack webhook verification where we don't know the tenant yet
func (s *SlackMessage) LoadByChannelAndTSUnscoped(
	ctx context.Context,
	conn pg.Conn,
	channelID string,
	messageTS string,
) error {
	q := `
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, created_at, updated_at, sent_at, error
FROM slack_messages
WHERE message_ts = @message_ts AND channel_id = @channel_id
LIMIT 1
	`

	args := pgx.StrictNamedArgs{
		"message_ts": messageTS,
		"channel_id": channelID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slack message: %w", err)
	}

	message, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackMessage])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("slack message not found")
		}
		return fmt.Errorf("cannot collect slack message: %w", err)
	}

	*s = message

	return nil
}

func (s *SlackMessage) Update(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
UPDATE slack_messages
SET body = @body, sent_at = @sent_at, updated_at = @updated_at, error = @error, message_ts = @message_ts, channel_id = @channel_id
WHERE id = @id
	`

	args := pgx.StrictNamedArgs{
		"id":         s.ID,
		"body":       s.Body,
		"sent_at":    s.SentAt,
		"updated_at": s.UpdatedAt,
		"error":      s.Error,
		"message_ts": s.MessageTS,
		"channel_id": s.ChannelID,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update slack message: %w", err)
	}

	return nil
}

func (s *SlackMessage) LoadById(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	slackMessageID gid.GID,
) error {
	q := `
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, created_at, updated_at, sent_at, error
FROM slack_messages
WHERE id = @id
AND %s
LIMIT 1
	`

	args := pgx.StrictNamedArgs{
		"id": slackMessageID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slack message: %w", err)
	}

	message, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackMessage])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("slack message not found")
		}
		return fmt.Errorf("cannot collect slack message: %w", err)
	}

	*s = message

	return nil
}

func (s *SlackMessage) LoadLatestByRequesterEmailAndType(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	requesterEmail string,
	messageType SlackMessageType,
	since time.Time,
) error {
	q := `
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, created_at, updated_at, sent_at, error
FROM slack_messages
WHERE %s
	AND organization_id = @organization_id
	AND requester_email = @requester_email
	AND type = @type
	AND created_at >= @since
ORDER BY created_at DESC
LIMIT 1
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"requester_email": requesterEmail,
		"type":            messageType,
		"since":           since,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slack message: %w", err)
	}

	message, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackMessage])
	if err != nil {
		return err
	}

	*s = message

	return nil
}
