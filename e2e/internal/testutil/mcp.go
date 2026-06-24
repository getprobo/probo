// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

// CreateOAuth2AccessToken creates a manual OAuth access token via the connect GraphQL API.
// It returns the raw bearer token string.
func (c *Client) CreateOAuth2AccessToken(name string, scopes []string) string {
	const query = `
		mutation($input: CreateOAuth2AccessTokenInput!) {
			createOAuth2AccessToken(input: $input) {
				token
			}
		}
	`

	var result struct {
		CreateOAuth2AccessToken struct {
			Token string `json:"token"`
		} `json:"createOAuth2AccessToken"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"name":      name,
			"expiresAt": time.Now().Add(90 * 24 * time.Hour).UTC().Format(time.RFC3339),
			"scopes":    scopes,
		},
	}, &result)
	require.NoError(c.T, err, "createOAuth2AccessToken failed")
	require.NotEmpty(c.T, result.CreateOAuth2AccessToken.Token, "OAuth access token is empty")

	return result.CreateOAuth2AccessToken.Token
}

// NewMCPClient creates an MCP client authenticated with an API key.
// It initializes an MCP session and stores the session ID.
func NewMCPClient(t require.TestingT, owner *Client) *MCPClient {
	return NewMCPClientWithAccessToken(t, owner, owner.CreateAPIKey("e2e-mcp-test"))
}

// NewMCPClientWithAccessToken creates an MCP client authenticated with a bearer token.
// It initializes an MCP session and stores the session ID.
func NewMCPClientWithAccessToken(t require.TestingT, owner *Client, accessToken string) *MCPClient {
	mc := &MCPClient{
		t:        t,
		baseURL:  owner.BaseURL() + "/mcp/v1",
		apiToken: accessToken,
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

// CallToolExpectToolError invokes an MCP tool and expects a tool-level error
// (isError: true in the result). It returns the error text content.
func (mc *MCPClient) CallToolExpectToolError(toolName string, args map[string]any) string {
	tr := mc.CallTool(toolName, args)
	require.True(mc.t, tr.IsError, "expected tool %s to return isError", toolName)
	require.NotEmpty(mc.t, tr.Content, "tool %s returned no content", toolName)

	var text string

	err := json.Unmarshal(tr.Content[0].Text, &text)
	require.NoError(mc.t, err, "cannot unmarshal error text for %s", toolName)

	return text
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
