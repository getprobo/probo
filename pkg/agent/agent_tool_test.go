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
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
		"empty JSON object returns tool error for missing input",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"sub",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("sub_tool", "A sub-agent tool.")
			result, err := tool.Execute(context.Background(), `{}`)

			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "Missing required parameters")
			assert.Contains(t, result.Content, "input")
		},
	)

	t.Run(
		"null input returns tool error for missing input",
		func(t *testing.T) {
			t.Parallel()

			ag := agent.New(
				"sub",
				newTestClient(&mockProvider{}),
				agent.WithModel("test-model"),
			)

			tool := ag.AsTool("sub_tool", "A sub-agent tool.")
			result, err := tool.Execute(context.Background(), `{"input":null}`)

			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "Missing required parameters")
			assert.Contains(t, result.Content, "input")
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

type nestedApprovalFixture struct {
	innerProvider *mockProvider
	outerProvider *mockProvider
	outerAgent    *agent.Agent
}

func noopDeleteFile(_ context.Context, _ struct{}) (agent.ToolResult, error) {
	return agent.ToolResult{Content: "file deleted"}, nil
}

func newNestedApprovalFixture(
	t *testing.T,
	deleteFunc func(context.Context, struct{}) (agent.ToolResult, error),
	innerResponses []*llm.ChatCompletionResponse,
	outerResponses []*llm.ChatCompletionResponse,
	outerOpts ...agent.Option,
) nestedApprovalFixture {
	t.Helper()

	deleteTool := agent.FunctionTool[struct{}](
		"delete_file",
		"Delete a file",
		deleteFunc,
	)

	innerProvider := &mockProvider{responses: innerResponses}

	innerAgent := agent.New(
		"file_manager",
		newTestClient(innerProvider),
		agent.WithModel("test-model"),
		agent.WithTools(deleteTool),
		agent.WithApproval(agent.ApprovalConfig{
			ToolNames: []string{"delete_file"},
		}),
	)

	outerProvider := &mockProvider{responses: outerResponses}

	opts := append(
		[]agent.Option{
			agent.WithModel("test-model"),
			agent.WithTools(innerAgent.AsTool("file_expert", "Manage files")),
		},
		outerOpts...,
	)

	outerAgent := agent.New(
		"assistant",
		newTestClient(outerProvider),
		opts...,
	)

	return nestedApprovalFixture{
		innerProvider: innerProvider,
		outerProvider: outerProvider,
		outerAgent:    outerAgent,
	}
}

func TestAgentTool_Execute_NestedApproval(t *testing.T) {
	t.Parallel()

	t.Run(
		"nested agent approval surfaces as InterruptedError",
		func(t *testing.T) {
			t.Parallel()

			f := newNestedApprovalFixture(
				t,
				noopDeleteFile,
				[]*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
				},
				[]*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "outer_tc1",
						Function: llm.FunctionCall{Name: "file_expert", Arguments: `{"input":"delete the file"}`},
					}),
				},
			)

			_, err := f.outerAgent.Run(
				context.Background(),
				[]llm.Message{userMessage("Delete the file")},
			)

			require.Error(t, err)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)
			assert.Len(t, interrupted.PendingApprovals, 1)
			assert.Equal(t, "delete_file", interrupted.PendingApprovals[0].Function.Name)
			assert.Equal(t, "file_manager", interrupted.Agent.Name())
		},
	)

	t.Run(
		"nested agent approval can be resumed with approve",
		func(t *testing.T) {
			t.Parallel()

			var toolExecuted bool

			f := newNestedApprovalFixture(
				t,
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					toolExecuted = true
					return agent.ToolResult{Content: "file deleted"}, nil
				},
				[]*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
					stopResponse("File has been deleted."),
				},
				[]*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "outer_tc1",
						Function: llm.FunctionCall{Name: "file_expert", Arguments: `{"input":"delete the file"}`},
					}),
					stopResponse("Done, the file has been deleted."),
				},
			)

			_, err := f.outerAgent.Run(
				context.Background(),
				[]llm.Message{userMessage("Delete the file")},
			)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)
			assert.False(t, toolExecuted)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"inner_tc1": {Approved: true},
					},
				},
			)

			require.NoError(t, err)
			assert.True(t, toolExecuted)
			assert.Equal(t, "Done, the file has been deleted.", result.FinalMessage().Text())
			assert.Equal(t, "assistant", result.LastAgent.Name())
		},
	)

	t.Run(
		"nested agent rejection resumes outer agent",
		func(t *testing.T) {
			t.Parallel()

			f := newNestedApprovalFixture(
				t,
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					t.Fatal("tool should not be called")
					return agent.ToolResult{}, nil
				},
				[]*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
					stopResponse("OK, I won't delete the file."),
				},
				[]*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "outer_tc1",
						Function: llm.FunctionCall{Name: "file_expert", Arguments: `{"input":"delete the file"}`},
					}),
					stopResponse("The file manager declined."),
				},
			)

			_, err := f.outerAgent.Run(
				context.Background(),
				[]llm.Message{userMessage("Delete the file")},
			)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"inner_tc1": {Approved: false, Message: "User denied deletion."},
					},
				},
			)

			require.NoError(t, err)
			assert.Equal(t, "The file manager declined.", result.FinalMessage().Text())
			assert.Equal(t, "assistant", result.LastAgent.Name())
		},
	)

	t.Run(
		"nested agent approval with parallel sibling tools",
		func(t *testing.T) {
			t.Parallel()

			var siblingCalled bool

			type Params struct{}

			siblingTool := agent.FunctionTool[Params](
				"list_files",
				"List files",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					siblingCalled = true
					return agent.ToolResult{Content: "file1.txt, file2.txt"}, nil
				},
			)

			f := newNestedApprovalFixture(
				t,
				noopDeleteFile,
				[]*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
					stopResponse("File has been deleted."),
				},
				[]*llm.ChatCompletionResponse{
					toolCallResponse(
						llm.ToolCall{
							ID:       "outer_tc1",
							Function: llm.FunctionCall{Name: "list_files", Arguments: `{}`},
						},
						llm.ToolCall{
							ID:       "outer_tc2",
							Function: llm.FunctionCall{Name: "file_expert", Arguments: `{"input":"delete the file"}`},
						},
					),
					stopResponse("Files listed and deleted."),
				},
				agent.WithTools(siblingTool),
			)

			_, err := f.outerAgent.Run(
				context.Background(),
				[]llm.Message{userMessage("List and delete files")},
			)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)
			assert.True(t, siblingCalled)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"inner_tc1": {Approved: true},
					},
				},
			)

			require.NoError(t, err)
			assert.Equal(t, "Files listed and deleted.", result.FinalMessage().Text())
		},
	)

	t.Run(
		"three-level nesting A to B to C preserves full chain",
		func(t *testing.T) {
			t.Parallel()

			var toolExecuted bool

			dangerTool := agent.FunctionTool[struct{}](
				"danger",
				"Dangerous operation",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					toolExecuted = true
					return agent.ToolResult{Content: "danger executed"}, nil
				},
			)

			cProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "c_tc1",
						Function: llm.FunctionCall{Name: "danger", Arguments: `{}`},
					}),
					stopResponse("C done."),
				},
			}

			agentC := agent.New(
				"agent_c",
				newTestClient(cProvider),
				agent.WithModel("test-model"),
				agent.WithTools(dangerTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"danger"},
				}),
			)

			bProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "b_tc1",
						Function: llm.FunctionCall{Name: "call_c", Arguments: `{"input":"do danger"}`},
					}),
					stopResponse("B done."),
				},
			}

			agentB := agent.New(
				"agent_b",
				newTestClient(bProvider),
				agent.WithModel("test-model"),
				agent.WithTools(agentC.AsTool("call_c", "Call agent C")),
			)

			aProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "a_tc1",
						Function: llm.FunctionCall{Name: "call_b", Arguments: `{"input":"delegate to C"}`},
					}),
					stopResponse("A done."),
				},
			}

			agentA := agent.New(
				"agent_a",
				newTestClient(aProvider),
				agent.WithModel("test-model"),
				agent.WithTools(agentB.AsTool("call_b", "Call agent B")),
			)

			_, err := agentA.Run(
				context.Background(),
				[]llm.Message{userMessage("start")},
			)

			require.Error(t, err)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)
			assert.Equal(t, "agent_c", interrupted.Agent.Name())
			assert.Equal(t, "danger", interrupted.PendingApprovals[0].Function.Name)
			assert.False(t, toolExecuted)

			result, err := agent.Resume(
				context.Background(),
				interrupted,
				agent.ResumeInput{
					Approvals: map[string]agent.ApprovalResult{
						"c_tc1": {Approved: true},
					},
				},
			)

			require.NoError(t, err)
			assert.True(t, toolExecuted)
			assert.Equal(t, "A done.", result.FinalMessage().Text())
			assert.Equal(t, "agent_a", result.LastAgent.Name())
		},
	)

	t.Run(
		"nested interruption emits paired OnToolStart and OnToolEnd on outer agent",
		func(t *testing.T) {
			t.Parallel()

			hook := &recordingHook{}

			f := newNestedApprovalFixture(
				t,
				noopDeleteFile,
				[]*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
				},
				[]*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "outer_tc1",
						Function: llm.FunctionCall{Name: "file_expert", Arguments: `{"input":"delete the file"}`},
					}),
				},
				agent.WithHooks(hook),
			)

			_, err := f.outerAgent.Run(
				context.Background(),
				[]llm.Message{userMessage("Delete the file")},
			)

			var interrupted *agent.InterruptedError
			require.ErrorAs(t, err, &interrupted)
			require.Len(t, hook.toolStartNames, 1)
			assert.Equal(t, hook.toolStartNames, hook.toolNames, "every OnToolStart must have a matching OnToolEnd")
		},
	)
}

