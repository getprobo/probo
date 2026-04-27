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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

func TestRun_CtxCancelGracefulSuspend(t *testing.T) {
	t.Parallel()

	t.Run(
		"cancel before first turn suspends with empty checkpoint",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("never called"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
			)

			store := newMemoryCheckpointer()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := ag.Run(
				ctx,
				[]llm.Message{userMessage("hi")},
				agent.WithCheckpointer(store, "run-cancel"),
			)

			var se *agent.SuspendedError
			require.ErrorAs(t, err, &se)
			assert.Equal(t, 0, provider.calls, "LLM must not be invoked when ctx was already cancelled at entry")

			// When a checkpointer is configured, the persistent store
			// is the source of truth — the error itself doesn't carry a
			// Checkpoint. Load from the store to verify.
			cp, loadErr := store.Load(context.Background(), "run-cancel")
			require.NoError(t, loadErr)
			require.NotNil(t, cp, "checkpoint should be persisted before SuspendedError surfaces")
			assert.Equal(t, agent.AgentStatusSuspended, cp.Status)
		},
	)

	t.Run(
		"cancel mid-run preserves the just-completed turn",
		func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					// First turn completes a tool call; the tool body
					// then cancels ctx so the next turn-boundary check
					// in coreLoop observes the cancellation.
					{
						Message: llm.Message{
							Role: llm.RoleAssistant,
							ToolCalls: []llm.ToolCall{{
								ID:       "tc_1",
								Function: llm.FunctionCall{Name: "noop", Arguments: `{}`},
							}},
						},
						FinishReason: llm.FinishReasonToolCalls,
					},
					stopResponse("never reached"),
				},
			}

			noopTool := agent.FunctionTool[struct{}](
				"noop",
				"no-op",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					cancel()
					return agent.ToolResult{Content: "ok"}, nil
				},
			)

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(noopTool),
			)

			store := newMemoryCheckpointer()

			_, err := ag.Run(
				ctx,
				[]llm.Message{userMessage("hi")},
				agent.WithCheckpointer(store, "run-mid"),
			)

			var se *agent.SuspendedError
			require.ErrorAs(t, err, &se)
			assert.Equal(t, 1, provider.calls, "second LLM call must not fire after cancel")

			cp, loadErr := store.Load(context.Background(), "run-mid")
			require.NoError(t, loadErr)
			require.NotNil(t, cp)
			assert.Equal(t, agent.AgentStatusSuspended, cp.Status)
			// The first LLM call completed; its output and the tool
			// reply must be in the checkpointed messages so a Restore
			// can resume from the next turn.
			assert.Equal(t, 1, cp.Turns)
			assert.GreaterOrEqual(t, len(cp.Messages), 3, "user + assistant tool-call + tool result")
		},
	)
}
