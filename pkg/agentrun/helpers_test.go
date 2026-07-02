// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package agentrun_test

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
	"go.probo.inc/probo/pkg/agentrun"
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
	opts ...agentrun.WorkerOption,
) *agentrun.Worker {
	store := coredata.NewPGCheckpointer(client)

	baseOpts := []agentrun.WorkerOption{
		agentrun.WithWorkerInterval(250 * time.Millisecond),
	}

	baseOpts = append(baseOpts, opts...)

	return agentrun.NewWorker(
		client,
		store,
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

func insertPendingRun(
	t *testing.T,
	client *pg.Client,
	agentName string,
	inputMessages []llm.Message,
) coredata.AgentRun {
	t.Helper()

	orgID := insertTestOrganization(t, client)

	return insertPendingRunInOrg(t, client, orgID, agentName, inputMessages)
}

func insertPendingRunInOrg(
	t *testing.T,
	client *pg.Client,
	organizationID gid.GID,
	agentName string,
	inputMessages []llm.Message,
) coredata.AgentRun {
	t.Helper()

	runID := gid.New(organizationID.TenantID(), coredata.AgentRunEntityType)

	inputJSON, err := json.Marshal(inputMessages)
	require.NoError(t, err)

	now := time.Now()

	run := coredata.AgentRun{
		ID:             runID,
		OrganizationID: organizationID,
		StartAgentName: agentName,
		Status:         coredata.AgentRunStatusPending,
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

func loadAgentRun(t *testing.T, client *pg.Client, id gid.GID) coredata.AgentRun {
	t.Helper()

	var run coredata.AgentRun

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return run.LoadByID(ctx, conn, coredata.NewNoScope(), id)
		},
	)
	require.NoError(t, err, "cannot load agent run %s", id)

	return run
}

func tryLoadAgentRun(client *pg.Client, id gid.GID) (coredata.AgentRun, error) {
	var run coredata.AgentRun

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return run.LoadByID(ctx, conn, coredata.NewNoScope(), id)
		},
	)

	return run, err
}

func resetRunToPending(t *testing.T, client *pg.Client, runID gid.GID) {
	t.Helper()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`UPDATE agent_runs
				 SET status = 'PENDING',
				     started_at = NULL,
				     updated_at = now()
				 WHERE id = $1`,
				runID.String(),
			)

			return err
		},
	)
	require.NoError(t, err)
}

func overwriteRunInputMessagesRaw(
	t *testing.T,
	client *pg.Client,
	runID gid.GID,
	rawJSON string,
) {
	t.Helper()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`UPDATE agent_runs
				 SET input_messages = $2::jsonb,
				     updated_at = now()
				 WHERE id = $1`,
				runID.String(),
				rawJSON,
			)

			return err
		},
	)
	require.NoError(t, err)
}
