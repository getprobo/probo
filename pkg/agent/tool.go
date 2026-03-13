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
	"strings"

	"go.probo.inc/probo/pkg/llm"
)

type (
	ToolResult struct {
		Content string
		IsError bool
	}

	ToolDescriptor interface {
		Name() string
		Definition() llm.Tool
	}

	Tool interface {
		ToolDescriptor
		Execute(ctx context.Context, arguments string) (ToolResult, error)
	}

	functionTool[P any] struct {
		name           string
		description    string
		fn             func(ctx context.Context, params P) (ToolResult, error)
		schema         json.RawMessage
		requiredFields []string
	}
)

func FunctionTool[P any](
	name string,
	description string,
	fn func(ctx context.Context, params P) (ToolResult, error),
) (Tool, error) {
	schema, err := jsonSchemaFor[P]()
	if err != nil {
		return nil, fmt.Errorf("cannot create tool %q: %w", name, err)
	}

	var parsed struct {
		Required []string `json:"required"`
	}
	_ = json.Unmarshal(schema, &parsed)

	return &functionTool[P]{
		name:           name,
		description:    description,
		fn:             fn,
		schema:         schema,
		requiredFields: parsed.Required,
	}, nil
}

func (t *functionTool[P]) Name() string { return t.name }

func (t *functionTool[P]) Definition() llm.Tool {
	return llm.Tool{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.schema,
	}
}

func (t *functionTool[P]) Execute(ctx context.Context, arguments string) (ToolResult, error) {
	if len(t.requiredFields) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(arguments), &fields); err != nil {
			return ToolResult{
				Content: fmt.Sprintf("Invalid parameters: %s", err.Error()),
				IsError: true,
			}, nil
		}

		var missing []string
		for _, f := range t.requiredFields {
			if _, ok := fields[f]; !ok {
				missing = append(missing, f)
			}
		}

		if len(missing) > 0 {
			return ToolResult{
				Content: fmt.Sprintf(
					"Missing required parameters: %s",
					strings.Join(missing, ", "),
				),
				IsError: true,
			}, nil
		}
	}

	var params P
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return ToolResult{
			Content: fmt.Sprintf("Invalid parameters: %s", err.Error()),
			IsError: true,
		}, nil
	}

	return t.fn(ctx, params)
}
