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

package mcputils_test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/server/api/mcp/mcputils"
)

func TestInstructionsMiddleware_FillsInitializeAndDiscover(t *testing.T) {
	t.Parallel()

	middleware := mcputils.InstructionsMiddleware("read documents with readDocument")
	handler := middleware(func(context.Context, string, mcp.Request) (mcp.Result, error) {
		return &mcp.InitializeResult{}, nil
	})

	result, err := handler(context.Background(), "initialize", &mcp.InitializeRequest{})
	require.NoError(t, err)
	assert.Equal(t, "read documents with readDocument", result.(*mcp.InitializeResult).Instructions)

	discover := middleware(func(context.Context, string, mcp.Request) (mcp.Result, error) {
		return &mcp.DiscoverResult{Instructions: "already set"}, nil
	})
	discovered, err := discover(context.Background(), "discover", &mcp.InitializeRequest{})
	require.NoError(t, err)
	assert.Equal(t, "already set", discovered.(*mcp.DiscoverResult).Instructions)
}

func TestRestrictTools_KeepsAllowlistedNames(t *testing.T) {
	t.Parallel()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "readDocument"}, func(context.Context, *mcp.CallToolRequest, map[string]any) (*mcp.CallToolResult, map[string]any, error) {
		return nil, map[string]any{}, nil
	})
	mcp.AddTool(server, &mcp.Tool{Name: "listRisks"}, func(context.Context, *mcp.CallToolRequest, map[string]any) (*mcp.CallToolResult, map[string]any, error) {
		return nil, map[string]any{}, nil
	})

	err := mcputils.RestrictTools(server, func(name string) bool {
		return name == "readDocument"
	})
	require.NoError(t, err)

	names, err := mcputils.RegisteredToolNames(server)
	require.NoError(t, err)
	assert.Equal(t, []string{"readDocument"}, names)
}
