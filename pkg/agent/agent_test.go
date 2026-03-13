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
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

type mockProvider struct {
	responses []*llm.ChatCompletionResponse
	calls     int
}

func (m *mockProvider) ChatCompletion(_ context.Context, _ *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
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

type mockChatStream struct {
	events []llm.ChatCompletionStreamEvent
	pos    int
}

func (s *mockChatStream) Next() bool {
	return s.pos < len(s.events)
}

func (s *mockChatStream) Event() llm.ChatCompletionStreamEvent {
	ev := s.events[s.pos]
	s.pos++
	return ev
}

func (s *mockChatStream) Err() error   { return nil }
func (s *mockChatStream) Close() error { return nil }

type mockStreamProvider struct {
	stream llm.ChatCompletionStream
	calls  int
}

func (p *mockStreamProvider) ChatCompletion(_ context.Context, _ *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *mockStreamProvider) ChatCompletionStream(_ context.Context, _ *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	p.calls++
	return p.stream, nil
}

type mockMultiStreamProvider struct {
	streams []llm.ChatCompletionStream
	calls   int
}

func (p *mockMultiStreamProvider) ChatCompletion(_ context.Context, _ *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *mockMultiStreamProvider) ChatCompletionStream(_ context.Context, _ *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	if p.calls >= len(p.streams) {
		return nil, errors.New("no more mock streams")
	}
	s := p.streams[p.calls]
	p.calls++
	return s, nil
}

type blockingGuardrail struct {
	keyword string
}

func (g *blockingGuardrail) Name() string { return "blocker" }

func (g *blockingGuardrail) Check(_ context.Context, messages []llm.Message) (*agent.GuardrailResult, error) {
	for _, m := range messages {
		if m.Role == llm.RoleUser {
			for _, p := range m.Parts {
				if tp, ok := p.(llm.TextPart); ok {
					if tp.Text == g.keyword {
						return &agent.GuardrailResult{
							Tripwire: true,
							Message:  "blocked content detected",
						}, nil
					}
				}
			}
		}
	}
	return nil, nil
}

type outputBlocker struct{}

func (g *outputBlocker) Name() string { return "output_blocker" }

func (g *outputBlocker) Check(_ context.Context, message llm.Message) (*agent.GuardrailResult, error) {
	if message.Text() == "bad response" {
		return &agent.GuardrailResult{
			Tripwire: true,
			Message:  "output blocked",
		}, nil
	}
	return nil, nil
}

type recordingHook struct {
	agent.NoOpHooks
	runStarted bool
	runEnded   bool
	toolNames  []string
	handoffs   []string
}

func (h *recordingHook) OnRunStart(_ context.Context, _ *agent.Agent, _ []llm.Message) {
	h.runStarted = true
}

func (h *recordingHook) OnRunEnd(_ context.Context, _ *agent.Agent, _ *agent.Result, _ error) {
	h.runEnded = true
}

func (h *recordingHook) OnToolEnd(_ context.Context, _ *agent.Agent, tool agent.Tool, _ agent.ToolResult, _ error) {
	h.toolNames = append(h.toolNames, tool.Name())
}

func (h *recordingHook) OnHandoff(_ context.Context, from *agent.Agent, to *agent.Agent) {
	h.handoffs = append(h.handoffs, from.Name()+"->"+to.Name())
}

type recordingAgentHook struct {
	agent.NoOpAgentHooks
	started   bool
	ended     bool
	handoffed bool
}

func (h *recordingAgentHook) OnStart(_ context.Context, _ *agent.Agent) {
	h.started = true
}

func (h *recordingAgentHook) OnEnd(_ context.Context, _ *agent.Agent, _ string) {
	h.ended = true
}

func (h *recordingAgentHook) OnHandoff(_ context.Context, _ *agent.Agent, _ *agent.Agent) {
	h.handoffed = true
}

type testSession struct {
	messages map[string][]llm.Message
}

func newTestSession() *testSession {
	return &testSession{messages: make(map[string][]llm.Message)}
}

func (s *testSession) Load(_ context.Context, sessionID string) ([]llm.Message, error) {
	msgs := s.messages[sessionID]
	cp := make([]llm.Message, len(msgs))
	copy(cp, msgs)
	return cp, nil
}

func (s *testSession) Save(_ context.Context, sessionID string, messages []llm.Message) error {
	cp := make([]llm.Message, len(messages))
	copy(cp, messages)
	s.messages[sessionID] = cp
	return nil
}

type failingSession struct{}

func (s *failingSession) Load(_ context.Context, _ string) ([]llm.Message, error) {
	return nil, nil
}

func (s *failingSession) Save(_ context.Context, _ string, _ []llm.Message) error {
	return errors.New("disk full")
}

type errorLoadSession struct{}

func (s *errorLoadSession) Load(_ context.Context, _ string) ([]llm.Message, error) {
	return nil, errors.New("storage unavailable")
}

func (s *errorLoadSession) Save(_ context.Context, _ string, _ []llm.Message) error {
	return nil
}

type errorGuardrail struct {
	err error
}

func (g *errorGuardrail) Name() string { return "error_guardrail" }

func (g *errorGuardrail) Check(_ context.Context, _ []llm.Message) (*agent.GuardrailResult, error) {
	return nil, g.err
}

type errorOutputGuardrail struct {
	err error
}

func (g *errorOutputGuardrail) Name() string { return "error_output_guardrail" }

func (g *errorOutputGuardrail) Check(_ context.Context, _ llm.Message) (*agent.GuardrailResult, error) {
	return nil, g.err
}

type errorChatStream struct {
	err error
}

func (s *errorChatStream) Next() bool                          { return false }
func (s *errorChatStream) Event() llm.ChatCompletionStreamEvent { return llm.ChatCompletionStreamEvent{} }
func (s *errorChatStream) Err() error                          { return s.err }
func (s *errorChatStream) Close() error                        { return nil }

type errorStreamProvider struct {
	err error
}

func (p *errorStreamProvider) ChatCompletion(_ context.Context, _ *llm.ChatCompletionRequest) (*llm.ChatCompletionResponse, error) {
	return nil, errors.New("not implemented")
}

func (p *errorStreamProvider) ChatCompletionStream(_ context.Context, _ *llm.ChatCompletionRequest) (llm.ChatCompletionStream, error) {
	return nil, p.err
}

func newTestClient(provider llm.Provider) *llm.Client {
	return llm.NewClient(provider, "test")
}

func userMessage(text string) llm.Message {
	return llm.Message{
		Role:  llm.RoleUser,
		Parts: []llm.Part{llm.TextPart{Text: text}},
	}
}

func assistantMessage(text string) llm.Message {
	return llm.Message{
		Role:  llm.RoleAssistant,
		Parts: []llm.Part{llm.TextPart{Text: text}},
	}
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

func finishReasonPtr(r llm.FinishReason) *llm.FinishReason {
	return &r
}

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run(
		"simple completion",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Hello!"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithInstructions("You are helpful."),
				agent.WithModel("test-model"),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("Hi")},
			)

			require.NoError(t, err)
			assert.Equal(t, 1, result.Turns)
			assert.Equal(t, "Hello!", result.FinalMessage().Text())
			assert.Equal(t, 10, result.Usage.InputTokens)
			assert.Equal(t, 5, result.Usage.OutputTokens)
			assert.Equal(t, "assistant", result.LastAgent.Name())
		},
	)

	t.Run(
		"tool call",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				City string `json:"city"`
			}

		weatherTool, err := agent.FunctionTool[Params](
			"get_weather",
			"Get weather for a city",
			func(_ context.Context, p Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "Sunny, 22°C in " + p.City}, nil
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID: "tc_1",
						Function: llm.FunctionCall{
							Name:      "get_weather",
							Arguments: `{"city":"Paris"}`,
						},
					}),
					stopResponse("It's sunny and 22°C in Paris!"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(weatherTool),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("What's the weather in Paris?")},
			)

			require.NoError(t, err)
			assert.Equal(t, 2, result.Turns)
			assert.Equal(t, "It's sunny and 22°C in Paris!", result.FinalMessage().Text())
			assert.Equal(t, 20, result.Usage.InputTokens)
		},
	)

	t.Run(
		"max turns exceeded",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "noop", Arguments: `{}`},
					}),
					toolCallResponse(llm.ToolCall{
						ID:       "tc_2",
						Function: llm.FunctionCall{Name: "noop", Arguments: `{}`},
					}),
				},
			}

		type Params struct{}
		noopTool, err := agent.FunctionTool[Params](
			"noop",
			"No-op",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "ok"}, nil
			},
		)
		require.NoError(t, err)

		ag := agent.New(
			"assistant",
			newTestClient(provider),
			agent.WithModel("test-model"),
			agent.WithTools(noopTool),
			agent.WithMaxTurns(2),
			)

		_, err = ag.Run(
			context.Background(),
			[]llm.Message{userMessage("loop")},
		)

		require.Error(t, err)
		var maxTurnsErr *agent.MaxTurnsExceededError
			require.ErrorAs(t, err, &maxTurnsErr)
			assert.Equal(t, 2, maxTurnsErr.MaxTurns)
		},
	)

	t.Run(
		"duplicate tool names",
		func(t *testing.T) {
			t.Parallel()

		type Params struct{}
		makeTool := func(name string) agent.Tool {
			tool, err := agent.FunctionTool[Params](
				name,
				"desc",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "ok"}, nil
				},
			)
			require.NoError(t, err)
			return tool
		}

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("ok"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(makeTool("search"), makeTool("search")),
			)

			_, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "duplicate tool name")
		},
	)

	t.Run(
		"context cancellation",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("should not reach"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
			)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := ag.Run(ctx, []llm.Message{userMessage("test")})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot complete")
			assert.Equal(t, 0, provider.calls)
		},
	)

	t.Run(
		"finish reason length treated as stop",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					{
						Model: "test-model",
						Message: llm.Message{
							Role:  llm.RoleAssistant,
							Parts: []llm.Part{llm.TextPart{Text: "Truncated response"}},
						},
						Usage:        llm.Usage{InputTokens: 10, OutputTokens: 50},
						FinishReason: llm.FinishReasonLength,
					},
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("Write a long essay")},
			)

			require.NoError(t, err)
			assert.Equal(t, 1, result.Turns)
			assert.Equal(t, "Truncated response", result.FinalMessage().Text())
			assert.Equal(t, 50, result.Usage.OutputTokens)
		},
	)

	t.Run(
		"parallel tool execution preserves ordering",
		func(t *testing.T) {
			t.Parallel()

			type Params struct{}

		tool1, err := agent.FunctionTool[Params](
			"first",
			"First tool",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "result_1"}, nil
			},
		)
		require.NoError(t, err)
		tool2, err := agent.FunctionTool[Params](
			"second",
			"Second tool",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "result_2"}, nil
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(
						llm.ToolCall{
							ID:       "tc_1",
							Function: llm.FunctionCall{Name: "first", Arguments: `{}`},
						},
						llm.ToolCall{
							ID:       "tc_2",
							Function: llm.FunctionCall{Name: "second", Arguments: `{}`},
						},
					),
					stopResponse("Both done."),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(tool1, tool2),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("do both")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Both done.", result.FinalMessage().Text())

			var toolMsgs []llm.Message
			for _, m := range result.Messages {
				if m.Role == llm.RoleTool {
					toolMsgs = append(toolMsgs, m)
				}
			}
			require.Len(t, toolMsgs, 2)
			assert.Equal(t, "tc_1", toolMsgs[0].ToolCallID)
			assert.Equal(t, "result_1", toolMsgs[0].Text())
			assert.Equal(t, "tc_2", toolMsgs[1].ToolCallID)
			assert.Equal(t, "result_2", toolMsgs[1].Text())
		},
	)

	t.Run(
		"parallel tool execution with partial failure",
		func(t *testing.T) {
			t.Parallel()

			type Params struct{}

		successTool, err := agent.FunctionTool[Params](
			"succeed",
			"Always succeeds",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "success_result"}, nil
			},
		)
		require.NoError(t, err)
		failTool, err := agent.FunctionTool[Params](
			"fail",
			"Always fails",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{}, errors.New("tool exploded")
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(
						llm.ToolCall{
							ID:       "tc_ok",
							Function: llm.FunctionCall{Name: "succeed", Arguments: `{}`},
						},
						llm.ToolCall{
							ID:       "tc_fail",
							Function: llm.FunctionCall{Name: "fail", Arguments: `{}`},
						},
					),
					stopResponse("Handled both."),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(successTool, failTool),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("do both")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Handled both.", result.FinalMessage().Text())

			var toolMsgs []llm.Message
			for _, m := range result.Messages {
				if m.Role == llm.RoleTool {
					toolMsgs = append(toolMsgs, m)
				}
			}
			require.Len(t, toolMsgs, 2)
			assert.Equal(t, "tc_ok", toolMsgs[0].ToolCallID)
			assert.Equal(t, "success_result", toolMsgs[0].Text())
			assert.Equal(t, "tc_fail", toolMsgs[1].ToolCallID)
			assert.Contains(t, toolMsgs[1].Text(), "Error:")
		},
	)

	t.Run(
		"tool accesses run context during execution",
		func(t *testing.T) {
			t.Parallel()

			type RequestContext struct {
				TenantID string
			}

			var capturedTenantID string

		type Params struct{}
		tool, err := agent.FunctionTool[Params](
			"check_tenant",
			"Check current tenant",
			func(ctx context.Context, _ Params) (agent.ToolResult, error) {
				rc := agent.RunContextFrom[*RequestContext](ctx)
				capturedTenantID = rc.TenantID
				return agent.ToolResult{Content: "tenant: " + rc.TenantID}, nil
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "check_tenant", Arguments: `{}`},
					}),
					stopResponse("Done"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(tool),
			)

			ctx := agent.WithRunContext(
				context.Background(),
				&RequestContext{TenantID: "t_456"},
			)

			result, err := ag.Run(
				ctx,
				[]llm.Message{userMessage("Check my tenant")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Done", result.FinalMessage().Text())
			assert.Equal(t, "t_456", capturedTenantID)
		},
	)
}

