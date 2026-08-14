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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	AgentExecutionStatus string

	AgentExecution struct {
		ID                    gid.GID              `db:"id"`
		OrganizationID        gid.GID              `db:"organization_id"`
		StartAgentName        string               `db:"start_agent_name"`
		Status                AgentExecutionStatus `db:"status"`
		Checkpoint            json.RawMessage      `db:"checkpoint"`
		ErrorMessage          *string              `db:"error_message"`
		StartedAt             *time.Time           `db:"started_at"`
		Source                *string              `db:"source"`
		SessionKey            *string              `db:"session_key"`
		SourceCoordinates     json.RawMessage      `db:"source_coordinates"`
		TrustedContext        json.RawMessage      `db:"trusted_context"`
		SessionMessages       json.RawMessage      `db:"session_messages"`
		ProcessingOwnerToken  *string              `db:"processing_owner_token"`
		ProcessingHeartbeatAt *time.Time           `db:"processing_heartbeat_at"`
		ProcessingInputIDs    []string             `db:"processing_input_ids"`
		AttemptCount          int                  `db:"attempt_count"`
		MaxAttempts           int                  `db:"max_attempts"`
		NextAttemptAt         *time.Time           `db:"next_attempt_at"`
		LastError             *string              `db:"last_error"`
		DeadLetteredAt        *time.Time           `db:"dead_lettered_at"`
		CreatedAt             time.Time            `db:"created_at"`
		UpdatedAt             time.Time            `db:"updated_at"`
	}

	AgentExecutions []*AgentExecution
)

const (
	AgentExecutionDefaultMaxAttempts = 5
	AgentExecutionStaleLeaseError    = "agent execution processing lease expired"

	AgentExecutionStatusIdle             AgentExecutionStatus = "IDLE"
	AgentExecutionStatusRunning          AgentExecutionStatus = "RUNNING"
	AgentExecutionStatusSuspended        AgentExecutionStatus = "SUSPENDED"
	AgentExecutionStatusAwaitingApproval AgentExecutionStatus = "AWAITING_APPROVAL"
	AgentExecutionStatusFailed           AgentExecutionStatus = "FAILED"
)

var (
	_ fmt.Stringer             = AgentExecutionStatus("")
	_ encoding.TextMarshaler   = AgentExecutionStatus("")
	_ encoding.TextUnmarshaler = (*AgentExecutionStatus)(nil)
)

func AgentExecutionStatuses() []AgentExecutionStatus {
	return []AgentExecutionStatus{
		AgentExecutionStatusIdle,
		AgentExecutionStatusRunning,
		AgentExecutionStatusSuspended,
		AgentExecutionStatusAwaitingApproval,
		AgentExecutionStatusFailed,
	}
}

func (v AgentExecutionStatus) IsValid() bool {
	switch v {
	case
		AgentExecutionStatusIdle,
		AgentExecutionStatusRunning,
		AgentExecutionStatusSuspended,
		AgentExecutionStatusAwaitingApproval,
		AgentExecutionStatusFailed:
		return true
	}

	return false
}

func (v AgentExecutionStatus) String() string {
	return string(v)
}

func (v AgentExecutionStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AgentExecutionStatus) UnmarshalText(text []byte) error {
	val := AgentExecutionStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AgentExecutionStatus value: %q", string(text))
	}

	*v = val

	return nil
}

func (e AgentExecution) CursorKey(orderBy AgentExecutionOrderField) page.CursorKey {
	switch orderBy {
	case AgentExecutionOrderFieldCreatedAt:
		return page.NewCursorKey(e.ID, e.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (e *AgentExecution) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM agent_executions WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (e *AgentExecution) LoadByIDForUpdate(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	error_message,
	started_at,
	source,
	session_key,
	source_coordinates,
	trusted_context,
	session_messages,
	processing_owner_token,
	processing_heartbeat_at,
	processing_input_ids,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
FROM
	agent_executions
WHERE
	%s
	AND id = @id
LIMIT 1
FOR UPDATE;
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	return e.loadExactlyOne(ctx, tx, q, args)
}

func (es *AgentExecutions) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AgentExecutionOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	error_message,
	started_at,
	source,
	session_key,
	source_coordinates,
	trusted_context,
	session_messages,
	processing_owner_token,
	processing_heartbeat_at,
	processing_input_ids,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
FROM
	agent_executions
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID.String()}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agent executions: %w", err)
	}

	entities, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AgentExecution])
	if err != nil {
		return fmt.Errorf("cannot collect agent executions: %w", err)
	}

	*es = entities

	return nil
}

