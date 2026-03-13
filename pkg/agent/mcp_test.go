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
	"sync"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/llm"
)

func TestWithMCPServers_SkipsNilEntries(t *testing.T) {
	t.Parallel()

	valid := NewMCPServer("valid", nil)
	a := &Agent{}

	opt := WithMCPServers(nil, valid, nil)
	opt(a)

	require.Len(t, a.mcpServers, 1)
	assert.Equal(t, "valid", a.mcpServers[0].Name())
}

func TestNewMCPServer(t *testing.T) {
	t.Parallel()

	s := NewMCPServer("my-server", nil)

	require.NotNil(t, s)
	assert.Equal(t, "my-server", s.name)
	assert.Nil(t, s.session)
	assert.False(t, s.toolsCached)
	assert.Nil(t, s.cachedTools)
}

func TestMCPServer_Name(t *testing.T) {
	t.Parallel()

	s := &MCPServer{name: "test-server"}
	assert.Equal(t, "test-server", s.Name())
}

func TestMCPServer_Tools(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns cached tools without calling session",
		func(t *testing.T) {
			t.Parallel()

			tool := &mcpTool{
				name:        "cached-tool",
				description: "A cached tool",
				inputSchema: json.RawMessage(`{"type":"object"}`),
			}

			s := &MCPServer{name: "s"}
			s.cachedTools = []Tool{tool}
			s.toolsCached = true

			tools, err := s.Tools(context.Background())
			require.NoError(t, err)
			require.Len(t, tools, 1)
			assert.Equal(t, "cached-tool", tools[0].Name())
		},
	)

	t.Run(
		"returns a defensive copy of cached tools",
		func(t *testing.T) {
			t.Parallel()

			tool := &mcpTool{
				name:        "tool-a",
				description: "Tool A",
				inputSchema: json.RawMessage(`{}`),
			}

			s := &MCPServer{name: "s"}
			s.cachedTools = []Tool{tool}
			s.toolsCached = true

			tools1, err := s.Tools(context.Background())
			require.NoError(t, err)

			tools2, err := s.Tools(context.Background())
			require.NoError(t, err)

			// Mutating one slice must not affect the other.
			tools1[0] = nil
			assert.NotNil(t, tools2[0])

			// Underlying cache must be untouched.
			s.mu.RLock()
			assert.NotNil(t, s.cachedTools[0])
			s.mu.RUnlock()
		},
	)

	t.Run(
		"concurrent reads return consistent results",
		func(t *testing.T) {
			t.Parallel()

			s := &MCPServer{name: "s"}
			s.cachedTools = []Tool{
				&mcpTool{
					name:        "echo",
					description: "Echoes the input",
					inputSchema: json.RawMessage(`{"type":"object"}`),
				},
			}
			s.toolsCached = true

			const goroutines = 50
			var wg sync.WaitGroup
			wg.Add(goroutines)

			for range goroutines {
				go func() {
					defer wg.Done()

					tools, err := s.Tools(context.Background())
					require.NoError(t, err)
					assert.Len(t, tools, 1)
					assert.Equal(t, "echo", tools[0].Name())
				}()
			}

			wg.Wait()
		},
	)
}

func TestMCPServer_ResetCache(t *testing.T) {
	t.Parallel()

	t.Run(
		"clears cached tools",
		func(t *testing.T) {
			t.Parallel()

			s := &MCPServer{name: "s"}
			s.cachedTools = []Tool{
				&mcpTool{
					name:        "tool1",
					description: "A tool",
					inputSchema: json.RawMessage(`{"type":"object"}`),
				},
			}
			s.toolsCached = true

			tools, err := s.Tools(context.Background())
			require.NoError(t, err)
			require.Len(t, tools, 1)

			s.ResetCache()

			s.mu.RLock()
			assert.False(t, s.toolsCached)
			assert.Nil(t, s.cachedTools)
			s.mu.RUnlock()
		},
	)

	t.Run(
		"is safe to call on fresh server",
		func(t *testing.T) {
			t.Parallel()

			s := &MCPServer{name: "s"}
			s.ResetCache()

			s.mu.RLock()
			assert.False(t, s.toolsCached)
			assert.Nil(t, s.cachedTools)
			s.mu.RUnlock()
		},
	)

	t.Run(
		"concurrent resets do not race",
		func(t *testing.T) {
			t.Parallel()

			s := &MCPServer{name: "s"}
			s.cachedTools = []Tool{
				&mcpTool{
					name:        "tool1",
					description: "A tool",
					inputSchema: json.RawMessage(`{"type":"object"}`),
				},
			}
			s.toolsCached = true

			const goroutines = 50
			var wg sync.WaitGroup
			wg.Add(goroutines)

			for range goroutines {
				go func() {
					defer wg.Done()
					s.ResetCache()
				}()
			}

			wg.Wait()

			s.mu.RLock()
			assert.False(t, s.toolsCached)
			assert.Nil(t, s.cachedTools)
			s.mu.RUnlock()
		},
	)
}

