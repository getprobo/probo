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

package coredata_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestAgentExecution_QueueLifecycle(t *testing.T) {
	ctx := t.Context()
	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5)`,
					organizationID,
					tenantID,
					"agent-execution-"+organizationID.String(),
					now,
					now,
				)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = client.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(ctx, "DELETE FROM organizations WHERE id = $1", organizationID)

					return err
				},
			)
		},
	)

	source := "provider"
	sessionKey := "opaque-session"
	execution := coredata.AgentExecution{
		ID:                gid.New(tenantID, coredata.AgentExecutionEntityType),
		OrganizationID:    organizationID,
		StartAgentName:    "assistant",
		Source:            &source,
		SessionKey:        &sessionKey,
		SourceCoordinates: []byte(`{"workspace":"opaque"}`),
		SessionMessages:   []byte(`[{"role":"user","text":"first"}]`),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	var inserted bool

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				inserted, err = execution.UpsertBySourceSession(ctx, tx, scope)

				return err
			},
		),
	)
	assert.True(t, inserted)
	assert.Equal(t, coredata.AgentExecutionDefaultMaxAttempts, execution.MaxAttempts)

	duplicate := coredata.AgentExecution{
		ID:                gid.New(tenantID, coredata.AgentExecutionEntityType),
		OrganizationID:    organizationID,
		StartAgentName:    "updated-assistant",
		Source:            &source,
		SessionKey:        &sessionKey,
		SourceCoordinates: []byte(`{"workspace":"updated"}`),
		SessionMessages:   []byte(`[{"role":"user","text":"replacement"}]`),
		CreatedAt:         now.Add(time.Second),
		UpdatedAt:         now.Add(time.Second),
	}

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				inserted, err = duplicate.UpsertBySourceSession(ctx, tx, scope)

				return err
			},
		),
	)
	assert.False(t, inserted)
	assert.Equal(t, execution.ID, duplicate.ID)
	assert.JSONEq(t, `[{"role":"user","text":"first"}]`, string(duplicate.SessionMessages))
	assert.JSONEq(t, `{"workspace":"updated"}`, string(duplicate.SourceCoordinates))

	eventID := "event-1"
	input := coredata.AgentInput{
		ID:               gid.New(tenantID, coredata.AgentInputEntityType),
		OrganizationID:   organizationID,
		AgentExecutionID: execution.ID,
		Source:           source,
		SourceEventID:    &eventID,
		Message:          []byte(`{"role":"user","text":"hello"}`),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				inserted, err = input.EnqueueIdempotently(ctx, tx, scope)

				return err
			},
		),
	)
	assert.True(t, inserted)

	duplicateInput := coredata.AgentInput{
		ID:               gid.New(tenantID, coredata.AgentInputEntityType),
		OrganizationID:   organizationID,
		AgentExecutionID: execution.ID,
		Source:           source,
		SourceEventID:    &eventID,
		Message:          []byte(`{"role":"user","text":"duplicate"}`),
		CreatedAt:        now.Add(time.Second),
		UpdatedAt:        now.Add(time.Second),
	}

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				inserted, err = duplicateInput.EnqueueIdempotently(ctx, tx, scope)

				return err
			},
		),
	)
	assert.False(t, inserted)
	assert.Equal(t, input.ID, duplicateInput.ID)
	assert.JSONEq(t, `{"role":"user","text":"hello"}`, string(duplicateInput.Message))

	ownerToken := "owner-one"

	var claimed coredata.AgentExecution

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return claimed.ClaimNextForUpdateSkipLocked(ctx, tx, now.Add(2*time.Second), ownerToken)
			},
		),
	)
	assert.Equal(t, execution.ID, claimed.ID)
	assert.Equal(t, coredata.AgentExecutionStatusRunning, claimed.Status)
	assert.Equal(t, 1, claimed.AttemptCount)
	require.NotNil(t, claimed.ProcessingOwnerToken)
	assert.Equal(t, ownerToken, *claimed.ProcessingOwnerToken)

	var pending coredata.AgentInputs

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return pending.LoadPendingByAgentExecutionID(
					ctx,
					conn,
					scope,
					claimed.ID,
					ownerToken,
					now.Add(2*time.Second),
					1,
				)
			},
		),
	)
	require.Len(t, pending, 1)
	assert.Equal(t, input.ID, pending[0].ID)

	err := client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return pending[0].MarkProcessed(ctx, conn, scope, "wrong-owner", now.Add(3*time.Second))
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := pending[0].MarkProcessed(ctx, tx, scope, ownerToken, now.Add(3*time.Second)); err != nil {
					return fmt.Errorf("cannot process input: %w", err)
				}

				return claimed.Release(ctx, tx, scope, ownerToken, now.Add(3*time.Second))
			},
		),
	)

	var persistedInput coredata.AgentInput

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return persistedInput.LoadByID(ctx, conn, scope, input.ID)
			},
		),
	)
	assert.NotNil(t, persistedInput.ProcessedAt)

	var persistedExecution coredata.AgentExecution

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return persistedExecution.LoadByID(ctx, conn, scope, execution.ID)
			},
		),
	)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, persistedExecution.Status)
	assert.Nil(t, persistedExecution.ProcessingOwnerToken)
	assert.Equal(t, 0, persistedExecution.AttemptCount)

	secondEventID := "event-2"
	secondInput := coredata.AgentInput{
		ID:               gid.New(tenantID, coredata.AgentInputEntityType),
		OrganizationID:   organizationID,
		AgentExecutionID: execution.ID,
		Source:           source,
		SourceEventID:    &secondEventID,
		Message:          []byte(`{"role":"user","text":"again"}`),
		CreatedAt:        now.Add(4 * time.Second),
		UpdatedAt:        now.Add(4 * time.Second),
	}

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if _, err := secondInput.EnqueueIdempotently(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot enqueue second input: %w", err)
				}

				return claimed.ClaimNextForUpdateSkipLocked(
					ctx,
					tx,
					now.Add(4*time.Second),
					"owner-two",
				)
			},
		),
	)
	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return claimed.Heartbeat(ctx, conn, scope, "owner-two", now.Add(5*time.Second))
			},
		),
	)
	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				ids, err := coredata.LockStaleAgentExecutionIDs(
					ctx,
					tx,
					now.Add(11*time.Second),
					5*time.Second,
				)
				if err != nil {
					return err
				}

				return coredata.ResetStaleAgentExecutionLeases(
					ctx,
					tx,
					ids,
					now.Add(11*time.Second),
				)
			},
		),
	)

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return persistedExecution.LoadByID(ctx, conn, scope, execution.ID)
			},
		),
	)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, persistedExecution.Status)
	assert.Nil(t, persistedExecution.ProcessingOwnerToken)
	require.NotNil(t, persistedExecution.LastError)
	assert.Equal(t, "agent execution processing lease expired", *persistedExecution.LastError)
}

func TestAgentInput_NullSourceEventIDsRemainDistinct(t *testing.T) {
	ctx := t.Context()
	client := test.PGClient(t)
	run := insertIdleExecution(t, client, "test-agent")
	scope := coredata.NewScope(run.ID.TenantID())
	now := time.Now().UTC()

	for range 2 {
		input := coredata.AgentInput{
			ID:               gid.New(run.ID.TenantID(), coredata.AgentInputEntityType),
			OrganizationID:   run.OrganizationID,
			AgentExecutionID: run.ID,
			Source:           "manual",
			Message:          []byte(`{"role":"user","text":"hello"}`),
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		var inserted bool

		require.NoError(
			t,
			client.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					var err error

					inserted, err = input.EnqueueIdempotently(ctx, tx, scope)

					return err
				},
			),
		)
		assert.True(t, inserted)
	}
}

func TestDeadLetterAgentInputsForStaleExecutions_DeadLettersAllPendingOnTerminal(t *testing.T) {
	ctx := t.Context()
	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()
	ownerToken := "stale-owner"

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5)`,
					organizationID,
					tenantID,
					"agent-execution-stale-"+organizationID.String(),
					now,
					now,
				)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = client.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(ctx, "DELETE FROM organizations WHERE id = $1", organizationID)

					return err
				},
			)
		},
	)

	source := "provider"
	sessionKey := "stale-all-pending"
	execution := coredata.AgentExecution{
		ID:              gid.New(tenantID, coredata.AgentExecutionEntityType),
		OrganizationID:  organizationID,
		StartAgentName:  "assistant",
		Source:          &source,
		SessionKey:      &sessionKey,
		SessionMessages: []byte("[]"),
		MaxAttempts:     1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				_, err := execution.UpsertBySourceSession(ctx, tx, scope)

				return err
			},
		),
	)

	batchInput := enqueueAgentInputForExecution(t, client, scope, execution, "batch-event", now)
	extraInput := enqueueAgentInputForExecution(t, client, scope, execution, "extra-event", now.Add(time.Millisecond))

	var claimed coredata.AgentExecution

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return claimed.ClaimNextForUpdateSkipLocked(ctx, tx, now.Add(time.Second), ownerToken)
			},
		),
	)
	require.Equal(t, execution.ID, claimed.ID)

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				if err := claimed.SetProcessingInputIDs(
					ctx,
					conn,
					scope,
					ownerToken,
					[]string{batchInput.ID.String()},
					now.Add(time.Second),
				); err != nil {
					return err
				}

				_, err := conn.Exec(
					ctx,
					`UPDATE agent_executions SET processing_heartbeat_at = $2 WHERE id = $1`,
					claimed.ID,
					now.Add(-time.Minute),
				)

				return err
			},
		),
	)

	recoverAt := now.Add(2 * time.Second)

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				ids, err := coredata.LockStaleAgentExecutionIDs(ctx, tx, recoverAt, 30*time.Second)
				if err != nil {
					return err
				}

				if err := coredata.DeadLetterAgentInputsForStaleExecutions(
					ctx,
					tx,
					ids,
					recoverAt,
					coredata.AgentExecutionStaleLeaseError,
				); err != nil {
					return err
				}

				return coredata.ResetStaleAgentExecutionLeases(ctx, tx, ids, recoverAt)
			},
		),
	)

	var loadedBatch, loadedExtra coredata.AgentInput

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				if err := loadedBatch.LoadByID(ctx, conn, scope, batchInput.ID); err != nil {
					return err
				}

				return loadedExtra.LoadByID(ctx, conn, scope, extraInput.ID)
			},
		),
	)
	require.NotNil(t, loadedBatch.DeadLetteredAt)
	require.NotNil(t, loadedExtra.DeadLetteredAt)
	assert.Nil(t, loadedBatch.ProcessedAt)
	assert.Nil(t, loadedExtra.ProcessedAt)

	var loadedExecution coredata.AgentExecution

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return loadedExecution.LoadByID(ctx, conn, scope, execution.ID)
			},
		),
	)
	assert.Equal(t, coredata.AgentExecutionStatusFailed, loadedExecution.Status)
	assert.Nil(t, loadedExecution.ProcessingOwnerToken)
	assert.NotNil(t, loadedExecution.DeadLetteredAt)
}

