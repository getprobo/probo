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

package vetting

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

type strictFunctionTool[P any] struct {
	name           string
	description    string
	fn             func(ctx context.Context, params P) (agent.ToolResult, error)
	schema         json.RawMessage
	requiredFields []string
}

// jsonSchemaForTool builds an OpenAI strict-mode JSON schema for vetting tools
// and structured outputs. The shared agent schema generator already emits
// additionalProperties=false and a complete required list on every struct and
// array-of-struct object, so enforceStrictJSONSchema only adds the remaining
// strict-mode normalisation the generator skips: it sorts required and locks
// down map[string]T fields (which serialise to a non-false additionalProperties).
// Output types that are flat or map-free (e.g. evidenceassessor.EvidenceAssessment)
// are already strict-valid straight from agent.NewOutputType and do not need this.
func jsonSchemaForTool[T any]() (json.RawMessage, error) {
	outputType, err := agent.NewOutputType[T]("_")
	if err != nil {
		return nil, fmt.Errorf("cannot generate schema: %w", err)
	}

	return enforceStrictJSONSchema(outputType.Schema)
}

func newVettingOutputType[T any](name string) (*agent.OutputType, error) {
	outputType, err := agent.NewOutputType[T](name)
	if err != nil {
		return nil, err
	}

	schema, err := enforceStrictJSONSchema(outputType.Schema)
	if err != nil {
		return nil, fmt.Errorf("cannot enforce strict schema for %q: %w", name, err)
	}

	outputType.Schema = schema

	return outputType, nil
}

func vettingFunctionTool[P any](
	name string,
	description string,
	fn func(ctx context.Context, params P) (agent.ToolResult, error),
) agent.Tool {
	schema, err := jsonSchemaForTool[P]()
	if err != nil {
		panic(fmt.Sprintf("vetting: cannot generate JSON schema for tool %q: %s", name, err))
	}

	var parsed struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(schema, &parsed); err != nil {
		panic(fmt.Sprintf("vetting: cannot parse generated schema for tool %q: %s", name, err))
	}

	return &strictFunctionTool[P]{
		name:           name,
		description:    description,
		fn:             fn,
		schema:         schema,
		requiredFields: parsed.Required,
	}
}

func (t *strictFunctionTool[P]) Name() string { return t.name }

func (t *strictFunctionTool[P]) Definition() llm.Tool {
	return llm.Tool{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.schema,
	}
}

func (t *strictFunctionTool[P]) Execute(ctx context.Context, arguments string) (agent.ToolResult, error) {
	if len(t.requiredFields) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(arguments), &fields); err != nil {
			return agent.ToolResult{
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
			return agent.ToolResult{
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
		return agent.ToolResult{
			Content: fmt.Sprintf("Invalid parameters: %s", err.Error()),
			IsError: true,
		}, nil
	}

	return t.fn(ctx, params)
}

func enforceStrictJSONSchema(raw json.RawMessage) (json.RawMessage, error) {
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, fmt.Errorf("cannot unmarshal schema: %w", err)
	}

	normalizeStrictObject(schema)

	data, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal strict schema: %w", err)
	}

	return json.RawMessage(data), nil
}

func normalizeStrictObject(schema map[string]any) {
	if schema == nil {
		return
	}

	if props, ok := schema["properties"].(map[string]any); ok && len(props) > 0 {
		required := make([]string, 0, len(props))
		for name, prop := range props {
			required = append(required, name)

			if nested, ok := prop.(map[string]any); ok {
				normalizeStrictObject(nested)
			}
		}

		slices.Sort(required)

		requiredAny := make([]any, len(required))
		for i, name := range required {
			requiredAny[i] = name
		}

		schema["required"] = requiredAny
		schema["additionalProperties"] = false
	}

	if items, ok := schema["items"].(map[string]any); ok {
		normalizeStrictObject(items)
	}

	if additional, ok := schema["additionalProperties"].(map[string]any); ok {
		normalizeStrictObject(additional)
	}
}
