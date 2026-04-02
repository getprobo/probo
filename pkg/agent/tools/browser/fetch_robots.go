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

package browser

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/agent"
)

type robotsParams struct {
	Domain string `json:"domain" jsonschema:"The domain to fetch robots.txt from (e.g. example.com)"`
}

type robotsResult struct {
	Found       bool     `json:"found"`
	Sitemaps    []string `json:"sitemaps,omitempty"`
	Disallowed  []string `json:"disallowed_paths,omitempty"`
	ErrorDetail string   `json:"error_detail,omitempty"`
}

func FetchRobotsTxtTool() (agent.Tool, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	return agent.FunctionTool[robotsParams](
		"fetch_robots_txt",
		"Fetch and parse the robots.txt file for a domain. Returns sitemap URLs and disallowed paths, which can reveal hidden pages the crawler might miss.",
		func(ctx context.Context, p robotsParams) (agent.ToolResult, error) {
			if err := validatePublicDomain(p.Domain); err != nil {
				data, _ := json.Marshal(robotsResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("domain not allowed: %s", err),
				})
				return agent.ToolResult{Content: string(data), IsError: true}, nil
			}

			u := "https://" + p.Domain + "/robots.txt"

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if err != nil {
				data, _ := json.Marshal(robotsResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("cannot create request: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			resp, err := client.Do(req)
			if err != nil {
				data, _ := json.Marshal(robotsResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("cannot fetch robots.txt: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				data, _ := json.Marshal(robotsResult{
					Found:       false,
					ErrorDetail: fmt.Sprintf("robots.txt returned status %d", resp.StatusCode),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			var result robotsResult
			result.Found = true

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())

				if after, ok := strings.CutPrefix(strings.ToLower(line), "sitemap:"); ok {
					result.Sitemaps = append(result.Sitemaps, strings.TrimSpace(line[len(line)-len(after):]))
				}

				if after, ok := strings.CutPrefix(strings.ToLower(line), "disallow:"); ok {
					path := strings.TrimSpace(after)
					if path != "" && len(result.Disallowed) < 50 {
						result.Disallowed = append(result.Disallowed, path)
					}
				}
			}

			data, _ := json.Marshal(result)
			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
