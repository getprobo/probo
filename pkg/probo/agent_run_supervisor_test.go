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

package probo_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/probo"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

var (
	sharedPGClient *pg.Client
	pgOnce         sync.Once
	pgInitErr      error
)

func testPGClient(t *testing.T) *pg.Client {
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

var ensureTableOnce sync.Once

func ensureAgentRunsTable(t *testing.T, client *pg.Client) {
	t.Helper()

	var tableErr error
	ensureTableOnce.Do(func() {
		ctx := context.Background()
		tableErr = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			// Check if the table already exists.
			var exists bool
			err := conn.QueryRow(
				ctx,
				`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'agent_runs')`,
			).Scan(&exists)
			if err != nil {
				return err
			}
			if exists {
				return nil
			}

			// Run the migration DDL.
			_, err = conn.Exec(ctx, `
				CREATE TABLE agent_runs (
					id              TEXT NOT NULL PRIMARY KEY,
					tenant_id       TEXT NOT NULL,
					organization_id TEXT NOT NULL,
					start_agent_name TEXT NOT NULL,
					status          TEXT NOT NULL DEFAULT 'PENDING',
					checkpoint      JSONB,
					input_messages  JSONB NOT NULL,
					result          JSONB,
					error_message   TEXT,
					stop_requested  BOOLEAN NOT NULL DEFAULT FALSE,
					started_at      TIMESTAMPTZ,
					lease_owner     TEXT,
					lease_expires_at TIMESTAMPTZ,
					created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
					updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
				);
				CREATE INDEX idx_agent_runs_status ON agent_runs (status)
					WHERE status IN ('PENDING', 'RUNNING', 'SUSPENDED');
				CREATE INDEX idx_agent_runs_organization_status ON agent_runs (organization_id, status, created_at);
				CREATE INDEX idx_agent_runs_running_lease ON agent_runs (lease_expires_at)
					WHERE status = 'RUNNING';
			`)
			return err
		})
	})
	require.NoError(t, tableErr, "cannot ensure agent_runs table")
}

func cleanupAgentRun(client *pg.Client, id gid.GID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, "DELETE FROM agent_runs WHERE id = $1", id.String())
		return err
	})
}

func insertPendingRun(
	t *testing.T,
	client *pg.Client,
	agentName string,
	inputMessages []llm.Message,
) coredata.AgentRun {
	t.Helper()

	tenantID := gid.NewTenantID()
	orgID := gid.New(tenantID, 1)
	runID := gid.New(tenantID, 2)

	inputJSON, err := json.Marshal(inputMessages)
	require.NoError(t, err)

	now := time.Now()

	run := coredata.AgentRun{
		ID:             runID,
		OrganizationID: orgID,
		StartAgentName: agentName,
		Status:         coredata.AgentRunStatusPending,
		InputMessages:  inputJSON,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = client.WithTx(
		context.Background(),
		func(ctx context.Context, tx pg.Tx) error {
			return run.Insert(ctx, tx, coredata.NewScope(tenantID))
		},
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupAgentRun(client, run.ID)
	})

	return run
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
	if err != nil {
		t.Fatalf("cannot load agent run %s: %v", id, err)
	}

	return run
}

// tryLoadAgentRun is a non-fatal variant safe for use inside
// require.Eventually callbacks (which recover panics).
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

func testLogger() *log.Logger {
	return log.NewLogger(log.WithFormat(log.FormatPretty))
}

// ---------------------------------------------------------------------------
// Mock LLM provider
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Simple agent registry
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Test 2: PGCheckpointStore integration
// ---------------------------------------------------------------------------