func TestAgentTool_Execute_DepthLimit(t *testing.T) {
	t.Parallel()

	t.Run(
		"deep agent-tool chain stops at depth limit",
		func(t *testing.T) {
			t.Parallel()

			innerProvider := &mockProvider{}

			innerAgent := agent.New(
				"inner",
				newTestClient(innerProvider),
				agent.WithModel("test-model"),
				agent.WithMaxToolDepth(1),
			)

			middleProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_mid",
						Function: llm.FunctionCall{Name: "call_inner", Arguments: `{"input":"ping"}`},
					}),
					stopResponse("inner was unreachable"),
				},
			}

			middleAgent := agent.New(
				"middle",
				newTestClient(middleProvider),
				agent.WithModel("test-model"),
				agent.WithTools(innerAgent.AsTool("call_inner", "Call inner")),
			)

			outerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_out",
						Function: llm.FunctionCall{Name: "call_middle", Arguments: `{"input":"start"}`},
					}),
					stopResponse("outer done"),
				},
			}

			outerAgent := agent.New(
				"outer",
				newTestClient(outerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(middleAgent.AsTool("call_middle", "Call middle")),
			)

			result, err := outerAgent.Run(
				context.Background(),
				[]llm.Message{userMessage("go")},
			)

			require.NoError(t, err)
			assert.Equal(t, "outer done", result.FinalMessage().Text())
			assert.Equal(t, 0, innerProvider.calls, "inner agent should never be called")
		},
	)

	t.Run(
		"delegation within depth limit succeeds",
		func(t *testing.T) {
			t.Parallel()

			innerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					stopResponse("inner result"),
				},
			}

			innerAgent := agent.New(
				"inner",
				newTestClient(innerProvider),
				agent.WithModel("test-model"),
				agent.WithMaxToolDepth(2),
			)

			outerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "tc_1",
						Function: llm.FunctionCall{Name: "call_inner", Arguments: `{"input":"hello"}`},
					}),
					stopResponse("outer result"),
				},
			}

			outerAgent := agent.New(
				"outer",
				newTestClient(outerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(innerAgent.AsTool("call_inner", "Call inner")),
			)

			result, err := outerAgent.Run(
				context.Background(),
				[]llm.Message{userMessage("go")},
			)

			require.NoError(t, err)
			assert.Equal(t, "outer result", result.FinalMessage().Text())
			assert.Equal(t, 1, innerProvider.calls, "inner agent should be called once")
		},
	)
}

