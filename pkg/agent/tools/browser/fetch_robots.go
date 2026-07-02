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

package browser

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/agent"
)

type (
	robotsParams struct {
		Domain string `json:"domain" jsonschema:"The domain to fetch robots.txt from (e.g. example.com)"`
	}

	robotsResult struct {
		Found       bool     `json:"found"`
		Sitemaps    []string `json:"sitemaps,omitempty"`
		Disallowed  []string `json:"disallowed_paths,omitempty"`
		ErrorDetail string   `json:"error_detail,omitempty"`
	}
)

func FetchRobotsTxtTool() agent.Tool {
	client := &http.Client{Timeout: 10 * time.Second}

	return agent.FunctionTool(
		"fetch_robots_txt",
		"Fetch and parse the robots.txt file for a domain. Returns sitemap URLs and disallowed paths, which can reveal hidden pages the crawler might miss.",
		func(ctx context.Context, p robotsParams) (agent.ToolResult, error) {
			if err := validatePublicDomain(p.Domain); err != nil {
				return agent.ResultJSON(
					robotsResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("domain not allowed: %s", err),
					},
				), nil
			}

			u := &url.URL{
				Scheme: "https",
				Host:   p.Domain,
				Path:   "/robots.txt",
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
			if err != nil {
				return agent.ResultJSON(
					robotsResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("cannot create request: %s", err),
					},
				), nil
			}

			resp, err := client.Do(req)
			if err != nil {
				return agent.ResultJSON(
					robotsResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("cannot fetch robots.txt: %s", err),
					},
				), nil
			}

			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return agent.ResultJSON(
					robotsResult{
						Found:       false,
						ErrorDetail: fmt.Sprintf("robots.txt returned status %d", resp.StatusCode),
					},
				), nil
			}

			var result robotsResult

			result.Found = true

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())

				// Directive names are case-insensitive but values
				// (URLs, paths) are case-sensitive, so extract the
				// original-case suffix from the raw line rather than
				// reading it off the lowercased copy.
				if after, ok := strings.CutPrefix(strings.ToLower(line), "sitemap:"); ok {
					result.Sitemaps = append(result.Sitemaps, strings.TrimSpace(line[len(line)-len(after):]))
				}

				if after, ok := strings.CutPrefix(strings.ToLower(line), "disallow:"); ok {
					path := strings.TrimSpace(line[len(line)-len(after):])
					if path != "" && len(result.Disallowed) < 50 {
						result.Disallowed = append(result.Disallowed, path)
					}
				}
			}

			return agent.ResultJSON(result), nil
		},
	)
}
