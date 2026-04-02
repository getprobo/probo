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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.probo.inc/probo/pkg/agent"
)

type (
	searchParams struct {
		Query      string `json:"query" jsonschema:"The search query to execute"`
		MaxResults int    `json:"max_results" jsonschema:"Maximum number of results to return (default 5, max 10)"`
	}

	searchResult struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	}

	searxngResponse struct {
		Results []searxngResult `json:"results"`
	}

	searxngResult struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	}
)

// WebSearchTool creates a tool that searches the web using a SearXNG instance.
// The endpoint should be the base URL of the SearXNG instance (e.g.
// "http://localhost:8888").
func WebSearchTool(endpoint string) (agent.Tool, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	return agent.FunctionTool[searchParams](
		"web_search",
		"Search the web for information about a topic. Returns a list of results with title, URL, and snippet. Use this to find news, reviews, breach reports, regulatory actions, and other external information about a vendor.",
		func(ctx context.Context, p searchParams) (agent.ToolResult, error) {
			maxResults := p.MaxResults
			if maxResults <= 0 {
				maxResults = 5
			}
			if maxResults > 10 {
				maxResults = 10
			}

			u, err := url.Parse(endpoint + "/search")
			if err != nil {
				return agent.ToolResult{
					Content: fmt.Sprintf("invalid search endpoint: %s", err),
					IsError: true,
				}, nil
			}

			q := u.Query()
			q.Set("q", p.Query)
			q.Set("format", "json")
			q.Set("categories", "general")
			u.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			if err != nil {
				return agent.ToolResult{
					Content: fmt.Sprintf("cannot create search request: %s", err),
					IsError: true,
				}, nil
			}

			resp, err := client.Do(req)
			if err != nil {
				return agent.ToolResult{
					Content: fmt.Sprintf("search request failed: %s", err),
					IsError: true,
				}, nil
			}
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return agent.ToolResult{
					Content: fmt.Sprintf("cannot read search response: %s", err),
					IsError: true,
				}, nil
			}

			if resp.StatusCode != http.StatusOK {
				return agent.ToolResult{
					Content: fmt.Sprintf("search returned status %d: %s", resp.StatusCode, string(body)),
					IsError: true,
				}, nil
			}

			var searxResp searxngResponse
			if err := json.Unmarshal(body, &searxResp); err != nil {
				return agent.ToolResult{
					Content: fmt.Sprintf("cannot parse search response: %s", err),
					IsError: true,
				}, nil
			}

			results := make([]searchResult, 0, maxResults)
			for i, r := range searxResp.Results {
				if i >= maxResults {
					break
				}
				results = append(results, searchResult{
					Title:   r.Title,
					URL:     r.URL,
					Snippet: r.Content,
				})
			}

			data, _ := json.Marshal(results)

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