func TestAgentTool_Execute_SuspendAndRestoreSingleLevel(t *testing.T) {
	t.Parallel()

	store := newMemoryCheckpointer()

	toolReady := make(chan struct{})
	toolRelease := make(chan struct{})

	var readyOnce sync.Once

	slowTool := agent.FunctionTool[struct{}](
		"slow_inner_work",
		"Slow inner work",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			readyOnce.Do(func() { close(toolReady) })
			<-toolRelease

			return agent.ToolResult{Content: "inner tool done"}, nil
		},
	)

	innerProvider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_inner",
					Function: llm.FunctionCall{Name: "slow_inner_work", Arguments: `{}`},
				},
			),
			stopResponse("inner completed"),
		},
	}

	innerAgent := agent.New(
		"inner-agent",
		newTestClient(innerProvider),
		agent.WithModel("test-model"),
		agent.WithTools(slowTool),
	)

	outerProvider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_outer",
					Function: llm.FunctionCall{Name: "call_inner", Arguments: `{"input":"delegate"}`},
				},
			),
			stopResponse("outer completed"),
		},
	}

	outerAgent := agent.New(
		"outer-agent",
		newTestClient(outerProvider),
		agent.WithModel("test-model"),
		agent.WithTools(innerAgent.AsTool("call_inner", "Call inner")),
	)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() {
		_, err := outerAgent.Run(
			ctx,
			[]llm.Message{userMessage("go")},
			agent.WithCheckpointer(store, "run-single-level"),
		)
		errCh <- err
	}()

	select {
	case <-toolReady:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for inner leaf tool to start")
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
	close(toolRelease)

	select {
	case err := <-errCh:
		var se *agent.SuspendedError
		require.ErrorAs(t, err, &se)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for run to suspend")
	}

	cp, err := store.Load(context.Background(), "run-single-level")
	require.NoError(t, err)
	require.NotNil(t, cp)
	assert.Equal(t, agent.AgentStatusSuspended, cp.Status)
	assert.Equal(t, "outer-agent", cp.AgentName)

	innerCP, ok := cp.InnerCheckpoints["tc_outer"]
	require.True(t, ok, "expected nested checkpoint keyed by outer tool call")
	require.NotNil(t, innerCP)
	assert.Equal(t, "inner-agent", innerCP.AgentName)
	assert.Equal(t, agent.AgentStatusSuspended, innerCP.Status)

	registry := &simpleRegistry{
		agents: map[string]*agent.Agent{
			"outer-agent": outerAgent,
			"inner-agent": innerAgent,
		},
	}

	result, err := agent.Restore(
		context.Background(),
		store,
		"run-single-level",
		registry,
	)
	require.NoError(t, err)
	assert.Equal(t, "outer completed", result.FinalMessage().Text())
	assert.Equal(t, "outer-agent", result.LastAgent.Name())
}

