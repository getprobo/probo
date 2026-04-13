// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	AgentRunStatus string

	AgentRun struct {
		ID             gid.GID         `db:"id"`
		OrganizationID gid.GID         `db:"organization_id"`
		StartAgentName string          `db:"start_agent_name"`
		Status         AgentRunStatus  `db:"status"`
		Checkpoint     json.RawMessage `db:"checkpoint"`
		InputMessages  json.RawMessage `db:"input_messages"`
		Result         json.RawMessage `db:"result"`
		ErrorMessage   *string         `db:"error_message"`
		StopRequested  bool            `db:"stop_requested"`
		StartedAt      *time.Time      `db:"started_at"`
		LeaseOwner     *string         `db:"lease_owner"`
		LeaseExpiresAt *time.Time      `db:"lease_expires_at"`
		CreatedAt      time.Time       `db:"created_at"`
		UpdatedAt      time.Time       `db:"updated_at"`
	}

	AgentRuns []*AgentRun
)

const (
	AgentRunStatusPending          AgentRunStatus = "PENDING"
	AgentRunStatusRunning          AgentRunStatus = "RUNNING"
	AgentRunStatusSuspended        AgentRunStatus = "SUSPENDED"
	AgentRunStatusAwaitingApproval AgentRunStatus = "AWAITING_APPROVAL"
	AgentRunStatusCompleted        AgentRunStatus = "COMPLETED"
	AgentRunStatusFailed           AgentRunStatus = "FAILED"
)

func (e AgentRun) CursorKey(orderBy AgentRunOrderField) page.CursorKey {
	switch orderBy {
	case AgentRunOrderFieldCreatedAt:
		return page.NewCursorKey(e.ID, e.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (e *AgentRun) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM agent_runs WHERE id = @id LIMIT 1;`

	args := pgx.StrictNamedArgs{"id": e.ID.String()}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query agent run authorization attributes: %w", err)
	}

	type row struct {
		OrganizationID gid.GID `db:"organization_id"`
	}

	r, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[row])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot load agent run authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": r.OrganizationID.String()}, nil
}

func (e *AgentRun) LoadByID(
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
	input_messages,
	result,
	error_message,
	stop_requested,
	started_at,
	lease_owner,
	lease_expires_at,
	created_at,
	updated_at
FROM
	agent_runs
WHERE
	%s
	AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id.String()}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agent run: %w", err)
	}

	entity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentRun])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot load agent run: %w", err)
	}

	*e = entity

	return nil
}

func (e *AgentRun) LoadByIDForUpdate(
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
	input_messages,
	result,
	error_message,
	stop_requested,
	started_at,
	lease_owner,
	lease_expires_at,
	created_at,
	updated_at
FROM
	agent_runs
WHERE
	%s
	AND id = @id
LIMIT 1
FOR UPDATE;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id.String()}
	maps.Copy(args, scope.SQLArguments())

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agent run: %w", err)
	}

	entity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentRun])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot load agent run: %w", err)
	}

	*e = entity

	return nil
}

func (rs *AgentRuns) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AgentRunOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	input_messages,
	result,
	error_message,
	stop_requested,
	started_at,
	lease_owner,
	lease_expires_at,
	created_at,
	updated_at
FROM
	agent_runs
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
		return fmt.Errorf("cannot query agent runs: %w", err)
	}

	entities, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AgentRun])
	if err != nil {
		return fmt.Errorf("cannot collect agent runs: %w", err)
	}

	*rs = entities

	return nil
}

func (rs *AgentRuns) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	agent_runs
WHERE
	%s
	AND organization_id = @organization_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID.String()}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count agent runs: %w", err)
	}

	return count, nil
}

func (e *AgentRun) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO agent_runs (
	id,
	tenant_id,
	organization_id,
	start_agent_name,
	status,
	input_messages,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@start_agent_name,
	@status,
	@input_messages,
	@created_at,
	@updated_at
)
RETURNING
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	input_messages,
	result,
	error_message,
	stop_requested,
	started_at,
	lease_owner,
	lease_expires_at,
	created_at,
	updated_at;
`

	args := pgx.StrictNamedArgs{
		"id":               e.ID.String(),
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  e.OrganizationID.String(),
		"start_agent_name": e.StartAgentName,
		"status":           e.Status,
		"input_messages":   e.InputMessages,
		"created_at":       e.CreatedAt,
		"updated_at":       e.UpdatedAt,
	}

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert agent run: %w", err)
	}

	entity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentRun])
	if err != nil {
		return fmt.Errorf("cannot insert agent run: %w", err)
	}

	*e = entity

	return nil
}