func TestPGCheckpointStore(t *testing.T) {
	t.Parallel()

	client := testPGClient(t)
	store := coredata.NewPGCheckpointStore(client)
	ctx := context.Background()

	// Insert a run so the checkpoint store has a row to update.
	run := insertPendingRun(
		t,
		client,
		"test-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}}},
	)
	runID := run.ID.String()

	t.Run(
		"load returns nil when no checkpoint exists",
		func(t *testing.T) {
			cp, err := store.Load(ctx, runID)
			require.NoError(t, err)
			assert.Nil(t, cp)
		},
	)

	t.Run(
		"save and load round-trip",
		func(t *testing.T) {
			original := &agent.Checkpoint{
				Version:   1,
				Status:    agent.CheckpointStatusSuspended,
				AgentName: "test-agent",
				Messages: []llm.Message{
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}},
					{Role: llm.RoleAssistant, Parts: []llm.Part{llm.TextPart{Text: "working..."}}},
				},
				Usage:         llm.Usage{InputTokens: 20, OutputTokens: 10},
				Turns:         1,
				ToolUsedInRun: true,
			}

			err := store.Save(ctx, runID, original)
			require.NoError(t, err)

			loaded, err := store.Load(ctx, runID)
			require.NoError(t, err)
			require.NotNil(t, loaded)

			assert.Equal(t, original.Version, loaded.Version)
			assert.Equal(t, original.Status, loaded.Status)
			assert.Equal(t, original.AgentName, loaded.AgentName)
			assert.Equal(t, original.Usage, loaded.Usage)
			assert.Equal(t, original.Turns, loaded.Turns)
			assert.Equal(t, original.ToolUsedInRun, loaded.ToolUsedInRun)
			assert.Len(t, loaded.Messages, 2)
		},
	)

	t.Run(
		"save overwrites previous checkpoint",
		func(t *testing.T) {
			updated := &agent.Checkpoint{
				Version:   1,
				Status:    agent.CheckpointStatusSuspended,
				AgentName: "test-agent",
				Messages: []llm.Message{
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "hello"}}},
					{Role: llm.RoleAssistant, Parts: []llm.Part{llm.TextPart{Text: "working..."}}},
					{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "continue"}}},
				},
				Usage: llm.Usage{InputTokens: 30, OutputTokens: 15},
				Turns: 2,
			}

			err := store.Save(ctx, runID, updated)
			require.NoError(t, err)

			loaded, err := store.Load(ctx, runID)
			require.NoError(t, err)
			require.NotNil(t, loaded)
			assert.Equal(t, 2, loaded.Turns)
			assert.Len(t, loaded.Messages, 3)
		},
	)

	t.Run(
		"delete clears checkpoint",
		func(t *testing.T) {
			err := store.Delete(ctx, runID)
			require.NoError(t, err)

			loaded, err := store.Load(ctx, runID)
			require.NoError(t, err)
			assert.Nil(t, loaded)
		},
	)

	t.Run(
		"delete is idempotent",
		func(t *testing.T) {
			err := store.Delete(ctx, runID)
			require.NoError(t, err)
		},
	)

	t.Run(
		"save to nonexistent run returns error",
		func(t *testing.T) {
			cp := &agent.Checkpoint{
				Version:   1,
				Status:    agent.CheckpointStatusSuspended,
				AgentName: "test-agent",
			}
			err := store.Save(ctx, "nonexistent-run-id", cp)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "not found")
		},
	)
}

// ---------------------------------------------------------------------------
// Test 3: Supervisor picks up a PENDING run and completes it
// ---------------------------------------------------------------------------

func TestAgentRunSupervisor_PicksUpAndCompletes(t *testing.T) {
	client := testPGClient(t)
	store := coredata.NewPGCheckpointStore(client)

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			stopResponse("Done."),
		},
	}

	ag := agent.New(
		"echo-agent",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithInstructions("Reply with done."),
	)

	registry := &simpleRegistry{
		agents: map[string]*agent.Agent{"echo-agent": ag},
	}

	run := insertPendingRun(
		t,
		client,
		"echo-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "go"}}}},
	)

	supervisor := probo.NewAgentRunSupervisor(
		client,
		store,
		registry,
		testLogger(),
		probo.WithAgentRunSupervisorInterval(500*time.Millisecond),
		probo.WithAgentRunSupervisorLeaseDuration(30*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go supervisor.Run(ctx)

	// Poll until the run is completed.
	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
		"run should reach COMPLETED status",
	)

	completed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint, "checkpoint should be cleared after completion")
	assert.Nil(t, completed.ErrorMessage)
	assert.False(t, completed.StopRequested)
}

// ---------------------------------------------------------------------------
// Test 4: Supervisor stop/resume cycle with checkpoint
// ---------------------------------------------------------------------------

