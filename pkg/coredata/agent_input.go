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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	AgentInputPurpose string

	AgentInput struct {
		ID                gid.GID           `db:"id"`
		OrganizationID    gid.GID           `db:"organization_id"`
		AgentExecutionID  gid.GID           `db:"agent_execution_id"`
		Source            string            `db:"source"`
		SourceEventID     *string           `db:"source_event_id"`
		Purpose           AgentInputPurpose `db:"purpose"`
		IdentityID        *gid.GID          `db:"identity_id"`
		Message           json.RawMessage   `db:"message"`
		SourceCoordinates json.RawMessage   `db:"source_coordinates"`
		ProcessedAt       *time.Time        `db:"processed_at"`
		AttemptCount      int               `db:"attempt_count"`
		MaxAttempts       int               `db:"max_attempts"`
		NextAttemptAt     *time.Time        `db:"next_attempt_at"`
		LastError         *string           `db:"last_error"`
		DeadLetteredAt    *time.Time        `db:"dead_lettered_at"`
		CreatedAt         time.Time         `db:"created_at"`
		UpdatedAt         time.Time         `db:"updated_at"`
	}

	AgentInputs []*AgentInput
)

const (
	AgentInputPurposeUser AgentInputPurpose = "USER"

	AgentInputDefaultMaxAttempts = 5
)

var (
	_ fmt.Stringer             = AgentInputPurpose("")
	_ encoding.TextMarshaler   = AgentInputPurpose("")
	_ encoding.TextUnmarshaler = (*AgentInputPurpose)(nil)
)

func AgentInputPurposes() []AgentInputPurpose {
	return []AgentInputPurpose{
		AgentInputPurposeUser,
	}
}

func (v AgentInputPurpose) IsValid() bool {
	switch v {
	case AgentInputPurposeUser:
		return true
	}

	return false
}

func (v AgentInputPurpose) String() string {
	return string(v)
}

func (v AgentInputPurpose) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AgentInputPurpose) UnmarshalText(text []byte) error {
	value := AgentInputPurpose(text)
	if !value.IsValid() {
		return fmt.Errorf("invalid AgentInputPurpose value: %q", string(text))
	}

	*v = value

	return nil
}

func (i *AgentInput) EnqueueIdempotently(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) (inserted bool, err error) {
	q := `
INSERT INTO agent_inputs (
	id,
	tenant_id,
	organization_id,
	agent_execution_id,
	source,
	source_event_id,
	purpose,
	identity_id,
	message,
	source_coordinates,
	processed_at,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
) SELECT
	@id,
	@tenant_id,
	@organization_id,
	@agent_execution_id,
	@source,
	@source_event_id,
	@purpose,
	@identity_id,
	@message,
	@source_coordinates,
	@processed_at,
	@attempt_count,
	@max_attempts,
	@next_attempt_at,
	@last_error,
	@dead_lettered_at,
	@created_at,
	@updated_at
FROM agent_executions
WHERE
	agent_executions.id = @agent_execution_id
	AND agent_executions.tenant_id = @tenant_id
	AND agent_executions.organization_id = @organization_id
ON CONFLICT (tenant_id, organization_id, source, source_event_id)
	WHERE source_event_id IS NOT NULL
DO UPDATE SET
	updated_at = agent_inputs.updated_at
RETURNING
	id,
	organization_id,
	agent_execution_id,
	source,
	source_event_id,
	purpose,
	identity_id,
	message,
	source_coordinates,
	processed_at,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
`

	if i.MaxAttempts <= 0 {
		i.MaxAttempts = AgentInputDefaultMaxAttempts
	}

	if i.Purpose == "" {
		i.Purpose = AgentInputPurposeUser
	}

	if !i.Purpose.IsValid() {
		return false, fmt.Errorf("cannot enqueue agent input: invalid purpose %q", i.Purpose)
	}

	originalID := i.ID
	args := pgx.StrictNamedArgs{
		"id":                 i.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    i.OrganizationID,
		"agent_execution_id": i.AgentExecutionID,
		"source":             i.Source,
		"source_event_id":    i.SourceEventID,
		"purpose":            i.Purpose,
		"identity_id":        i.IdentityID,
		"message":            i.Message,
		"source_coordinates": i.SourceCoordinates,
		"processed_at":       i.ProcessedAt,
		"attempt_count":      i.AttemptCount,
		"max_attempts":       i.MaxAttempts,
		"next_attempt_at":    i.NextAttemptAt,
		"last_error":         i.LastError,
		"dead_lettered_at":   i.DeadLetteredAt,
		"created_at":         i.CreatedAt,
		"updated_at":         i.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot enqueue agent input: %w", err)
	}

	input, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentInput])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrResourceNotFound
		}

		return false, fmt.Errorf("cannot collect enqueued agent input: %w", err)
	}

	*i = input

	return originalID == i.ID, nil
}

