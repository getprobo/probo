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

package agentrun_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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

var (
	sharedPGClient  *pg.Client
	pgOnce          sync.Once
	pgInitErr       error
	ensureTableOnce sync.Once
	ensureTableErr  error
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
		agentrun.WithWorkerLeaseDuration(30 * time.Second),
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

func pgClient(t *testing.T) *pg.Client {
	t.Helper()

	pgOnce.Do(func() {
		addr := os.Getenv("PROBO_TEST_PG_ADDR")
		if addr == "" {
			addr = "localhost:5432"
		}

		user := os.Getenv("PROBO_TEST_PG_USER")
		if user == "" {
			user = "probod"
		}

		password := os.Getenv("PROBO_TEST_PG_PASSWORD")
		if password == "" {
			password = "probod"
		}

		database := os.Getenv("PROBO_TEST_PG_DATABASE")
		if database == "" {
			database = "probod_test"
		}

		sharedPGClient, pgInitErr = pg.NewClient(
			pg.WithAddr(addr),
			pg.WithUser(user),
			pg.WithPassword(password),
			pg.WithDatabase(database),
			pg.WithPoolSize(5),
		)
		if pgInitErr != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pgInitErr = sharedPGClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(ctx, "SELECT 1")
			return err
		})
	})

	if pgInitErr != nil {
		t.Skipf("cannot connect to test database: %v", pgInitErr)
	}

	ensureAgentRunsTable(t, sharedPGClient)

	return sharedPGClient
}

func ensureAgentRunsTable(t *testing.T, client *pg.Client) {
	t.Helper()

	ensureTableOnce.Do(func() {
		ctx := context.Background()
		ensureTableErr = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			var exists bool
			if err := conn.QueryRow(
				ctx,
				`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'agent_runs')`,
			).Scan(&exists); err != nil {
				return fmt.Errorf("cannot check agent_runs existence: %w", err)
			}

			if !exists {
				ddl, err := coredata.Migrations.ReadFile("migrations/20260424T173529Z.sql")
				if err != nil {
					return fmt.Errorf("cannot read agent_runs base migration: %w", err)
				}

				if _, err := conn.Exec(ctx, string(ddl)); err != nil {
					return fmt.Errorf("cannot apply agent_runs base migration: %w", err)
				}
			}

			var hasLeaseGeneration bool
			if err := conn.QueryRow(
				ctx,
				`SELECT EXISTS (
					SELECT 1
					FROM information_schema.columns
					WHERE table_name = 'agent_runs'
						AND column_name = 'lease_generation'
				)`,
			).Scan(&hasLeaseGeneration); err != nil {
				return fmt.Errorf("cannot check lease_generation column: %w", err)
			}

			if !hasLeaseGeneration {
				ddl, err := coredata.Migrations.ReadFile("migrations/20260607T060000Z.sql")
				if err != nil {
					return fmt.Errorf("cannot read agent_runs lease generation migration: %w", err)
				}

				if _, err := conn.Exec(ctx, string(ddl)); err != nil {
					return fmt.Errorf("cannot apply agent_runs lease generation migration: %w", err)
				}
			}

			return nil
		})
	})
	require.NoError(t, ensureTableErr, "cannot ensure agent_runs table")
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
				     lease_expires_at = NULL,
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

func bumpRunLeaseGeneration(t *testing.T, client *pg.Client, runID gid.GID) {
	t.Helper()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`UPDATE agent_runs
				 SET lease_generation = lease_generation + 1,
				     updated_at = now()
				 WHERE id = $1`,
				runID.String(),
			)
			return err
		},
	)
	require.NoError(t, err)
}
