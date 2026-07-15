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
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// trustCenterHTTPSAddr is the loopback address of the dedicated trust-center
// HTTPS listener started by the e2e probod (see generateConfig). Compliance
// pages are served here exclusively, routed by TLS SNI / Host header.
const trustCenterHTTPSAddr = "127.0.0.1:443"

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// DataString returns the GraphQL data payload as JSON text. Absent and JSON null
// responses both normalize to an empty string for assert.Empty checks.
func (r *GraphQLResponse) DataString() string {
	if len(r.Data) == 0 || string(r.Data) == "null" {
		return ""
	}

	return string(r.Data)
}

type GraphQLError struct {
	Message    string         `json:"message"`
	Path       []any          `json:"path,omitempty"`
	Extensions map[string]any `json:"extensions,omitempty"`
}

func (e GraphQLError) Error() string {
	return e.Message
}

func (e GraphQLError) Code() string {
	if e.Extensions == nil {
		return ""
	}

	if code, ok := e.Extensions["code"].(string); ok {
		return code
	}

	return ""
}

type GraphQLErrors []GraphQLError

func (e GraphQLErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	if len(e) == 1 {
		return e[0].Message
	}

	return fmt.Sprintf("%s (and %d more errors)", e[0].Message, len(e)-1)
}

func (c *Client) doWithEndpoint(endpoint string, query string, variables map[string]any) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
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

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return &gqlResp, GraphQLErrors(gqlResp.Errors)
	}

	return &gqlResp, nil
}

func (c *Client) Do(query string, variables map[string]any) (*GraphQLResponse, error) {
	return c.doWithEndpoint("/api/console/v1/graphql", query, variables)
}

func (c *Client) DoConnect(query string, variables map[string]any) (*GraphQLResponse, error) {
	return c.doWithEndpoint("/api/connect/v1/graphql", query, variables)
}

// ConsoleGraphQLWithAccessToken posts to the console GraphQL endpoint using a
// bearer access token and no session cookies.
func ConsoleGraphQLWithAccessToken(
	t testing.TB,
	accessToken string,
	query string,
	variables map[string]any,
) (*GraphQLResponse, error) {
	t.Helper()

	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		GetBaseURL()+"/api/console/v1/graphql",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
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

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return &gqlResp, GraphQLErrors(gqlResp.Errors)
	}

	return &gqlResp, nil
}

// trustHTTPClient builds an HTTP client that always dials the dedicated
// trust-center HTTPS listener on loopback while presenting the compliance
// page's host as TLS SNI. Certificates are step-ca-issued for e2e, so
// verification is skipped when the root is not installed in the test runner.
func trustHTTPClient(serverName string) *http.Client {
	return trustHTTPClientWithJar(serverName, nil)
}

func trustHTTPClientWithJar(serverName string, jar http.CookieJar) *http.Client {
	dialer := &net.Dialer{Timeout: 5 * time.Second}

	return &http.Client{
		Jar:     jar,
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, "tcp", trustCenterHTTPSAddr)
			},
			TLSClientConfig: &tls.Config{
				ServerName:         serverName,
				InsecureSkipVerify: true, //nolint:gosec // e2e talks to step-ca-issued certs on loopback.
			},
		},
	}
}

// WaitForTrustCenterHTTPS blocks until the dedicated trust-center listener
// serves the page over TLS. Managed domains provision certificates
// asynchronously after activation.
func WaitForTrustCenterHTTPS(t testing.TB, host string) {
	t.Helper()

	client := TrustHTTPClient(host)

	require.Eventually(
		t,
		func() bool {
			resp, err := client.Get("https://" + host + complianceportalOAuthMetadataPath())
			if err != nil {
				return false
			}

			defer func() { _ = resp.Body.Close() }()

			return resp.StatusCode == http.StatusOK
		},
		30*time.Second,
		500*time.Millisecond,
		"trust center did not become servable on the dedicated listener",
	)
}

func complianceportalOAuthMetadataPath() string {
	return "/.well-known/oauth-client-metadata"
}

// DoTrust posts a GraphQL query to a compliance page served on the dedicated
// listener. host is the page's serving domain (a customer custom domain or a
// managed {slug}.probopage.localhost subdomain).
func (c *Client) DoTrust(host string, query string, variables map[string]any) (*GraphQLResponse, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request: %w", err)
	}

	endpoint := fmt.Sprintf("https://%s/graphql", host)

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := trustHTTPClient(host)
	if c.trustClient != nil && host == c.trustHost {
		client = c.trustClient
	}

	resp, err := client.Do(req)
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

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("cannot decode response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return &gqlResp, GraphQLErrors(gqlResp.Errors)
	}

	return &gqlResp, nil
}

