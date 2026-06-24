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
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
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
		StartedAt      *time.Time      `db:"started_at"`
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

var (
	_ fmt.Stringer             = AgentRunStatus("")
	_ encoding.TextMarshaler   = AgentRunStatus("")
	_ encoding.TextUnmarshaler = (*AgentRunStatus)(nil)
)

func AgentRunStatuses() []AgentRunStatus {
	return []AgentRunStatus{
		AgentRunStatusPending,
		AgentRunStatusRunning,
		AgentRunStatusSuspended,
		AgentRunStatusAwaitingApproval,
		AgentRunStatusCompleted,
		AgentRunStatusFailed,
	}
}

func (v AgentRunStatus) IsValid() bool {
	switch v {
	case
		AgentRunStatusPending,
		AgentRunStatusRunning,
		AgentRunStatusSuspended,
		AgentRunStatusAwaitingApproval,
		AgentRunStatusCompleted,
		AgentRunStatusFailed:
		return true
	}

	return false
}

func (v AgentRunStatus) String() string {
	return string(v)
}

func (v AgentRunStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AgentRunStatus) UnmarshalText(text []byte) error {
	val := AgentRunStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AgentRunStatus value: %q", string(text))
	}

	*v = val

	return nil
}

func (e AgentRun) CursorKey(orderBy AgentRunOrderField) page.CursorKey {
	switch orderBy {
	case AgentRunOrderFieldCreatedAt:
		return page.NewCursorKey(e.ID, e.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (e *AgentRun) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM agent_runs WHERE id = ANY(@resource_ids::text[])`

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
	started_at,
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
	started_at,
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
	started_at,
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
	started_at,
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

// Update intentionally does not write the checkpoint column. Status
// commits and checkpoint persistence are split: PGCheckpointer.Save is
// the only writer of checkpoint and ClearCheckpoint is the only path
// to remove it. This prevents a status update from accidentally erasing
// an in-flight checkpoint saved between Load and Update.
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
	started_at = @started_at,
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
	started_at,
	created_at,
	updated_at;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":            e.ID.String(),
		"status":        e.Status,
		"result":        e.Result,
		"error_message": e.ErrorMessage,
		"started_at":    e.StartedAt,
		"updated_at":    e.UpdatedAt,
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

// CommitAgentRunResult writes the terminal or resting state of a run
// (COMPLETED, FAILED, PENDING on graceful suspend, or AWAITING_APPROVAL)
// guarded on the row still being RUNNING. The guard is a lightweight
// safety net: it discards a commit for a run a human moved out of
// RUNNING manually (the only way a run leaves RUNNING out from under an
// active worker now that lease-based recovery is gone). Returns the
// number of rows affected (0 when the guard rejected the write).
func CommitAgentRunResult(
	ctx context.Context,
	tx pg.Tx,
	e *AgentRun,
) (int64, error) {
	q := `
UPDATE agent_runs
SET
	status = @status,
	result = @result,
	error_message = @error_message,
	started_at = @started_at,
	updated_at = @updated_at
WHERE
	id = @id
	AND status = 'RUNNING';
`

	args := pgx.StrictNamedArgs{
		"id":            e.ID.String(),
		"status":        e.Status,
		"result":        e.Result,
		"error_message": e.ErrorMessage,
		"started_at":    e.StartedAt,
		"updated_at":    e.UpdatedAt,
	}

	tag, err := tx.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot commit agent run result: %w", err)
	}

	return tag.RowsAffected(), nil
}

// RequeueForApprovalResume persists the run's checkpoint and returns it
// to PENDING so a worker resumes from the approval boundary. The caller
// populates e.Checkpoint (with the approval decisions merged in),
// e.Status, e.StartedAt, and e.UpdatedAt before calling. The write is
// guarded on the row still being AWAITING_APPROVAL; ErrResourceNotFound
// is returned when the guard rejects it. This is the one path that writes
// the checkpoint alongside a status change — Update deliberately omits the
// checkpoint column.
func (e *AgentRun) RequeueForApprovalResume(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE agent_runs
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
		"expected_status": AgentRunStatusAwaitingApproval,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot requeue agent run for approval resume: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

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
	started_at,
	created_at,
	updated_at
FROM
	agent_runs
WHERE
	status = 'PENDING'
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

// PGCheckpointer implements agent.Checkpointer backed by the
// agent_runs table checkpoint column. The runID is validated as a GID
// up front so malformed identifiers fail closed; rows are then scoped
// by primary key.
type PGCheckpointer struct {
	pg                 *pg.Client
	maxCheckpointBytes int
}

type PGCheckpointerOption func(*PGCheckpointer)

// WithMaxCheckpointBytes overrides the default per-checkpoint size cap
// enforced on both Save and Load.
func WithMaxCheckpointBytes(n int) PGCheckpointerOption {
	return func(s *PGCheckpointer) {
		if n > 0 {
			s.maxCheckpointBytes = n
		}
	}
}

func NewPGCheckpointer(pgClient *pg.Client, opts ...PGCheckpointerOption) *PGCheckpointer {
	s := &PGCheckpointer{
		pg:                 pgClient,
		maxCheckpointBytes: 10 * 1024 * 1024,
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *PGCheckpointer) Save(ctx context.Context, runID string, cp *agent.Checkpoint) error {
	if _, err := gid.ParseGID(runID); err != nil {
		return fmt.Errorf("cannot parse agent run id: %w", err)
	}

	data, err := s.marshalAgentCheckpoint(cp)
	if err != nil {
		return err
	}

	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			q := `
UPDATE agent_runs
SET
	checkpoint = @checkpoint,
	updated_at = now()
WHERE
	id = @id;
`

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
	if _, err := gid.ParseGID(runID); err != nil {
		return nil, fmt.Errorf("cannot parse agent run id: %w", err)
	}

	var cp *agent.Checkpoint

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			q := `
SELECT checkpoint
FROM agent_runs
WHERE
	id = @id;
`

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

			if len(r.Checkpoint) > s.maxCheckpointBytes {
				return fmt.Errorf("cannot load checkpoint: size %d exceeds limit %d", len(r.Checkpoint), s.maxCheckpointBytes)
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

func (s *PGCheckpointer) marshalAgentCheckpoint(cp *agent.Checkpoint) ([]byte, error) {
	if cp == nil {
		return nil, fmt.Errorf("cannot marshal checkpoint: checkpoint is required")
	}

	data, err := json.Marshal(cp)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal checkpoint: %w", err)
	}

	if len(data) > s.maxCheckpointBytes {
		return nil, fmt.Errorf("cannot marshal checkpoint: size %d exceeds limit %d", len(data), s.maxCheckpointBytes)
	}

	return data, nil
}
