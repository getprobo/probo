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

type waybackParams struct {
	URL string `json:"url" jsonschema:"The URL to check in the Wayback Machine (e.g. https://example.com/privacy)"`
}

type waybackSnapshot struct {
	Timestamp string `json:"timestamp"`
	URL       string `json:"url"`
}

type waybackResult struct {
	Available      bool             `json:"available"`
	OldestSnapshot *waybackSnapshot `json:"oldest_snapshot,omitempty"`
	NewestSnapshot *waybackSnapshot `json:"newest_snapshot,omitempty"`
	ErrorDetail    string           `json:"error_detail,omitempty"`
}

type waybackAvailabilityResponse struct {
	ArchivedSnapshots struct {
		Closest struct {
			Available bool   `json:"available"`
			URL       string `json:"url"`
			Timestamp string `json:"timestamp"`
		} `json:"closest"`
	} `json:"archived_snapshots"`
}

type waybackCDXResponse = [][]string

func CheckWaybackTool() (agent.Tool, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	return agent.FunctionTool[waybackParams](
		"check_wayback",
		"Check the Internet Archive Wayback Machine for archived versions of a URL. Useful for detecting changes in privacy policies, trust pages, or terms of service over time.",
		func(ctx context.Context, p waybackParams) (agent.ToolResult, error) {
			var result waybackResult

			// Check availability.
			availURL := "https://archive.org/wayback/available?url=" + url.QueryEscape(p.URL)
			body, err := httpGet(ctx, client, availURL)
			if err != nil {
				result.ErrorDetail = fmt.Sprintf("cannot check Wayback Machine availability: %s", err)
				data, _ := json.Marshal(result)
				return agent.ToolResult{Content: string(data)}, nil
			}

			var avail waybackAvailabilityResponse
			if err := json.Unmarshal(body, &avail); err == nil {
				result.Available = avail.ArchivedSnapshots.Closest.Available
			}

			if !result.Available {
				data, _ := json.Marshal(result)
				return agent.ToolResult{Content: string(data)}, nil
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

			data, _ := json.Marshal(result)
			return agent.ToolResult{Content: string(data)}, nil
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