func (e *AgentRun) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE agent_runs
SET
	status = @status,
	result = @result,
	error_message = @error_message,
	stop_requested = @stop_requested,
	started_at = @started_at,
	lease_owner = @lease_owner,
	lease_expires_at = @lease_expires_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
RETURNING
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	input_messages,
	result,
	error_message,
	stop_requested,
	started_at,
	lease_owner,
	lease_expires_at,
	created_at,
	updated_at;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":               e.ID.String(),
		"status":           e.Status,
		"result":           e.Result,
		"error_message":    e.ErrorMessage,
		"stop_requested":   e.StopRequested,
		"started_at":       e.StartedAt,
		"lease_owner":      e.LeaseOwner,
		"lease_expires_at": e.LeaseExpiresAt,
		"updated_at":       e.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update agent run: %w", err)
	}

	entity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentRun])
	if err != nil {
		return fmt.Errorf("cannot update agent run: %w", err)
	}

	*e = entity

	return nil
}

// ClearCheckpoint is the explicit path for removing persisted checkpoint
// data. AgentRun.Update intentionally does not write checkpoint so status
// commits cannot erase a checkpoint saved by PGCheckpointer.Save.
func (e *AgentRun) ClearCheckpoint(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE agent_runs
SET
	checkpoint = NULL,
	updated_at = now()
WHERE
	%s
	AND id = @id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": e.ID.String()}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot clear agent run checkpoint: %w", err)
	}

	e.Checkpoint = nil

	return nil
}

// SaveCheckpoint writes checkpoint inside the caller's transaction. Use
// this from service methods that must update checkpoint and status
// atomically, such as Approve.
func (e *AgentRun) SaveCheckpoint(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cp *agent.Checkpoint,
) error {
	data, err := marshalAgentCheckpoint(cp)
	if err != nil {
		return err
	}

	q := `
UPDATE agent_runs
SET
	checkpoint = @checkpoint,
	updated_at = now()
WHERE
	%s
	AND id = @id
RETURNING
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	input_messages,
	result,
	error_message,
	stop_requested,
	started_at,
	lease_owner,
	lease_expires_at,
	created_at,
	updated_at;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         e.ID.String(),
		"checkpoint": json.RawMessage(data),
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot save agent run checkpoint: %w", err)
	}

	entity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentRun])
	if err != nil {
		return fmt.Errorf("cannot save agent run checkpoint: %w", err)
	}

	*e = entity

	return nil
}

func (e *AgentRun) LoadNextPendingForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
	id,
	organization_id,
	start_agent_name,
	status,
	checkpoint,
	input_messages,
	result,
	error_message,
	stop_requested,
	started_at,
	lease_owner,
	lease_expires_at,
	created_at,
	updated_at
FROM
	agent_runs
WHERE
	status = 'PENDING'
	AND stop_requested = false
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
`

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query pending agent run: %w", err)
	}

	entity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgentRun])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot load pending agent run: %w", err)
	}

	*e = entity

	return nil
}

// ResetStaleAgentRuns resets agent runs whose worker lease has expired.
// The worker refreshes lease_expires_at from a separate heartbeat goroutine,
// so a long LLM or tool call is not considered stale while the process is alive.
// Stale recovery returns rows to PENDING so the supervisor auto-resumes
// from checkpoint when one exists. User-requested suspension remains
// SUSPENDED and is resumed only by the Resume service method.
func ResetStaleAgentRuns(ctx context.Context, conn pg.Querier) error {
	q := `
UPDATE agent_runs
SET
	status = 'PENDING',
	started_at = NULL,
	lease_owner = NULL,
	lease_expires_at = NULL,
	stop_requested = false,
	updated_at = now()
WHERE
	status = 'RUNNING'
	AND lease_expires_at IS NOT NULL
	AND lease_expires_at < now();
`

	_, err := conn.Exec(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot reset stale agent runs: %w", err)
	}

	return nil
}