func TestRun_Handoff(t *testing.T) {
	t.Parallel()

	t.Run(
		"basic handoff",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "transfer_to_billing", Arguments: `{}`},
					}),
					stopResponse("Your invoice is $42."),
				},
			}

			client := newTestClient(provider)

			billing := agent.New(
				"billing",
				client,
				agent.WithInstructions("You handle billing questions."),
				agent.WithModel("test-model"),
			)

			triage := agent.New(
				"triage",
				client,
				agent.WithInstructions("Route users to the right agent."),
				agent.WithModel("test-model"),
				agent.WithHandoffs(billing),
			)

			result, err := triage.Run(
				context.Background(),
				[]llm.Message{userMessage("How much is my invoice?")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Your invoice is $42.", result.FinalMessage().Text())
			assert.Equal(t, "billing", result.LastAgent.Name())
		},
	)

	t.Run(
		"custom tool name and description",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "ask_billing", Arguments: `{}`},
					}),
					stopResponse("Your invoice is $42."),
				},
			}

			client := newTestClient(provider)

			billing := agent.New(
				"billing",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffDescription("Handles all billing questions."),
			)

			triage := agent.New(
				"triage",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffConfigs(
					agent.HandoffTo(
						billing,
						agent.WithHandoffToolName("ask_billing"),
						agent.WithHandoffToolDescription("Route to billing"),
					),
				),
			)

			result, err := triage.Run(
				context.Background(),
				[]llm.Message{userMessage("How much is my invoice?")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Your invoice is $42.", result.FinalMessage().Text())
			assert.Equal(t, "billing", result.LastAgent.Name())
		},
	)

	t.Run(
		"on_handoff callback fires",
		func(t *testing.T) {
			t.Parallel()

			var callbackFired bool

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "transfer_to_billing", Arguments: `{}`},
					}),
					stopResponse("Done."),
				},
			}

			client := newTestClient(provider)

			billing := agent.New(
				"billing",
				client,
				agent.WithModel("test-model"),
			)

			triage := agent.New(
				"triage",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffConfigs(
					agent.HandoffTo(
						billing,
						agent.WithOnHandoff(func(_ context.Context) error {
							callbackFired = true
							return nil
						}),
					),
				),
			)

			_, err := triage.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)

			require.NoError(t, err)
			assert.True(t, callbackFired)
		},
	)

	t.Run(
		"input filter filters messages",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "transfer_to_specialist", Arguments: `{}`},
					}),
					stopResponse("Filtered."),
				},
			}

			client := newTestClient(provider)

			specialist := agent.New(
				"specialist",
				client,
				agent.WithModel("test-model"),
			)

			router := agent.New(
				"router",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffConfigs(
					agent.HandoffTo(
						specialist,
						agent.WithHandoffInputFilter(func(data agent.HandoffInputData) []llm.Message {
							var filtered []llm.Message
							for _, m := range data.NewItems {
								if m.Role == llm.RoleUser {
									filtered = append(filtered, m)
								}
							}
							return filtered
						}),
					),
				),
			)

			result, err := router.Run(
				context.Background(),
				[]llm.Message{userMessage("help me")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Filtered.", result.FinalMessage().Text())
		},
	)

	t.Run(
		"tool name sanitized from agent name",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "transfer_to_billing_support", Arguments: `{}`},
					}),
					stopResponse("Done."),
				},
			}

			client := newTestClient(provider)

			billing := agent.New(
				"billing support",
				client,
				agent.WithModel("test-model"),
			)

			triage := agent.New(
				"triage",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffs(billing),
			)

			result, err := triage.Run(
				context.Background(),
				[]llm.Message{userMessage("help")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Done.", result.FinalMessage().Text())
			assert.Equal(t, "billing support", result.LastAgent.Name())
		},
	)

	t.Run(
		"handoff description getter",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"billing",
				newTestClient(&mockProvider{}),
				agent.WithHandoffDescription("Handles billing and invoicing."),
			)

			assert.Equal(t, "Handles billing and invoicing.", ag.HandoffDescription())
		},
	)
}

