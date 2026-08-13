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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

const (
	SlackbotMessageMetadataDedupKey      = "dedup_key"
	SlackbotMessageMetadataSourceEventID = "source_event_id"
	SlackbotMessageDefaultMaxAttempts    = 5
)

type SlackbotMessage struct {
	ID                       gid.GID        `db:"id"`
	OrganizationID           gid.GID        `db:"organization_id"`
	MessageType              string         `db:"message_type"`
	Body                     map[string]any `db:"body"`
	Metadata                 map[string]any `db:"metadata"`
	ChannelID                *string        `db:"channel_id"`
	MessageTS                *string        `db:"message_ts"`
	InitialSlackbotMessageID gid.GID        `db:"initial_slackbot_message_id"`
	CreatedAt                time.Time      `db:"created_at"`
	UpdatedAt                time.Time      `db:"updated_at"`
	SentAt                   *time.Time     `db:"sent_at"`
	ProcessingStartedAt      *time.Time     `db:"processing_started_at"`
	AttemptCount             int            `db:"attempt_count"`
	MaxAttempts              int            `db:"max_attempts"`
	LastAttemptedAt          *time.Time     `db:"last_attempted_at"`
	NextAttemptAt            *time.Time     `db:"next_attempt_at"`
	LastError                *string        `db:"last_error"`
	ClientMsgID              string         `db:"client_msg_id"`
	Error                    *string        `db:"error"`
}

func NewSlackbotMessage(
	scope Scoper,
	organizationID gid.GID,
	messageType string,
	body map[string]any,
	metadata map[string]any,
) *SlackbotMessage {
	now := time.Now()
	id := gid.New(scope.GetTenantID(), SlackbotMessageEntityType)

	return &SlackbotMessage{
		ID:                       id,
		OrganizationID:           organizationID,
		MessageType:              messageType,
		Body:                     body,
		Metadata:                 metadata,
		InitialSlackbotMessageID: id,
		MaxAttempts:              SlackbotMessageDefaultMaxAttempts,
		ClientMsgID:              uuid.MustNewV4().String(),
		CreatedAt:                now,
		UpdatedAt:                now,
	}
}

func (m *SlackbotMessage) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO slackbot_messages (
	id,
	tenant_id,
	organization_id,
	message_type,
	body,
	metadata,
	channel_id,
	message_ts,
	initial_slackbot_message_id,
	created_at,
	updated_at,
	sent_at,
	processing_started_at,
	attempt_count,
	max_attempts,
	last_attempted_at,
	next_attempt_at,
	last_error,
	client_msg_id,
	error
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@message_type,
	@body,
	@metadata,
	@channel_id,
	@message_ts,
	@initial_slackbot_message_id,
	@created_at,
	@updated_at,
	@sent_at,
	@processing_started_at,
	@attempt_count,
	@max_attempts,
	@last_attempted_at,
	@next_attempt_at,
	@last_error,
	@client_msg_id,
	@error
)
`

	args := pgx.StrictNamedArgs{
		"id":                          m.ID,
		"tenant_id":                   scope.GetTenantID(),
		"organization_id":             m.OrganizationID,
		"message_type":                m.MessageType,
		"body":                        m.Body,
		"metadata":                    m.Metadata,
		"channel_id":                  m.ChannelID,
		"message_ts":                  m.MessageTS,
		"initial_slackbot_message_id": m.InitialSlackbotMessageID,
		"created_at":                  m.CreatedAt,
		"updated_at":                  m.UpdatedAt,
		"sent_at":                     m.SentAt,
		"processing_started_at":       m.ProcessingStartedAt,
		"attempt_count":               m.AttemptCount,
		"max_attempts":                m.MaxAttempts,
		"last_attempted_at":           m.LastAttemptedAt,
		"next_attempt_at":             m.NextAttemptAt,
		"last_error":                  m.LastError,
		"client_msg_id":               m.ClientMsgID,
		"error":                       m.Error,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "slackbot_messages_source_event_id_unique" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert Slackbot message: %w", err)
	}

	return nil
}

func (m *SlackbotMessage) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	message_type,
	body,
	metadata,
	channel_id,
	message_ts,
	initial_slackbot_message_id,
	created_at,
	updated_at,
	sent_at,
	processing_started_at,
	attempt_count,
	max_attempts,
	last_attempted_at,
	next_attempt_at,
	last_error,
	client_msg_id,
	error
FROM slackbot_messages
WHERE %s
	AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	return m.loadExactlyOne(ctx, conn, q, args)
}

func (m *SlackbotMessage) LoadBySourceEventID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceEventID string,
) error {
	q := `
SELECT
	id,
	organization_id,
	message_type,
	body,
	metadata,
	channel_id,
	message_ts,
	initial_slackbot_message_id,
	created_at,
	updated_at,
	sent_at,
	processing_started_at,
	attempt_count,
	max_attempts,
	last_attempted_at,
	next_attempt_at,
	last_error,
	client_msg_id,
	error
