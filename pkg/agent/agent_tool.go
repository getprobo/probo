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

package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"go.probo.inc/probo/pkg/llm"
)

const DefaultMaxToolDepth = 16

type (
	agentTool struct {
		agent       *Agent
		toolName    string
		description string
		schema      json.RawMessage
	}

	agentToolParams struct {
		Input string `json:"input" jsonschema:"The input to send to the agent"`
	}

	agentToolDepthKey struct{}
)

var (
	agentToolParamsSchema = mustJSONSchemaFor[agentToolParams]()

	_ SuspendableTool = (*agentTool)(nil)
)

func agentToolDepth(ctx context.Context) int {
	if v, ok := ctx.Value(agentToolDepthKey{}).(int); ok {
		return v
	}

	return 0
}

func newAgentTool(agent *Agent, name, description string) *agentTool {
	return &agentTool{
		agent:       agent,
		toolName:    name,
		description: description,
		schema:      agentToolParamsSchema,
	}
}

func (t *agentTool) Name() string { return t.toolName }

func (t *agentTool) Suspendable() {}

func (t *agentTool) Definition() llm.Tool {
	return llm.Tool{
		Name:        t.toolName,
		Description: t.description,
		Parameters:  t.schema,
	}
}

func (t *agentTool) Execute(ctx context.Context, arguments string) (ToolResult, error) {
	depth := agentToolDepth(ctx)
	if depth >= t.agent.maxToolDepth {
		return ToolResult{}, &MaxToolDepthExceededError{MaxDepth: t.agent.maxToolDepth}
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(arguments), &fields); err != nil {
		return ToolResult{
			Content: fmt.Sprintf("Invalid parameters: %s", err.Error()),
			IsError: true,
		}, nil
	}

	raw, ok := fields["input"]
	if !ok || string(raw) == "null" {
		return ToolResult{
			Content: "Missing required parameters: input",
			IsError: true,
		}, nil
	}

	var params agentToolParams
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return ToolResult{
			Content: fmt.Sprintf("Invalid parameters: %s", err.Error()),
			IsError: true,
		}, nil
	}

	ctx = context.WithValue(ctx, agentToolDepthKey{}, depth+1)

	result, err := t.agent.Run(
		ctx,
		[]llm.Message{
			{
				Role: llm.RoleUser,
				Parts: []llm.Part{
					llm.TextPart{Text: params.Input},
				},
			},
		},
	)
	if err != nil {
		return ToolResult{}, err
	}

	text := result.FinalMessage().Text()

	if t.agent.outputType != nil {
		if !json.Valid([]byte(text)) {
			preview := text
			if len(preview) > 500 {
				preview = preview[:500] + "... (truncated)"
			}

			return ToolResult{
				Content: fmt.Sprintf("Sub-agent %q returned invalid JSON. Raw output:\n%s", t.agent.name, preview),
				IsError: true,
			}, nil
		}
	}

	return ToolResult{Content: text}, nil
}