func TestRun_Guardrails(t *testing.T) {
	t.Parallel()

	t.Run(
		"input guardrail trips",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("should not reach"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithInputGuardrails(&blockingGuardrail{keyword: "forbidden"}),
			)

			_, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("forbidden")},
			)

			require.Error(t, err)
			var tripErr *agent.InputGuardrailTrippedError
			require.ErrorAs(t, err, &tripErr)
			assert.Equal(t, "blocker", tripErr.Guardrail)
			assert.Equal(t, 0, provider.calls)
		},
	)

	t.Run(
		"output guardrail trips",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("bad response"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithOutputGuardrails(&outputBlocker{}),
			)

			_, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)

			require.Error(t, err)
			var tripErr *agent.OutputGuardrailTrippedError
			require.ErrorAs(t, err, &tripErr)
			assert.Equal(t, "output_blocker", tripErr.Guardrail)
		},
	)
}

func TestRun_Hooks(t *testing.T) {
	t.Parallel()

	t.Run(
		"run hooks fire during tool use",
		func(t *testing.T) {
			t.Parallel()

		type Params struct{}
		noopTool, err := agent.FunctionTool[Params](
			"noop",
			"No-op",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "ok"}, nil
			},
		)
		require.NoError(t, err)

		provider := &mockProvider{
			responses: []*llm.ChatCompletionResponse{
				toolCallResponse(llm.ToolCall{
					ID:       "tc_1",
					Function: llm.FunctionCall{Name: "noop", Arguments: `{}`},
				}),
				stopResponse("done"),
			},
		}

		hook := &recordingHook{}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(noopTool),
				agent.WithHooks(hook),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)

			require.NoError(t, err)
			assert.Equal(t, "done", result.FinalMessage().Text())
			assert.True(t, hook.runStarted)
			assert.True(t, hook.runEnded)
			assert.Equal(t, []string{"noop"}, hook.toolNames)
		},
	)

	t.Run(
		"agent hooks fire during run",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Hello!"),
				},
			}

			hook := &recordingAgentHook{}
			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithAgentHooks(hook),
			)

			_, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("Hi")},
			)

			require.NoError(t, err)
			assert.True(t, hook.started)
			assert.True(t, hook.ended)
		},
	)

	t.Run(
		"target agent OnHandoff fires on handoff",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "transfer_to_billing", Arguments: `{}`},
					}),
					stopResponse("Done."),
				},
			}

			client := newTestClient(provider)

			billingHook := &recordingAgentHook{}
			billing := agent.New(
				"billing",
				client,
				agent.WithModel("test-model"),
				agent.WithAgentHooks(billingHook),
			)

			triage := agent.New(
				"triage",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffs(billing),
			)

			_, err := triage.Run(
				context.Background(),
				[]llm.Message{userMessage("invoice?")},
			)

			require.NoError(t, err)
			assert.True(t, billingHook.handoffed)
			assert.True(t, billingHook.started)
		},
	)

	t.Run(
		"OnRunEnd fires even on session save failure",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Hello!"),
				},
			}

			hook := &recordingHook{}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithHooks(hook),
				agent.WithSession(&failingSession{}, "sess-1"),
			)

			_, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("Hi")},
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot save session")
			assert.True(t, hook.runStarted, "OnRunStart should have fired")
			assert.True(t, hook.runEnded, "OnRunEnd should fire even when session save fails")
		},
	)
}

func TestRun_Session(t *testing.T) {
	t.Parallel()

	t.Run(
		"round trip across two runs",
		func(t *testing.T) {
			t.Parallel()

			store := newTestSession()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Hi Alice!"),
					stopResponse("Your name is Alice."),
				},
			}

			client := newTestClient(provider)

			ag1 := agent.New(
				"assistant",
				client,
				agent.WithModel("test-model"),
				agent.WithInstructions("You are a helpful assistant."),
				agent.WithSession(store, "session-1"),
			)

			_, err := ag1.Run(
				context.Background(),
				[]llm.Message{userMessage("I'm Alice")},
			)
			require.NoError(t, err)

			ag2 := agent.New(
				"assistant",
				client,
				agent.WithModel("test-model"),
				agent.WithInstructions("You are a helpful assistant."),
				agent.WithSession(store, "session-1"),
			)

			result, err := ag2.Run(
				context.Background(),
				[]llm.Message{userMessage("What's my name?")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Your name is Alice.", result.FinalMessage().Text())
			assert.Len(t, result.Messages, 4)
		},
	)
}

func TestRun_DynamicInstructions(t *testing.T) {
	t.Parallel()

	t.Run(
		"instructions from function with run context",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Hello, Alice!"),
				},
			}

			type userInfo struct {
				Name string
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithInstructionsFunc(func(ctx context.Context, _ *agent.Agent) string {
					info := agent.RunContextFrom[*userInfo](ctx)
					return "You are helping " + info.Name + ". Be concise."
				}),
			)

			ctx := agent.WithRunContext(
				context.Background(),
				&userInfo{Name: "Alice"},
			)
			result, err := ag.Run(
				ctx,
				[]llm.Message{userMessage("Hi")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Hello, Alice!", result.FinalMessage().Text())
		},
	)

	t.Run(
		"instructionsFunc overrides static instructions",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("dynamic"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithInstructions("static"),
				agent.WithInstructionsFunc(func(_ context.Context, _ *agent.Agent) string {
					return "dynamic"
				}),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)

			require.NoError(t, err)
			assert.Equal(t, "dynamic", result.FinalMessage().Text())
		},
	)
}