func TestAgentRunSupervisor_StopAndResume(t *testing.T) {
	client := testPGClient(t)
	store := coredata.NewPGCheckpointStore(client)

	// The tool blocks until signaled, giving us time to set stop_requested.
	toolReady := make(chan struct{})
	toolRelease := make(chan struct{})

	slowTool, err := agent.FunctionTool[struct{}](
		"slow_work",
		"Does slow work",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			close(toolReady)
			<-toolRelease
			return agent.ToolResult{Content: "work done"}, nil
		},
	)
	require.NoError(t, err)

	// Provider sequence:
	// Call 1: request tool call (first execution)
	// Call 2: final stop response (after restoration)
	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			// First execution: LLM asks to call the tool.
			toolCallResponse(llm.ToolCall{
				ID:       "tc_1",
				Function: llm.FunctionCall{Name: "slow_work", Arguments: `{}`},
			}),
			// After resume: the incremental checkpoint saved after tool completion
			// means restore continues with these messages; LLM returns final answer.
			stopResponse("All done after resume."),
		},
	}

	ag := agent.New(
		"worker-agent",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(slowTool),
	)

	registry := &simpleRegistry{
		agents: map[string]*agent.Agent{"worker-agent": ag},
	}

	run := insertPendingRun(
		t,
		client,
		"worker-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "do work"}}}},
	)

	supervisor := probo.NewAgentRunSupervisor(
		client,
		store,
		registry,
		testLogger(),
		probo.WithAgentRunSupervisorInterval(500*time.Millisecond),
		probo.WithAgentRunSupervisorLeaseDuration(30*time.Second),
	)

	// --- Phase 1: Start and let the supervisor pick up the run ---

	ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()

	go supervisor.Run(ctx1)

	// Wait for the tool to start executing — this confirms the supervisor
	// claimed the run and the agent called the tool.
	select {
	case <-toolReady:
	case <-ctx1.Done():
		t.Fatal("timed out waiting for tool to start")
	}

	// The run should now be RUNNING.
	running := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusRunning, running.Status)

	// Set stop_requested in the database WHILE the tool is still blocked.
	// The supervisor polls for this on each tick and signals the run's
	// stop channel.
	err = client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				"UPDATE agent_runs SET stop_requested = true WHERE id = $1",
				run.ID.String(),
			)
			return err
		},
	)
	require.NoError(t, err)

	// Give the supervisor at least one tick to poll stop requests and
	// close the run's stop channel before the tool finishes.
	time.Sleep(1 * time.Second)

	// Now release the tool. After completion the coreLoop saves an
	// incremental checkpoint and checks the stop signal at the next
	// turn boundary — it should already be closed.
	close(toolRelease)

	// Wait for the checkpoint to appear. The supervisor leaves the row
	// in RUNNING because SuspendedError triggers the "leaving for stale
	// recovery" path.
	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
		"checkpoint should be saved after stop",
	)

	// Stop the first supervisor.
	cancel1()
	time.Sleep(500 * time.Millisecond)

	// Verify checkpoint content.
	cp, err := store.Load(context.Background(), run.ID.String())
	require.NoError(t, err)
	require.NotNil(t, cp, "checkpoint must exist after suspension")
	assert.Equal(t, agent.CheckpointStatusSuspended, cp.Status)
	assert.Equal(t, "worker-agent", cp.AgentName)
	assert.True(t, len(cp.Messages) > 0, "checkpoint should contain messages")

	// --- Phase 2: Simulate resume by resetting to PENDING ---

	err = client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`UPDATE agent_runs
				 SET status = 'PENDING',
				     stop_requested = false,
				     started_at = NULL,
				     lease_owner = NULL,
				     lease_expires_at = NULL,
				     updated_at = now()
				 WHERE id = $1`,
				run.ID.String(),
			)
			return err
		},
	)
	require.NoError(t, err)

	// Start a fresh supervisor to pick up the resumed run.
	supervisor2 := probo.NewAgentRunSupervisor(
		client,
		store,
		registry,
		testLogger(),
		probo.WithAgentRunSupervisorInterval(500*time.Millisecond),
		probo.WithAgentRunSupervisorLeaseDuration(30*time.Second),
	)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()

	go supervisor2.Run(ctx2)

	// The resumed run should load the checkpoint, call Restore, get the
	// second LLM response (stopResponse), and complete.
	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
		"run should reach COMPLETED after resume",
	)

	completed := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint, "checkpoint should be cleared after completion")
	assert.Nil(t, completed.ErrorMessage)
}

