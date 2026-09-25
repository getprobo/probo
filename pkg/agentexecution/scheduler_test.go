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

package agentexecution_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agentexecution"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

type recordingIdentityPreparer struct {
	mu         sync.Mutex
	identities []gid.GID
}

func (p *recordingIdentityPreparer) Prepare(
	ctx context.Context,
	_ *coredata.AgentExecution,
	registry agent.AgentRegistry,
	input *coredata.AgentInput,
) (context.Context, agent.AgentRegistry, error) {
	if input != nil && input.IdentityID != nil {
		p.mu.Lock()
		p.identities = append(p.identities, *input.IdentityID)
		p.mu.Unlock()
	}

	return ctx, registry, nil
}

func TestScheduler_PreparesEachUserInputWithItsIdentity(t *testing.T) {
	client := test.PGClient(t)
	organizationID := insertTestOrganization(t, client)
	execution := insertExecution(
		t,
		client,
		organizationID,
		"conversation-agent",
		"provider",
		"per-input-identity",
		3,
	)

	aliceID := gid.New(execution.ID.TenantID(), coredata.IdentityEntityType)
	bobID := gid.New(execution.ID.TenantID(), coredata.IdentityEntityType)
	aliceInput := enqueueAgentInputWithIdentity(
		t,
		client,
		execution,
		"alice-ping",
		userMessage("grant the report"),
		aliceID,
	)
	bobInput := enqueueAgentInputWithIdentity(
		t,
		client,
		execution,
		"bob-ping",
		userMessage("do not grant"),
		bobID,
	)

	preparer := &recordingIdentityPreparer{}
	ag := newDummyAgent(
		"conversation-agent",
		[]*llm.ChatCompletionResponse{
			stopResponse("granted"),
			stopResponse("declined"),
		},
	)
	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"conversation-agent": ag}},
		agentexecution.WithWorkerInterval(20*time.Millisecond),
		agentexecution.WithExecutionPreparer(preparer),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			return loadAgentInput(t, client, aliceInput.ID).ProcessedAt != nil &&
				loadAgentInput(t, client, bobInput.ID).ProcessedAt != nil
		},
		8*time.Second,
		50*time.Millisecond,
	)

	preparer.mu.Lock()
	identities := append([]gid.GID(nil), preparer.identities...)
	preparer.mu.Unlock()
	require.Equal(t, []gid.GID{aliceID, bobID}, identities)
}

func TestScheduler_ConversationalTurnsProcessOneInputAtATime(t *testing.T) {
	client := test.PGClient(t)
	organizationID := insertTestOrganization(t, client)
	execution := insertExecution(
		t,
		client,
		organizationID,
		"conversation-agent",
		"provider",
		"session-batch",
		3,
	)
	inputs := []coredata.AgentInput{
		enqueueAgentInput(t, client, execution, "event-1", userMessage("one")),
		enqueueAgentInput(t, client, execution, "event-2", userMessage("two")),
		enqueueAgentInput(t, client, execution, "event-3", userMessage("three")),
	}

	ag := newDummyAgent(
		"conversation-agent",
		[]*llm.ChatCompletionResponse{
			stopResponse("first turn complete"),
			stopResponse("second turn complete"),
			stopResponse("third turn complete"),
		},
	)
	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"conversation-agent": ag}},
		agentexecution.WithWorkerInterval(20*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			for _, input := range inputs {
				if loadAgentInput(t, client, input.ID).ProcessedAt == nil {
					return false
				}
			}

			return true
		},
		8*time.Second,
		50*time.Millisecond,
	)

	persisted := loadAgentExecution(t, client, execution.ID)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, persisted.Status)
	assert.Nil(t, persisted.Checkpoint)
	assert.Empty(t, persisted.ProcessingInputIDs)

	var sessionMessages []llm.Message
	require.NoError(t, json.Unmarshal(persisted.SessionMessages, &sessionMessages))
	require.Len(t, sessionMessages, 6)
	assert.Equal(t, "one", sessionMessages[0].Text())
	assert.Equal(t, "first turn complete", sessionMessages[1].Text())
	assert.Equal(t, "two", sessionMessages[2].Text())
	assert.Equal(t, "second turn complete", sessionMessages[3].Text())
	assert.Equal(t, "three", sessionMessages[4].Text())
	assert.Equal(t, "third turn complete", sessionMessages[5].Text())
}