func (es *AgentExecutions) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	agent_executions
WHERE
	%s
	AND organization_id = @organization_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID.String()}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count agent executions: %w", err)
	}

	return count, nil
}

func (e *AgentExecution) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO agent_executions (
	id,
	tenant_id,
	organization_id,
	start_agent_name,
	status,
	session_messages,
	processing_input_ids,
	attempt_count,
	max_attempts,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@start_agent_name,
	@status,
	@session_messages,
	@processing_input_ids,
	@attempt_count,
	@max_attempts,
	@created_at,
	@updated_at
)
RETURNING
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	error_message,
	started_at,
	source,
	session_key,
	source_coordinates,
	trusted_context,
	session_messages,
	processing_owner_token,
	processing_heartbeat_at,
	processing_input_ids,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at;
`

	if e.SessionMessages == nil {
		e.SessionMessages = json.RawMessage("[]")
	}

	if e.MaxAttempts <= 0 {
		e.MaxAttempts = AgentExecutionDefaultMaxAttempts
	}

	args := pgx.StrictNamedArgs{
		"id":                   e.ID.String(),
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      e.OrganizationID.String(),
		"start_agent_name":     e.StartAgentName,
		"status":               e.Status,
		"session_messages":     e.SessionMessages,
		"processing_input_ids": []string{},
		"attempt_count":        e.AttemptCount,
		"max_attempts":         e.MaxAttempts,
		"created_at":           e.CreatedAt,
		"updated_at":           e.UpdatedAt,
	}

	return e.loadExactlyOne(ctx, tx, q, args)
}

func (e *AgentExecution) RequeueForApprovalResume(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE agent_executions
SET
	checkpoint = @checkpoint,
	status = @status,
	started_at = @started_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND status = @expected_status;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":              e.ID.String(),
		"checkpoint":      e.Checkpoint,
		"status":          e.Status,
		"started_at":      e.StartedAt,
		"updated_at":      e.UpdatedAt,
		"expected_status": AgentExecutionStatusAwaitingApproval,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot requeue agent execution for approval resume: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (e *AgentExecution) UpsertBySourceSession(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) (inserted bool, err error) {
	if e.Source == nil || *e.Source == "" {
		return false, fmt.Errorf("cannot upsert agent execution: source is required")
	}

	if e.SessionKey == nil || *e.SessionKey == "" {
		return false, fmt.Errorf("cannot upsert agent execution: session key is required")
	}

	q := `
INSERT INTO agent_executions (
	id,
	tenant_id,
	organization_id,
	start_agent_name,
	status,
	source,
	session_key,
	source_coordinates,
	trusted_context,
	session_messages,
	processing_input_ids,
	attempt_count,
	max_attempts,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@start_agent_name,
	@status,
	@source,
	@session_key,
	@source_coordinates,
	@trusted_context,
	@session_messages,
	@processing_input_ids,
	@attempt_count,
	@max_attempts,
	@created_at,
	@updated_at
)
ON CONFLICT (tenant_id, organization_id, source, session_key)
	WHERE source IS NOT NULL AND session_key IS NOT NULL
DO UPDATE SET
	start_agent_name = EXCLUDED.start_agent_name,
	source_coordinates = EXCLUDED.source_coordinates,
	trusted_context = COALESCE(EXCLUDED.trusted_context, agent_executions.trusted_context),
	updated_at = EXCLUDED.updated_at
RETURNING
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	error_message,
	started_at,
	source,
	session_key,
	source_coordinates,
	trusted_context,
	session_messages,
	processing_owner_token,
	processing_heartbeat_at,
	processing_input_ids,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
