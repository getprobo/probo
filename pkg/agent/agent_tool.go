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

	return ToolResult{Content: result.FinalMessage().Text()}, nil
}