func TestScheduler_HeartbeatPreventsStaleRecovery(t *testing.T) {
	client := test.PGClient(t)
	organizationID := insertTestOrganization(t, client)
	execution := insertExecution(
		t,
		client,
		organizationID,
		"heartbeat-agent",
		"provider",
		"session-heartbeat",
		3,
	)
	input := enqueueAgentInput(t, client, execution, "heartbeat-event", userMessage("work"))

	toolStarted := make(chan struct{})
	toolRelease := make(chan struct{})
	slowTool := agent.FunctionTool[struct{}](
		"slow_work",
		"Waits for test release",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			close(toolStarted)
			<-toolRelease

			return agent.ToolResult{Content: "done"}, nil
		},
	)
	ag := newDummyAgent(
		"heartbeat-agent",
		[]*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "heartbeat-call",
					Function: llm.FunctionCall{Name: "slow_work", Arguments: `{}`},
				},
			),
			stopResponse("done"),
		},
		slowTool,
	)
	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"heartbeat-agent": ag}},
		agentexecution.WithWorkerInterval(20*time.Millisecond),
		agentexecution.WithWorkerHeartbeatInterval(20*time.Millisecond),
		agentexecution.WithWorkerStaleAfter(100*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	select {
	case <-toolStarted:
	case <-ctx.Done():
		t.Fatal("timed out waiting for heartbeat test tool")
	}

	time.Sleep(150 * time.Millisecond)
	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				ids, err := coredata.LockStaleAgentExecutionIDs(
					ctx,
					tx,
					time.Now(),
					100*time.Millisecond,
				)
				if err != nil {
					return err
				}

				return coredata.ResetStaleAgentExecutionLeases(ctx, tx, ids, time.Now())
			},
		),
	)

	running := loadAgentExecution(t, client, execution.ID)
	assert.Equal(t, coredata.AgentExecutionStatusRunning, running.Status)
	assert.NotNil(t, running.ProcessingOwnerToken)

	close(toolRelease)
	require.Eventually(
		t,
		func() bool {
			return loadAgentInput(t, client, input.ID).ProcessedAt != nil
		},
		8*time.Second,
		50*time.Millisecond,
	)
}

func TestScheduler_ConversationalFailuresRetryThenDeadLetter(t *testing.T) {
	client := test.PGClient(t)
	organizationID := insertTestOrganization(t, client)
	execution := insertExecution(
		t,
		client,
		organizationID,
		"missing-agent",
		"provider",
		"session-dead-letter",
		2,
	)
	input := enqueueAgentInput(t, client, execution, "failure-event", userMessage("fail"))

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{}},
		agentexecution.WithWorkerInterval(10*time.Millisecond),
		agentexecution.WithWorkerRetryBackoff(20*time.Millisecond, 20*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			return loadAgentExecution(t, client, execution.ID).DeadLetteredAt != nil
		},
		8*time.Second,
		50*time.Millisecond,
	)

	failedExecution := loadAgentExecution(t, client, execution.ID)
	failedInput := loadAgentInput(t, client, input.ID)
	assert.Equal(t, coredata.AgentExecutionStatusFailed, failedExecution.Status)
	assert.Equal(t, 2, failedExecution.AttemptCount)
	assert.NotNil(t, failedExecution.LastError)
	assert.NotNil(t, failedInput.DeadLetteredAt)
	assert.Equal(t, 2, failedInput.AttemptCount)
	assert.Nil(t, failedInput.ProcessedAt)
}

func TestScheduler_StaleProcessingInputIDsFallThroughToPendingInput(t *testing.T) {
	client := test.PGClient(t)
	organizationID := insertTestOrganization(t, client)
	execution := insertExecution(
		t,
		client,
		organizationID,
		"conversation-agent",
		"provider",
		"stale-processing-ids",
		3,
	)
	staleInput := enqueueAgentInput(t, client, execution, "stale-event", userMessage("already done"))
	pendingInput := enqueueAgentInput(t, client, execution, "pending-event", userMessage("still pending"))

	require.NoError(
		t,
		client.WithConn(
			context.Background(),
			func(ctx context.Context, conn pg.Querier) error {
				now := time.Now()
				if _, err := conn.Exec(
					ctx,
					`UPDATE agent_inputs
					 SET processed_at = $2, updated_at = $2
					 WHERE id = $1`,
					staleInput.ID.String(),
					now,
				); err != nil {
					return err
				}

				_, err := conn.Exec(
					ctx,
					`UPDATE agent_executions
					 SET processing_input_ids = ARRAY[$2]::text[],
					     status = 'IDLE',
					     processing_owner_token = NULL,
					     processing_heartbeat_at = NULL,
					     updated_at = $3
					 WHERE id = $1`,
					execution.ID.String(),
					staleInput.ID.String(),
					now,
				)

				return err
			},
		),
	)

	ag := newDummyAgent(
		"conversation-agent",
		[]*llm.ChatCompletionResponse{
			stopResponse("handled pending turn"),
		},
	)
	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{"conversation-agent": ag}},
		agentexecution.WithWorkerInterval(20*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go func() { _ = runWorker.Run(ctx) }()

	require.Eventually(
		t,
		func() bool {
			return loadAgentInput(t, client, pendingInput.ID).ProcessedAt != nil
		},
		8*time.Second,
		50*time.Millisecond,
	)

	assert.NotNil(t, loadAgentInput(t, client, staleInput.ID).ProcessedAt)
	assert.NotEqual(t, coredata.AgentExecutionStatusFailed, loadAgentExecution(t, client, execution.ID).Status)
}

