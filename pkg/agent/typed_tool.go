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

type typedTool[In, Out any] struct {
	name           string
	description    string
	fn             func(ctx context.Context, params In) (Out, error)
	schema         json.RawMessage
	requiredFields []string
}

// TypedTool creates a tool with typed input and output. The output is
// automatically marshaled to JSON. If the function returns an error,
// it becomes a ToolResult with IsError set to true.
func TypedTool[In, Out any](
	name string,
	description string,
	fn func(ctx context.Context, params In) (Out, error),
) (Tool, error) {
	schema, err := jsonSchemaFor[In]()
	if err != nil {
		return nil, fmt.Errorf("cannot create tool %q: %w", name, err)
	}

	var parsed struct {
		Required []string `json:"required"`
	}
	_ = json.Unmarshal(schema, &parsed)

	return &typedTool[In, Out]{
		name:           name,
		description:    description,
		fn:             fn,
		schema:         schema,
		requiredFields: parsed.Required,
	}, nil
}

func (t *typedTool[In, Out]) Name() string { return t.name }

func (t *typedTool[In, Out]) Definition() llm.Tool {
	return llm.Tool{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.schema,
	}
}

func (t *typedTool[In, Out]) Execute(ctx context.Context, arguments string) (ToolResult, error) {
	if len(t.requiredFields) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(arguments), &fields); err != nil {
			return ResultErrorf("Invalid parameters: %s", err), nil
		}

		var missing []string
		for _, f := range t.requiredFields {
			if _, ok := fields[f]; !ok {
				missing = append(missing, f)
			}
		}

		if len(missing) > 0 {
			return ResultError(
				fmt.Sprintf(
					"Missing required parameters: %s",
					strings.Join(missing, ", "),
				),
			), nil
		}
	}

	var params In
	if err := json.Unmarshal([]byte(arguments), &params); err != nil {
		return ResultErrorf("Invalid parameters: %s", err), nil
	}

	out, err := t.fn(ctx, params)
	if err != nil {
		return ResultError(err.Error()), nil
	}

	return ResultJSON(out), nil
}