func TestRun_ToolUseBehavior(t *testing.T) {
	t.Parallel()

	t.Run(
		"stop on first tool",
		func(t *testing.T) {
			t.Parallel()

		type Params struct{}
		tool, err := agent.FunctionTool[Params](
			"compute",
			"Compute something",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "computed_result"}, nil
			},
		)
		require.NoError(t, err)

		provider := &mockProvider{
			responses: []*llm.ChatCompletionResponse{
				toolCallResponse(llm.ToolCall{
					ID:       "tc_1",
					Function: llm.FunctionCall{Name: "compute", Arguments: `{}`},
				}),
			},
		}

		ag := agent.New(
			"assistant",
			newTestClient(provider),
			agent.WithModel("test-model"),
			agent.WithTools(tool),
			agent.WithToolUseBehavior(agent.StopOnFirstTool()),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("compute")},
			)

			require.NoError(t, err)
			assert.Equal(t, "computed_result", result.FinalMessage().Text())
			assert.Equal(t, 1, result.Turns)
			assert.Equal(t, 1, provider.calls)
		},
	)

	t.Run(
		"stop at specific tools",
		func(t *testing.T) {
			t.Parallel()

			type Params struct{}

		tool1, err := agent.FunctionTool[Params](
			"search",
			"Search",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "search_result"}, nil
			},
		)
		require.NoError(t, err)
		tool2, err := agent.FunctionTool[Params](
			"submit",
			"Submit",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "submitted"}, nil
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "submit", Arguments: `{}`},
					}),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(tool1, tool2),
				agent.WithToolUseBehavior(agent.StopAtTools("submit")),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("submit it")},
			)

			require.NoError(t, err)
			assert.Equal(t, "submitted", result.FinalMessage().Text())
		},
	)

	t.Run(
		"default run_llm_again continues loop",
		func(t *testing.T) {
			t.Parallel()

		type Params struct{}
		tool, err := agent.FunctionTool[Params](
			"noop",
			"No-op",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "ok"}, nil
			},
		)
		require.NoError(t, err)

		provider := &mockProvider{
			responses: []*llm.ChatCompletionResponse{
				toolCallResponse(llm.ToolCall{
					ID:       "tc_1",
					Function: llm.FunctionCall{Name: "noop", Arguments: `{}`},
				}),
				stopResponse("Final answer."),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(tool),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Final answer.", result.FinalMessage().Text())
			assert.Equal(t, 2, result.Turns)
		},
	)

	t.Run(
		"custom behavior error propagation",
		func(t *testing.T) {
			t.Parallel()

		type Params struct{}
		tool, err := agent.FunctionTool[Params](
			"compute",
			"Compute something",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "result"}, nil
			},
		)
		require.NoError(t, err)

		provider := &mockProvider{
			responses: []*llm.ChatCompletionResponse{
				toolCallResponse(llm.ToolCall{
					ID:       "tc_1",
					Function: llm.FunctionCall{Name: "compute", Arguments: `{}`},
				}),
			},
		}

		ag := agent.New(
			"assistant",
			newTestClient(provider),
			agent.WithModel("test-model"),
			agent.WithTools(tool),
			agent.WithToolUseBehavior(agent.ToolUseBehavior(func(_ context.Context, _ []agent.ToolCallResult) (string, bool, error) {
					return "", false, errors.New("custom behavior failed")
				})),
			)

		_, err = ag.Run(
			context.Background(),
			[]llm.Message{userMessage("compute")},
		)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "custom behavior failed")
		},
	)
}

func TestRun_OutputType(t *testing.T) {
	t.Parallel()

	type Info struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	}

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			stopResponse(`{"name":"Probo","country":"FR"}`),
		},
	}

	infoType, err := agent.NewOutputType[Info]("info")
	require.NoError(t, err)

	ag := agent.New(
		"assistant",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithOutputType(infoType),
	)

	result, err := ag.Run(
		context.Background(),
		[]llm.Message{userMessage("Tell me about Probo")},
	)

	require.NoError(t, err)
	assert.Contains(t, result.FinalMessage().Text(), "Probo")
}

func TestRun_Approval(t *testing.T) {
	t.Parallel()

	t.Run(
		"tool requiring approval interrupts the run",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc1",
						Function: llm.FunctionCall{Name: "delete_account", Arguments: `{}`},
					}),
				},
			}

		deleteTool, err := agent.FunctionTool[struct{}](
			"delete_account",
			"Deletes the user account",
			func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "deleted"}, nil
			},
		)
		require.NoError(t, err)

		ag := agent.New(
			"assistant",
			newTestClient(provider),
			agent.WithModel("test-model"),
			agent.WithTools(deleteTool),
			agent.WithApproval(agent.ApprovalConfig{
				ToolNames: []string{"delete_account"},
			}),
		)

		_, err = ag.Run(
			context.Background(),
			[]llm.Message{userMessage("Delete my account")},
		)

		require.Error(t, err)
		var interrupted *agent.InterruptedError
		require.ErrorAs(t, err, &interrupted)
		assert.Len(t, interrupted.ToolCalls, 1)
		assert.Equal(t, "delete_account", interrupted.ToolCalls[0].Function.Name)
		assert.Len(t, interrupted.PendingApprovals, 1)
		assert.Equal(t, "delete_account", interrupted.PendingApprovals[0].Function.Name)
		assert.Equal(t, 1, interrupted.Turns)
		},
	)

	t.Run(
		"resume with approval executes the tool",
		func(t *testing.T) {
			t.Parallel()

		var toolExecuted bool

		deleteTool, err := agent.FunctionTool[struct{}](
			"delete_account",
			"Deletes the user account",
			func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
				toolExecuted = true
				return agent.ToolResult{Content: "account deleted"}, nil
			},
		)
		require.NoError(t, err)

		provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc1",
						Function: llm.FunctionCall{Name: "delete_account", Arguments: `{}`},
					}),
					stopResponse("Your account has been deleted."),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(deleteTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"delete_account"},
				}),
			)

		_, err = ag.Run(
			context.Background(),
			[]llm.Message{userMessage("Delete my account")},
		)

		var interrupted *agent.InterruptedError
		require.ErrorAs(t, err, &interrupted)
		assert.False(t, toolExecuted)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"tc1": {Approved: true},
					},
				},
			)

			require.NoError(t, err)
			assert.True(t, toolExecuted)
			assert.Equal(t, "Your account has been deleted.", result.FinalMessage().Text())
		},
	)

	t.Run(
		"resume with rejection denies the tool",
		func(t *testing.T) {
			t.Parallel()

		deleteTool, err := agent.FunctionTool[struct{}](
			"delete_account",
			"Deletes the user account",
			func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
				t.Fatal("tool should not be executed")
				return agent.ToolResult{}, nil
			},
		)
		require.NoError(t, err)

		provider := &mockProvider{
			responses: []*llm.ChatCompletionResponse{
				toolCallResponse(llm.ToolCall{
					ID:       "tc1",
					Function: llm.FunctionCall{Name: "delete_account", Arguments: `{}`},
				}),
				stopResponse("OK, I won't delete your account."),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(deleteTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"delete_account"},
				}),
			)

			_, err = ag.Run(
				context.Background(),
				[]llm.Message{userMessage("Delete my account")},
			)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"tc1": {Approved: false, Message: "User cancelled the operation."},
					},
				},
			)

			require.NoError(t, err)
			assert.Equal(t, "OK, I won't delete your account.", result.FinalMessage().Text())
		},
	)

	t.Run(
		"ShouldApprove function takes priority",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc1",
						Function: llm.FunctionCall{Name: "safe_tool", Arguments: `{}`},
					}),
					stopResponse("Done"),
				},
			}

		safeTool, err := agent.FunctionTool[struct{}](
			"safe_tool",
			"A safe tool",
			func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "safe result"}, nil
			},
		)
		require.NoError(t, err)

		ag := agent.New(
			"assistant",
			newTestClient(provider),
			agent.WithModel("test-model"),
			agent.WithTools(safeTool),
			agent.WithApproval(agent.ApprovalConfig{
				ShouldApprove: func(_ context.Context, tc llm.ToolCall) bool {
						return tc.Function.Name == "dangerous_tool"
					},
				}),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("Do safe thing")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Done", result.FinalMessage().Text())
		},
	)

	t.Run(
		"multi-tool batch with partial approval",
		func(t *testing.T) {
			t.Parallel()

			var safeExecuted, dangerExecuted bool

		safeTool, err := agent.FunctionTool[struct{}](
			"safe_action",
			"A safe action",
			func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
				safeExecuted = true
				return agent.ToolResult{Content: "safe done"}, nil
			},
		)
		require.NoError(t, err)

		dangerTool, err := agent.FunctionTool[struct{}](
			"danger_action",
			"A dangerous action",
			func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
				dangerExecuted = true
				return agent.ToolResult{Content: "danger done"}, nil
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(
						llm.ToolCall{
							ID:       "tc_safe",
							Function: llm.FunctionCall{Name: "safe_action", Arguments: `{}`},
						},
						llm.ToolCall{
							ID:       "tc_danger",
							Function: llm.FunctionCall{Name: "danger_action", Arguments: `{}`},
						},
					),
					stopResponse("Both actions completed."),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(safeTool, dangerTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"danger_action"},
				}),
			)

			_, err = ag.Run(
				context.Background(),
				[]llm.Message{userMessage("Do both")},
			)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)
			assert.Len(t, interrupted.ToolCalls, 2)
			assert.Len(t, interrupted.PendingApprovals, 1)
			assert.Equal(t, "danger_action", interrupted.PendingApprovals[0].Function.Name)
			assert.False(t, safeExecuted)
			assert.False(t, dangerExecuted)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"tc_danger": {Approved: true},
					},
				},
			)

			require.NoError(t, err)
			assert.True(t, safeExecuted)
			assert.True(t, dangerExecuted)
			assert.Equal(t, "Both actions completed.", result.FinalMessage().Text())
		},
	)
}