FROM slackbot_messages
WHERE %s
	AND metadata->>@metadata_key = @source_event_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"metadata_key":    SlackbotMessageMetadataSourceEventID,
		"source_event_id": sourceEventID,
	}
	maps.Copy(args, scope.SQLArguments())

	return m.loadExactlyOne(ctx, conn, q, args)
}

func (m *SlackbotMessage) LoadInitialByChannelAndTS(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	channelID string,
	messageTS string,
) error {
	q := `
SELECT
	id,
	organization_id,
	message_type,
	body,
	metadata,
	channel_id,
	message_ts,
	initial_slackbot_message_id,
	created_at,
	updated_at,
	sent_at,
	processing_started_at,
	attempt_count,
	max_attempts,
	last_attempted_at,
	next_attempt_at,
	last_error,
	client_msg_id,
	error
FROM slackbot_messages
WHERE %s
	AND channel_id = @channel_id
	AND message_ts = @message_ts
	AND id = initial_slackbot_message_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"channel_id": channelID,
		"message_ts": messageTS,
	}
	maps.Copy(args, scope.SQLArguments())

	return m.loadExactlyOne(ctx, conn, q, args)
}

func (m *SlackbotMessage) LoadLatestDeliveredByChannelAndTS(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	channelID string,
	messageTS string,
) error {
	q := `
SELECT
	id,
	organization_id,
	message_type,
	body,
	metadata,
	channel_id,
	message_ts,
	initial_slackbot_message_id,
	created_at,
	updated_at,
	sent_at,
	processing_started_at,
	attempt_count,
	max_attempts,
	last_attempted_at,
	next_attempt_at,
	last_error,
	client_msg_id,
	error
FROM slackbot_messages
WHERE %s
	AND channel_id = @channel_id
	AND message_ts = @message_ts
	AND sent_at IS NOT NULL
	AND error IS NULL
ORDER BY created_at DESC, id DESC
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"channel_id": channelID,
		"message_ts": messageTS,
	}
	maps.Copy(args, scope.SQLArguments())

	return m.loadExactlyOne(ctx, conn, q, args)
}

func (m *SlackbotMessage) LoadLatestDeliveredByOrganizationIDChannelAndTS(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	channelID string,
	messageTS string,
) error {
	q := `
SELECT
	id,
	organization_id,
	message_type,
	body,
	metadata,
	channel_id,
	message_ts,
	initial_slackbot_message_id,
	created_at,
	updated_at,
	sent_at,
	processing_started_at,
	attempt_count,
	max_attempts,
	last_attempted_at,
	next_attempt_at,
	last_error,
	client_msg_id,
	error
FROM slackbot_messages
WHERE %s
	AND organization_id = @organization_id
	AND channel_id = @channel_id
	AND message_ts = @message_ts
	AND sent_at IS NOT NULL
	AND error IS NULL
ORDER BY created_at DESC, id DESC
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"channel_id":      channelID,
		"message_ts":      messageTS,
	}
	maps.Copy(args, scope.SQLArguments())

	return m.loadExactlyOne(ctx, conn, q, args)
}

func (m *SlackbotMessage) LoadLatestByOrganizationIDChannelIDMessageTypeAndDedupKey(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	channelID string,
	messageType string,
	dedupKey string,
	createdAfter time.Time,
) error {
	q := `
SELECT
	id,
	organization_id,
	message_type,
	body,
	metadata,
	channel_id,
	message_ts,
	initial_slackbot_message_id,
	created_at,
	updated_at,
	sent_at,
	processing_started_at,
	attempt_count,
	max_attempts,
	last_attempted_at,
	next_attempt_at,
	last_error,
	client_msg_id,
	error
FROM slackbot_messages
WHERE %s
	AND organization_id = @organization_id
	AND channel_id = @channel_id
	AND message_type = @message_type
	AND metadata ->> @metadata_key = @dedup_key
	AND created_at >= @created_after
ORDER BY created_at DESC
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"channel_id":      channelID,
		"message_type":    messageType,
		"metadata_key":    SlackbotMessageMetadataDedupKey,
		"dedup_key":       dedupKey,
		"created_after":   createdAfter,
	}
	maps.Copy(args, scope.SQLArguments())

	return m.loadExactlyOne(ctx, conn, q, args)
}

