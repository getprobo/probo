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

package agent_test

import (
	"context"
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

			dangerousTool, err := agent.FunctionTool[struct{}](
				"dangerous_tool",
				"A dangerous operation",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "executed"}, nil
				},
			)
			require.NoError(t, err)

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
			err = store.Save(context.Background(), "run-approved", &agent.Checkpoint{
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
}