func TestResume(t *testing.T) {
	t.Parallel()

	t.Run(
		"handoff tool call with approval",
		func(t *testing.T) {
			t.Parallel()

			var handoffCallbackFired bool

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc1",
						Function: llm.FunctionCall{Name: "transfer_to_specialist", Arguments: `{}`},
					}),
					stopResponse("Specialist handled it."),
				},
			}

			client := newTestClient(provider)

			specialist := agent.New(
				"specialist",
				client,
				agent.WithModel("test-model"),
			)

			triage := agent.New(
				"triage",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffConfigs(
					agent.HandoffTo(
						specialist,
						agent.WithOnHandoff(func(_ context.Context) error {
							handoffCallbackFired = true
							return nil
						}),
					),
				),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"transfer_to_specialist"},
				}),
			)

			_, err := triage.Run(
				context.Background(),
				[]llm.Message{userMessage("Help me")},
			)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)
			assert.False(t, handoffCallbackFired)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"tc1": {Approved: true},
					},
				},
			)

			require.NoError(t, err)
			assert.True(t, handoffCallbackFired)
			assert.Equal(t, "Specialist handled it.", result.FinalMessage().Text())
			assert.Equal(t, "specialist", result.LastAgent.Name())
		},
	)

	t.Run(
		"carries forward usage and turns",
		func(t *testing.T) {
			t.Parallel()

		deleteTool, err := agent.FunctionTool[struct{}](
			"delete_account",
			"Deletes the user account",
			func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "deleted"}, nil
			},
		)
		require.NoError(t, err)

		provider := &mockProvider{
			responses: []*llm.ChatCompletionResponse{
				toolCallResponse(llm.ToolCall{
					ID:       "tc1",
					Function: llm.FunctionCall{Name: "delete_account", Arguments: `{}`},
				}),
				stopResponse("Account deleted."),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(deleteTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"delete_account"},
				}),
			)

			_, err = ag.Run(
				context.Background(),
				[]llm.Message{userMessage("Delete my account")},
			)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)
			assert.Equal(t, 1, interrupted.Turns)
			assert.Equal(t, 10, interrupted.Usage.InputTokens)
			assert.Equal(t, 5, interrupted.Usage.OutputTokens)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"tc1": {Approved: true},
					},
				},
			)

			require.NoError(t, err)
			assert.Equal(t, 2, result.Turns, "turns should include the interrupted turn")
			assert.Equal(t, 20, result.Usage.InputTokens, "usage should include pre-interruption tokens")
			assert.Equal(t, 10, result.Usage.OutputTokens, "usage should include pre-interruption tokens")
		},
	)

	t.Run(
		"rejected handoff stays with current agent",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc1",
						Function: llm.FunctionCall{Name: "transfer_to_specialist", Arguments: `{}`},
					}),
					stopResponse("OK, I will handle it myself."),
				},
			}

			client := newTestClient(provider)

			specialist := agent.New(
				"specialist",
				client,
				agent.WithModel("test-model"),
			)

			triage := agent.New(
				"triage",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffs(specialist),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"transfer_to_specialist"},
				}),
			)

			_, err := triage.Run(
				context.Background(),
				[]llm.Message{userMessage("Help me")},
			)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"tc1": {Approved: false, Message: "User declined the transfer."},
					},
				},
			)

			require.NoError(t, err)
			assert.Equal(t, "OK, I will handle it myself.", result.FinalMessage().Text())
			assert.Equal(t, "triage", result.LastAgent.Name())
		},
	)
}

func TestRunTyped(t *testing.T) {
	t.Parallel()

	type Info struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	}

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			stopResponse(`{"name":"Stripe","country":"US"}`),
		},
	}

	ag := agent.New(
		"assistant",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithInstructions("Return structured info."),
	)

	result, err := agent.RunTyped[Info](
		context.Background(),
		ag,
		[]llm.Message{userMessage("Tell me about Stripe")},
	)

	require.NoError(t, err)
	assert.Equal(t, "Stripe", result.Output.Name)
	assert.Equal(t, "US", result.Output.Country)
}