func (m *SlackbotMessage) ClaimNextForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
WITH candidate AS (
	SELECT m.id
	FROM slackbot_messages m
	LEFT JOIN slackbot_messages initial
		ON initial.id = m.initial_slackbot_message_id
	WHERE m.sent_at IS NULL
		AND m.error IS NULL
		AND m.processing_started_at IS NULL
		AND m.attempt_count < m.max_attempts
		AND (m.next_attempt_at IS NULL OR m.next_attempt_at <= @now)
		AND (
			m.id = m.initial_slackbot_message_id
			OR (
				initial.sent_at IS NOT NULL
				AND initial.error IS NULL
				AND NOT EXISTS (
					SELECT 1
					FROM slackbot_messages earlier
					WHERE earlier.initial_slackbot_message_id = m.initial_slackbot_message_id
						AND earlier.id != earlier.initial_slackbot_message_id
						AND earlier.sent_at IS NULL
						AND earlier.error IS NULL
						AND (
							earlier.created_at < m.created_at
							OR (
								earlier.created_at = m.created_at
								AND earlier.id < m.id
							)
						)
				)
			)
		)
	ORDER BY m.created_at ASC, m.id ASC
	LIMIT 1
	FOR UPDATE OF m SKIP LOCKED
)
UPDATE slackbot_messages m
SET
	attempt_count = m.attempt_count + 1,
	last_attempted_at = @now,
	processing_started_at = @now,
	updated_at = @now
FROM candidate
WHERE m.id = candidate.id
RETURNING
	m.id,
	m.organization_id,
	m.message_type,
	m.body,
	m.metadata,
	m.channel_id,
	m.message_ts,
	m.initial_slackbot_message_id,
	m.created_at,
	m.updated_at,
	m.sent_at,
	m.processing_started_at,
	m.attempt_count,
	m.max_attempts,
	m.last_attempted_at,
	m.next_attempt_at,
	m.last_error,
	m.client_msg_id,
	m.error
`

	return m.loadExactlyOne(ctx, conn, q, pgx.StrictNamedArgs{"now": now})
}

func (m *SlackbotMessage) UpdateDeliveryState(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE slackbot_messages
SET
	sent_at = @sent_at,
	processing_started_at = @processing_started_at,
	next_attempt_at = @next_attempt_at,
	last_error = @last_error,
	error = @error,
	updated_at = @updated_at
WHERE %s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                    m.ID,
		"sent_at":               m.SentAt,
		"processing_started_at": m.ProcessingStartedAt,
		"next_attempt_at":       m.NextAttemptAt,
		"last_error":            m.LastError,
		"error":                 m.Error,
		"updated_at":            m.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update Slackbot message delivery state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (m *SlackbotMessage) PropagateDeliveryReferenceToRevisions(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE slackbot_messages
SET
	channel_id = @channel_id,
	message_ts = @message_ts,
	updated_at = @updated_at
WHERE %s
	AND initial_slackbot_message_id = @initial_slackbot_message_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"initial_slackbot_message_id": m.InitialSlackbotMessageID,
		"channel_id":                  m.ChannelID,
		"message_ts":                  m.MessageTS,
		"updated_at":                  m.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot propagate Slackbot message delivery reference: %w", err)
	}

	return nil
}

func ResetStaleProcessingSlackbotMessages(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
) error {
	q := `
UPDATE slackbot_messages
SET
	processing_started_at = NULL,
	next_attempt_at = CASE
		WHEN attempt_count >= max_attempts THEN NULL
		ELSE @updated_at::timestamptz
	END,
	last_error = @stale_error,
	error = CASE
		WHEN attempt_count >= max_attempts THEN @stale_error
		ELSE error
	END,
	updated_at = @updated_at
WHERE sent_at IS NULL
	AND error IS NULL
	AND processing_started_at IS NOT NULL
	AND processing_started_at < @stale_threshold
`

	now := time.Now()
	args := pgx.StrictNamedArgs{
		"updated_at":      now,
		"stale_threshold": now.Add(-staleAfter),
		"stale_error":     "Slackbot delivery processing lease expired",
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot reset stale processing Slackbot messages: %w", err)
	}

	return nil
}

func DeleteRetiredSlackbotMessageThreadsBefore(
	ctx context.Context,
	conn pg.Querier,
	completedBefore time.Time,
	deadLetterBefore time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT initial_slackbot_message_id AS id
	FROM slackbot_messages
	GROUP BY initial_slackbot_message_id
	HAVING bool_and(
		(sent_at IS NOT NULL AND sent_at < @completed_before)
		OR (error IS NOT NULL AND updated_at < @dead_letter_before)
	)
	ORDER BY max(updated_at) ASC, initial_slackbot_message_id ASC
	LIMIT @limit
)
DELETE FROM slackbot_messages
WHERE id IN (SELECT id FROM doomed)
	AND id = initial_slackbot_message_id
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"completed_before":   completedBefore,
			"dead_letter_before": deadLetterBefore,
			"limit":              limit,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete retired Slackbot message threads: %w", err)
	}

	return result.RowsAffected(), nil
}

func (m *SlackbotMessage) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Slackbot message: %w", err)
	}

	message, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackbotMessage])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect Slackbot message: %w", err)
	}

	*m = message

	return nil
}