func (i *AgentInput) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	agent_execution_id,
	source,
	source_event_id,
	purpose,
	identity_id,
	message,
	source_coordinates,
	processed_at,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
FROM agent_inputs
WHERE
	%s
	AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agent input: %w", err)
	}

	input, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentInput])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect agent input: %w", err)
	}

	*i = input

	return nil
}

func AgentInputExistsBySourceEventID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	source string,
	sourceEventID string,
) (bool, error) {
	q := `
SELECT EXISTS (
	SELECT 1
	FROM agent_inputs
	WHERE
		%s
		AND organization_id = @organization_id
		AND source = @source
		AND source_event_id = @source_event_id
)
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"source":          source,
		"source_event_id": sourceEventID,
	}
	maps.Copy(args, scope.SQLArguments())

	var exists bool
	if err := conn.QueryRow(ctx, q, args).Scan(&exists); err != nil {
		return false, fmt.Errorf("cannot check agent input source event: %w", err)
	}

	return exists, nil
}

func (is *AgentInputs) LoadPendingByAgentExecutionID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	agentExecutionID gid.GID,
	ownerToken string,
	now time.Time,
	limit int,
) error {
	q := `
SELECT
	id,
	organization_id,
	agent_execution_id,
	source,
	source_event_id,
	purpose,
	identity_id,
	message,
	source_coordinates,
	processed_at,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
FROM agent_inputs
WHERE
	%s
	AND agent_execution_id = @agent_execution_id
	AND processed_at IS NULL
	AND dead_lettered_at IS NULL
	AND attempt_count < max_attempts
	AND (next_attempt_at IS NULL OR next_attempt_at <= @now)
	AND EXISTS (
		SELECT 1
		FROM agent_executions
		WHERE
			agent_executions.id = agent_inputs.agent_execution_id
			AND agent_executions.processing_owner_token = @owner_token
	)
ORDER BY created_at ASC, id ASC
LIMIT @limit
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"agent_execution_id": agentExecutionID,
		"owner_token":        ownerToken,
		"now":                now,
		"limit":              limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query pending agent inputs: %w", err)
	}

	inputs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AgentInput])
	if err != nil {
		return fmt.Errorf("cannot collect pending agent inputs: %w", err)
	}

	*is = inputs

	return nil
}