`

	if e.MaxAttempts <= 0 {
		e.MaxAttempts = AgentExecutionDefaultMaxAttempts
	}

	if e.SessionMessages == nil {
		e.SessionMessages = json.RawMessage("[]")
	}

	originalID := e.ID
	args := pgx.StrictNamedArgs{
		"id":                   e.ID,
		"tenant_id":            scope.GetTenantID(),
		"organization_id":      e.OrganizationID,
		"start_agent_name":     e.StartAgentName,
		"status":               AgentExecutionStatusIdle,
		"source":               e.Source,
		"session_key":          e.SessionKey,
		"source_coordinates":   e.SourceCoordinates,
		"trusted_context":      e.TrustedContext,
		"session_messages":     e.SessionMessages,
		"processing_input_ids": []string{},
		"attempt_count":        e.AttemptCount,
		"max_attempts":         e.MaxAttempts,
		"created_at":           e.CreatedAt,
		"updated_at":           e.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert agent execution: %w", err)
	}

	execution, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentExecution])
	if err != nil {
		return false, fmt.Errorf("cannot collect agent execution: %w", err)
	}

	*e = execution

	return originalID == e.ID, nil
}

func (e *AgentExecution) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	error_message,
	started_at,
	source,
	session_key,
	source_coordinates,
	trusted_context,
	session_messages,
	processing_owner_token,
	processing_heartbeat_at,
	processing_input_ids,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
FROM agent_executions
WHERE
	%s
	AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	return e.loadExactlyOne(ctx, conn, q, args)
}

func (e *AgentExecution) UpdateSourceCoordinates(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	coordinates json.RawMessage,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	source_coordinates = @source_coordinates,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":                 e.ID,
		"source_coordinates": coordinates,
		"updated_at":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	commandTag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update agent execution source coordinates: %w", err)
	}

	if commandTag.RowsAffected() != 1 {
		return ErrResourceNotFound
	}

	e.SourceCoordinates = coordinates
	e.UpdatedAt = now

	return nil
}

func (e *AgentExecution) ClaimNextForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
	ownerToken string,
) error {
	q := `
WITH candidate AS (
	SELECT id
	FROM agent_executions
	WHERE
		status IN (@idle_status, @suspended_status)
		AND processing_owner_token IS NULL
		AND dead_lettered_at IS NULL
		AND attempt_count < max_attempts
		AND (next_attempt_at IS NULL OR next_attempt_at <= @now)
		AND (
			checkpoint IS NOT NULL
			OR cardinality(processing_input_ids) > 0
			OR EXISTS (
				SELECT 1
				FROM agent_inputs
				WHERE
					agent_inputs.agent_execution_id = agent_executions.id
					AND agent_inputs.processed_at IS NULL
					AND agent_inputs.dead_lettered_at IS NULL
					AND agent_inputs.attempt_count < agent_inputs.max_attempts
					AND (
						agent_inputs.next_attempt_at IS NULL
						OR agent_inputs.next_attempt_at <= @now
					)
			)
		)
	ORDER BY updated_at ASC, id ASC
	LIMIT 1
	FOR UPDATE SKIP LOCKED
)
UPDATE agent_executions
SET
	status = @running_status,
	started_at = @now,
	processing_owner_token = @owner_token,
	processing_heartbeat_at = @now,
	attempt_count = attempt_count + 1,
	updated_at = @now
WHERE id = (SELECT id FROM candidate)
RETURNING
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	error_message,
	started_at,
	source,
	session_key,
	source_coordinates,
	trusted_context,
	session_messages,
	processing_owner_token,
	processing_heartbeat_at,
	processing_input_ids,
	attempt_count,
	max_attempts,
	next_attempt_at,
	last_error,
	dead_lettered_at,
	created_at,
	updated_at
