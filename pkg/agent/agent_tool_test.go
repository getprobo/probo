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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

func TestAgentTool_Name(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns custom tool name not agent name",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"geography",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("geo_expert", "Ask geography questions")
			assert.Equal(t, "geo_expert", tool.Name())
		},
	)

	t.Run(
		"different AsTool calls return different names",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"helper",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			tool1 := ag.AsTool("tool_a", "First tool")
			tool2 := ag.AsTool("tool_b", "Second tool")

			assert.Equal(t, "tool_a", tool1.Name())
			assert.Equal(t, "tool_b", tool2.Name())
		},
	)
}

func TestAgentTool_Definition(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns name and description",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"sub",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("my_tool", "Does something useful.")
			def := tool.Definition()

			assert.Equal(t, "my_tool", def.Name)
			assert.Equal(t, "Does something useful.", def.Description)
		},
	)

	t.Run(
		"schema contains input string property",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"sub",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("delegate", "Delegate work.")
			def := tool.Definition()

			require.NotNil(t, def.Parameters)

			var schema map[string]any
			require.NoError(t, json.Unmarshal(def.Parameters, &schema))

			assert.Equal(t, "object", schema["type"])

			props, ok := schema["properties"].(map[string]any)
			require.True(t, ok)
			assert.Contains(t, props, "input")

			inputProp := props["input"].(map[string]any)
			assert.Equal(t, "string", inputProp["type"])
		},
	)

	t.Run(
		"schema requires input field",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"sub",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("ask", "Ask a question.")
			def := tool.Definition()

			var schema map[string]any
			require.NoError(t, json.Unmarshal(def.Parameters, &schema))

			required, ok := schema["required"].([]any)
			require.True(t, ok)
			assert.Contains(t, required, "input")
		},
	)
}

func TestAgentTool_Execute(t *testing.T) {
	t.Parallel()

	t.Run(
		"runs sub-agent and returns final message",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("The capital of France is Paris."),
				},
			}

			ag := agent.New(
				"geography",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithInstructions("You are a geography expert."),
			)

			tool := ag.AsTool("geo_expert", "Ask geography questions.")
			result, err := tool.Execute(
				context.Background(),
				`{"input":"What is the capital of France?"}`,
			)

			require.NoError(t, err)
			assert.Equal(t, "The capital of France is Paris.", result.Content)
			assert.False(t, result.IsError)
			assert.Equal(t, 1, provider.calls)
		},
	)

	t.Run(
		"invalid JSON returns tool error not Go error",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"sub",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("sub_tool", "A sub-agent tool.")
			result, err := tool.Execute(context.Background(), `{bad json}`)

			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "Invalid parameters")
		},
	)

	t.Run(
		"empty JSON object returns tool error",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("ok"),
				},
			}

			ag := agent.New(
				"sub",
				newTestClient(provider),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("sub_tool", "A sub-agent tool.")
			result, err := tool.Execute(context.Background(), `{}`)

			require.NoError(t, err)
			assert.False(t, result.IsError)
		},
	)

	t.Run(
		"sub-agent error propagates as Go error",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{},
			}

			ag := agent.New(
				"sub",
				newTestClient(provider),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("helper", "A helper agent.")
			_, err := tool.Execute(
				context.Background(),
				`{"input":"hello"}`,
			)

			require.Error(t, err)
		},
	)

	t.Run(
		"context is forwarded to sub-agent",
		func(t *testing.T) {
			t.Parallel()

			type AppCtx struct {
				TenantID string
			}

			var captured string

			type Params struct{}
			tenantTool := agent.FunctionTool[Params](
				"get_tenant",
				"Get tenant",
				func(ctx context.Context, _ Params) (agent.ToolResult, error) {
					rc := agent.RunContextFrom[*AppCtx](ctx)
					captured = rc.TenantID
					return agent.ToolResult{Content: rc.TenantID}, nil
				},
			)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "get_tenant", Arguments: `{}`},
					}),
					stopResponse("tenant is t_789"),
				},
			}

			subAgent := agent.New(
				"tenant_agent",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(tenantTool),
			)

			tool := subAgent.AsTool("check_tenant", "Check tenant info.")

			ctx := agent.WithRunContext(
				context.Background(),
				&AppCtx{TenantID: "t_789"},
			)

			result, err := tool.Execute(ctx, `{"input":"what tenant?"}`)

			require.NoError(t, err)
			assert.Equal(t, "tenant is t_789", result.Content)
			assert.Equal(t, "t_789", captured)
		},
	)

	t.Run(
		"sub-agent with tool calls completes multi-turn",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				Expr string `json:"expr"`
			}

			calcTool := agent.FunctionTool[Params](
				"calc",
				"Calculate expression",
				func(_ context.Context, p Params) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "42"}, nil
				},
			)

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "calc", Arguments: `{"expr":"6*7"}`},
					}),
					stopResponse("The answer is 42."),
				},
			}

			subAgent := agent.New(
				"math",
				newTestClient(provider),
				agent.WithModel("test-model"),
				agent.WithTools(calcTool),
			)

			tool := subAgent.AsTool("math_expert", "Ask math questions.")
			result, err := tool.Execute(
				context.Background(),
				`{"input":"What is 6 times 7?"}`,
			)

			require.NoError(t, err)
			assert.Equal(t, "The answer is 42.", result.Content)
			assert.False(t, result.IsError)
			assert.Equal(t, 2, provider.calls)
		},
	)

	t.Run(
		"extra JSON fields are ignored",
		func(t *testing.T) {
			t.Parallel()

			provider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("ok"),
				},
			}

			ag := agent.New(
				"sub",
				newTestClient(provider),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("sub_tool", "Sub tool.")
			result, err := tool.Execute(
				context.Background(),
				`{"input":"hello","extra":"ignored"}`,
			)

			require.NoError(t, err)
			assert.False(t, result.IsError)
			assert.Equal(t, "ok", result.Content)
		},
	)
}

func TestAgentTool_InterfaceSatisfaction(t *testing.T) {
	t.Parallel()

	ag := agent.New(
		"sub",
		newTestClient(&mockProvider{}),
		agent.WithModel("test-model"),
	)

	tool := ag.AsTool("test_tool", "Test tool")

	assert.Implements(t, (*agent.Tool)(nil), tool)
	assert.Implements(t, (*agent.ToolDescriptor)(nil), tool)
}
