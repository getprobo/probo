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

package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/stretchr/testify/require"
)

type GraphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
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

func (c *Client) Execute(query string, variables map[string]any, result any) error {
	resp, err := c.Do(query, variables)
	if err != nil {
		return err
	}

	if result != nil && resp.Data != nil {
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

	if result != nil && resp.Data != nil {
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

type UploadFile struct {
	Filename    string
	ContentType string
	Content     []byte
}

func (c *Client) ExecuteWithFile(query string, variables map[string]any, variablePath string, file UploadFile, result any) error {
	return c.executeMultipart(query, variables, map[string]UploadFile{variablePath: file}, result)
}

func (c *Client) ExecuteWithFiles(query string, variables map[string]any, files map[string]UploadFile, result any) error {
	return c.executeMultipart(query, variables, files, result)
}

func (c *Client) executeMultipart(query string, variables map[string]any, files map[string]UploadFile, result any) error {
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
	req, err := http.NewRequest("POST", c.baseURL+"/api/console/v1/graphql", &buf)
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

	if result != nil && gqlResp.Data != nil {
		if err := json.Unmarshal(gqlResp.Data, result); err != nil {
			return fmt.Errorf("cannot unmarshal data: %w", err)
		}
	}

	return nil
}