func TestAgentTool_Execute_SuspendAndRestoreMultiLevel(t *testing.T) {
	t.Parallel()

	store := newMemoryCheckpointer()

	toolReady := make(chan struct{})
	toolRelease := make(chan struct{})

	var readyOnce sync.Once

	slowTool := agent.FunctionTool[struct{}](
		"slow_grandchild_work",
		"Slow grandchild work",
		func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
			readyOnce.Do(func() { close(toolReady) })
			<-toolRelease

			return agent.ToolResult{Content: "grandchild tool done"}, nil
		},
	)

	grandchildProvider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_grandchild",
					Function: llm.FunctionCall{Name: "slow_grandchild_work", Arguments: `{}`},
				},
			),
			stopResponse("grandchild completed"),
		},
	}

	grandchildAgent := agent.New(
		"grandchild-agent",
		newTestClient(grandchildProvider),
		agent.WithModel("test-model"),
		agent.WithTools(slowTool),
	)

	childProvider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_child",
					Function: llm.FunctionCall{Name: "call_grandchild", Arguments: `{"input":"delegate deeper"}`},
				},
			),
			stopResponse("child completed"),
		},
	}

	childAgent := agent.New(
		"child-agent",
		newTestClient(childProvider),
		agent.WithModel("test-model"),
		agent.WithTools(grandchildAgent.AsTool("call_grandchild", "Call grandchild")),
	)

	outerProvider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_outer",
					Function: llm.FunctionCall{Name: "call_child", Arguments: `{"input":"delegate"}`},
				},
			),
			stopResponse("outer completed"),
		},
	}

	outerAgent := agent.New(
		"outer-agent",
		newTestClient(outerProvider),
		agent.WithModel("test-model"),
		agent.WithTools(childAgent.AsTool("call_child", "Call child")),
	)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)

	go func() {
		_, err := outerAgent.Run(
			ctx,
			[]llm.Message{userMessage("go")},
			agent.WithCheckpointer(store, "run-multi-level"),
		)
		errCh <- err
	}()

	select {
	case <-toolReady:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for grandchild leaf tool to start")
	}

	cancel()
	time.Sleep(50 * time.Millisecond)
	close(toolRelease)

	select {
	case err := <-errCh:
		var se *agent.SuspendedError
		require.ErrorAs(t, err, &se)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for multi-level run to suspend")
	}

	cp, err := store.Load(context.Background(), "run-multi-level")
	require.NoError(t, err)
	require.NotNil(t, cp)
	assert.Equal(t, "outer-agent", cp.AgentName)

	childCP, ok := cp.InnerCheckpoints["tc_outer"]
	require.True(t, ok)
	require.NotNil(t, childCP)
	assert.Equal(t, "child-agent", childCP.AgentName)

	grandchildCP, ok := childCP.InnerCheckpoints["tc_child"]
	require.True(t, ok, "child checkpoint keys: %v", childCP.InnerCheckpoints)
	require.NotNil(t, grandchildCP)
	assert.Equal(t, "grandchild-agent", grandchildCP.AgentName)
	assert.Equal(t, agent.AgentStatusSuspended, grandchildCP.Status)

	registry := &simpleRegistry{
		agents: map[string]*agent.Agent{
			"outer-agent":      outerAgent,
			"child-agent":      childAgent,
			"grandchild-agent": grandchildAgent,
		},
	}

	result, err := agent.Restore(
		context.Background(),
		store,
		"run-multi-level",
		registry,
	)
	require.NoError(t, err)
	assert.Equal(t, "outer completed", result.FinalMessage().Text())
	assert.Equal(t, "outer-agent", result.LastAgent.Name())
}

