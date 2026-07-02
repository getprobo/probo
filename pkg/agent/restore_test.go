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

package agent_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

type memoryCheckpointer struct {
	mu          sync.Mutex
	checkpoints map[string]*agent.Checkpoint
}

func newMemoryCheckpointer() *memoryCheckpointer {
	return &memoryCheckpointer{
		checkpoints: make(map[string]*agent.Checkpoint),
	}
}

func (s *memoryCheckpointer) Save(_ context.Context, runID string, cp *agent.Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clone := *cp
	s.checkpoints[runID] = &clone

	return nil
}

func (s *memoryCheckpointer) Load(_ context.Context, runID string) (*agent.Checkpoint, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp, ok := s.checkpoints[runID]
	if !ok {
		return nil, nil
	}

	clone := *cp

	return &clone, nil
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

type saveFailCheckpointer struct {
	cp *agent.Checkpoint
}

func (s *saveFailCheckpointer) Save(_ context.Context, _ string, _ *agent.Checkpoint) error {
	return errors.New("save exploded")
}

func (s *saveFailCheckpointer) Load(_ context.Context, _ string) (*agent.Checkpoint, error) {
	clone := *s.cp

	return &clone, nil
}

func TestRestore(t *testing.T) {
	t.Parallel()

	t.Run(
		"no checkpoint returns error",
		func(t *testing.T) {
			t.Parallel()

			store := newMemoryCheckpointer()
			registry := &simpleRegistry{agents: map[string]*agent.Agent{}}

			_, err := agent.Restore(
				context.Background(),
				store,
				"nonexistent-run",
				registry,
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "no checkpoint")
		},
	)

	t.Run(
		"suspended checkpoint restores and completes",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Restored successfully."),
				},
			}

			ag := agent.New(
				"test-agent",
				newTestClient(provider),
				agent.WithInstructions("You are a test agent."),
				agent.WithModel("test-model"),
			)

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-suspended", &agent.Checkpoint{
				Status:    agent.AgentStatusSuspended,
				AgentName: "test-agent",
				Messages: []llm.Message{
					{
						Role:  llm.RoleUser,
						Parts: []llm.Part{llm.TextPart{Text: "Hello"}},
					},
					{
						Role:  llm.RoleAssistant,
						Parts: []llm.Part{llm.TextPart{Text: "Working on it..."}},
					},
				},
				Usage: llm.Usage{InputTokens: 20, OutputTokens: 10},
				Turns: 1,
			})
			require.NoError(t, err)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"test-agent": ag,
				},
			}

			result, err := agent.Restore(
				context.Background(),
				store,
				"run-suspended",
				registry,
			)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, "Restored successfully.", result.FinalMessage().Text())
			assert.Equal(t, 2, result.Turns, "turns should include initial plus restored")
			assert.Equal(t, 30, result.Usage.InputTokens, "usage should accumulate")
			assert.Equal(t, 15, result.Usage.OutputTokens, "usage should accumulate")
		},
	)

	t.Run(
		"awaiting approval without input returns InterruptedError",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Done."),
				},
			}

			ag := agent.New(
				"test-agent",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"dangerous_tool"},
				}),
			)

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-approval", &agent.Checkpoint{
				Status:    agent.AgentStatusAwaitingApproval,
				AgentName: "test-agent",
				Messages: []llm.Message{
					{
						Role:  llm.RoleUser,
						Parts: []llm.Part{llm.TextPart{Text: "Do the thing"}},
					},
				},
				PendingToolCalls: []llm.ToolCall{
					{
						ID: "tc_1",
						Function: llm.FunctionCall{
							Name:      "dangerous_tool",
							Arguments: `{}`,
						},
					},
				},
				PendingApprovals: []llm.ToolCall{
					{
						ID: "tc_1",
						Function: llm.FunctionCall{
							Name:      "dangerous_tool",
							Arguments: `{}`,
						},
					},
				},
				Usage: llm.Usage{InputTokens: 10, OutputTokens: 5},
				Turns: 1,
			})
			require.NoError(t, err)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"test-agent": ag,
				},
			}

			_, err = agent.Restore(
				context.Background(),
				store,
				"run-approval",
				registry,
			)

			require.Error(t, err)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)
			assert.Len(t, interrupted.PendingApprovals, 1)
			assert.Equal(t, "dangerous_tool", interrupted.PendingApprovals[0].Function.Name)
			assert.Equal(t, 1, interrupted.Turns)
			assert.Equal(t, 10, interrupted.Usage.InputTokens)
		},
	)

	t.Run(
		"awaiting approval with input resumes execution",
		func(t *testing.T) {
			t.Parallel()

			dangerousTool := agent.FunctionTool[struct{}](
				"dangerous_tool",
				"A dangerous operation",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "executed"}, nil
				},
			)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Operation approved and completed."),
				},
			}

			ag := agent.New(
				"test-agent",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(dangerousTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"dangerous_tool"},
				}),
			)

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-approved", &agent.Checkpoint{
				Status:    agent.AgentStatusAwaitingApproval,
				AgentName: "test-agent",
				Messages: []llm.Message{
					{
						Role:  llm.RoleUser,
						Parts: []llm.Part{llm.TextPart{Text: "Do the thing"}},
					},
				},
				PendingToolCalls: []llm.ToolCall{
					{
						ID: "tc_1",
						Function: llm.FunctionCall{
							Name:      "dangerous_tool",
							Arguments: `{}`,
						},
					},
				},
				PendingApprovals: []llm.ToolCall{
					{
						ID: "tc_1",
						Function: llm.FunctionCall{
							Name:      "dangerous_tool",
							Arguments: `{}`,
						},
					},
				},
				ApprovalInput: map[string]agent.ApprovalResult{
					"tc_1": {Approved: true},
				},
				Usage: llm.Usage{InputTokens: 10, OutputTokens: 5},
				Turns: 1,
			})
			require.NoError(t, err)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"test-agent": ag,
				},
			}

			result, err := agent.Restore(
				context.Background(),
				store,
				"run-approved",
				registry,
			)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, "Operation approved and completed.", result.FinalMessage().Text())
		},
	)

	t.Run(
		"nested approval rejects multiple inner checkpoints",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"test-agent",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-nested", &agent.Checkpoint{
				Status:    agent.AgentStatusAwaitingApproval,
				AgentName: "test-agent",
				Messages: []llm.Message{
					{
						Role:  llm.RoleUser,
						Parts: []llm.Part{llm.TextPart{Text: "Do things"}},
					},
				},
				PendingToolCalls: []llm.ToolCall{
					{
						ID: "tc_1",
						Function: llm.FunctionCall{
							Name:      "inner_tool",
							Arguments: `{}`,
						},
					},
				},
				PendingApprovals: []llm.ToolCall{
					{
						ID: "tc_1",
						Function: llm.FunctionCall{
							Name:      "inner_tool",
							Arguments: `{}`,
						},
					},
				},
				InnerCheckpoints: map[string]*agent.Checkpoint{
					"tc_inner_1": {
						Status:    agent.AgentStatusAwaitingApproval,
						AgentName: "inner-agent-1",
					},
					"tc_inner_2": {
						Status:    agent.AgentStatusAwaitingApproval,
						AgentName: "inner-agent-2",
					},
				},
				Usage: llm.Usage{InputTokens: 10, OutputTokens: 5},
				Turns: 1,
			})
			require.NoError(t, err)

			innerAgent1 := agent.New(
				"inner-agent-1",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)
			innerAgent2 := agent.New(
				"inner-agent-2",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"test-agent":    ag,
					"inner-agent-1": innerAgent1,
					"inner-agent-2": innerAgent2,
				},
			}

			_, err = agent.Restore(
				context.Background(),
				store,
				"run-nested",
				registry,
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "expected one inner checkpoint")
		},
	)

	t.Run(
		"unknown agent name returns error",
		func(t *testing.T) {
			t.Parallel()

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-unknown", &agent.Checkpoint{
				Status:    agent.AgentStatusSuspended,
				AgentName: "missing-agent",
			})
			require.NoError(t, err)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{},
			}

			_, err = agent.Restore(
				context.Background(),
				store,
				"run-unknown",
				registry,
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot resolve agent")
		},
	)

	t.Run(
		"unknown checkpoint status returns error",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"test-agent",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-bad-status", &agent.Checkpoint{
				Status:    agent.AgentStatus("bogus"),
				AgentName: "test-agent",
			})
			require.NoError(t, err)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"test-agent": ag,
				},
			}

			_, err = agent.Restore(
				context.Background(),
				store,
				"run-bad-status",
				registry,
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "unknown checkpoint status")
		},
	)

	t.Run(
		"buildCheckpoint captures MaxTurns in config snapshot",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"producer-agent",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
				agent.WithMaxTurns(7),
			)

			store := newMemoryCheckpointer()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := ag.Run(
				ctx,
				[]llm.Message{
					{
						Role:  llm.RoleUser,
						Parts: []llm.Part{llm.TextPart{Text: "begin"}},
					},
				},
				agent.WithCheckpointer(store, "run-save-side"),
			)

			var se *agent.SuspendedError
			require.ErrorAs(t, err, &se)

			cp, loadErr := store.Load(context.Background(), "run-save-side")
			require.NoError(t, loadErr)
			require.NotNil(t, cp)
			assert.Equal(t, 7, cp.Config.MaxTurns)
		},
	)

	t.Run(
		"checkpoint config supersedes live agent config on restore",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Completed after resume."),
				},
			}

			// Agent registered at restore time has a bound tighter
			// than the count of turns already taken. Without the
			// checkpoint config snapshot, coreLoop would immediately
			// trip MaxTurnsExceededError on its first iteration.
			restoreAgent := agent.New(
				"test-agent",
				newTestClient(provider),
				agent.WithInstructions("Test."),
				agent.WithModel("test-model"),
				agent.WithMaxTurns(5),
			)

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-config", &agent.Checkpoint{
				Status:    agent.AgentStatusSuspended,
				AgentName: "test-agent",
				Config:    agent.AgentConfig{MaxTurns: 20},
				Messages: []llm.Message{
					{
						Role:  llm.RoleUser,
						Parts: []llm.Part{llm.TextPart{Text: "hi"}},
					},
				},
				Turns: 15,
			})
			require.NoError(t, err)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"test-agent": restoreAgent,
				},
			}

			result, err := agent.Restore(
				context.Background(),
				store,
				"run-config",
				registry,
			)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, "Completed after resume.", result.FinalMessage().Text())
		},
	)

	t.Run(
		"nested suspended restore persists progress when one inner agent missing",
		func(t *testing.T) {
			t.Parallel()

			outerAgent := agent.New(
				"outer-agent",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)
			// done-agent resolves and completes on restore; its
			// progress must be persisted even though inner-agent
			// (below) cannot be resolved.
			doneAgent := agent.New(
				"done-agent",
				newTestClient(&mockProvider{
					responses: []*llm.ChatCompletionResponse{
						stopResponse("inner done"),
					},
				}),
				agent.WithModel("test-model"),
			)

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-nested-missing-inner", &agent.Checkpoint{
				Status:    agent.AgentStatusSuspended,
				AgentName: "outer-agent",
				Messages: []llm.Message{
					{
						Role:  llm.RoleUser,
						Parts: []llm.Part{llm.TextPart{Text: "continue"}},
					},
				},
				AllToolCalls: []llm.ToolCall{
					{
						ID: "tc_done",
						Function: llm.FunctionCall{
							Name:      "call_done",
							Arguments: `{"input":"go"}`,
						},
					},
					{
						ID: "tc_missing",
						Function: llm.FunctionCall{
							Name:      "call_inner",
							Arguments: `{"input":"go"}`,
						},
					},
				},
				InnerCheckpoints: map[string]*agent.Checkpoint{
					"tc_done": {
						Status:    agent.AgentStatusSuspended,
						AgentName: "done-agent",
						Messages: []llm.Message{
							{
								Role:  llm.RoleUser,
								Parts: []llm.Part{llm.TextPart{Text: "go"}},
							},
						},
					},
					"tc_missing": {
						Status:    agent.AgentStatusSuspended,
						AgentName: "inner-agent",
					},
				},
			})
			require.NoError(t, err)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"outer-agent": outerAgent,
					"done-agent":  doneAgent,
				},
			}

			_, err = agent.Restore(
				context.Background(),
				store,
				"run-nested-missing-inner",
				registry,
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `cannot resolve inner agent "inner-agent"`)

			cp, loadErr := store.Load(context.Background(), "run-nested-missing-inner")
			require.NoError(t, loadErr)
			require.NotNil(t, cp)

			// The unresolved tool call is retained for a future retry.
			require.Contains(t, cp.InnerCheckpoints, "tc_missing")

			// The resolved tool call's progress was persisted: its
			// inner checkpoint is dropped and its result recorded as a
			// completed call, so a later retry does not re-run it.
			require.NotContains(t, cp.InnerCheckpoints, "tc_done")
			require.Len(t, cp.CompletedCalls, 1)
			require.Equal(t, "tc_done", cp.CompletedCalls[0].ToolCallID)
			require.Equal(t, "inner done", cp.CompletedCalls[0].Result.Content)
		},
	)

	t.Run(
		"nested suspended restore returns suspended when inner stays suspended",
		func(t *testing.T) {
			t.Parallel()

			outerAgent := agent.New(
				"outer-agent",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)
			innerAgent := agent.New(
				"inner-agent",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-nested-still-suspended", &agent.Checkpoint{
				Status:    agent.AgentStatusSuspended,
				AgentName: "outer-agent",
				Messages: []llm.Message{
					{
						Role:  llm.RoleUser,
						Parts: []llm.Part{llm.TextPart{Text: "continue"}},
					},
				},
				AllToolCalls: []llm.ToolCall{
					{
						ID: "tc_inner",
						Function: llm.FunctionCall{
							Name:      "call_inner",
							Arguments: `{"input":"go"}`,
						},
					},
				},
				InnerCheckpoints: map[string]*agent.Checkpoint{
					"tc_inner": {
						Status:    agent.AgentStatusSuspended,
						AgentName: "inner-agent",
					},
				},
			})
			require.NoError(t, err)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"outer-agent": outerAgent,
					"inner-agent": innerAgent,
				},
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err = agent.Restore(
				ctx,
				store,
				"run-nested-still-suspended",
				registry,
			)

			var se *agent.SuspendedError
			require.ErrorAs(t, err, &se)
			require.NotNil(t, se.Checkpoint)
			require.Contains(t, se.Checkpoint.InnerCheckpoints, "tc_inner")
		},
	)

	t.Run(
		"nested suspended restore joins save failure with restore error",
		func(t *testing.T) {
			t.Parallel()

			outerAgent := agent.New(
				"outer-agent",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			store := &saveFailCheckpointer{
				cp: &agent.Checkpoint{
					Status:    agent.AgentStatusSuspended,
					AgentName: "outer-agent",
					AllToolCalls: []llm.ToolCall{
						{
							ID: "tc_missing",
							Function: llm.FunctionCall{
								Name:      "call_inner",
								Arguments: `{"input":"go"}`,
							},
						},
					},
					InnerCheckpoints: map[string]*agent.Checkpoint{
						"tc_missing": {
							Status:    agent.AgentStatusSuspended,
							AgentName: "inner-agent",
						},
					},
				},
			}

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"outer-agent": outerAgent,
				},
			}

			_, err := agent.Restore(
				context.Background(),
				store,
				"run-nested-save-fail",
				registry,
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `cannot resolve inner agent "inner-agent"`)
			assert.Contains(t, err.Error(), "cannot save nested restore progress")
		},
	)

	t.Run(
		"nested awaiting approval with unknown inner agent returns error",
		func(t *testing.T) {
			t.Parallel()

			outerAgent := agent.New(
				"outer-agent",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			store := newMemoryCheckpointer()
			err := store.Save(context.Background(), "run-awaiting-missing-inner", &agent.Checkpoint{
				Status:    agent.AgentStatusAwaitingApproval,
				AgentName: "outer-agent",
				InnerCheckpoints: map[string]*agent.Checkpoint{
					"tc_inner": {
						Status:    agent.AgentStatusAwaitingApproval,
						AgentName: "inner-agent",
					},
				},
			})
			require.NoError(t, err)

			registry := &simpleRegistry{
				agents: map[string]*agent.Agent{
					"outer-agent": outerAgent,
				},
			}

			_, err = agent.Restore(
				context.Background(),
				store,
				"run-awaiting-missing-inner",
				registry,
			)
			require.Error(t, err)
			assert.Contains(t, err.Error(), `cannot resolve inner agent "inner-agent"`)
		},
	)
}