func TestRunStreamed(t *testing.T) {
	t.Parallel()

	t.Run(
		"streams delta events and completes",
		func(t *testing.T) {
			t.Parallel()

			mockStream := &mockChatStream{
				events: []llm.ChatCompletionStreamEvent{
					{Delta: llm.MessageDelta{Content: "Hello"}},
					{Delta: llm.MessageDelta{Content: " world"}},
					{
						Delta:        llm.MessageDelta{Content: "!"},
						Usage:        &llm.Usage{InputTokens: 10, OutputTokens: 3},
						FinishReason: finishReasonPtr(llm.FinishReasonStop),
					},
				},
			}

			streamProvider := &mockStreamProvider{stream: mockStream}
			client := llm.NewClient(streamProvider, "test")

			ag := agent.New(
				"assistant",
				client,
				agent.WithModel("test-model"),
				agent.WithInstructions("Be brief."),
			)

			sr := ag.RunStreamed(
				context.Background(),
				[]llm.Message{userMessage("Hi")},
			)

			var deltas []string
			var gotComplete bool

			for ev := range sr.Events {
				switch ev.Type {
				case agent.StreamEventLLMDelta:
					deltas = append(deltas, ev.Delta)
				case agent.StreamEventComplete:
					gotComplete = true
				}
			}

			result, err := sr.Wait()
			require.NoError(t, err)
			assert.True(t, gotComplete)
			assert.Equal(t, []string{"Hello", " world", "!"}, deltas)
			assert.Equal(t, 1, result.Turns)
			assert.Equal(t, "Hello world!", result.FinalMessage().Text())
		},
	)

	t.Run(
		"with tool calls",
		func(t *testing.T) {
			t.Parallel()

		type Params struct{}
		tool, err := agent.FunctionTool[Params](
			"noop",
			"No-op",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "ok"}, nil
			},
		)
		require.NoError(t, err)

		stream1 := &mockChatStream{
				events: []llm.ChatCompletionStreamEvent{
					{
						Delta: llm.MessageDelta{
							ToolCalls: []llm.ToolCallDelta{
								{Index: 0, ID: "tc_1", Name: "noop", Arguments: `{}`},
							},
						},
						Usage:        &llm.Usage{InputTokens: 10, OutputTokens: 5},
						FinishReason: finishReasonPtr(llm.FinishReasonToolCalls),
					},
				},
			}

			stream2 := &mockChatStream{
				events: []llm.ChatCompletionStreamEvent{
					{Delta: llm.MessageDelta{Content: "Done!"}},
					{
						Delta:        llm.MessageDelta{},
						Usage:        &llm.Usage{InputTokens: 15, OutputTokens: 3},
						FinishReason: finishReasonPtr(llm.FinishReasonStop),
					},
				},
			}

			streamProvider := &mockMultiStreamProvider{
				streams: []llm.ChatCompletionStream{stream1, stream2},
			}

			client := llm.NewClient(streamProvider, "test")

			ag := agent.New(
				"assistant",
				client,
				agent.WithModel("test-model"),
				agent.WithTools(tool),
			)

			sr := ag.RunStreamed(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)

			var gotToolStart, gotToolEnd, gotComplete bool
			for ev := range sr.Events {
				switch ev.Type {
				case agent.StreamEventToolStart:
					gotToolStart = true
				case agent.StreamEventToolEnd:
					gotToolEnd = true
				case agent.StreamEventComplete:
					gotComplete = true
				}
			}

			result, err := sr.Wait()
			require.NoError(t, err)
			assert.True(t, gotToolStart)
			assert.True(t, gotToolEnd)
			assert.True(t, gotComplete)
			assert.Equal(t, 2, result.Turns)
			assert.Equal(t, "Done!", result.FinalMessage().Text())
		},
	)

	t.Run(
		"session save failure emits error not complete",
		func(t *testing.T) {
			t.Parallel()

			mockStream := &mockChatStream{
				events: []llm.ChatCompletionStreamEvent{
					{Delta: llm.MessageDelta{Content: "Hello"}},
					{
						Delta:        llm.MessageDelta{Content: "!"},
						Usage:        &llm.Usage{InputTokens: 10, OutputTokens: 2},
						FinishReason: finishReasonPtr(llm.FinishReasonStop),
					},
				},
			}

			streamProvider := &mockStreamProvider{stream: mockStream}

			ag := agent.New(
				"assistant",
				llm.NewClient(streamProvider, "test"),
				agent.WithModel("test-model"),
				agent.WithSession(&failingSession{}, "sess-1"),
			)

			sr := ag.RunStreamed(
				context.Background(),
				[]llm.Message{userMessage("Hi")},
			)

			var gotComplete, gotError bool
			for ev := range sr.Events {
				switch ev.Type {
				case agent.StreamEventComplete:
					gotComplete = true
				case agent.StreamEventError:
					gotError = true
				}
			}

			_, err := sr.Wait()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot save session")
			assert.False(t, gotComplete, "StreamEventComplete should not be emitted when session save fails")
			assert.True(t, gotError, "StreamEventError should be emitted when session save fails")
		},
	)

	t.Run(
		"concurrent consumer receives all events before Wait returns",
		func(t *testing.T) {
			t.Parallel()

			mockStream := &mockChatStream{
				events: []llm.ChatCompletionStreamEvent{
					{Delta: llm.MessageDelta{Content: "Hello"}},
					{Delta: llm.MessageDelta{Content: " world"}},
					{
						Delta:        llm.MessageDelta{Content: "!"},
						Usage:        &llm.Usage{InputTokens: 10, OutputTokens: 3},
						FinishReason: finishReasonPtr(llm.FinishReasonStop),
					},
				},
			}

			streamProvider := &mockStreamProvider{stream: mockStream}
			client := llm.NewClient(streamProvider, "test")

			ag := agent.New(
				"assistant",
				client,
				agent.WithModel("test-model"),
			)

			sr := ag.RunStreamed(
				context.Background(),
				[]llm.Message{userMessage("Hi")},
			)

			var collected []agent.StreamEvent
			done := make(chan struct{})
			go func() {
				defer close(done)
				for ev := range sr.Events {
					collected = append(collected, ev)
				}
			}()

			result, err := sr.Wait()
			<-done

			require.NoError(t, err)
			assert.Equal(t, "Hello world!", result.FinalMessage().Text())

			var deltaCount int
			var gotAgentStart, gotAgentEnd, gotComplete bool
			for _, ev := range collected {
				switch ev.Type {
				case agent.StreamEventLLMDelta:
					deltaCount++
				case agent.StreamEventAgentStart:
					gotAgentStart = true
				case agent.StreamEventAgentEnd:
					gotAgentEnd = true
				case agent.StreamEventComplete:
					gotComplete = true
				}
			}

			assert.Equal(t, 3, deltaCount, "all delta events should reach the consumer")
			assert.True(t, gotAgentStart, "agent_start event should reach the consumer")
			assert.True(t, gotAgentEnd, "agent_end event should reach the consumer")
			assert.True(t, gotComplete, "complete event should reach the consumer")
		},
	)
}

func TestClone(t *testing.T) {
	t.Parallel()

	t.Run(
		"preserves name and overrides instructions",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("ok"),
				},
			}

			original := agent.New(
				"original",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithInstructions("original instructions"),
			)

			cloned := original.Clone(
				agent.WithInstructions("cloned instructions"),
			)

			assert.Equal(t, "original", cloned.Name())
		},
	)

	t.Run(
		"does not mutate original tool list",
		func(t *testing.T) {
			t.Parallel()

		type Params struct{}
		tool1, err := agent.FunctionTool[Params](
			"t1",
			"desc",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "ok"}, nil
			},
		)
		require.NoError(t, err)
		tool2, err := agent.FunctionTool[Params](
			"t2",
			"desc",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "ok"}, nil
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("ok"),
					stopResponse("ok"),
				},
			}

			original := agent.New(
				"original",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(tool1),
			)

			cloned := original.Clone(agent.WithTools(tool2))

			originalResult, err := original.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)
			require.NoError(t, err)
			assert.Equal(t, "ok", originalResult.FinalMessage().Text())

			clonedResult, err := cloned.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)
			require.NoError(t, err)
			assert.Equal(t, "ok", clonedResult.FinalMessage().Text())
		},
	)
}

func TestWithMaxTurns(t *testing.T) {
	t.Parallel()

	t.Run(
		"zero clamps to minimum",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("ok"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithMaxTurns(0),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)

			require.NoError(t, err)
			assert.Equal(t, 1, result.Turns)
		},
	)

	t.Run(
		"negative clamps to minimum",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("ok"),
				},
			}

			ag := agent.New(
				"assistant",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithMaxTurns(-5),
			)

			result, err := ag.Run(
				context.Background(),
				[]llm.Message{userMessage("test")},
			)

			require.NoError(t, err)
			assert.Equal(t, 1, result.Turns)
		},
	)
}