// ---------------------------------------------------------------------------
// Test 5: SIGTERM battle test — realistic multi-turn security audit
// with parallel tool calls, long-running operations, thinking turns,
// and multiple kill/resume cycles.
//
// Simulated workflow (10 tool-call turns + 1 final response):
//
//	Turn 0: [think] scan_repos (single, 800ms)
//	Turn 1: [think] fetch_config ×3 (parallel, 300-500ms)
//	Turn 2: [think] analyze (single, 1000ms)
//	Turn 3: [think] check ×3 (parallel, 400-800ms)
//	Turn 4: [think] deep_analysis (single, 1500ms — long running)
//	Turn 5: [think] generate ×2 (parallel, 500-600ms)
//	Turn 6: [think] cve_lookup (single, 700ms)
//	Turn 7: [think] compile (single, 600ms)
//	Turn 8: [think] validate + format (parallel, 300-400ms)
//	Turn 9: [think] publish (single, 400ms)
//	Turn 10: final response — "Security audit complete..."
//
// SIGTERM is sent 3 times at different points, each time interrupting
// during tool execution (sometimes single, sometimes parallel).
// After each kill the checkpoint is verified to show progressive
// accumulation. A final in-process resume runs the remaining turns
// to completion.
// ---------------------------------------------------------------------------

// workInput is the shared parameter type for all battle-test tools.
type workInput struct {
	Task       string `json:"task"`
	DurationMs int    `json:"duration_ms"`
}

// battleTestResponses returns the full LLM response sequence for a
// simulated security-audit agent. Each tool-call turn includes
// thinking text so the checkpoint messages are realistic.
func battleTestResponses() []*llm.ChatCompletionResponse {
	tc := func(id, name, task string, ms int) llm.ToolCall {
		return llm.ToolCall{
			ID: id,
			Function: llm.FunctionCall{
				Name:      name,
				Arguments: fmt.Sprintf(`{"task":%q,"duration_ms":%d}`, task, ms),
			},
		}
	}

	think := func(text string, calls ...llm.ToolCall) *llm.ChatCompletionResponse {
		return &llm.ChatCompletionResponse{
			Model: "test-model",
			Message: llm.Message{
				Role:      llm.RoleAssistant,
				Parts:     []llm.Part{llm.TextPart{Text: text}},
				ToolCalls: calls,
			},
			Usage:        llm.Usage{InputTokens: 50, OutputTokens: 30},
			FinishReason: llm.FinishReasonToolCalls,
		}
	}

	return []*llm.ChatCompletionResponse{
		// Turn 0 — single long scan
		think(
			"I'll begin the security audit by scanning all repositories to identify codebases, dependency manifests, and access-control configurations.",
			tc("tc_0_1", "scan", "scan_repos", 800),
		),

		// Turn 1 — 3 parallel fetches
		think(
			"Found 3 repositories: api-gateway, auth-service, data-pipeline. Fetching their configurations in parallel to save time.",
			tc("tc_1_1", "fetch", "fetch_api_config", 300),
			tc("tc_1_2", "fetch", "fetch_auth_config", 400),
			tc("tc_1_3", "fetch", "fetch_data_config", 500),
		),

		// Turn 2 — single analysis
		think(
			"All configurations retrieved. Running a comprehensive vulnerability analysis against the OWASP Top-10 checklist.",
			tc("tc_2_1", "analyze", "analyze_configs", 1000),
		),

		// Turn 3 — 3 parallel security checks
		think(
			"Analysis flagged several areas of concern. Running dependency audit, secret scanning, and IAM permission checks in parallel.",
			tc("tc_3_1", "check", "check_dependencies", 600),
			tc("tc_3_2", "check", "check_secrets", 400),
			tc("tc_3_3", "check", "check_permissions", 800),
		),

		// Turn 4 — single very long deep-dive
		think(
			"Multiple issues found: 3 outdated dependencies with known CVEs, 2 overly permissive IAM roles. Performing a deep analysis on the critical findings to determine exploitability and blast radius.",
			tc("tc_4_1", "analyze", "deep_analysis", 1500),
		),

		// Turn 5 — 2 parallel report sections
		think(
			"Deep analysis complete. auth-service uses deprecated TLS 1.1 and data-pipeline stores PII unencrypted. Generating the executive summary and detailed findings sections in parallel.",
			tc("tc_5_1", "generate", "generate_summary", 500),
			tc("tc_5_2", "generate", "generate_findings", 600),
		),

		// Turn 6 — single CVE lookup
		think(
			"Report sections drafted. Cross-referencing all findings against the NVD and GitHub Advisory databases for known CVE identifiers.",
			tc("tc_6_1", "lookup", "cve_lookup", 700),
		),

		// Turn 7 — single compile
		think(
			"CVE-2026-1234 matches the auth-service TLS vulnerability (CVSS 9.1). Compiling all sections, references, and remediation steps into the final report.",
			tc("tc_7_1", "compile", "compile_report", 600),
		),

		// Turn 8 — 2 parallel validation + formatting
		think(
			"Draft report assembled (12 pages). Running structural validation and PDF formatting concurrently.",
			tc("tc_8_1", "validate", "validate_report", 400),
			tc("tc_8_2", "format", "format_pdf", 300),
		),

		// Turn 9 — single publish
		think(
			"Validation passed, PDF formatted. Publishing the finalized audit report to the internal portal.",
			tc("tc_9_1", "publish", "publish_report", 400),
		),

		// Turn 10 — final text response
		{
			Model: "test-model",
			Message: llm.Message{
				Role:  llm.RoleAssistant,
				Parts: []llm.Part{llm.TextPart{Text: "Security audit complete.\n\nFindings:\n- 3 critical (auth-service TLS 1.1, unencrypted PII, CVE-2026-1234)\n- 5 medium (outdated deps, permissive IAM)\n- 4 low (missing rate-limiting, verbose logging)\n\nFull report: https://audits.internal/report-2026-04"}},
			},
			Usage:        llm.Usage{InputTokens: 50, OutputTokens: 30},
			FinishReason: llm.FinishReasonStop,
		},
	}
}

