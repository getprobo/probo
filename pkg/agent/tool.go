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
		Content string `json:"content"`
		IsError bool   `json:"is_error"`
	}

	// ToolDescriptor describes a tool's name and LLM definition.
	ToolDescriptor interface {
		Name() string
		Definition() llm.Tool
	}

	Tool interface {
		ToolDescriptor
		Execute(ctx context.Context, arguments string) (ToolResult, error)
	}
)

// ResultJSON marshals v to JSON and returns a successful ToolResult.
func ResultJSON(v any) ToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return ToolResult{
			Content: fmt.Sprintf("cannot marshal tool result: %s", err),
			IsError: true,
		}
	}
	return ToolResult{Content: string(data)}
}

// ResultError returns an error ToolResult with the given message.
func ResultError(msg string) ToolResult {
	return ToolResult{Content: msg, IsError: true}
}

// ResultErrorf returns an error ToolResult with a formatted message.
func ResultErrorf(format string, args ...any) ToolResult {
	return ToolResult{Content: fmt.Sprintf(format, args...), IsError: true}
}

type (
	functionTool[P any] struct {
		name           string
		description    string
		fn             func(ctx context.Context, params P) (ToolResult, error)
		schema         json.RawMessage
		requiredFields []string
	}
)

// FunctionTool creates a tool whose parameters are typed by P. The JSON
// schema advertised to the LLM is generated from P at construction time.
//
// Schema generation is derived from a compile-time Go type: a failure
// here is a programmer error (bad struct tag, unsupported type), not a
// runtime condition, so we panic rather than returning an error. The
// same applies to the required-fields metadata parsed back out of the
// generated schema.
func FunctionTool[P any](
	name string,
	description string,
	fn func(ctx context.Context, params P) (ToolResult, error),
) Tool {
	schema, err := jsonSchemaFor[P]()
	if err != nil {
		panic(fmt.Sprintf("agent: cannot generate JSON schema for tool %q: %s", name, err))
	}

	var parsed struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(schema, &parsed); err != nil {
		panic(fmt.Sprintf("agent: cannot parse generated schema for tool %q: %s", name, err))
	}

	return &functionTool[P]{
		name:           name,
		description:    description,
		fn:             fn,
		schema:         schema,
		requiredFields: parsed.Required,
	}
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
