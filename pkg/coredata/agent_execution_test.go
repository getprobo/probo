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

func TestAgentExecution_ConversationalQueueLifecycle(t *testing.T) {
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
		ID:                gid.New(tenantID, coredata.AgentRunEntityType),
		OrganizationID:    organizationID,
		StartAgentName:    "assistant",
		Source:            &source,
		SessionKey:        &sessionKey,
		SourceCoordinates: []byte(`{"workspace":"opaque"}`),
		TrustedContext:    []byte(`{"organization":"trusted"}`),
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

				inserted, err = execution.UpsertConversationalBySourceSession(ctx, tx, scope)

				return err
			},
		),
	)
	assert.True(t, inserted)
	assert.Equal(t, coredata.AgentExecutionKindConversational, execution.ExecutionKind)
	assert.Equal(t, coredata.AgentExecutionDefaultMaxAttempts, execution.MaxAttempts)

	duplicate := coredata.AgentExecution{
		ID:                gid.New(tenantID, coredata.AgentRunEntityType),
		OrganizationID:    organizationID,
		StartAgentName:    "updated-assistant",
		Source:            &source,
		SessionKey:        &sessionKey,
		SourceCoordinates: []byte(`{"workspace":"updated"}`),
		TrustedContext:    []byte(`{"organization":"updated"}`),
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

				inserted, err = duplicate.UpsertConversationalBySourceSession(ctx, tx, scope)

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
		ID:             gid.New(tenantID, coredata.AgentInputEntityType),
		OrganizationID: organizationID,
		AgentRunID:     execution.ID,
		Source:         source,
		SourceEventID:  &eventID,
		Message:        []byte(`{"role":"user","text":"hello"}`),
		CreatedAt:      now,
		UpdatedAt:      now,
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
		ID:             gid.New(tenantID, coredata.AgentInputEntityType),
		OrganizationID: organizationID,
		AgentRunID:     execution.ID,
		Source:         source,
		SourceEventID:  &eventID,
		Message:        []byte(`{"role":"user","text":"duplicate"}`),
		CreatedAt:      now.Add(time.Second),
		UpdatedAt:      now.Add(time.Second),
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
	assert.Equal(t, coredata.AgentRunStatusRunning, claimed.Status)
	assert.Equal(t, 1, claimed.AttemptCount)
	require.NotNil(t, claimed.ProcessingOwnerToken)
	assert.Equal(t, ownerToken, *claimed.ProcessingOwnerToken)

	var pending coredata.AgentInputs

	require.NoError(
		t,
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return pending.LoadPendingByAgentRunID(
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
	assert.Equal(t, coredata.AgentRunStatusPending, persistedExecution.Status)
	assert.Nil(t, persistedExecution.ProcessingOwnerToken)
	assert.Equal(t, 0, persistedExecution.AttemptCount)

	secondEventID := "event-2"
	secondInput := coredata.AgentInput{
		ID:             gid.New(tenantID, coredata.AgentInputEntityType),
		OrganizationID: organizationID,
		AgentRunID:     execution.ID,
		Source:         source,
		SourceEventID:  &secondEventID,
		Message:        []byte(`{"role":"user","text":"again"}`),
		CreatedAt:      now.Add(4 * time.Second),
		UpdatedAt:      now.Add(4 * time.Second),
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
		client.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				return coredata.ResetStaleAgentExecutionLeases(
					ctx,
					conn,
					now.Add(11*time.Second),
					5*time.Second,
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
	assert.Equal(t, coredata.AgentRunStatusPending, persistedExecution.Status)
	assert.Nil(t, persistedExecution.ProcessingOwnerToken)
	require.NotNil(t, persistedExecution.LastError)
	assert.Equal(t, "agent execution processing lease expired", *persistedExecution.LastError)
}

func TestAgentInput_NullSourceEventIDsRemainDistinct(t *testing.T) {
	ctx := t.Context()
	client := test.PGClient(t)
	run := insertPendingRun(t, client, "test-agent", nil)
	scope := coredata.NewScope(run.ID.TenantID())
	now := time.Now().UTC()

	for range 2 {
		input := coredata.AgentInput{
			ID:             gid.New(run.ID.TenantID(), coredata.AgentInputEntityType),
			OrganizationID: run.OrganizationID,
			AgentRunID:     run.ID,
			Source:         "manual",
			Message:        []byte(`{"role":"user","text":"hello"}`),
			CreatedAt:      now,
			UpdatedAt:      now,
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
