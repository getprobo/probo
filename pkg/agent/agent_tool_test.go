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
			tenantTool, err := agent.FunctionTool[Params](
				"get_tenant",
				"Get tenant",
				func(ctx context.Context, _ Params) (agent.ToolResult, error) {
					rc := agent.RunContextFrom[*AppCtx](ctx)
					captured = rc.TenantID
					return agent.ToolResult{Content: rc.TenantID}, nil
				},
			)
			require.NoError(t, err)

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

			calcTool, err := agent.FunctionTool[Params](
				"calc",
				"Calculate expression",
				func(_ context.Context, p Params) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "42"}, nil
				},
			)
			require.NoError(t, err)

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

func TestAgentTool_Execute_NestedApproval(t *testing.T) {
	t.Parallel()

	t.Run(
		"nested agent approval surfaces as InterruptedError",
		func(t *testing.T) {
			t.Parallel()

			deleteTool, err := agent.FunctionTool[struct{}](
				"delete_file",
				"Delete a file",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "file deleted"}, nil
				},
			)
			require.NoError(t, err)

			innerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
				},
			}

			innerAgent := agent.New(
				"file_manager",
				newTestClient(innerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(deleteTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"delete_file"},
				}),
			)

			outerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "outer_tc1",
						Function: llm.FunctionCall{Name: "file_expert", Arguments: `{"input":"delete the file"}`},
					}),
				},
			}

			outerAgent := agent.New(
				"assistant",
				newTestClient(outerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(innerAgent.AsTool("file_expert", "Manage files")),
			)

			_, err = outerAgent.Run(
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

			deleteTool, err := agent.FunctionTool[struct{}](
				"delete_file",
				"Delete a file",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					toolExecuted = true
					return agent.ToolResult{Content: "file deleted"}, nil
				},
			)
			require.NoError(t, err)

			innerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
					stopResponse("File has been deleted."),
				},
			}

			innerAgent := agent.New(
				"file_manager",
				newTestClient(innerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(deleteTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"delete_file"},
				}),
			)

			outerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "outer_tc1",
						Function: llm.FunctionCall{Name: "file_expert", Arguments: `{"input":"delete the file"}`},
					}),
					stopResponse("Done, the file has been deleted."),
				},
			}

			outerAgent := agent.New(
				"assistant",
				newTestClient(outerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(innerAgent.AsTool("file_expert", "Manage files")),
			)

			_, err = outerAgent.Run(
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

			deleteTool, err := agent.FunctionTool[struct{}](
				"delete_file",
				"Delete a file",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					t.Fatal("tool should not be called")
					return agent.ToolResult{}, nil
				},
			)
			require.NoError(t, err)

			innerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
					stopResponse("OK, I won't delete the file."),
				},
			}

			innerAgent := agent.New(
				"file_manager",
				newTestClient(innerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(deleteTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"delete_file"},
				}),
			)

			outerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "outer_tc1",
						Function: llm.FunctionCall{Name: "file_expert", Arguments: `{"input":"delete the file"}`},
					}),
					stopResponse("The file manager declined."),
				},
			}

			outerAgent := agent.New(
				"assistant",
				newTestClient(outerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(innerAgent.AsTool("file_expert", "Manage files")),
			)

			_, err = outerAgent.Run(
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
			siblingTool, err := agent.FunctionTool[Params](
				"list_files",
				"List files",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					siblingCalled = true
					return agent.ToolResult{Content: "file1.txt, file2.txt"}, nil
				},
			)
			require.NoError(t, err)

			deleteTool, err := agent.FunctionTool[struct{}](
				"delete_file",
				"Delete a file",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "file deleted"}, nil
				},
			)
			require.NoError(t, err)

			innerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
					stopResponse("File has been deleted."),
				},
			}

			innerAgent := agent.New(
				"file_manager",
				newTestClient(innerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(deleteTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"delete_file"},
				}),
			)

			outerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
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
			}

			outerAgent := agent.New(
				"assistant",
				newTestClient(outerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(
					siblingTool,
					innerAgent.AsTool("file_expert", "Manage files"),
				),
			)

			_, err = outerAgent.Run(
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

			dangerTool, err := agent.FunctionTool[struct{}](
				"danger",
				"Dangerous operation",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					toolExecuted = true
					return agent.ToolResult{Content: "danger executed"}, nil
				},
			)
			require.NoError(t, err)

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

			_, err = agentA.Run(
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

			deleteTool, err := agent.FunctionTool[struct{}](
				"delete_file",
				"Delete a file",
				func(_ context.Context, _ struct{}) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "file deleted"}, nil
				},
			)
			require.NoError(t, err)

			innerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "inner_tc1",
						Function: llm.FunctionCall{Name: "delete_file", Arguments: `{}`},
					}),
				},
			}

			innerAgent := agent.New(
				"file_manager",
				newTestClient(innerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(deleteTool),
				agent.WithApproval(agent.ApprovalConfig{
					ToolNames: []string{"delete_file"},
				}),
			)

			outerProvider := &mockProvider{
				responses: []*llm.ChatCompletionResponse{
					toolCallResponse(llm.ToolCall{
						ID:       "outer_tc1",
						Function: llm.FunctionCall{Name: "file_expert", Arguments: `{"input":"delete the file"}`},
					}),
				},
			}

			hook := &recordingHook{}

			outerAgent := agent.New(
				"assistant",
				newTestClient(outerProvider),
				agent.WithModel("test-model"),
				agent.WithTools(innerAgent.AsTool("file_expert", "Manage files")),
				agent.WithHooks(hook),
			)

			_, err = outerAgent.Run(
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