// makeBattleTools creates the tool set for the battle test. Every tool
// shares the same handler that sleeps for the requested duration and
// records progress to a shared file.
func makeBattleTools(progressFile string) []agent.Tool {
	var mu sync.Mutex

	handler := func(_ context.Context, input workInput) (agent.ToolResult, error) {
		// Simulate real work.
		time.Sleep(time.Duration(input.DurationMs) * time.Millisecond)

		// Record completion — written AFTER the sleep so the parent's
		// step count reflects truly-finished work.
		mu.Lock()
		f, err := os.OpenFile(progressFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			mu.Unlock()
			return agent.ToolResult{}, err
		}
		fmt.Fprintln(f, input.Task)
		f.Close()
		mu.Unlock()

		return agent.ToolResult{Content: fmt.Sprintf("completed: %s", input.Task)}, nil
	}

	names := []struct{ name, desc string }{
		{"scan", "Scan repositories for audit targets"},
		{"fetch", "Fetch configuration or source files"},
		{"analyze", "Run vulnerability analysis"},
		{"check", "Execute a specific security check"},
		{"generate", "Generate a report section"},
		{"lookup", "Query external vulnerability databases"},
		{"compile", "Compile report sections into final document"},
		{"validate", "Validate report structure"},
		{"format", "Apply output formatting"},
		{"publish", "Publish report to internal portal"},
	}

	tools := make([]agent.Tool, len(names))
	for i, n := range names {
		tool, err := agent.FunctionTool[workInput](n.name, n.desc, handler)
		if err != nil {
			panic(fmt.Sprintf("cannot create tool %q: %v", n.name, err))
		}
		tools[i] = tool
	}

	return tools
}

