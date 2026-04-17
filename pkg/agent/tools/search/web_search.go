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
	"net/http"
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

	return agent.FunctionTool(
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

			results, err := searxngSearch(ctx, client, endpoint, p.Query, maxResults)
			if err != nil {
				return agent.ResultErrorf("search request failed: %s", err), nil
			}

			return agent.ResultJSON(results), nil
		},
	)
}