func TestGenerateSchema_EmbeddedStruct(t *testing.T) {
	t.Parallel()

	type Base struct {
		ID   string `json:"id" jsonschema:"unique identifier"`
		Kind string `json:"kind"`
	}
	type Params struct {
		Base
		Name string `json:"name"`
	}

	tool, err := agent.FunctionTool[Params](
		"create",
		"Create item",
		func(_ context.Context, _ Params) (agent.ToolResult, error) {
			return agent.ToolResult{Content: "ok"}, nil
		},
	)
	require.NoError(t, err)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(tool.Definition().Parameters, &schema))
	assert.Equal(t, "object", schema["type"])

	props := schema["properties"].(map[string]any)
	assert.Contains(t, props, "id")
	assert.Contains(t, props, "kind")
	assert.Contains(t, props, "name")
	assert.NotContains(t, props, "Base")

	idProp := props["id"].(map[string]any)
	assert.Equal(t, "string", idProp["type"])
	assert.Equal(t, "unique identifier", idProp["description"])

	required := schema["required"].([]any)
	assert.Contains(t, required, "id")
	assert.Contains(t, required, "kind")
	assert.Contains(t, required, "name")
}

func TestRun_InputGuardrailError(t *testing.T) {
	t.Parallel()

	guardrail := &errorGuardrail{err: errors.New("check failed")}

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			stopResponse("should not reach"),
		},
	}

	ag := agent.New(
		"assistant",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithInputGuardrails(guardrail),
	)

	_, err := ag.Run(
		context.Background(),
		[]llm.Message{userMessage("test")},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot run input guardrail")
	assert.Contains(t, err.Error(), "check failed")
	assert.Equal(t, 0, provider.calls)
}

func TestRun_OutputGuardrailError(t *testing.T) {
	t.Parallel()

	guardrail := &errorOutputGuardrail{err: errors.New("output check failed")}

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			stopResponse("hello"),
		},
	}

	ag := agent.New(
		"assistant",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithOutputGuardrails(guardrail),
	)

	_, err := ag.Run(
		context.Background(),
		[]llm.Message{userMessage("test")},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot run output guardrail")
	assert.Contains(t, err.Error(), "output check failed")
}

func TestRun_HandoffCallbackError(t *testing.T) {
	t.Parallel()

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc_1",
				Function: llm.FunctionCall{Name: "transfer_to_billing", Arguments: `{}`},
			}),
		},
	}

	client := newTestClient(provider)

	billing := agent.New(
		"billing",
		client,
		agent.WithModel("test-model"),
	)

	triage := agent.New(
		"triage",
		client,
		agent.WithModel("test-model"),
		agent.WithHandoffConfigs(
			agent.HandoffTo(
				billing,
				agent.WithOnHandoff(func(_ context.Context) error {
					return errors.New("handoff setup failed")
				}),
			),
		),
	)

	_, err := triage.Run(
		context.Background(),
		[]llm.Message{userMessage("help")},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot execute handoff callback")
	assert.Contains(t, err.Error(), "handoff setup failed")
}

func TestRun_ContentFilterFinishReason(t *testing.T) {
	t.Parallel()

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			{
				Model: "test-model",
				Message: llm.Message{
					Role:  llm.RoleAssistant,
					Parts: []llm.Part{llm.TextPart{Text: ""}},
				},
				Usage:        llm.Usage{InputTokens: 10, OutputTokens: 0},
				FinishReason: llm.FinishReasonContentFilter,
			},
		},
	}

	ag := agent.New(
		"assistant",
		newTestClient(provider),
		agent.WithModel("test-model"),
	)

	_, err := ag.Run(
		context.Background(),
		[]llm.Message{userMessage("bad prompt")},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "content was filtered")
}

func TestRun_UnknownToolCall(t *testing.T) {
	t.Parallel()

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc_1",
				Function: llm.FunctionCall{Name: "nonexistent_tool", Arguments: `{}`},
			}),
		},
	}

	type Params struct{}
	tool, err := agent.FunctionTool[Params](
		"real_tool",
		"A real tool",
		func(_ context.Context, _ Params) (agent.ToolResult, error) {
			return agent.ToolResult{Content: "ok"}, nil
		},
	)
	require.NoError(t, err)

	ag := agent.New(
		"assistant",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(tool),
	)

	_, err = ag.Run(
		context.Background(),
		[]llm.Message{userMessage("test")},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool")
	assert.Contains(t, err.Error(), "nonexistent_tool")
}

func TestRun_SessionLoadFailure(t *testing.T) {
	t.Parallel()

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			stopResponse("should not reach"),
		},
	}

	ag := agent.New(
		"assistant",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithSession(&errorLoadSession{}, "sess-1"),
	)

	_, err := ag.Run(
		context.Background(),
		[]llm.Message{userMessage("test")},
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot load session")
	assert.Equal(t, 0, provider.calls)
}

func TestClone_WithApprovalConfig(t *testing.T) {
	t.Parallel()

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc1",
				Function: llm.FunctionCall{Name: "delete", Arguments: `{}`},
			}),
			toolCallResponse(llm.ToolCall{
				ID:       "tc2",
				Function: llm.FunctionCall{Name: "delete", Arguments: `{}`},
			}),
		},
	}

	deleteTool, err := agent.FunctionTool[struct{}](
		"delete",
		"Delete something",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			return agent.ToolResult{Content: "deleted"}, nil
		},
	)
	require.NoError(t, err)

	original := agent.New(
		"assistant",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(deleteTool),
		agent.WithApproval(agent.ApprovalConfig{
			ToolNames: []string{"delete"},
		}),
	)

	cloned := original.Clone()

	_, err = cloned.Run(
		context.Background(),
		[]llm.Message{userMessage("delete it")},
	)

	require.Error(t, err)
	var interrupted *agent.InterruptedError
	require.ErrorAs(t, err, &interrupted)
	assert.Len(t, interrupted.PendingApprovals, 1)

	_, err = original.Run(
		context.Background(),
		[]llm.Message{userMessage("delete it")},
	)

	require.Error(t, err)
	require.ErrorAs(t, err, &interrupted)
	assert.Len(t, interrupted.PendingApprovals, 1)
}

