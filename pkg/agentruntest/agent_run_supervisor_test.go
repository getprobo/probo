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

package agentruntest_test

import (
	"context"
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
	"go.probo.inc/probo/pkg/agentruntest"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/probo"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

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
// Test 3: Supervisor picks up a PENDING run and completes it
// ---------------------------------------------------------------------------

// Supervisor tests are intentionally sequential. The supervisor claims
// runs cross-tenant via LoadNextPendingForUpdateSkipLocked; running two
// supervisors against the same test database would steal each other's
// runs. If a per-tenant claim filter is ever added, these can go back
// to t.Parallel().
func TestAgentRunSupervisor_PicksUpAndCompletes(t *testing.T) {
	client := agentruntest.PGClient(t)
	store := coredata.NewPGCheckpointer(client)

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

	run := agentruntest.InsertPendingRun(
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
			r, err := agentruntest.TryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
		"run should reach COMPLETED status",
	)

	completed := agentruntest.LoadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusCompleted, completed.Status)
	assert.NotNil(t, completed.Result)
	assert.Nil(t, completed.Checkpoint, "checkpoint should be cleared after completion")
	assert.Nil(t, completed.ErrorMessage)
}

// ---------------------------------------------------------------------------
// Test 4: Supervisor stop/resume cycle with checkpoint
// ---------------------------------------------------------------------------

func TestAgentRunSupervisor_StopAndResume(t *testing.T) {
	client := agentruntest.PGClient(t)
	store := coredata.NewPGCheckpointer(client)

	// The tool blocks until signaled, giving us time to trigger graceful
	// shutdown via the supervisor context while the agent is mid-turn.
	toolReady := make(chan struct{})
	toolRelease := make(chan struct{})

	slowTool := agent.FunctionTool[struct{}](
		"slow_work",
		"Does slow work",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			close(toolReady)
			<-toolRelease
			return agent.ToolResult{Content: "work done"}, nil
		},
	)

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

	run := agentruntest.InsertPendingRun(
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
	running := agentruntest.LoadAgentRun(t, client, run.ID)
	assert.Equal(t, coredata.AgentRunStatusRunning, running.Status)

	// Trigger graceful shutdown of the supervisor: context.AfterFunc
	// registered in Run() fires signalShutdown, closing the shutdown
	// broadcast channel; the per-run forwarder goroutine closes the
	// agent's stop channel.
	cancel1()

	// Wait for the shutdown broadcast to be observed (the AfterFunc
	// goroutine closes it) before releasing the tool. This is
	// deterministic: no wall-clock sleep. The per-run forwarder
	// goroutine observes the same close synchronously and closes the
	// agent stop channel.
	select {
	case <-supervisor.ShutdownBroadcast():
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for supervisor shutdown broadcast")
	}

	// Now release the tool. When the coreLoop resumes control at the
	// next turn boundary it observes the closed stop channel, saves
	// the suspension checkpoint, and returns SuspendedError.
	close(toolRelease)

	// Wait for the checkpoint to appear. The supervisor leaves the row
	// in RUNNING because SuspendedError triggers the "leaving for stale
	// recovery" path.
	require.Eventually(
		t,
		func() bool {
			r, err := agentruntest.TryLoadAgentRun(client, run.ID)
			return err == nil && r.Checkpoint != nil
		},
		10*time.Second,
		200*time.Millisecond,
		"checkpoint should be saved after stop",
	)

	// Verify checkpoint content.
	cp, err := store.Load(context.Background(), run.ID.String())
	require.NoError(t, err)
	require.NotNil(t, cp, "checkpoint must exist after suspension")
	assert.Equal(t, agent.AgentStatusSuspended, cp.Status)
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
			r, err := agentruntest.TryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		10*time.Second,
		200*time.Millisecond,
		"run should reach COMPLETED after resume",
	)

	completed := agentruntest.LoadAgentRun(t, client, run.ID)
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
		tools[i] = agent.FunctionTool[workInput](n.name, n.desc, handler)
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
	client := agentruntest.PGClient(t)
	store := coredata.NewPGCheckpointer(client)

	run := agentruntest.InsertPendingRun(
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
			if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
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
		assert.Equal(t, agent.AgentStatusSuspended, cp.Status, "phase %d", phase)
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
			r, err := agentruntest.TryLoadAgentRun(client, run.ID)
			return err == nil && r.Status == coredata.AgentRunStatusCompleted
		},
		25*time.Second,
		200*time.Millisecond,
		"run should complete after final resume",
	)

	stepsFinal := countSteps()
	final := agentruntest.LoadAgentRun(t, client, run.ID)
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

	store := coredata.NewPGCheckpointer(pgClient)
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
