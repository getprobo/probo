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
	"go.probo.inc/probo/pkg/gid"
)

const SlackbotInteractiveCommandDefaultMaxAttempts = 5

// SlackbotInteractiveCommand is the Slack interactive-command inbox.
// OrganizationID is optional until the worker decrypts and resolves the payload.
type SlackbotInteractiveCommand struct {
	ID                  gid.GID    `db:"id"`
	OrganizationID      *gid.GID   `db:"organization_id"`
	RequestDigest       []byte     `db:"request_digest"`
	EncryptedPayload    []byte     `db:"encrypted_payload"`
	ProcessingStartedAt *time.Time `db:"processing_started_at"`
	ProcessedAt         *time.Time `db:"processed_at"`
	AttemptCount        int        `db:"attempt_count"`
	MaxAttempts         int        `db:"max_attempts"`
	NextAttemptAt       *time.Time `db:"next_attempt_at"`
	LastError           *string    `db:"last_error"`
	DeadLetteredAt      *time.Time `db:"dead_lettered_at"`
	CreatedAt           time.Time  `db:"created_at"`
	UpdatedAt           time.Time  `db:"updated_at"`
}

func NewSlackbotInteractiveCommand(
	requestDigest []byte,
	encryptedPayload []byte,
) *SlackbotInteractiveCommand {
	now := time.Now()

	return &SlackbotInteractiveCommand{
		ID:               gid.New(gid.NilTenant, SlackbotInteractiveCommandEntityType),
		RequestDigest:    requestDigest,
		EncryptedPayload: encryptedPayload,
		MaxAttempts:      SlackbotInteractiveCommandDefaultMaxAttempts,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (c *SlackbotInteractiveCommand) Insert(ctx context.Context, conn pg.Querier) (bool, error) {
	q := `
INSERT INTO slackbot_interactive_commands (
	id, tenant_id, organization_id, request_digest, encrypted_payload,
	processing_started_at, processed_at, attempt_count, max_attempts,
	next_attempt_at, last_error, dead_lettered_at, created_at, updated_at
) VALUES (
	@id, NULL, NULL, @request_digest, @encrypted_payload,
	@processing_started_at, @processed_at, @attempt_count, @max_attempts,
	@next_attempt_at, @last_error, @dead_lettered_at, @created_at, @updated_at
)
ON CONFLICT (request_digest) DO NOTHING
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":                    c.ID,
			"request_digest":        c.RequestDigest,
			"encrypted_payload":     c.EncryptedPayload,
			"processing_started_at": c.ProcessingStartedAt,
			"processed_at":          c.ProcessedAt,
			"attempt_count":         c.AttemptCount,
			"max_attempts":          c.MaxAttempts,
			"next_attempt_at":       c.NextAttemptAt,
			"last_error":            c.LastError,
			"dead_lettered_at":      c.DeadLetteredAt,
			"created_at":            c.CreatedAt,
			"updated_at":            c.UpdatedAt,
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot insert Slackbot interactive command: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (c *SlackbotInteractiveCommand) ResetDeadLetteredByRequestDigest(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
) (bool, error) {
	q := `
UPDATE slackbot_interactive_commands
SET
	processing_started_at = NULL,
	processed_at = NULL,
	dead_lettered_at = NULL,
	next_attempt_at = NULL,
	attempt_count = 0,
	last_error = NULL,
	updated_at = @now
WHERE request_digest = @request_digest
	AND dead_lettered_at IS NOT NULL
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"request_digest": c.RequestDigest,
			"now":            now,
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot reset dead-lettered Slackbot interactive command: %w", err)
	}

	return result.RowsAffected() > 0, nil
}

func (c *SlackbotInteractiveCommand) LoadByRequestDigest(
	ctx context.Context,
	conn pg.Querier,
	requestDigest []byte,
) error {
	q := `
SELECT
	id,
	organization_id,
	request_digest,
	encrypted_payload,
	processing_started_at,
	processed_at,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
FROM slackbot_interactive_commands
WHERE request_digest = @request_digest
`

	return c.loadExactlyOne(
		ctx,
		conn,
		q,
		pgx.StrictNamedArgs{"request_digest": requestDigest},
	)
}

func (c *SlackbotInteractiveCommand) ClaimNextForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
WITH candidate AS (
	SELECT id
	FROM slackbot_interactive_commands
	WHERE processed_at IS NULL
		AND processing_started_at IS NULL
		AND dead_lettered_at IS NULL
		AND attempt_count < max_attempts
		AND (next_attempt_at IS NULL OR next_attempt_at <= @now)
	ORDER BY created_at ASC, id ASC
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
UPDATE slackbot_interactive_commands command
SET
	attempt_count = command.attempt_count + 1,
	processing_started_at = @now,
	updated_at = @now
FROM candidate
WHERE command.id = candidate.id
RETURNING
	command.id, command.organization_id, command.request_digest,
	command.encrypted_payload, command.processing_started_at,
	command.processed_at, command.attempt_count, command.max_attempts,
	command.next_attempt_at, command.last_error, command.dead_lettered_at,
	command.created_at, command.updated_at
`

	return c.loadExactlyOne(ctx, conn, q, pgx.StrictNamedArgs{"now": now})
}

func (c *SlackbotInteractiveCommand) UpdateProcessingState(
	ctx context.Context,
	conn pg.Querier,
) error {
	var tenantID *gid.TenantID

	if c.OrganizationID != nil {
		value := c.OrganizationID.TenantID()
		tenantID = &value
	}

	q := `
UPDATE slackbot_interactive_commands
SET
	tenant_id = @tenant_id,
	organization_id = @organization_id,
	processing_started_at = @processing_started_at,
	processed_at = @processed_at,
	next_attempt_at = @next_attempt_at,
	last_error = @last_error,
	dead_lettered_at = @dead_lettered_at,
	updated_at = @updated_at
WHERE id = @id
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":                    c.ID,
			"tenant_id":             tenantID,
			"organization_id":       c.OrganizationID,
			"processing_started_at": c.ProcessingStartedAt,
			"processed_at":          c.ProcessedAt,
			"next_attempt_at":       c.NextAttemptAt,
			"last_error":            c.LastError,
			"dead_lettered_at":      c.DeadLetteredAt,
			"updated_at":            c.UpdatedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot update Slackbot interactive command: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func ResetStaleSlackbotInteractiveCommands(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	staleAfter time.Duration,
) error {
	q := `
UPDATE slackbot_interactive_commands
SET
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
			"stale_error":     "Slack interactive command processing lease expired",
		},
	)
	if err != nil {
		return fmt.Errorf("cannot reset stale Slackbot interactive commands: %w", err)
	}

	return nil
}

func DeleteCompletedSlackbotInteractiveCommandsBefore(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM slackbot_interactive_commands
	WHERE processed_at IS NOT NULL
		AND processed_at < @before
	ORDER BY processed_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM slackbot_interactive_commands
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
		return 0, fmt.Errorf("cannot delete completed Slackbot interactive commands: %w", err)
	}

	return result.RowsAffected(), nil
}

func DeleteDeadLetteredSlackbotInteractiveCommandsBefore(
	ctx context.Context,
	conn pg.Querier,
	before time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM slackbot_interactive_commands
	WHERE dead_lettered_at IS NOT NULL
		AND dead_lettered_at < @before
	ORDER BY dead_lettered_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM slackbot_interactive_commands
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{"before": before, "limit": limit},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete dead-lettered Slackbot interactive commands: %w", err)
	}

	return result.RowsAffected(), nil
}

func (c *SlackbotInteractiveCommand) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Slackbot interactive command: %w", err)
	}

	command, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[SlackbotInteractiveCommand],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect Slackbot interactive command: %w", err)
	}

	*c = command

	return nil
}