`

	args := pgx.StrictNamedArgs{
		"idle_status":      AgentExecutionStatusIdle,
		"suspended_status": AgentExecutionStatusSuspended,
		"running_status":   AgentExecutionStatusRunning,
		"now":              now,
		"owner_token":      ownerToken,
	}

	return e.loadExactlyOne(ctx, conn, q, args)
}

func (e *AgentExecution) UpdateSessionState(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	checkpoint = @checkpoint,
	trusted_context = @trusted_context,
	session_messages = @session_messages,
	processing_heartbeat_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processing_owner_token = @owner_token
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":               e.ID,
		"checkpoint":       e.Checkpoint,
		"trusted_context":  e.TrustedContext,
		"session_messages": e.SessionMessages,
		"owner_token":      ownerToken,
		"now":              now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update agent execution session state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	e.ProcessingHeartbeatAt = &now
	e.UpdatedAt = now

	return nil
}

func (e *AgentExecution) SaveCheckpoint(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	checkpoint json.RawMessage,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	checkpoint = @checkpoint,
	processing_heartbeat_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processing_owner_token = @owner_token
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          e.ID,
		"checkpoint":  checkpoint,
		"owner_token": ownerToken,
		"now":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot save agent execution checkpoint: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	e.Checkpoint = checkpoint
	e.ProcessingHeartbeatAt = &now
	e.UpdatedAt = now

	return nil
}

func (e *AgentExecution) SetProcessingInputIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	inputIDs []string,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	processing_input_ids = @input_ids,
	processing_heartbeat_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processing_owner_token = @owner_token
	AND cardinality(processing_input_ids) = 0
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          e.ID,
		"input_ids":   inputIDs,
		"owner_token": ownerToken,
		"now":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot set agent execution processing inputs: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	e.ProcessingInputIDs = inputIDs
	e.ProcessingHeartbeatAt = &now
	e.UpdatedAt = now

	return nil
}

func (e *AgentExecution) ClearProcessingInputIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	processing_input_ids = '{}',
	processing_heartbeat_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processing_owner_token = @owner_token
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          e.ID,
		"owner_token": ownerToken,
		"now":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot clear agent execution processing inputs: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	e.ProcessingInputIDs = []string{}
	e.ProcessingHeartbeatAt = &now
	e.UpdatedAt = now

	return nil
}

func (e *AgentExecution) CommitSuccess(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	status = @status,
	checkpoint = NULL,
	session_messages = @session_messages,
	processing_input_ids = '{}',
	started_at = NULL,
	processing_owner_token = NULL,
	processing_heartbeat_at = NULL,
	attempt_count = 0,
	next_attempt_at = NULL,
	last_error = NULL,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processing_owner_token = @owner_token
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":               e.ID,
		"status":           AgentExecutionStatusIdle,
		"session_messages": e.SessionMessages,
		"owner_token":      ownerToken,
		"now":              now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot commit agent execution: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	e.Status = AgentExecutionStatusIdle
	e.Checkpoint = nil
	e.ProcessingInputIDs = nil
	e.StartedAt = nil
	e.ProcessingOwnerToken = nil
	e.ProcessingHeartbeatAt = nil
	e.AttemptCount = 0
	e.NextAttemptAt = nil
	e.LastError = nil
	e.UpdatedAt = now

	return nil
}

func (e *AgentExecution) Heartbeat(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	processing_heartbeat_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processing_owner_token = @owner_token
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          e.ID,
		"owner_token": ownerToken,
		"now":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot heartbeat agent execution: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	e.ProcessingHeartbeatAt = &now
	e.UpdatedAt = now

	return nil
}

func (e *AgentExecution) Release(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	now time.Time,
) error {
	return e.ReleaseToStatus(
		ctx,
		conn,
		scope,
		ownerToken,
		AgentExecutionStatusIdle,
		now,
	)
}

func (e *AgentExecution) ReleaseToStatus(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	status AgentExecutionStatus,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	status = @status,
	started_at = NULL,
	processing_owner_token = NULL,
	processing_heartbeat_at = NULL,
	attempt_count = 0,
	next_attempt_at = NULL,
	last_error = NULL,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processing_owner_token = @owner_token
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          e.ID,
		"owner_token": ownerToken,
		"status":      status,
		"now":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot release agent execution: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	e.Status = status
	e.StartedAt = nil
	e.ProcessingOwnerToken = nil
	e.ProcessingHeartbeatAt = nil
	e.AttemptCount = 0
	e.NextAttemptAt = nil
	e.LastError = nil
	e.UpdatedAt = now

	return nil
}

func (e *AgentExecution) RecordFailure(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	lastError string,
	nextAttemptAt time.Time,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	status = @status,
	started_at = NULL,
	processing_owner_token = NULL,
	processing_heartbeat_at = NULL,
	next_attempt_at = @next_attempt_at,
	last_error = @last_error,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processing_owner_token = @owner_token
	AND attempt_count < max_attempts
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":              e.ID,
		"owner_token":     ownerToken,
		"status":          AgentExecutionStatusIdle,
		"next_attempt_at": nextAttemptAt,
		"last_error":      lastError,
		"now":             now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot record agent execution failure: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	e.Status = AgentExecutionStatusIdle
	e.StartedAt = nil
	e.ProcessingOwnerToken = nil
	e.ProcessingHeartbeatAt = nil
	e.NextAttemptAt = &nextAttemptAt
	e.LastError = &lastError
	e.UpdatedAt = now

	return nil
}

func (e *AgentExecution) DeadLetter(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ownerToken string,
	lastError string,
	now time.Time,
) error {
	q := `