func (is *AgentInputs) LoadPendingByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	agentExecutionID gid.GID,
	ownerToken string,
	inputIDs []string,
	limit int,
) error {
	q := `
SELECT
	id,
	organization_id,
	agent_execution_id,
	source,
	source_event_id,
	purpose,
	identity_id,
	message,
	source_coordinates,
	processed_at,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
FROM agent_inputs
WHERE
	%s
	AND agent_execution_id = @agent_execution_id
	AND id = ANY(@input_ids::text[])
	AND processed_at IS NULL
	AND dead_lettered_at IS NULL
	AND EXISTS (
		SELECT 1
		FROM agent_executions
		WHERE
			agent_executions.id = agent_inputs.agent_execution_id
			AND agent_executions.processing_owner_token = @owner_token
	)
ORDER BY created_at ASC, id ASC
LIMIT @limit
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"agent_execution_id": agentExecutionID,
		"owner_token":        ownerToken,
		"input_ids":          inputIDs,
		"limit":              limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agent inputs by processing IDs: %w", err)
	}

	inputs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AgentInput])
	if err != nil {
		return fmt.Errorf("cannot collect agent inputs by processing IDs: %w", err)
	}

	*is = inputs

	return nil
}

func (i *AgentInput) MarkProcessed(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	now time.Time,
) error {
	q := `
UPDATE agent_inputs
SET
	processed_at = @now,
	next_attempt_at = NULL,
	last_error = NULL,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processed_at IS NULL
	AND dead_lettered_at IS NULL
	AND EXISTS (
		SELECT 1
		FROM agent_executions
		WHERE
			agent_executions.id = agent_inputs.agent_execution_id
			AND agent_executions.processing_owner_token = @owner_token
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          i.ID,
		"owner_token": ownerToken,
		"now":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark agent input processed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	i.ProcessedAt = &now
	i.NextAttemptAt = nil
	i.LastError = nil
	i.UpdatedAt = now

	return nil
}

func (i *AgentInput) RecordFailure(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	lastError string,
	nextAttemptAt time.Time,
	now time.Time,
) error {
	q := `
UPDATE agent_inputs
SET
	attempt_count = attempt_count + 1,
	next_attempt_at = @next_attempt_at,
	last_error = @last_error,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processed_at IS NULL
	AND dead_lettered_at IS NULL
	AND attempt_count + 1 < max_attempts
	AND EXISTS (
		SELECT 1
		FROM agent_executions
		WHERE
			agent_executions.id = agent_inputs.agent_execution_id
			AND agent_executions.processing_owner_token = @owner_token
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":              i.ID,
		"owner_token":     ownerToken,
		"next_attempt_at": nextAttemptAt,
		"last_error":      lastError,
		"now":             now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot record agent input failure: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	i.AttemptCount++
	i.NextAttemptAt = &nextAttemptAt
	i.LastError = &lastError
	i.UpdatedAt = now

	return nil
}

func (i *AgentInput) DeadLetter(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	lastError string,
	now time.Time,
) error {
	q := `
UPDATE agent_inputs
SET
	attempt_count = attempt_count + 1,
	next_attempt_at = NULL,
	last_error = @last_error,
	dead_lettered_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processed_at IS NULL
	AND dead_lettered_at IS NULL
	AND EXISTS (
		SELECT 1
		FROM agent_executions
		WHERE
			agent_executions.id = agent_inputs.agent_execution_id
			AND agent_executions.processing_owner_token = @owner_token
	)
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          i.ID,
		"owner_token": ownerToken,
		"last_error":  lastError,
		"now":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot dead-letter agent input: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	i.AttemptCount++
	i.NextAttemptAt = nil
	i.LastError = &lastError
	i.DeadLetteredAt = &now
	i.UpdatedAt = now

	return nil
}

func DeadLetterAgentInputsForStaleExecutions(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	staleAfter time.Duration,
	lastError string,
) error {
	q := `
UPDATE agent_inputs
SET
	attempt_count = max_attempts,
	next_attempt_at = NULL,
	last_error = @last_error,
	dead_lettered_at = @now,
	updated_at = @now
WHERE
	processed_at IS NULL
	AND dead_lettered_at IS NULL
	AND EXISTS (
		SELECT 1
		FROM agent_executions
		WHERE
			agent_executions.id = agent_inputs.agent_execution_id
			AND agent_inputs.id = ANY(agent_executions.processing_input_ids)
			AND agent_executions.processing_owner_token IS NOT NULL
			AND agent_executions.processing_heartbeat_at <= @stale_before
			AND agent_executions.attempt_count >= agent_executions.max_attempts
	)
`

	args := pgx.StrictNamedArgs{
		"now":          now,
		"stale_before": now.Add(-staleAfter),
		"last_error":   lastError,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot dead-letter inputs for stale agent executions: %w", err)
	}

	return nil
}

func DeleteRetiredAgentInputsBefore(
	ctx context.Context,
	conn pg.Querier,
	processedBefore time.Time,
	deadLetterBefore time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM agent_inputs
	WHERE
		(processed_at IS NOT NULL AND processed_at < @processed_before)
		OR (dead_lettered_at IS NOT NULL AND dead_lettered_at < @dead_letter_before)
	ORDER BY updated_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM agent_inputs
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"processed_before":   processedBefore,
			"dead_letter_before": deadLetterBefore,
			"limit":              limit,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete retired agent inputs: %w", err)
	}

	return result.RowsAffected(), nil
}
