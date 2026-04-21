// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stretchr/testify/require"
)

// MCPClient wraps an authenticated MCP session for e2e testing.
type MCPClient struct {
	t          require.TestingT
	baseURL    string
	apiToken   string
	sessionID  string
	httpClient *http.Client
}

// CreateAPIKey creates a personal API key via the connect GraphQL API.
// It returns the raw bearer token string.
func (c *Client) CreateAPIKey(name string) string {
	const query = `
		mutation($input: CreatePersonalAPIKeyInput!) {
			createPersonalAPIKey(input: $input) {
				token
			}
		}
	`

	var result struct {
		CreatePersonalAPIKey struct {
			Token string `json:"token"`
		} `json:"createPersonalAPIKey"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"name":      name,
			"expiresAt": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		},
	}, &result)
	require.NoError(c.T, err, "createPersonalAPIKey failed")
	require.NotEmpty(c.T, result.CreatePersonalAPIKey.Token, "API key token is empty")

	return result.CreatePersonalAPIKey.Token
}

// NewMCPClient creates an MCP client authenticated with an API key.
// It initializes an MCP session and stores the session ID.
func NewMCPClient(t require.TestingT, owner *Client) *MCPClient {
	token := owner.CreateAPIKey("e2e-mcp-test")

	mc := &MCPClient{
		t:        t,
		baseURL:  owner.BaseURL() + "/mcp/v1",
		apiToken: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	mc.initialize()

	return mc
}

// jsonrpcRequest is a JSON-RPC 2.0 request.
type jsonrpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// jsonrpcResponse is a JSON-RPC 2.0 response.
type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *jsonrpcError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

func (mc *MCPClient) doRequest(method string, params any) (json.RawMessage, error) {
	reqBody := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", mc.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("Authorization", "Bearer "+mc.apiToken)
	if mc.sessionID != "" {
		req.Header.Set("Mcp-Session-Id", mc.sessionID)
	}

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	// Store session ID from response
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		mc.sessionID = sid
	}

	var rpcResp jsonrpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %w (body: %s)", err, string(respBody))
	}

	if rpcResp.Error != nil {
		return nil, rpcResp.Error
	}

	return rpcResp.Result, nil
}

func (mc *MCPClient) initialize() {
	result, err := mc.doRequest("initialize", map[string]any{
		"protocolVersion": "2025-03-26",
		"capabilities":    map[string]any{},
		"clientInfo": map[string]any{
			"name":    "probo-e2e-test",
			"version": "1.0.0",
		},
	})
	require.NoError(mc.t, err, "MCP initialize failed")
	require.NotNil(mc.t, result, "MCP initialize returned nil result")
}

// MCPToolResult represents the result of a tools/call response.
type MCPToolResult struct {
	Content []MCPContent `json:"content"`
	IsError bool         `json:"isError"`
}

// MCPContent represents content within a tool result.
type MCPContent struct {
	Type string          `json:"type"`
	Text json.RawMessage `json:"text"`
}

// CallTool invokes an MCP tool and returns the parsed result.
func (mc *MCPClient) CallTool(toolName string, args map[string]any) *MCPToolResult {
	result, err := mc.doRequest("tools/call", map[string]any{
		"name":      toolName,
		"arguments": args,
	})
	require.NoError(mc.t, err, "MCP tools/call %s failed", toolName)

	var toolResult MCPToolResult
	err = json.Unmarshal(result, &toolResult)
	require.NoError(mc.t, err, "cannot unmarshal tool result for %s", toolName)

	return &toolResult
}

// CallToolInto invokes an MCP tool and unmarshals the first text content into dest.
func (mc *MCPClient) CallToolInto(toolName string, args map[string]any, dest any) {
	tr := mc.CallTool(toolName, args)
	require.False(mc.t, tr.IsError, "tool %s returned error: %v", toolName, tr.Content)
	require.NotEmpty(mc.t, tr.Content, "tool %s returned no content", toolName)

	// The text field in MCP content is a JSON-encoded string of the output.
	// First unmarshal the raw JSON to get the string.
	var textStr string
	err := json.Unmarshal(tr.Content[0].Text, &textStr)
	require.NoError(mc.t, err, "cannot unmarshal text content for %s", toolName)

	// Then unmarshal that string as JSON into the destination.
	err = json.Unmarshal([]byte(textStr), dest)
	require.NoError(mc.t, err, "cannot unmarshal tool output for %s", toolName)
}
