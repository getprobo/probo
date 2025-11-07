// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package v1

import (
	"context"
	"time"

	"go.probo.inc/probo/pkg/gid"
)

type (
	// Config holds configuration for the MCP server
	Config struct {
		// Version is the MCP server version
		Version string
		// RequestTimeout is the maximum duration for a request
		RequestTimeout time.Duration
		// MaxRequestSize is the maximum size of a request body in bytes
		MaxRequestSize int64
	}

	// MCPContext holds the authenticated context for MCP requests
	MCPContext struct {
		UserID    gid.GID
		TenantIDs []gid.TenantID
	}

	ctxKey struct{ name string }
)

var (
	mcpContextKey = &ctxKey{name: "mcp_context"}
)

// MCPContextFromContext extracts the MCP context from the request context
func MCPContextFromContext(ctx context.Context) *MCPContext {
	mcpCtx, _ := ctx.Value(mcpContextKey).(*MCPContext)
	return mcpCtx
}

// ContextWithMCPContext adds the MCP context to the request context
func ContextWithMCPContext(ctx context.Context, mcpCtx *MCPContext) context.Context {
	return context.WithValue(ctx, mcpContextKey, mcpCtx)
}

// DefaultConfig returns a default MCP configuration
func DefaultConfig() Config {
	return Config{
		Version:        "1.0.0",
		RequestTimeout: 30 * time.Second,
		MaxRequestSize: 10 * 1024 * 1024, // 10MB
	}
}
