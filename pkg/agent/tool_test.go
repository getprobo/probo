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
)

func TestFunctionTool_Name(t *testing.T) {
	t.Parallel()

	type Params struct{}

	tool, err := agent.FunctionTool(
		"my_tool",
		"does things",
		func(_ context.Context, _ Params) (agent.ToolResult, error) {
			return agent.ToolResult{}, nil
		},
	)
	require.NoError(t, err)

	assert.Equal(t, "my_tool", tool.Name())
}

func TestFunctionTool_Definition(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns name and description",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				Query string `json:"query"`
			}

			tool, err := agent.FunctionTool(
				"search",
				"Search for items",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					return agent.ToolResult{}, nil
				},
			)
			require.NoError(t, err)

			def := tool.Definition()
			assert.Equal(t, "search", def.Name)
			assert.Equal(t, "Search for items", def.Description)
		},
	)

	t.Run(
		"generates valid JSON schema from params type",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				Name  string `json:"name" jsonschema:"The item name"`
				Count int    `json:"count"`
			}

			tool, err := agent.FunctionTool(
				"create",
				"Create items",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					return agent.ToolResult{}, nil
				},
			)
			require.NoError(t, err)

			def := tool.Definition()
			require.NotNil(t, def.Parameters)

			var schema map[string]any
			require.NoError(t, json.Unmarshal(def.Parameters, &schema))

			assert.Equal(t, "object", schema["type"])

			props := schema["properties"].(map[string]any)
			assert.Contains(t, props, "name")
			assert.Contains(t, props, "count")

			nameProp := props["name"].(map[string]any)
			assert.Equal(t, "string", nameProp["type"])
			assert.Equal(t, "The item name", nameProp["description"])

			countProp := props["count"].(map[string]any)
			assert.Equal(t, "integer", countProp["type"])
		},
	)

	t.Run(
		"empty struct produces object schema with no properties",
		func(t *testing.T) {
			t.Parallel()

			type Params struct{}

			tool, err := agent.FunctionTool(
				"noop",
				"No-op",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					return agent.ToolResult{}, nil
				},
			)
			require.NoError(t, err)

			var schema map[string]any
			require.NoError(t, json.Unmarshal(tool.Definition().Parameters, &schema))
			assert.Equal(t, "object", schema["type"])
		},
	)

	t.Run(
		"pointer fields are not nullable in schema",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				Title *string `json:"title,omitempty"`
			}

			tool, err := agent.FunctionTool(
				"update",
				"Update",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					return agent.ToolResult{}, nil
				},
			)
			require.NoError(t, err)

			var schema map[string]any
			require.NoError(t, json.Unmarshal(tool.Definition().Parameters, &schema))

			props := schema["properties"].(map[string]any)
			titleProp := props["title"].(map[string]any)
			assert.Equal(t, "string", titleProp["type"])
			assert.Nil(t, titleProp["types"])
		},
	)
}