func (c *Client) Execute(query string, variables map[string]any, result any) error {
	resp, err := c.Do(query, variables)
	if err != nil {
		return err
	}

	if result != nil && resp.DataString() != "" {
		if err := json.Unmarshal(resp.Data, result); err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}
	}

	return nil
}

func (c *Client) ExecuteConnect(query string, variables map[string]any, result any) error {
	resp, err := c.DoConnect(query, variables)
	if err != nil {
		return err
	}

	if result != nil && resp.DataString() != "" {
		if err := json.Unmarshal(resp.Data, result); err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}
	}

	return nil
}

func (c *Client) ExecuteTrust(host string, query string, variables map[string]any, result any) error {
	resp, err := c.DoTrust(host, query, variables)
	if err != nil {
		return err
	}

	if result != nil && resp.DataString() != "" {
		if err := json.Unmarshal(resp.Data, result); err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}
	}

	return nil
}

func (c *Client) MustExecute(query string, variables map[string]any, result any) {
	c.T.Helper()
	err := c.Execute(query, variables, result)
	require.NoError(c.T, err, "GraphQL request failed")
}

func (c *Client) ExecuteShouldFail(query string, variables map[string]any) error {
	c.T.Helper()
	_, err := c.Do(query, variables)
	require.Error(c.T, err, "expected GraphQL request to fail but it succeeded")

	return err
}

func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func TrustHTTPClient(trustHost string) *http.Client {
	return trustHTTPClient(trustHost)
}

type UploadFile struct {
	Filename    string
	ContentType string
	Content     []byte
}

func (c *Client) ExecuteWithFile(query string, variables map[string]any, variablePath string, file UploadFile, result any) error {
	return c.executeMultipart("/api/console/v1/graphql", query, variables, map[string]UploadFile{variablePath: file}, result)
}

func (c *Client) ExecuteConnectWithFile(query string, variables map[string]any, variablePath string, file UploadFile, result any) error {
	return c.executeMultipart("/api/connect/v1/graphql", query, variables, map[string]UploadFile{variablePath: file}, result)
}

func (c *Client) ExecuteWithFiles(query string, variables map[string]any, files map[string]UploadFile, result any) error {
	return c.executeMultipart("/api/console/v1/graphql", query, variables, files, result)
}

func (c *Client) executeMultipart(endpoint string, query string, variables map[string]any, files map[string]UploadFile, result any) error {
	// Create multipart writer using standard library
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	// Build the operations JSON
	operations := map[string]any{
		"query":     query,
		"variables": variables,
	}

	operationsJSON, err := json.Marshal(operations)
	if err != nil {
		return fmt.Errorf("cannot marshal operations: %w", err)
	}

	// Add operations part
	if err := writer.WriteField("operations", string(operationsJSON)); err != nil {
		return fmt.Errorf("cannot write operations field: %w", err)
	}

	// Build the map for file variables (sorted for deterministic order)
	fileMap := make(map[string][]string)

	fileOrder := make([]string, 0, len(files))
	for path := range files {
		fileOrder = append(fileOrder, path)
	}

	// Sort for deterministic ordering
	for i, path := range fileOrder {
		fileMap[fmt.Sprintf("%d", i)] = []string{"variables." + path}
	}

	mapJSON, err := json.Marshal(fileMap)
	if err != nil {
		return fmt.Errorf("cannot marshal map: %w", err)
	}

	// Add map part
	if err := writer.WriteField("map", string(mapJSON)); err != nil {
		return fmt.Errorf("cannot write map field: %w", err)
	}

	// Add file parts
	for i, path := range fileOrder {
		file := files[path]
		fieldName := fmt.Sprintf("%d", i)

		// Create form file part with proper headers
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, file.Filename))
		h.Set("Content-Type", file.ContentType)

		part, err := writer.CreatePart(h)
		if err != nil {
			return fmt.Errorf("cannot create file part %s: %w", path, err)
		}

		if _, err := part.Write(file.Content); err != nil {
			return fmt.Errorf("cannot write file content %s: %w", path, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("cannot close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", c.baseURL+endpoint, &buf)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp GraphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return fmt.Errorf("cannot decode response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return GraphQLErrors(gqlResp.Errors)
	}

	if result != nil && gqlResp.DataString() != "" {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}
	}

	return nil
}