func TestAgentRunSupervisor_SIGTERM(t *testing.T) {
	// ---- Subprocess mode ----
	if os.Getenv("TEST_SIGTERM_SUBPROCESS") == "1" {
		runSIGTERMSubprocess()
		return
	}

	// ---- Parent mode ----
	client := testPGClient(t)
	store := coredata.NewPGCheckpointStore(client)

	run := insertPendingRun(
		t,
		client,
		"battle-agent",
		[]llm.Message{{Role: llm.RoleUser, Parts: []llm.Part{llm.TextPart{Text: "Run a full security audit on all repositories."}}}},
	)

	progressFile := filepath.Join(t.TempDir(), "progress")

	// ---- Helpers ----

	countSteps := func() int {
		data, err := os.ReadFile(progressFile)
		if err != nil {
			return 0
		}
		n := 0
		for _, b := range data {
			if b == '\n' {
				n++
			}
		}
		return n
	}

	startSubprocess := func(skipResponses int) *exec.Cmd {
		cmd := exec.Command(
			os.Args[0],
			"-test.run=^TestAgentRunSupervisor_SIGTERM$",
			"-test.v",
		)
		cmd.Env = append(os.Environ(),
			"TEST_SIGTERM_SUBPROCESS=1",
			"TEST_SIGTERM_PROGRESS_FILE="+progressFile,
			"TEST_SIGTERM_SKIP_RESPONSES="+strconv.Itoa(skipResponses),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		require.NoError(t, cmd.Start())
		return cmd
	}

	killAndWait := func(cmd *exec.Cmd) {
		require.NoError(t, cmd.Process.Signal(syscall.SIGTERM))
		err := cmd.Wait()
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				t.Logf("subprocess exited: %v", exitErr)
			} else {
				t.Fatalf("subprocess error: %v", err)
			}
		}
	}

	resetToPending := func() {
		err := client.WithConn(
			context.Background(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(ctx, `
					UPDATE agent_runs
					SET status = 'PENDING',
					    stop_requested = false,
					    started_at = NULL,
					    lease_owner = NULL,
					    lease_expires_at = NULL,
					    updated_at = now()
					WHERE id = $1`,
					run.ID.String(),
				)
				return err
			},
		)
		require.NoError(t, err)
	}

	waitForSteps := func(target int) {
		require.Eventually(
			t,
			func() bool { return countSteps() >= target },
			30*time.Second,
			100*time.Millisecond,
			fmt.Sprintf("expected at least %d completed tool executions", target),
		)
	}

	verifyCheckpoint := func(phase int) *agent.Checkpoint {
		cp, err := store.Load(context.Background(), run.ID.String())
		require.NoError(t, err)
		require.NotNil(t, cp, "phase %d: checkpoint must exist", phase)
		assert.Equal(t, agent.CheckpointStatusSuspended, cp.Status, "phase %d", phase)
		assert.Equal(t, "battle-agent", cp.AgentName, "phase %d", phase)
		assert.Greater(t, len(cp.Messages), 1, "phase %d: checkpoint should have messages", phase)
		assert.Greater(t, cp.Turns, 0, "phase %d: checkpoint should have turns", phase)
		t.Logf(
			"  checkpoint: %d messages, %d turns, usage=%+v",
			len(cp.Messages), cp.Turns, cp.Usage,
		)
		return cp
	}

	// ============================================================
	// Phase 1: SIGTERM during the parallel fetch (turn 1)
	//   Steps so far: turn0=1(scan) + turn1=3(fetch×3) = 4
	// ============================================================
	t.Log("=== Phase 1: SIGTERM after scan + parallel fetch (4 steps) ===")
	cmd1 := startSubprocess(0)
	waitForSteps(4)
	killAndWait(cmd1)

	steps1 := countSteps()
	t.Logf("  %d tool executions completed", steps1)
	require.GreaterOrEqual(t, steps1, 4)

	cp1 := verifyCheckpoint(1)
	require.GreaterOrEqual(t, cp1.Turns, 2, "should have completed at least turns 0-1")

	resetToPending()

	// ============================================================
	// Phase 2: SIGTERM during the parallel security checks (turn 3)
	//   New steps: turn2=1(analyze) + turn3=3(check×3) = 4
	// ============================================================
	t.Log("=== Phase 2: SIGTERM after analyze + parallel checks (4 more steps) ===")
	cmd2 := startSubprocess(cp1.Turns)
	waitForSteps(steps1 + 4)
	killAndWait(cmd2)

	steps2 := countSteps()
	t.Logf("  %d tool executions completed (total)", steps2)
	require.GreaterOrEqual(t, steps2, steps1+4)

	cp2 := verifyCheckpoint(2)
	assert.Greater(t, cp2.Turns, cp1.Turns, "turns should grow")
	assert.Greater(t, len(cp2.Messages), len(cp1.Messages), "messages should grow")

	resetToPending()

	// ============================================================
	// Phase 3: SIGTERM during deep_analysis (turn 4, long-running)
	//   or after generate ×2 (turn 5)
	//   New steps: turn4=1(deep) + turn5=2(generate×2) = 3
	// ============================================================
	t.Log("=== Phase 3: SIGTERM during long-running deep analysis (3 more steps) ===")
	cmd3 := startSubprocess(cp2.Turns)
	waitForSteps(steps2 + 3)
	killAndWait(cmd3)

	steps3 := countSteps()
	t.Logf("  %d tool executions completed (total)", steps3)
	require.GreaterOrEqual(t, steps3, steps2+3)

	cp3 := verifyCheckpoint(3)
	assert.Greater(t, cp3.Turns, cp2.Turns, "turns should grow again")
	assert.Greater(t, len(cp3.Messages), len(cp2.Messages), "messages should grow again")
	assert.Greater(t, cp3.Usage.InputTokens, 0, "usage should accumulate")
	assert.Greater(t, cp3.Usage.OutputTokens, 0, "usage should accumulate")

	t.Logf(
		"  after 3 SIGTERM cycles: %d steps, %d turns, %d messages, usage=%+v",
		steps3, cp3.Turns, len(cp3.Messages), cp3.Usage,
	)

	resetToPending()

	// ============================================================
	// Phase 4: final in-process resume — run remaining turns to
	// completion (lookup, compile, validate+format, publish, done)
	// ============================================================
	t.Log("=== Phase 4: in-process resume to completion ===")

	remaining := battleTestResponses()[cp3.Turns:]
	t.Logf("  %d LLM responses remaining (turns %d–10)", len(remaining), cp3.Turns)

	tools := makeBattleTools(progressFile)

	resumeAgent := agent.New(
		"battle-agent",
		newTestClient(&mockProvider{responses: remaining}),
		agent.WithModel("test-model"),
		agent.WithTools(tools...),
		agent.WithMaxTurns(25),
	)

	supervisor := probo.NewAgentRunSupervisor(
		client,
		store,
		&simpleRegistry{agents: map[string]*agent.Agent{"battle-agent": resumeAgent}},
		testLogger(),
		probo.WithAgentRunSupervisorInterval(500*time.Millisecond),
		probo.WithAgentRunSupervisorLeaseDuration(30*time.Second),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go supervisor.Run(ctx)

	require.Eventually(
		t,
		func() bool {
			r, err := tryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		25*time.Second,
		200*time.Millisecond,
		"run should complete after final resume",
	)

	stepsFinal := countSteps()
	final := loadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, final.Status)
	assert.NotNil(t, final.Result)
	assert.Nil(t, final.Checkpoint, "checkpoint should be cleared")
	assert.Nil(t, final.ErrorMessage)
	assert.Contains(t, string(final.Result), "Security audit complete")

	t.Logf(
		"  battle test done: %d total tool executions across 3 SIGTERM cycles + final resume",
		stepsFinal,
	)
}

