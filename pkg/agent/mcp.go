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

	var allTools []*mcp.Tool
	var cursor string

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