// HeartbeatAgentRunLease refreshes the lease for a running agent run.
// Returns the number of rows affected (0 if lease was lost).
func HeartbeatAgentRunLease(
	ctx context.Context,
	conn pg.Querier,
	runID string,
	leaseOwner string,
	expiresAt time.Time,
) (int64, error) {
	q := `
UPDATE agent_runs
SET
	lease_expires_at = @lease_expires_at,
	updated_at = now()
WHERE
	id = @id
	AND status = 'RUNNING'
	AND lease_owner = @lease_owner;
`

	args := pgx.StrictNamedArgs{
		"id":               runID,
		"lease_owner":      leaseOwner,
		"lease_expires_at": expiresAt,
	}

	tag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot heartbeat agent run lease: %w", err)
	}

	return tag.RowsAffected(), nil
}

// LoadRunningStopRequestedIDs returns IDs of running agent runs with
// stop_requested = true.
func LoadRunningStopRequestedIDs(ctx context.Context, conn pg.Querier) ([]string, error) {
	q := `SELECT id FROM agent_runs WHERE status = 'RUNNING' AND stop_requested = true;`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("cannot query stop-requested agent runs: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("cannot scan stop-requested agent run ID: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate stop-requested agent runs: %w", err)
	}

	return ids, nil
}

// PGCheckpointer implements agent.Checkpointer backed by the
// agent_runs table checkpoint column. It is supervisor-internal and
// intentionally uses raw run IDs with no tenant scope; public service/API
// methods must load AgentRun through scoped coredata methods before invoking
// lifecycle transitions.
type PGCheckpointer struct {
	pg *pg.Client
}

func NewPGCheckpointer(pgClient *pg.Client) *PGCheckpointer {
	return &PGCheckpointer{pg: pgClient}
}

func (s *PGCheckpointer) Save(ctx context.Context, runID string, cp *agent.Checkpoint) error {
	data, err := marshalAgentCheckpoint(cp)
	if err != nil {
		return err
	}

	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			q := `UPDATE agent_runs SET checkpoint = @checkpoint, updated_at = now() WHERE id = @id;`

			args := pgx.StrictNamedArgs{
				"id":         runID,
				"checkpoint": json.RawMessage(data),
			}

			tag, err := conn.Exec(ctx, q, args)
			if err != nil {
				return fmt.Errorf("cannot save checkpoint: %w", err)
			}

			if tag.RowsAffected() == 0 {
				return fmt.Errorf("cannot save checkpoint: agent run %s not found", runID)
			}

			return nil
		},
	)
}

func (s *PGCheckpointer) Load(ctx context.Context, runID string) (*agent.Checkpoint, error) {
	var cp *agent.Checkpoint

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			q := `SELECT checkpoint FROM agent_runs WHERE id = @id;`

			args := pgx.StrictNamedArgs{"id": runID}

			rows, err := conn.Query(ctx, q, args)
			if err != nil {
				return fmt.Errorf("cannot query checkpoint: %w", err)
			}

			type row struct {
				Checkpoint json.RawMessage `db:"checkpoint"`
			}

			r, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[row])
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return ErrResourceNotFound
				}
				return fmt.Errorf("cannot load checkpoint: %w", err)
			}

			if r.Checkpoint == nil {
				return nil
			}

			cp = new(agent.Checkpoint)
			if err := json.Unmarshal(r.Checkpoint, cp); err != nil {
				return fmt.Errorf("cannot unmarshal checkpoint: %w", err)
			}

			return nil
		},
	)

	return cp, err
}

func marshalAgentCheckpoint(cp *agent.Checkpoint) ([]byte, error) {
	if cp == nil {
		return nil, fmt.Errorf("cannot marshal checkpoint: checkpoint is required")
	}

	data, err := json.Marshal(cp)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal checkpoint: %w", err)
	}

	if len(data) > agent.MaxCheckpointBytes {
		return nil, fmt.Errorf("cannot marshal checkpoint: size %d exceeds limit %d", len(data), agent.MaxCheckpointBytes)
	}

	return data, nil
}