func TestFunctionTool_Execute(t *testing.T) {
	t.Parallel()

	t.Run(
		"unmarshals params and calls function",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				X int `json:"x"`
				Y int `json:"y"`
			}

			tool, err := agent.FunctionTool(
				"add",
				"Add two numbers",
				func(_ context.Context, p Params) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "42"}, nil
				},
			)
			require.NoError(t, err)

			result, err := tool.Execute(context.Background(), `{"x": 1, "y": 2}`)
			require.NoError(t, err)
			assert.Equal(t, "42", result.Content)
			assert.False(t, result.IsError)
		},
	)

	t.Run(
		"passes received params to function",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				City string `json:"city"`
			}

			var received string
			tool, err := agent.FunctionTool(
				"weather",
				"Get weather",
				func(_ context.Context, p Params) (agent.ToolResult, error) {
					received = p.City
					return agent.ToolResult{Content: "sunny"}, nil
				},
			)
			require.NoError(t, err)

			_, err = tool.Execute(context.Background(), `{"city":"Paris"}`)
			require.NoError(t, err)
			assert.Equal(t, "Paris", received)
		},
	)

	t.Run(
		"invalid JSON returns tool error not Go error",
		func(t *testing.T) {
			t.Parallel()

			type Params struct{}

			tool, err := agent.FunctionTool(
				"noop",
				"No-op",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "ok"}, nil
				},
			)
			require.NoError(t, err)

			result, err := tool.Execute(context.Background(), `{invalid`)
			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "Invalid parameters")
		},
	)

	t.Run(
		"infrastructure error propagated as Go error",
		func(t *testing.T) {
			t.Parallel()

			type Params struct{}

			tool, err := agent.FunctionTool(
				"fail",
				"Always fails",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					return agent.ToolResult{}, errors.New("db down")
				},
			)
			require.NoError(t, err)

			_, err = tool.Execute(context.Background(), `{}`)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "db down")
		},
	)

	t.Run(
		"context is forwarded to function",
		func(t *testing.T) {
			t.Parallel()

			type ctxKey struct{}
			type Params struct{}

			tool, err := agent.FunctionTool(
				"ctx_check",
				"Check context",
				func(ctx context.Context, _ Params) (agent.ToolResult, error) {
					val := ctx.Value(ctxKey{}).(string)
					return agent.ToolResult{Content: val}, nil
				},
			)
			require.NoError(t, err)

			ctx := context.WithValue(context.Background(), ctxKey{}, "hello")
			result, err := tool.Execute(ctx, `{}`)
			require.NoError(t, err)
			assert.Equal(t, "hello", result.Content)
		},
	)

	t.Run(
		"missing single required field returns tool error",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				City string `json:"city"`
			}

			tool, err := agent.FunctionTool(
				"weather",
				"Get weather",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					t.Fatal("function should not be called")
					return agent.ToolResult{}, nil
				},
			)
			require.NoError(t, err)

			result, err := tool.Execute(context.Background(), `{}`)
			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "city")
		},
	)

	t.Run(
		"missing multiple required fields lists all of them",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				City    string `json:"city"`
				Country string `json:"country"`
			}

			tool, err := agent.FunctionTool(
				"weather",
				"Get weather",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					t.Fatal("function should not be called")
					return agent.ToolResult{}, nil
				},
			)
			require.NoError(t, err)

			result, err := tool.Execute(context.Background(), `{}`)
			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "city")
			assert.Contains(t, result.Content, "country")
		},
	)

	t.Run(
		"partially missing required fields detected",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				City    string `json:"city"`
				Country string `json:"country"`
			}

			tool, err := agent.FunctionTool(
				"weather",
				"Get weather",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					t.Fatal("function should not be called")
					return agent.ToolResult{}, nil
				},
			)
			require.NoError(t, err)

			result, err := tool.Execute(context.Background(), `{"city":"Paris"}`)
			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "country")
		},
	)

	t.Run(
		"optional fields can be omitted",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				City  string  `json:"city"`
				Units *string `json:"units,omitempty"`
			}

			tool, err := agent.FunctionTool(
				"weather",
				"Get weather",
				func(_ context.Context, p Params) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "sunny in " + p.City}, nil
				},
			)
			require.NoError(t, err)

			result, err := tool.Execute(context.Background(), `{"city":"Paris"}`)
			require.NoError(t, err)
			assert.False(t, result.IsError)
			assert.Equal(t, "sunny in Paris", result.Content)
		},
	)

	t.Run(
		"extra JSON fields are ignored",
		func(t *testing.T) {
			t.Parallel()

			type Params struct {
				Name string `json:"name"`
			}

			tool, err := agent.FunctionTool(
				"greet",
				"Greet",
				func(_ context.Context, p Params) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "hi " + p.Name}, nil
				},
			)
			require.NoError(t, err)

			result, err := tool.Execute(context.Background(), `{"name":"Alice","extra":"ignored"}`)
			require.NoError(t, err)
			assert.Equal(t, "hi Alice", result.Content)
			assert.False(t, result.IsError)
		},
	)

	t.Run(
		"empty JSON object works for empty params",
		func(t *testing.T) {
			t.Parallel()

			type Params struct{}

			tool, err := agent.FunctionTool(
				"ping",
				"Ping",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "pong"}, nil
				},
			)
			require.NoError(t, err)

			result, err := tool.Execute(context.Background(), `{}`)
			require.NoError(t, err)
			assert.Equal(t, "pong", result.Content)
		},
	)

	t.Run(
		"function can return IsError true",
		func(t *testing.T) {
			t.Parallel()

			type Params struct{}

			tool, err := agent.FunctionTool(
				"validate",
				"Validate input",
				func(_ context.Context, _ Params) (agent.ToolResult, error) {
					return agent.ToolResult{Content: "validation failed", IsError: true}, nil
				},
			)
			require.NoError(t, err)

			result, err := tool.Execute(context.Background(), `{}`)
			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Equal(t, "validation failed", result.Content)
		},
	)
}

func TestFunctionTool_InterfaceSatisfaction(t *testing.T) {
	t.Parallel()

	type Params struct{}

	tool, err := agent.FunctionTool(
		"test",
		"test tool",
		func(_ context.Context, _ Params) (agent.ToolResult, error) {
			return agent.ToolResult{}, nil
		},
	)
	require.NoError(t, err)

	assert.Implements(t, (*agent.Tool)(nil), tool)
	assert.Implements(t, (*agent.ToolDescriptor)(nil), tool)
}
