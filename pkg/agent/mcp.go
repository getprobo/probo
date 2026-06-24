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
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.probo.inc/probo/pkg/llm"
)

type (
	MCPServer struct {
		name    string
		session *mcp.ClientSession

		mu          sync.RWMutex
		cachedTools []Tool
		toolsCached bool
	}

	mcpTool struct {
		server      *MCPServer
		name        string
		description string
		inputSchema json.RawMessage
	}
)

func NewMCPServer(name string, session *mcp.ClientSession) *MCPServer {
	return &MCPServer{
		name:    name,
		session: session,
	}
}

func (s *MCPServer) Name() string {
	return s.name
}

func (s *MCPServer) Tools(ctx context.Context) ([]Tool, error) {
	s.mu.RLock()

	if s.toolsCached {
		cp := make([]Tool, len(s.cachedTools))
		copy(cp, s.cachedTools)
		s.mu.RUnlock()

		return cp, nil
	}

	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.toolsCached {
		cp := make([]Tool, len(s.cachedTools))
		copy(cp, s.cachedTools)

		return cp, nil
	}

	var (
		allTools []*mcp.Tool
		cursor   string
	)

	for {
		params := &mcp.ListToolsParams{}
		if cursor != "" {
			params.Cursor = cursor
		}

		result, err := s.session.ListTools(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("cannot list tools from MCP server %q: %w", s.name, err)
		}

		allTools = append(allTools, result.Tools...)

		if result.NextCursor == "" {
			break
		}

		cursor = result.NextCursor
	}

	tools := make([]Tool, len(allTools))
	for i, t := range allTools {
		schema, err := json.Marshal(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal input schema for tool %q: %w", t.Name, err)
		}

		tools[i] = &mcpTool{
			server:      s,
			name:        t.Name,
			description: t.Description,
			inputSchema: schema,
		}
	}

	s.cachedTools = tools
	s.toolsCached = true

	return tools, nil
}

// ResetCache clears the cached tool definitions, forcing the next call to
// Tools to re-fetch from the MCP server.
func (s *MCPServer) ResetCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cachedTools = nil
	s.toolsCached = false
}

func (t *mcpTool) Name() string { return t.name }

func (t *mcpTool) Definition() llm.Tool {
	return llm.Tool{
		Name:        t.name,
		Description: t.description,
		Parameters:  t.inputSchema,
	}
}

func (t *mcpTool) Execute(ctx context.Context, arguments string) (ToolResult, error) {
	var args map[string]any
	if arguments != "" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return ToolResult{
				Content: fmt.Sprintf("Invalid arguments: %s", err.Error()),
				IsError: true,
			}, nil
		}
	}

	result, err := t.server.session.CallTool(
		ctx,
		&mcp.CallToolParams{
			Name:      t.name,
			Arguments: args,
		},
	)
	if err != nil {
		return ToolResult{}, fmt.Errorf("cannot call MCP tool %q: %w", t.name, err)
	}

	content := extractMCPContent(result)

	return ToolResult{
		Content: content,
		IsError: result.IsError,
	}, nil
}

func extractMCPContent(result *mcp.CallToolResult) string {
	if result == nil {
		return ""
	}

	var parts []string

	for _, c := range result.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}

	if len(parts) > 0 {
		return strings.Join(parts, "\n")
	}

	if result.StructuredContent != nil {
		data, err := json.Marshal(result.StructuredContent)
		if err == nil {
			return string(data)
		}
	}

	return ""
}