func TestLockStaleAgentExecutionIDs_BlocksHeartbeatUntilReset(t *testing.T) {
	ctx := t.Context()
	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()
	ownerToken := "heartbeat-owner"

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5)`,
					organizationID,
					tenantID,
					"agent-execution-lock-"+organizationID.String(),
					now,
					now,
				)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = client.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(ctx, "DELETE FROM organizations WHERE id = $1", organizationID)

					return err
				},
			)
		},
	)

	source := "provider"
	sessionKey := "stale-lock"
	execution := coredata.AgentExecution{
		ID:              gid.New(tenantID, coredata.AgentExecutionEntityType),
		OrganizationID:  organizationID,
		StartAgentName:  "assistant",
		Source:          &source,
		SessionKey:      &sessionKey,
		SessionMessages: []byte("[]"),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				_, err := execution.UpsertBySourceSession(ctx, tx, scope)

				return err
			},
		),
	)

	input := enqueueAgentInputForExecution(t, client, scope, execution, "lock-event", now)

	var claimed coredata.AgentExecution

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return claimed.ClaimNextForUpdateSkipLocked(ctx, tx, now.Add(time.Second), ownerToken)
			},
		),
	)
	require.Equal(t, execution.ID, claimed.ID)

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				if err := claimed.SetProcessingInputIDs(
					ctx,
					conn,
					scope,
					ownerToken,
					[]string{input.ID.String()},
					now.Add(time.Second),
				); err != nil {
					return err
				}

				_, err := conn.Exec(
					ctx,
					`UPDATE agent_executions SET processing_heartbeat_at = $2 WHERE id = $1`,
					claimed.ID,
					now.Add(-time.Minute),
				)

				return err
			},
		),
	)

	recoverAt := now.Add(2 * time.Second)
	heartbeatDone := make(chan error, 1)

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				ids, err := coredata.LockStaleAgentExecutionIDs(ctx, tx, recoverAt, 30*time.Second)
				if err != nil {
					return err
				}

				go func() {
					heartbeatDone <- client.WithConn(
						context.Background(),
						func(ctx context.Context, conn pg.Querier) error {
							return claimed.Heartbeat(ctx, conn, scope, ownerToken, recoverAt)
						},
					)
				}()

				time.Sleep(150 * time.Millisecond)

				select {
				case err := <-heartbeatDone:
					return fmt.Errorf("heartbeat completed while stale recovery held the lock: %w", err)
				default:
				}

				if err := coredata.DeadLetterAgentInputsForStaleExecutions(
					ctx,
					tx,
					ids,
					recoverAt,
					coredata.AgentExecutionStaleLeaseError,
				); err != nil {
					return err
				}

				return coredata.ResetStaleAgentExecutionLeases(ctx, tx, ids, recoverAt)
			},
		),
	)

	select {
	case err := <-heartbeatDone:
		require.ErrorIs(t, err, coredata.ErrResourceNotFound)
	case <-time.After(2 * time.Second):
		t.Fatal("heartbeat did not unblock after stale recovery committed")
	}

	var loadedExecution coredata.AgentExecution

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return loadedExecution.LoadByID(ctx, conn, scope, execution.ID)
			},
		),
	)
	assert.Nil(t, loadedExecution.ProcessingOwnerToken)
}

func enqueueAgentInputForExecution(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	execution coredata.AgentExecution,
	eventID string,
	now time.Time,
) coredata.AgentInput {
	t.Helper()

	input := coredata.AgentInput{
		ID:               gid.New(execution.ID.TenantID(), coredata.AgentInputEntityType),
		OrganizationID:   execution.OrganizationID,
		AgentExecutionID: execution.ID,
		Source:           *execution.Source,
		SourceEventID:    &eventID,
		Message:          []byte(`{"role":"user","text":"hello"}`),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				_, err := input.EnqueueIdempotently(ctx, tx, scope)

				return err
			},
		),
	)

	return input
}