func TestScheduler_ConversationalShutdownRestoresCheckpoint(t *testing.T) {
	client := test.PGClient(t)
	organizationID := insertTestOrganization(t, client)
	execution := insertExecution(
		t,
		client,
		organizationID,
		"restore-agent",
		"provider",
		"session-restore",
		3,
	)
	input := enqueueAgentInput(t, client, execution, "restore-event", userMessage("start"))

	toolStarted := make(chan struct{})
	toolRelease := make(chan struct{})
	slowTool := agent.FunctionTool[struct{}](
		"checkpoint_work",
		"Waits for shutdown",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			close(toolStarted)
			<-toolRelease

			return agent.ToolResult{Content: "checkpointed"}, nil
		},
	)
	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "restore-call",
					Function: llm.FunctionCall{Name: "checkpoint_work", Arguments: `{}`},
				},
			),
			stopResponse("restored"),
		},
	}
	ag := agent.New(
		"restore-agent",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(slowTool),
	)
	registry := &simpleRegistry{agents: map[string]*agent.Agent{"restore-agent": ag}}
	firstWorker := newTestWorker(
		client,
		registry,
		agentexecution.WithWorkerInterval(20*time.Millisecond),
		agentexecution.WithWorkerHeartbeatInterval(20*time.Millisecond),
	)

	firstCtx, stopFirst := context.WithCancel(context.Background())

	firstDone := make(chan error, 1)
	go func() { firstDone <- firstWorker.Run(firstCtx) }()

	select {
	case <-toolStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for checkpoint tool")
	}

	stopFirst()
	close(toolRelease)

	select {
	case <-firstDone:
	case <-time.After(5 * time.Second):
		t.Fatal("first worker did not stop gracefully")
	}

	require.Eventually(
		t,
		func() bool {
			persisted := loadAgentExecution(t, client, execution.ID)

			return persisted.Status == coredata.AgentExecutionStatusSuspended &&
				persisted.Checkpoint != nil &&
				persisted.ProcessingOwnerToken == nil
		},
		5*time.Second,
		50*time.Millisecond,
	)
	assert.Nil(t, loadAgentInput(t, client, input.ID).ProcessedAt)

	secondWorker := newTestWorker(
		client,
		registry,
		agentexecution.WithWorkerInterval(20*time.Millisecond),
	)

	secondCtx, stopSecond := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopSecond()

	go func() { _ = secondWorker.Run(secondCtx) }()

	require.Eventually(
		t,
		func() bool {
			return loadAgentInput(t, client, input.ID).ProcessedAt != nil
		},
		8*time.Second,
		50*time.Millisecond,
	)

	persisted := loadAgentExecution(t, client, execution.ID)
	assert.Nil(t, persisted.Checkpoint)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, persisted.Status)
}

func TestRecoverStale_DeadLettersAllPendingInputsOnTerminalExecution(t *testing.T) {
	client := test.PGClient(t)
	organizationID := insertTestOrganization(t, client)
	execution := insertExecution(
		t,
		client,
		organizationID,
		"stale-agent",
		"provider",
		"stale-all-pending",
		1,
	)
	batchInput := enqueueAgentInput(t, client, execution, "batch-event", userMessage("batch"))
	extraInput := enqueueAgentInput(t, client, execution, "extra-event", userMessage("extra"))
	ownerToken := "stale-owner"
	now := time.Now()

	var claimed coredata.AgentExecution

	require.NoError(
		t,
		client.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return claimed.ClaimNextForUpdateSkipLocked(ctx, tx, now, ownerToken)
			},
		),
	)
	require.Equal(t, execution.ID, claimed.ID)

	require.NoError(
		t,
		client.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				if err := claimed.SetProcessingInputIDs(
					ctx,
					conn,
					coredata.NewScope(organizationID.TenantID()),
					ownerToken,
					[]string{batchInput.ID.String()},
					now,
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

	runWorker := newTestWorker(
		client,
		&simpleRegistry{agents: map[string]*agent.Agent{}},
		agentexecution.WithWorkerStaleAfter(10*time.Second),
	)
	require.NoError(t, runWorker.RecoverStale(t.Context()))

	loadedBatch := loadAgentInput(t, client, batchInput.ID)
	loadedExtra := loadAgentInput(t, client, extraInput.ID)
	require.NotNil(t, loadedBatch.DeadLetteredAt)
	require.NotNil(t, loadedExtra.DeadLetteredAt)
	assert.Nil(t, loadedBatch.ProcessedAt)
	assert.Nil(t, loadedExtra.ProcessedAt)

	loadedExecution := loadAgentExecution(t, client, execution.ID)
	assert.Equal(t, coredata.AgentExecutionStatusFailed, loadedExecution.Status)
	assert.Nil(t, loadedExecution.ProcessingOwnerToken)
	assert.NotNil(t, loadedExecution.DeadLetteredAt)
}
