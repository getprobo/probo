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

package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"go.probo.inc/probo/pkg/agent"
)

type (
	waybackParams struct {
		URL string `json:"url" jsonschema:"The URL to check in the Wayback Machine (e.g. https://example.com/privacy)"`
	}

	waybackSnapshot struct {
		Timestamp string `json:"timestamp"`
		URL       string `json:"url"`
	}

	waybackResult struct {
		Available      bool             `json:"available"`
		OldestSnapshot *waybackSnapshot `json:"oldest_snapshot,omitempty"`
		NewestSnapshot *waybackSnapshot `json:"newest_snapshot,omitempty"`
		ErrorDetail    string           `json:"error_detail,omitempty"`
	}

	waybackAvailabilityResponse struct {
		ArchivedSnapshots struct {
			Closest struct {
				Available bool   `json:"available"`
				URL       string `json:"url"`
				Timestamp string `json:"timestamp"`
			} `json:"closest"`
		} `json:"archived_snapshots"`
	}

	waybackCDXResponse = [][]string
)

func CheckWaybackTool() agent.Tool {
	client := newHTTPClient()

	return agent.FunctionTool(
		"check_wayback",
		"Check the Internet Archive Wayback Machine for archived versions of a URL. Useful for detecting changes in privacy policies, trust pages, or terms of service over time.",
		func(ctx context.Context, p waybackParams) (agent.ToolResult, error) {
			var result waybackResult

			// Check availability.
			availURL, err := url.Parse("https://archive.org/wayback/available")
			if err != nil {
				result.ErrorDetail = fmt.Sprintf("cannot parse Wayback Machine URL: %s", err)
				return agent.ResultJSON(result), nil
			}

			q := availURL.Query()
			q.Set("url", p.URL)
			availURL.RawQuery = q.Encode()

			body, err := httpGet(ctx, client, availURL.String())
			if err != nil {
				result.ErrorDetail = fmt.Sprintf("cannot check Wayback Machine availability: %s", err)
				return agent.ResultJSON(result), nil
			}

			var avail waybackAvailabilityResponse
			if err := json.Unmarshal(body, &avail); err == nil {
				result.Available = avail.ArchivedSnapshots.Closest.Available
			}

			if !result.Available {
				return agent.ResultJSON(result), nil
			}

			// Get oldest snapshot.
			oldestURL := fmt.Sprintf(
				"https://web.archive.org/cdx/search/cdx?url=%s&output=json&fl=timestamp,original&limit=1",
				url.QueryEscape(p.URL),
			)
			if body, err := httpGet(ctx, client, oldestURL); err == nil {
				if snap := parseCDXSnapshot(body); snap != nil {
					result.OldestSnapshot = snap
				}
			}

			// Get newest snapshot.
			newestURL := fmt.Sprintf(
				"https://web.archive.org/cdx/search/cdx?url=%s&output=json&fl=timestamp,original&limit=1&sort=reverse",
				url.QueryEscape(p.URL),
			)
			if body, err := httpGet(ctx, client, newestURL); err == nil {
				if snap := parseCDXSnapshot(body); snap != nil {
					result.NewestSnapshot = snap
				}
			}

			return agent.ResultJSON(result), nil
		},
	)
}

func httpGet(ctx context.Context, client *http.Client, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024))
}

func parseCDXSnapshot(body []byte) *waybackSnapshot {
	var rows waybackCDXResponse
	if err := json.Unmarshal(body, &rows); err != nil || len(rows) < 2 {
		return nil
	}

	// First row is headers ["timestamp", "original"], data starts at row 1.
	row := rows[1]
	if len(row) < 2 {
		return nil
	}

	return &waybackSnapshot{
		Timestamp: row[0],
		URL:       row[1],
	}
}
