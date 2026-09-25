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
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

const BotMessageDefaultMaxAttempts = 5

type (
	BotMessagePurpose string

	BotMessage struct {
		ID                   gid.GID           `db:"id"`
		OrganizationID       gid.GID           `db:"organization_id"`
		Capability           string            `db:"capability"`
		MessageType          string            `db:"message_type"`
		Attributes           json.RawMessage   `db:"attributes"`
		SubjectNamespace     string            `db:"subject_namespace"`
		SubjectKey           string            `db:"subject_key"`
		EventKey             string            `db:"event_key"`
		Purpose              BotMessagePurpose `db:"purpose"`
		ProcessingOwnerToken *string           `db:"processing_owner_token"`
		ProcessingStartedAt  *time.Time        `db:"processing_started_at"`
		ProcessedAt          *time.Time        `db:"processed_at"`
		AttemptCount         int               `db:"attempt_count"`
		MaxAttempts          int               `db:"max_attempts"`
		NextAttemptAt        *time.Time        `db:"next_attempt_at"`
		LastError            *string           `db:"last_error"`
		DeadLetteredAt       *time.Time        `db:"dead_lettered_at"`
		CreatedAt            time.Time         `db:"created_at"`
		UpdatedAt            time.Time         `db:"updated_at"`
	}
)

const (
	BotMessagePurposePost   BotMessagePurpose = "POST"
	BotMessagePurposeUpdate BotMessagePurpose = "UPDATE"
)

var (
	_ fmt.Stringer             = BotMessagePurpose("")
	_ encoding.TextMarshaler   = BotMessagePurpose("")
	_ encoding.TextUnmarshaler = (*BotMessagePurpose)(nil)
)

func BotMessagePurposes() []BotMessagePurpose {
	return []BotMessagePurpose{
		BotMessagePurposePost,
		BotMessagePurposeUpdate,
	}
}

func (v BotMessagePurpose) IsValid() bool {
	switch v {
	case BotMessagePurposePost, BotMessagePurposeUpdate:
		return true
	}

	return false
}

func (v BotMessagePurpose) String() string {
	return string(v)
}

func (v BotMessagePurpose) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *BotMessagePurpose) UnmarshalText(text []byte) error {
	value := BotMessagePurpose(text)
	if !value.IsValid() {
		return fmt.Errorf("invalid BotMessagePurpose value: %q", string(text))
	}

	*v = value

	return nil
}

func (m *BotMessage) EnqueueIdempotently(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) (inserted bool, err error) {
	q := `
INSERT INTO bot_messages (
	id,
	tenant_id,
	organization_id,
	capability,
	message_type,
	attributes,
	subject_namespace,
	subject_key,
	event_key,
	purpose,
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
	@tenant_id,
	@organization_id,
	@capability,
	@message_type,
	@attributes,
	@subject_namespace,
	@subject_key,
	@event_key,
	@purpose,
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
ON CONFLICT (tenant_id, organization_id, subject_namespace, subject_key, event_key)
DO UPDATE SET
	updated_at = bot_messages.updated_at
RETURNING
	id,
	organization_id,
	capability,
	message_type,
	attributes,
	subject_namespace,
	subject_key,
	event_key,
	purpose,
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
`

	if m.MaxAttempts <= 0 {
		m.MaxAttempts = BotMessageDefaultMaxAttempts
	}

	if !m.Purpose.IsValid() {
		return false, fmt.Errorf("cannot enqueue bot message: invalid purpose %q", m.Purpose)
	}

	if m.Attributes == nil {
		m.Attributes = json.RawMessage("{}")
	}

	originalID := m.ID

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":                     m.ID,
			"tenant_id":              scope.GetTenantID(),
			"organization_id":        m.OrganizationID,
			"capability":             m.Capability,
			"message_type":           m.MessageType,
			"attributes":             m.Attributes,
			"subject_namespace":      m.SubjectNamespace,
			"subject_key":            m.SubjectKey,
			"event_key":              m.EventKey,
			"purpose":                m.Purpose,
			"processing_owner_token": m.ProcessingOwnerToken,
			"processing_started_at":  m.ProcessingStartedAt,
			"processed_at":           m.ProcessedAt,
			"attempt_count":          m.AttemptCount,
			"max_attempts":           m.MaxAttempts,
			"next_attempt_at":        m.NextAttemptAt,
			"last_error":             m.LastError,
			"dead_lettered_at":       m.DeadLetteredAt,
			"created_at":             m.CreatedAt,
			"updated_at":             m.UpdatedAt,
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot enqueue bot message: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[BotMessage])
	if err != nil {
		return false, fmt.Errorf("cannot collect enqueued bot message: %w", err)
	}

	*m = row

	return originalID == m.ID, nil
}