func TestMCPTool_Name(t *testing.T) {
	t.Parallel()

	tool := &mcpTool{name: "get_weather"}
	assert.Equal(t, "get_weather", tool.Name())
}

func TestMCPTool_Definition(t *testing.T) {
	t.Parallel()

	schema := json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`)

	tool := &mcpTool{
		name:        "get_weather",
		description: "Returns current weather for a city",
		inputSchema: schema,
	}

	def := tool.Definition()

	assert.Equal(t, llm.Tool{
		Name:        "get_weather",
		Description: "Returns current weather for a city",
		Parameters:  schema,
	}, def)
}

func TestMCPTool_Execute(t *testing.T) {
	t.Parallel()

	t.Run(
		"invalid JSON arguments returns error result",
		func(t *testing.T) {
			t.Parallel()

			tool := &mcpTool{name: "my_tool"}

			result, err := tool.Execute(context.Background(), "not-json")
			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "Invalid arguments")
		},
	)

	t.Run(
		"malformed JSON arguments returns error result",
		func(t *testing.T) {
			t.Parallel()

			tool := &mcpTool{name: "my_tool"}

			result, err := tool.Execute(context.Background(), `{"key": }`)
			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "Invalid arguments")
		},
	)

	t.Run(
		"JSON array arguments returns error result",
		func(t *testing.T) {
			t.Parallel()

			tool := &mcpTool{name: "my_tool"}

			// Execute expects a JSON object, not an array.
			result, err := tool.Execute(context.Background(), `[1,2,3]`)
			require.NoError(t, err)
			assert.True(t, result.IsError)
			assert.Contains(t, result.Content, "Invalid arguments")
		},
	)
}

func TestExtractMCPContent(t *testing.T) {
	t.Parallel()

	t.Run(
		"nil result returns empty",
		func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "", extractMCPContent(nil))
		},
	)

	t.Run(
		"empty content returns empty",
		func(t *testing.T) {
			t.Parallel()
			result := &mcp.CallToolResult{Content: []mcp.Content{}}
			assert.Equal(t, "", extractMCPContent(result))
		},
	)

	t.Run(
		"single text content",
		func(t *testing.T) {
			t.Parallel()
			result := &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "hello world"},
				},
			}
			assert.Equal(t, "hello world", extractMCPContent(result))
		},
	)

	t.Run(
		"multiple text contents joined by newline",
		func(t *testing.T) {
			t.Parallel()
			result := &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "line one"},
					&mcp.TextContent{Text: "line two"},
				},
			}
			assert.Equal(t, "line one\nline two", extractMCPContent(result))
		},
	)

	t.Run(
		"non-text content is skipped",
		func(t *testing.T) {
			t.Parallel()
			result := &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "text part"},
					&mcp.ImageContent{Data: []byte("base64data"), MIMEType: "image/png"},
				},
			}
			assert.Equal(t, "text part", extractMCPContent(result))
		},
	)

	t.Run(
		"only non-text content returns empty",
		func(t *testing.T) {
			t.Parallel()
			result := &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{Data: []byte("base64data"), MIMEType: "image/png"},
				},
			}
			assert.Equal(t, "", extractMCPContent(result))
		},
	)

	t.Run(
		"falls back to structured content when no text content",
		func(t *testing.T) {
			t.Parallel()
			result := &mcp.CallToolResult{
				StructuredContent: map[string]any{
					"status": "ok",
					"count":  float64(42),
				},
			}
			got := extractMCPContent(result)
			assert.Contains(t, got, `"status":"ok"`)
			assert.Contains(t, got, `"count":42`)
		},
	)

	t.Run(
		"text content takes precedence over structured content",
		func(t *testing.T) {
			t.Parallel()
			result := &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "text wins"},
				},
				StructuredContent: map[string]any{"key": "value"},
			}
			assert.Equal(t, "text wins", extractMCPContent(result))
		},
	)

	t.Run(
		"structured content used when content has only non-text",
		func(t *testing.T) {
			t.Parallel()
			result := &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.ImageContent{Data: []byte("img"), MIMEType: "image/png"},
				},
				StructuredContent: map[string]any{"fallback": true},
			}
			assert.Equal(t, `{"fallback":true}`, extractMCPContent(result))
		},
	)
}
