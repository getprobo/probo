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

package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.probo.inc/probo/pkg/agent"
)

type (
	searchResult struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	}

	firecrawlParams struct {
		Query      string `json:"query" jsonschema:"The search query to execute"`
		MaxResults int    `json:"max_results" jsonschema:"Maximum number of results to return (default 5, max 10)"`
	}

	firecrawlRequest struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}

	firecrawlResponse struct {
		Success bool `json:"success"`
		Data    struct {
			Web []firecrawlWebResult `json:"web"`
		} `json:"data"`
	}

	firecrawlWebResult struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
	}
)

// FirecrawlSearchTool creates a tool that searches the web using the Firecrawl
// API. The endpoint should be the base URL of the Firecrawl instance (e.g.
// "https://api.firecrawl.dev/v2"). The apiKey is used for Bearer authentication.
func FirecrawlSearchTool(endpoint, apiKey string) agent.Tool {
	client := newHTTPClient()

	return agent.FunctionTool(
		"web_search",
		"Search the web for information about a topic. Returns a list of results with title, URL, and snippet. Use this to find news, reviews, breach reports, regulatory actions, and other external information about a vendor.",
		func(ctx context.Context, p firecrawlParams) (agent.ToolResult, error) {
			maxResults := p.MaxResults
			if maxResults <= 0 {
				maxResults = 5
			}
			if maxResults > 10 {
				maxResults = 10
			}

			results, err := firecrawlSearch(ctx, client, endpoint, apiKey, p.Query, maxResults)
			if err != nil {
				return agent.ResultErrorf("search request failed: %s", err), nil
			}

			return agent.ResultJSON(results), nil
		},
	)
}

func firecrawlSearch(
	ctx context.Context,
	client *http.Client,
	endpoint, apiKey, query string,
	maxResults int,
) ([]searchResult, error) {
	u, err := url.JoinPath(endpoint, "search")
	if err != nil {
		return nil, fmt.Errorf("cannot build search URL: %w", err)
	}

	body, err := json.Marshal(
		firecrawlRequest{
			Query: query,
			Limit: maxResults,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot execute search request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	var fcResp firecrawlResponse
	if err := json.Unmarshal(respBody, &fcResp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal response: %w", err)
	}

	if !fcResp.Success {
		return nil, fmt.Errorf("search returned success=false")
	}

	results := make([]searchResult, 0, len(fcResp.Data.Web))
	for _, r := range fcResp.Data.Web {
		results = append(
			results,
			searchResult{
				Title:   r.Title,
				URL:     r.URL,
				Snippet: r.Description,
			},
		)
	}

	return results, nil
}