// runSIGTERMSubprocess is the child-process entry point. It sets up a
// supervisor with the full security-audit agent (10 distinct tools,
// thinking text, parallel calls, varying durations) and handles
// SIGTERM via signal.NotifyContext — identical to production probod.
func runSIGTERMSubprocess() {
	progressFile := os.Getenv("TEST_SIGTERM_PROGRESS_FILE")
	skip, _ := strconv.Atoi(os.Getenv("TEST_SIGTERM_SKIP_RESPONSES"))

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

	pgClient, err := pg.NewClient(
		pg.WithAddr(addr),
		pg.WithUser(user),
		pg.WithPassword(password),
		pg.WithDatabase(database),
		pg.WithPoolSize(5),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "subprocess: cannot create pg client: %v\n", err)
		os.Exit(1)
	}
	defer pgClient.Close()

	store := coredata.NewPGCheckpointStore(pgClient)
	tools := makeBattleTools(progressFile)

	responses := battleTestResponses()
	if skip > 0 && skip < len(responses) {
		responses = responses[skip:]
	}

	ag := agent.New(
		"battle-agent",
		newTestClient(&mockProvider{responses: responses}),
		agent.WithModel("test-model"),
		agent.WithTools(tools...),
		agent.WithMaxTurns(25),
	)

	supervisor := probo.NewAgentRunSupervisor(
		pgClient,
		store,
		&simpleRegistry{agents: map[string]*agent.Agent{"battle-agent": ag}},
		log.NewLogger(log.WithFormat(log.FormatPretty)),
		probo.WithAgentRunSupervisorInterval(500*time.Millisecond),
		probo.WithAgentRunSupervisorLeaseDuration(5*time.Second),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	err = supervisor.Run(ctx)
	if err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "subprocess: supervisor error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