func (m *BotMessage) ClaimNextForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
WITH candidate AS (
	SELECT id
	FROM bot_messages
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
UPDATE bot_messages m
SET
	attempt_count = m.attempt_count + 1,
	processing_owner_token = @owner_token,
	processing_started_at = @now,
	updated_at = @now
FROM candidate
WHERE m.id = candidate.id
RETURNING
	m.id,
	m.organization_id,
	m.capability,
	m.message_type,
	m.attributes,
	m.subject_namespace,
	m.subject_key,
	m.event_key,
	m.purpose,
	m.processing_owner_token,
	m.processing_started_at,
	m.processed_at,
	m.attempt_count,
	m.max_attempts,
	m.next_attempt_at,
	m.last_error,
	m.dead_lettered_at,
	m.created_at,
	m.updated_at
`

	return m.loadExactlyOne(
		ctx,
		conn,
		q,
		pgx.StrictNamedArgs{
			"now":         now,
			"owner_token": uuid.MustNewV4().String(),
		},
	)
}

func (m *BotMessage) UpdateProcessingState(ctx context.Context, conn pg.Querier) error {
	if m.ProcessingOwnerToken == nil || *m.ProcessingOwnerToken == "" {
		panic("bot message processing update requires an owner token")
	}

	q := `
UPDATE bot_messages
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
			"id":                     m.ID,
			"processing_owner_token": m.ProcessingOwnerToken,
			"processing_started_at":  m.ProcessingStartedAt,
			"processed_at":           m.ProcessedAt,
			"attempt_count":          m.AttemptCount,
			"next_attempt_at":        m.NextAttemptAt,
			"last_error":             m.LastError,
			"dead_lettered_at":       m.DeadLetteredAt,
			"updated_at":             m.UpdatedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot update bot message processing state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrProcessingLeaseLost
	}

	return nil
}

func ResetStaleProcessingBotMessages(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	staleAfter time.Duration,
) error {
	q := `
UPDATE bot_messages
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
			"stale_error":     "bot message processing lease expired",
		},
	)
	if err != nil {
		return fmt.Errorf("cannot reset stale processing bot messages: %w", err)
	}

	return nil
}

func DeleteProcessedBotMessagesBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM bot_messages
	WHERE processed_at IS NOT NULL
		AND processed_at < @before
	ORDER BY processed_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM bot_messages
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
		return 0, fmt.Errorf("cannot delete processed bot messages: %w", err)
	}

	return result.RowsAffected(), nil
}

func DeleteDeadLetteredBotMessagesBeforeBatch(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM bot_messages
	WHERE dead_lettered_at IS NOT NULL
		AND dead_lettered_at < @before
	ORDER BY dead_lettered_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM bot_messages
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
		return 0, fmt.Errorf("cannot delete dead-lettered bot messages: %w", err)
	}

	return result.RowsAffected(), nil
}

func (m *BotMessage) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	capability,
	message_type,
	attributes,
	subject_namespace,
	subject_key,
	event_key,
	purpose,
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
FROM bot_messages
WHERE
	%s
	AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	return m.loadExactlyOne(ctx, conn, q, args)
}

func (m *BotMessage) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query bot message: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[BotMessage])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect bot message: %w", err)
	}

	*m = row

	return nil
}