func TestAgentTool_Execute_LeafToolsRemainDetachedOnSuspend(t *testing.T) {
	t.Parallel()

	var leafCtxCanceled atomic.Bool

	leafStarted := make(chan struct{})
	leafRelease := make(chan struct{})

	leafTool := agent.FunctionTool[struct{}](
		"slow_leaf",
		"Slow leaf tool",
		func(ctx context.Context, _ struct{}) (agent.ToolResult, error) {
			close(leafStarted)

			select {
			case <-ctx.Done():
				leafCtxCanceled.Store(true)
				return agent.ToolResult{Content: "leaf cancelled", IsError: true}, nil
			case <-leafRelease:
				if ctx.Err() != nil {
					leafCtxCanceled.Store(true)
				}
				return agent.ToolResult{Content: "leaf completed"}, nil
			}
		},
	)

	provider := &mockProvider{
		responses: []*llm.ChatCompletionResponse{
			toolCallResponse(
				llm.ToolCall{
					ID:       "tc_leaf",
					Function: llm.FunctionCall{Name: "slow_leaf", Arguments: `{}`},
				},
			),
			stopResponse("done"),
		},
	}

	ag := agent.New(
		"leaf-agent",
		newTestClient(provider),
		agent.WithModel("test-model"),
		agent.WithTools(leafTool),
	)

	ctx, cancel := context.WithCancel(context.Background())
	type runResult struct {
		result *agent.Result
		err    error
	}
	runDone := make(chan runResult, 1)

	go func() {
		result, err := ag.Run(ctx, []llm.Message{userMessage("go")})
		runDone <- runResult{result: result, err: err}
	}()

	select {
	case <-leafStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for leaf tool to start")
	}

	cancel()

	select {
	case outcome := <-runDone:
		t.Fatalf(
			"run returned before leaf tool release: err=%v result=%v",
			outcome.err,
			outcome.result,
		)
	case <-time.After(250 * time.Millisecond):
	}

	close(leafRelease)

	select {
	case outcome := <-runDone:
		var se *agent.SuspendedError
		require.ErrorAs(t, outcome.err, &se)
		assert.Nil(t, outcome.result)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for run completion after leaf release")
	}

	assert.False(
		t,
		leafCtxCanceled.Load(),
		"leaf tool ctx should remain detached from suspend cancellation",
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
	assert.Implements(t, (*agent.SuspendableTool)(nil), tool)
}
