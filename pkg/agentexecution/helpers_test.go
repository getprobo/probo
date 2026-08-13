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
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agentexecution"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

func testLogger() *log.Logger {
	return log.NewLogger(log.WithFormat(log.FormatPretty))
}

type mockProvider struct {
	mu        sync.Mutex
	responses []*llm.ChatCompletionResponse
	calls     int
}

func (m *mockProvider) ChatCompletion(_ context.Context, _ *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.calls >= len(m.responses) {
		return nil, errors.New("no more mock responses")
	}

	resp := m.responses[m.calls]
	m.calls++

	return resp, nil
}

func (m *mockProvider) ChatCompletionStream(_ context.Context, _ *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	return nil, errors.New("not implemented")
}

func newTestClient(provider llm.Provider) *llm.Client {
	return llm.NewClient(provider, "test")
}

func newDummyAgent(name string, responses []*llm.ChatCompletionResponse, tools ...agent.Tool) *agent.Agent {
	provider := &mockProvider{
		responses: responses,
	}

	opts := []agent.Option{
		agent.WithModel("test-model"),
	}
	if len(tools) > 0 {
		opts = append(opts, agent.WithTools(tools...))
	}

	return agent.New(
		name,
		newTestClient(provider),
		opts...,
	)
}

func newTestWorker(
	client *pg.Client,
	registry agent.AgentRegistry,
	opts ...agentexecution.WorkerOption,
) *agentexecution.Worker {
	baseOpts := []agentexecution.WorkerOption{
		agentexecution.WithWorkerInterval(250 * time.Millisecond),
	}

	baseOpts = append(baseOpts, opts...)

	return agentexecution.NewWorker(
		client,
		registry,
		testLogger(),
		baseOpts...,
	)
}

func stopResponse(text string) *llm.ChatCompletionResponse {
	return &llm.ChatCompletionResponse{
		Model: "test-model",
		Message: llm.Message{
			Role:  llm.RoleAssistant,
			Parts: []llm.Part{llm.TextPart{Text: text}},
		},
		Usage:        llm.Usage{InputTokens: 10, OutputTokens: 5},
		FinishReason: llm.FinishReasonStop,
	}
}

func toolCallResponse(toolCalls ...llm.ToolCall) *llm.ChatCompletionResponse {
	return &llm.ChatCompletionResponse{
		Model: "test-model",
		Message: llm.Message{
			Role:      llm.RoleAssistant,
			ToolCalls: toolCalls,
		},
		Usage:        llm.Usage{InputTokens: 10, OutputTokens: 5},
		FinishReason: llm.FinishReasonToolCalls,
	}
}

type simpleRegistry struct {
	agents map[string]*agent.Agent
}

func (r *simpleRegistry) Agent(name string) (*agent.Agent, error) {
	a, ok := r.agents[name]
	if !ok {
		return nil, fmt.Errorf("agent %q not found", name)
	}

	return a, nil
}

func insertTestOrganization(t *testing.T, client *pg.Client) gid.GID {
	t.Helper()

	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
				orgID.String(),
				tenantID.String(),
				"test-org-"+orgID.String(),
				now,
				now,
			)

			return err
		},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupOrganization(client, orgID)
	})

	return orgID
}

func insertPendingExecution(
	t *testing.T,
	client *pg.Client,
	agentName string,
	inputMessages []llm.Message,
) coredata.AgentExecution {
	t.Helper()

	orgID := insertTestOrganization(t, client)

	return insertPendingExecutionInOrg(t, client, orgID, agentName, inputMessages)
}

func insertPendingExecutionInOrg(
	t *testing.T,
	client *pg.Client,
	organizationID gid.GID,
	agentName string,
	inputMessages []llm.Message,
) coredata.AgentExecution {
	t.Helper()

	runID := gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType)

	inputJSON, err := json.Marshal(inputMessages)
	require.NoError(t, err)

	now := time.Now()

	run := coredata.AgentExecution{
		ID:             runID,
		OrganizationID: organizationID,
		StartAgentName: agentName,
		Status:         coredata.AgentExecutionStatusPending,
		InputMessages:  inputJSON,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = client.WithTx(
		context.Background(),
		func(ctx context.Context, tx pg.Tx) error {
			return run.Insert(ctx, tx, coredata.NewScope(organizationID.TenantID()))
		},
	)
	require.NoError(t, err)

	return run
}

func cleanupOrganization(client *pg.Client, id gid.GID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, "DELETE FROM organizations WHERE id = $1", id.String())
		return err
	})
}

func tryLoadAgentExecution(client *pg.Client, id gid.GID) (coredata.AgentExecution, error) {
	var run coredata.AgentExecution

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return run.LoadByID(ctx, conn, coredata.NewNoScope(), id)
		},
	)

	return run, err
}

func resetExecutionToPending(t *testing.T, client *pg.Client, runID gid.GID) {
	t.Helper()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`UPDATE agent_executions
				 SET status = 'PENDING',
				     started_at = NULL,
				     processing_owner_token = NULL,
				     processing_heartbeat_at = NULL,
				     attempt_count = 0,
				     updated_at = now()
				 WHERE id = $1`,
				runID.String(),
			)

			return err
		},
	)
	require.NoError(t, err)
}

func userMessage(text string) llm.Message {
	return llm.Message{
		Role:  llm.RoleUser,
		Parts: []llm.Part{llm.TextPart{Text: text}},
	}
}

