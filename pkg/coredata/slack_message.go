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

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type (
	SlackMessage struct {
		ID                    gid.GID          `db:"id"`
		OrganizationID        gid.GID          `db:"organization_id"`
		Type                  SlackMessageType `db:"type"`
		Body                  map[string]any   `db:"body"`
		MessageTS             *string          `db:"message_ts"`
		ChannelID             *string          `db:"channel_id"`
		RequesterEmail        *mail.Addr       `db:"requester_email"`
		Metadata              map[string]any   `db:"metadata"`
		InitialSlackMessageID gid.GID          `db:"initial_slack_message_id"`
		CreatedAt             time.Time        `db:"created_at"`
		UpdatedAt             time.Time        `db:"updated_at"`
		SentAt                *time.Time       `db:"sent_at"`
		Error                 *string          `db:"error"`
	}

	ErrNoUnsentSlackMessage struct{}

	ErrSlackMessageNotFound struct{}
)

func (e ErrNoUnsentSlackMessage) Error() string {
	return "no unsent slack message found"
}

func (e ErrSlackMessageNotFound) Error() string {
	return "slack message not found"
}

func (sm *SlackMessage) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM slack_messages WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, sm.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query slack message authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func NewSlackMessage(
	scope Scoper,
	organizationID gid.GID,
	messageType SlackMessageType,
	body map[string]any,
) *SlackMessage {
	now := time.Now()
	id := gid.New(scope.GetTenantID(), SlackMessageEntityType)
	return &SlackMessage{
		ID:                    id,
		OrganizationID:        organizationID,
		Type:                  messageType,
		Body:                  body,
		InitialSlackMessageID: id,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

func (s *SlackMessage) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO slack_messages (
	id,
	tenant_id,
	organization_id,
	type,
	body,
	requester_email,
	metadata,
	initial_slack_message_id,
	message_ts,
	channel_id,
	created_at,
	updated_at
)
VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@type,
	@body,
	@requester_email,
	@metadata,
	@initial_slack_message_id,
	@message_ts,
	@channel_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                       s.ID,
		"tenant_id":                scope.GetTenantID(),
		"organization_id":          s.OrganizationID,
		"type":                     s.Type,
		"body":                     s.Body,
		"requester_email":          s.RequesterEmail,
		"metadata":                 s.Metadata,
		"initial_slack_message_id": s.InitialSlackMessageID,
		"message_ts":               s.MessageTS,
		"channel_id":               s.ChannelID,
		"created_at":               s.CreatedAt,
		"updated_at":               s.UpdatedAt,
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
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, metadata, initial_slack_message_id, created_at, updated_at, sent_at, error
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

func (s *SlackMessage) LoadNextInitalUnsentForUpdate(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, metadata, initial_slack_message_id, created_at, updated_at, sent_at, error
FROM slack_messages
WHERE sent_at IS NULL AND error IS NULL AND id = initial_slack_message_id
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

func (s *SlackMessage) LoadNextUpdateUnsentForUpdate(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT
	sm.id,
	sm.organization_id,
	sm.type,
	sm.body,
	COALESCE(sm.message_ts, original.message_ts) as message_ts,
	COALESCE(sm.channel_id, original.channel_id) as channel_id,
	sm.requester_email,
	sm.metadata,
	sm.initial_slack_message_id,
	sm.created_at,
	sm.updated_at,
	sm.sent_at,
	sm.error
FROM slack_messages sm
INNER JOIN slack_messages original ON sm.initial_slack_message_id = original.id
WHERE sm.sent_at IS NULL
	AND sm.error IS NULL
	AND sm.id != sm.initial_slack_message_id
	AND original.sent_at IS NOT NULL
	AND original.error IS NULL
ORDER BY sm.created_at ASC
LIMIT 1
FOR UPDATE OF sm
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

func (s *SlackMessage) LoadInitialByChannelAndTS(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	channelID string,
	messageTS string,
) error {
	q := `
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, metadata, initial_slack_message_id, created_at, updated_at, sent_at, error
FROM slack_messages
WHERE message_ts = @message_ts AND channel_id = @channel_id AND id = initial_slack_message_id AND %s
LIMIT 1
	`

	args := pgx.StrictNamedArgs{
		"message_ts": messageTS,
		"channel_id": channelID,
	}
	maps.Copy(args, scope.SQLArguments())

	q = fmt.Sprintf(q, scope.SQLFragment())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slack message: %w", err)
	}

	message, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackMessage])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSlackMessageNotFound{}
		}
		return fmt.Errorf("cannot collect slack message: %w", err)
	}

	*s = message

	return nil
}

func (s *SlackMessage) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE slack_messages
SET sent_at = @sent_at, updated_at = @updated_at, error = @error
WHERE id = @id AND %s
	`

	args := pgx.StrictNamedArgs{
		"id":         s.ID,
		"sent_at":    s.SentAt,
		"updated_at": s.UpdatedAt,
		"error":      s.Error,
	}

	q = fmt.Sprintf(q, scope.SQLFragment())

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update slack message: %w", err)
	}

	return nil
}

func (s *SlackMessage) UpdateChannelAndTSByInitialMessageID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	initialSlackMessageID gid.GID,
	channelID string,
	messageTS string,
	updatedAt time.Time,
) error {
	q := `
UPDATE slack_messages
SET channel_id = @channel_id, message_ts = @message_ts, updated_at = @updated_at
WHERE initial_slack_message_id = @initial_slack_message_id AND %s
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"initial_slack_message_id": initialSlackMessageID,
		"channel_id":               channelID,
		"message_ts":               messageTS,
		"updated_at":               updatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update slack messages with initial message id: %w", err)
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
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, metadata, initial_slack_message_id, created_at, updated_at, sent_at, error
FROM slack_messages
WHERE id = @id
AND %s
LIMIT 1
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

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
			return ErrSlackMessageNotFound{}
		}
		return fmt.Errorf("cannot collect slack message: %w", err)
	}

	*s = message

	return nil
}

func (s *SlackMessage) LoadLatestByInitialMessageID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	initialSlackMessageID gid.GID,
) error {
	q := `
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, metadata, initial_slack_message_id, created_at, updated_at, sent_at, error
FROM slack_messages
WHERE %s
	AND initial_slack_message_id = @initial_slack_message_id
ORDER BY created_at DESC
LIMIT 1
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"initial_slack_message_id": initialSlackMessageID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query slack message: %w", err)
	}

	message, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackMessage])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSlackMessageNotFound{}
		}
		return err
	}

	*s = message

	return nil
}

func (s *SlackMessage) LoadLatestByRequesterEmailAndType(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	requesterEmail mail.Addr,
	messageType SlackMessageType,
	since time.Time,
) error {
	q := `
SELECT id, organization_id, type, body, message_ts, channel_id, requester_email, metadata, initial_slack_message_id, created_at, updated_at, sent_at, error
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
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrSlackMessageNotFound{}
		}
		return err
	}

	*s = message

	return nil
}