UPDATE agent_executions
SET
	status = @status,
	started_at = NULL,
	processing_owner_token = NULL,
	processing_heartbeat_at = NULL,
	processing_input_ids = '{}',
	next_attempt_at = NULL,
	last_error = @last_error,
	error_message = @last_error,
	dead_lettered_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND processing_owner_token = @owner_token
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":          e.ID,
		"owner_token": ownerToken,
		"status":      AgentExecutionStatusFailed,
		"last_error":  lastError,
		"now":         now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot dead-letter agent execution: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	e.Status = AgentExecutionStatusFailed
	e.StartedAt = nil
	e.ProcessingOwnerToken = nil
	e.ProcessingHeartbeatAt = nil
	e.ProcessingInputIDs = nil
	e.NextAttemptAt = nil
	e.LastError = &lastError
	e.ErrorMessage = &lastError
	e.DeadLetteredAt = &now
	e.UpdatedAt = now

	return nil
}

func ResetStaleAgentExecutionLeases(
	ctx context.Context,
	conn pg.Querier,
	now time.Time,
	staleAfter time.Duration,
) error {
	q := `
UPDATE agent_executions
SET
	status = CASE
		WHEN attempt_count >= max_attempts THEN @failed_status
		WHEN checkpoint IS NOT NULL THEN @suspended_status
		ELSE @idle_status
	END,
	started_at = NULL,
	processing_owner_token = NULL,
	processing_heartbeat_at = NULL,
	processing_input_ids = CASE
		WHEN attempt_count >= max_attempts THEN '{}'
		ELSE processing_input_ids
	END,
	next_attempt_at = CASE
		WHEN attempt_count >= max_attempts THEN NULL
		ELSE @now::timestamptz
	END,
	last_error = @last_error,
	error_message = CASE
		WHEN attempt_count >= max_attempts THEN @last_error
		ELSE error_message
	END,
	dead_lettered_at = CASE
		WHEN attempt_count >= max_attempts THEN @now::timestamptz
		ELSE NULL
	END,
	updated_at = @now
WHERE
	processing_owner_token IS NOT NULL
	AND processing_heartbeat_at <= @stale_before
`

	args := pgx.StrictNamedArgs{
		"failed_status":    AgentExecutionStatusFailed,
		"idle_status":      AgentExecutionStatusIdle,
		"suspended_status": AgentExecutionStatusSuspended,
		"now":              now,
		"last_error":       AgentExecutionStaleLeaseError,
		"stale_before":     now.Add(-staleAfter),
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot reset stale agent execution leases: %w", err)
	}

	return nil
}

func DeleteRetiredAgentExecutionsBefore(
	ctx context.Context,
	conn pg.Querier,
	idleBefore time.Time,
	deadLetterBefore time.Time,
	limit int,
) (int64, error) {
	q := `
WITH doomed AS (
	SELECT id
	FROM agent_executions
	WHERE
		processing_owner_token IS NULL
		AND checkpoint IS NULL
		AND (
			(dead_lettered_at IS NOT NULL AND dead_lettered_at < @dead_letter_before)
			OR (
				dead_lettered_at IS NULL
				AND updated_at < @idle_before
				AND NOT EXISTS (
					SELECT 1
					FROM agent_inputs
					WHERE
						agent_inputs.agent_execution_id = agent_executions.id
						AND agent_inputs.processed_at IS NULL
						AND agent_inputs.dead_lettered_at IS NULL
				)
			)
		)
	ORDER BY updated_at ASC, id ASC
	LIMIT @limit
)
DELETE FROM agent_executions
WHERE id IN (SELECT id FROM doomed)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"idle_before":        idleBefore,
			"dead_letter_before": deadLetterBefore,
			"limit":              limit,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot delete retired agent executions: %w", err)
	}

	return result.RowsAffected(), nil
}

func (e *AgentExecution) loadExactlyOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agent execution: %w", err)
	}

	execution, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentExecution])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect agent execution: %w", err)
	}

	*e = execution

	return nil
}