func insertPendingTurn(
	t *testing.T,
	client *pg.Client,
	agentName string,
	prompt string,
) (coredata.AgentExecution, coredata.AgentInput) {
	t.Helper()

	orgID := insertTestOrganization(t, client)
	execution := insertConversationalExecution(
		t,
		client,
		orgID,
		agentName,
		"test",
		t.Name(),
		1,
	)
	input := enqueueAgentInput(t, client, execution, t.Name(), userMessage(prompt))

	return loadAgentExecution(t, client, execution.ID), input
}

func conversationSettled(client *pg.Client, executionID, inputID gid.GID) bool {
	var input coredata.AgentInput

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return input.LoadByID(ctx, conn, coredata.NewNoScope(), inputID)
		},
	)
	if err != nil || input.ProcessedAt == nil {
		return false
	}

	var execution coredata.AgentExecution

	err = client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return execution.LoadByID(ctx, conn, coredata.NewNoScope(), executionID)
		},
	)

	return err == nil &&
		execution.Status == coredata.AgentExecutionStatusPending &&
		execution.Checkpoint == nil
}

func insertConversationalExecution(
	t *testing.T,
	client *pg.Client,
	organizationID gid.GID,
	agentName string,
	source string,
	sessionKey string,
	maxAttempts int,
) coredata.AgentExecution {
	t.Helper()

	now := time.Now()
	execution := coredata.AgentExecution{
		ID:              gid.New(organizationID.TenantID(), coredata.AgentExecutionEntityType),
		OrganizationID:  organizationID,
		StartAgentName:  agentName,
		Source:          &source,
		SessionKey:      &sessionKey,
		SessionMessages: json.RawMessage("[]"),
		MaxAttempts:     maxAttempts,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	require.NoError(
		t,
		client.WithTx(
			context.Background(),
			func(ctx context.Context, tx pg.Tx) error {
				_, err := execution.UpsertConversationalBySourceSession(
					ctx,
					tx,
					coredata.NewScope(organizationID.TenantID()),
				)

				return err
			},
		),
	)

	return execution
}

func enqueueAgentInput(
	t *testing.T,
	client *pg.Client,
	execution coredata.AgentExecution,
	eventID string,
	message llm.Message,
) coredata.AgentInput {
	t.Helper()

	data, err := json.Marshal(message)
	require.NoError(t, err)

	now := time.Now()
	input := coredata.AgentInput{
		ID:               gid.New(execution.ID.TenantID(), coredata.AgentInputEntityType),
		OrganizationID:   execution.OrganizationID,
		AgentExecutionID: execution.ID,
		Source:           *execution.Source,
		SourceEventID:    &eventID,
		Message:          data,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(
		t,
		client.WithTx(
			context.Background(),
			func(ctx context.Context, tx pg.Tx) error {
				_, err := input.EnqueueIdempotently(
					ctx,
					tx,
					coredata.NewScope(execution.ID.TenantID()),
				)

				return err
			},
		),
	)

	return input
}

func enqueueAgentInputWithIdentity(
	t *testing.T,
	client *pg.Client,
	execution coredata.AgentExecution,
	eventID string,
	message llm.Message,
	identityID gid.GID,
) coredata.AgentInput {
	t.Helper()

	data, err := json.Marshal(message)
	require.NoError(t, err)

	now := time.Now()
	input := coredata.AgentInput{
		ID:               gid.New(execution.ID.TenantID(), coredata.AgentInputEntityType),
		OrganizationID:   execution.OrganizationID,
		AgentExecutionID: execution.ID,
		Source:           *execution.Source,
		SourceEventID:    &eventID,
		IdentityID:       &identityID,
		Message:          data,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	require.NoError(
		t,
		client.WithTx(
			context.Background(),
			func(ctx context.Context, tx pg.Tx) error {
				_, err := input.EnqueueIdempotently(
					ctx,
					tx,
					coredata.NewScope(execution.ID.TenantID()),
				)

				return err
			},
		),
	)

	return input
}

func loadAgentExecution(
	t *testing.T,
	client *pg.Client,
	id gid.GID,
) coredata.AgentExecution {
	t.Helper()

	var execution coredata.AgentExecution

	require.NoError(
		t,
		client.WithConn(
			context.Background(),
			func(ctx context.Context, conn pg.Querier) error {
				return execution.LoadByID(ctx, conn, coredata.NewNoScope(), id)
			},
		),
	)

	return execution
}

func loadCheckpoint(t *testing.T, client *pg.Client, id gid.GID) *agent.Checkpoint {
	t.Helper()

	execution := loadAgentExecution(t, client, id)
	if execution.Checkpoint == nil {
		return nil
	}

	cp := new(agent.Checkpoint)
	require.NoError(t, json.Unmarshal(execution.Checkpoint, cp))

	return cp
}

func loadAgentInput(t *testing.T, client *pg.Client, id gid.GID) coredata.AgentInput {
	t.Helper()

	var input coredata.AgentInput

	require.NoError(
		t,
		client.WithConn(
			context.Background(),
			func(ctx context.Context, conn pg.Querier) error {
				return input.LoadByID(ctx, conn, coredata.NewNoScope(), id)
			},
		),
	)

	return input
}

func overwriteAgentInputMessageRaw(
	t *testing.T,
	client *pg.Client,
	inputID gid.GID,
	rawJSON string,
) {
	t.Helper()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`UPDATE agent_inputs
				 SET message = $2::jsonb,
				     updated_at = now()
				 WHERE id = $1`,
				inputID.String(),
				rawJSON,
			)

			return err
		},
	)
	require.NoError(t, err)
}