func TestRun_HandoffWithPreHandoffTools(t *testing.T) {
	t.Parallel()

	t.Run(
		"tools before handoff are executed in order",
		func(t *testing.T) {
			t.Parallel()

			var executionOrder []string

		type Params struct{}
		tool1, err := agent.FunctionTool[Params](
			"prepare",
			"Prepare data",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				executionOrder = append(executionOrder, "prepare")
				return agent.ToolResult{Content: "prepared"}, nil
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(
						llm.ToolCall{
							ID:       "tc_1",
							Function: llm.FunctionCall{Name: "prepare", Arguments: `{}`},
						},
						llm.ToolCall{
							ID:       "tc_2",
							Function: llm.FunctionCall{Name: "transfer_to_specialist", Arguments: `{}`},
						},
					),
					stopResponse("Specialist here."),
				},
			}

			client := newTestClient(provider)

			specialist := agent.New(
				"specialist",
				client,
				agent.WithModel("test-model"),
			)

			router := agent.New(
				"router",
				client,
				agent.WithModel("test-model"),
				agent.WithTools(tool1),
				agent.WithHandoffs(specialist),
			)

			result, err := router.Run(
				context.Background(),
				[]llm.Message{userMessage("prepare and transfer")},
			)

			require.NoError(t, err)
			assert.Equal(t, []string{"prepare"}, executionOrder)
			assert.Equal(t, "Specialist here.", result.FinalMessage().Text())
			assert.Equal(t, "specialist", result.LastAgent.Name())
		},
	)

	t.Run(
		"tools after handoff get tool-result messages",
		func(t *testing.T) {
			t.Parallel()

		type Params struct{}
		tool1, err := agent.FunctionTool[Params](
			"prepare",
			"Prepare data",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{Content: "prepared"}, nil
			},
		)
		require.NoError(t, err)
		tool2, err := agent.FunctionTool[Params](
			"finalize",
			"Finalize data",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				t.Fatal("tool after handoff should not be executed")
				return agent.ToolResult{}, nil
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(
						llm.ToolCall{
							ID:       "tc_1",
							Function: llm.FunctionCall{Name: "prepare", Arguments: `{}`},
						},
						llm.ToolCall{
							ID:       "tc_2",
							Function: llm.FunctionCall{Name: "transfer_to_specialist", Arguments: `{}`},
						},
						llm.ToolCall{
							ID:       "tc_3",
							Function: llm.FunctionCall{Name: "finalize", Arguments: `{}`},
						},
					),
					stopResponse("Specialist here."),
				},
			}

			client := newTestClient(provider)

			specialist := agent.New(
				"specialist",
				client,
				agent.WithModel("test-model"),
			)

			router := agent.New(
				"router",
				client,
				agent.WithModel("test-model"),
				agent.WithTools(tool1, tool2),
				agent.WithHandoffs(specialist),
			)

			result, err := router.Run(
				context.Background(),
				[]llm.Message{userMessage("prepare, transfer, and finalize")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Specialist here.", result.FinalMessage().Text())
			assert.Equal(t, "specialist", result.LastAgent.Name())

			var toolMsgs []llm.Message
			for _, m := range result.Messages {
				if m.Role == llm.RoleTool {
					toolMsgs = append(toolMsgs, m)
				}
			}

			require.Len(t, toolMsgs, 3)
			assert.Equal(t, "tc_1", toolMsgs[0].ToolCallID)
			assert.Equal(t, "prepared", toolMsgs[0].Text())
			assert.Equal(t, "tc_2", toolMsgs[1].ToolCallID)
			assert.Contains(t, toolMsgs[1].Text(), "Transferred to specialist")
			assert.Equal(t, "tc_3", toolMsgs[2].ToolCallID)
			assert.Contains(t, toolMsgs[2].Text(), "not executed")
		},
	)

	t.Run(
		"pre-handoff tool error aborts handoff",
		func(t *testing.T) {
			t.Parallel()

		type Params struct{}
		failingTool, err := agent.FunctionTool[Params](
			"prepare",
			"Prepare data",
			func(_ context.Context, _ Params) (agent.ToolResult, error) {
				return agent.ToolResult{}, errors.New("preparation failed")
			},
		)
		require.NoError(t, err)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(
						llm.ToolCall{
							ID:       "tc_1",
							Function: llm.FunctionCall{Name: "prepare", Arguments: `{}`},
						},
						llm.ToolCall{
							ID:       "tc_2",
							Function: llm.FunctionCall{Name: "transfer_to_specialist", Arguments: `{}`},
						},
					),
				},
			}

			client := newTestClient(provider)

			specialist := agent.New(
				"specialist",
				client,
				agent.WithModel("test-model"),
			)

			router := agent.New(
				"router",
				client,
				agent.WithModel("test-model"),
				agent.WithTools(failingTool),
				agent.WithHandoffs(specialist),
			)

			_, err = router.Run(
				context.Background(),
				[]llm.Message{userMessage("prepare and transfer")},
			)

			require.Error(t, err)
			assert.Contains(t, err.Error(), "cannot execute tool")
			assert.Contains(t, err.Error(), "preparation failed")
		},
	)
}

func TestRunStreamed_StreamError(t *testing.T) {
	t.Parallel()

	errStream := &errorChatStream{err: errors.New("stream broke")}
	streamProvider := &mockStreamProvider{stream: errStream}

	ag := agent.New(
		"assistant",
		llm.NewClient(streamProvider, "test"),
		agent.WithModel("test-model"),
	)

	sr := ag.RunStreamed(
		context.Background(),
		[]llm.Message{userMessage("test")},
	)

	for range sr.Events {
	}

	_, err := sr.Wait()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stream broke")
}

func TestRunStreamed_StreamProviderError(t *testing.T) {
	t.Parallel()

	errProvider := &errorStreamProvider{err: errors.New("connection refused")}

	ag := agent.New(
		"assistant",
		llm.NewClient(errProvider, "test"),
		agent.WithModel("test-model"),
	)

	sr := ag.RunStreamed(
		context.Background(),
		[]llm.Message{userMessage("test")},
	)

	for range sr.Events {
	}

	_, err := sr.Wait()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestResume_HandoffWithInputFilter(t *testing.T) {
	t.Parallel()

	var receivedMessages int

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(llm.ToolCall{
				ID:       "tc1",
				Function: llm.FunctionCall{Name: "transfer_to_specialist", Arguments: `{}`},
			}),
			stopResponse("Filtered specialist here."),
		},
	}

	client := newTestClient(provider)

	specialist := agent.New(
		"specialist",
		client,
		agent.WithModel("test-model"),
	)

	triage := agent.New(
		"triage",
		client,
		agent.WithModel("test-model"),
		agent.WithHandoffConfigs(
			agent.HandoffTo(
				specialist,
				agent.WithHandoffInputFilter(func(data agent.HandoffInputData) []llm.Message {
					receivedMessages = len(data.InputHistory) + len(data.NewItems)
					return data.NewItems
				}),
			),
		),
		agent.WithApproval(agent.ApprovalConfig{
			ToolNames: []string{"transfer_to_specialist"},
		}),
	)

	_, err := triage.Run(
		context.Background(),
		[]llm.Message{userMessage("Help me")},
	)

	var interrupted *agent.InterruptedError
	require.ErrorAs(t, err, &interrupted)

	result, err := agent.Resume(
		context.Background(),
		interrupted,
		agent.ResumeInput{
			Approvals: map[string]agent.ApprovalResult{
				"tc1": {Approved: true},
			},
		},
	)

	require.NoError(t, err)
	assert.Greater(t, receivedMessages, 0)
	assert.Equal(t, "Filtered specialist here.", result.FinalMessage().Text())
	assert.Equal(t, "specialist", result.LastAgent.Name())
}

func TestRun_NilHandoffsAreSkipped(t *testing.T) {
	t.Parallel()

	t.Run(
		"nil agent in WithHandoffs",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Hello."),
				},
			}

			client := newTestClient(provider)

			billing := agent.New(
				"billing",
				client,
				agent.WithModel("test-model"),
			)

			a := agent.New(
				"triage",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffs(nil, billing, nil),
			)

			result, err := a.Run(
				context.Background(),
				[]llm.Message{userMessage("hi")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Hello.", result.FinalMessage().Text())
		},
	)

	t.Run(
		"nil handoff in WithHandoffConfigs",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Hello."),
				},
			}

			client := newTestClient(provider)

			billing := agent.New(
				"billing",
				client,
				agent.WithModel("test-model"),
			)

			a := agent.New(
				"triage",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffConfigs(
					nil,
					agent.HandoffTo(billing),
					nil,
				),
			)

			result, err := a.Run(
				context.Background(),
				[]llm.Message{userMessage("hi")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Hello.", result.FinalMessage().Text())
		},
	)

	t.Run(
		"HandoffTo with nil agent in WithHandoffConfigs",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("Hello."),
				},
			}

			client := newTestClient(provider)

			a := agent.New(
				"triage",
				client,
				agent.WithModel("test-model"),
				agent.WithHandoffConfigs(agent.HandoffTo(nil)),
			)

			result, err := a.Run(
				context.Background(),
				[]llm.Message{userMessage("hi")},
			)

			require.NoError(t, err)
			assert.Equal(t, "Hello.", result.FinalMessage().Text())
		},
	)
}
