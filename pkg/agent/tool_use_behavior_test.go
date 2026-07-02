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

package agent_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
)

func TestRunLLMAgain(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns not final with no results",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.RunLLMAgain()
			output, isFinal, err := behavior(context.Background(), nil)

			require.NoError(t, err)
			assert.False(t, isFinal)
			assert.Equal(t, "", output)
		},
	)

	t.Run(
		"returns not final with results present",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.RunLLMAgain()
			results := []agent.ToolCallResult{
				{
					ToolName:  "search",
					Arguments: `{"q":"test"}`,
					Result:    agent.ToolResult{Content: "found 3 items"},
				},
			}

			output, isFinal, err := behavior(context.Background(), results)

			require.NoError(t, err)
			assert.False(t, isFinal)
			assert.Equal(t, "", output)
		},
	)
}

func TestStopOnFirstTool(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns not final when results are empty",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopOnFirstTool()
			output, isFinal, err := behavior(context.Background(), nil)

			require.NoError(t, err)
			assert.False(t, isFinal)
			assert.Equal(t, "", output)
		},
	)

	t.Run(
		"returns first tool output as final",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopOnFirstTool()
			results := []agent.ToolCallResult{
				{
					ToolName: "lookup",
					Result:   agent.ToolResult{Content: "first result"},
				},
				{
					ToolName: "search",
					Result:   agent.ToolResult{Content: "second result"},
				},
			}

			output, isFinal, err := behavior(context.Background(), results)

			require.NoError(t, err)
			assert.True(t, isFinal)
			assert.Equal(t, "first result", output)
		},
	)

	t.Run(
		"returns single result as final",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopOnFirstTool()
			results := []agent.ToolCallResult{
				{
					ToolName: "compute",
					Result:   agent.ToolResult{Content: "42"},
				},
			}

			output, isFinal, err := behavior(context.Background(), results)

			require.NoError(t, err)
			assert.True(t, isFinal)
			assert.Equal(t, "42", output)
		},
	)

	t.Run(
		"returns empty content when first tool has empty output",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopOnFirstTool()
			results := []agent.ToolCallResult{
				{
					ToolName: "noop",
					Result:   agent.ToolResult{Content: ""},
				},
			}

			output, isFinal, err := behavior(context.Background(), results)

			require.NoError(t, err)
			assert.True(t, isFinal)
			assert.Equal(t, "", output)
		},
	)
}

func TestStopAtTools(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns not final when results are empty",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopAtTools("done")
			output, isFinal, err := behavior(context.Background(), nil)

			require.NoError(t, err)
			assert.False(t, isFinal)
			assert.Equal(t, "", output)
		},
	)

	t.Run(
		"returns not final when no tool matches",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopAtTools("done", "finish")
			results := []agent.ToolCallResult{
				{
					ToolName: "search",
					Result:   agent.ToolResult{Content: "results"},
				},
				{
					ToolName: "compute",
					Result:   agent.ToolResult{Content: "42"},
				},
			}

			output, isFinal, err := behavior(context.Background(), results)

			require.NoError(t, err)
			assert.False(t, isFinal)
			assert.Equal(t, "", output)
		},
	)

	t.Run(
		"stops on matching tool and returns its output",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopAtTools("submit")
			results := []agent.ToolCallResult{
				{
					ToolName: "search",
					Result:   agent.ToolResult{Content: "search result"},
				},
				{
					ToolName: "submit",
					Result:   agent.ToolResult{Content: "submitted"},
				},
			}

			output, isFinal, err := behavior(context.Background(), results)

			require.NoError(t, err)
			assert.True(t, isFinal)
			assert.Equal(t, "submitted", output)
		},
	)

	t.Run(
		"returns first matching tool when multiple match",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopAtTools("done", "finish")
			results := []agent.ToolCallResult{
				{
					ToolName: "finish",
					Result:   agent.ToolResult{Content: "finished"},
				},
				{
					ToolName: "done",
					Result:   agent.ToolResult{Content: "done"},
				},
			}

			output, isFinal, err := behavior(context.Background(), results)

			require.NoError(t, err)
			assert.True(t, isFinal)
			assert.Equal(t, "finished", output)
		},
	)

	t.Run(
		"matches any of the listed tool names",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopAtTools("alpha", "beta", "gamma")
			results := []agent.ToolCallResult{
				{
					ToolName: "unrelated",
					Result:   agent.ToolResult{Content: "ignored"},
				},
				{
					ToolName: "gamma",
					Result:   agent.ToolResult{Content: "gamma output"},
				},
			}

			output, isFinal, err := behavior(context.Background(), results)

			require.NoError(t, err)
			assert.True(t, isFinal)
			assert.Equal(t, "gamma output", output)
		},
	)

	t.Run(
		"no stop names means never final",
		func(t *testing.T) {
			t.Parallel()

			behavior := agent.StopAtTools()
			results := []agent.ToolCallResult{
				{
					ToolName: "anything",
					Result:   agent.ToolResult{Content: "value"},
				},
			}

			output, isFinal, err := behavior(context.Background(), results)

			require.NoError(t, err)
			assert.False(t, isFinal)
			assert.Equal(t, "", output)
		},
	)
}
