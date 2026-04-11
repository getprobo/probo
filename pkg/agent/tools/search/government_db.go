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
	govDBParams struct {
		CompanyName string `json:"company_name" jsonschema:"The company name to search for in government databases"`
		Domain      string `json:"domain" jsonschema:"The company domain for additional search context (optional)"`
	}

	govDBEntry struct {
		Source  string `json:"source"`
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet,omitempty"`
	}

	govDBResult struct {
		SECFilings   []govDBEntry `json:"sec_filings,omitempty"`
		FTCActions   []govDBEntry `json:"ftc_actions,omitempty"`
		GDPRFines    []govDBEntry `json:"gdpr_fines,omitempty"`
		OtherActions []govDBEntry `json:"other_regulatory_actions,omitempty"`
		ErrorDetail  string       `json:"error_detail,omitempty"`
	}
)

func CheckGovernmentDBTool(searchEndpoint string) (agent.Tool, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	return agent.FunctionTool(
		"check_government_databases",
		"Search government and regulatory databases for enforcement actions, SEC filings, FTC actions, and GDPR fines related to a company.",
		func(ctx context.Context, p govDBParams) (agent.ToolResult, error) {
			var result govDBResult

			name := p.CompanyName
			if p.Domain != "" {
				name = name + " " + p.Domain
			}

			type searchSpec struct {
				query  string
				source string
				target *[]govDBEntry
			}

			searches := []searchSpec{
				{
					query:  fmt.Sprintf(`site:sec.gov "%s"`, p.CompanyName),
					source: "SEC",
					target: &result.SECFilings,
				},
				{
					query:  fmt.Sprintf(`site:ftc.gov "%s"`, p.CompanyName),
					source: "FTC",
					target: &result.FTCActions,
				},
				{
					query:  fmt.Sprintf(`site:enforcementtracker.com "%s"`, p.CompanyName),
					source: "GDPR Enforcement Tracker",
					target: &result.GDPRFines,
				},
				{
					query:  fmt.Sprintf(`"%s" regulatory action OR enforcement OR fine OR penalty OR sanction`, name),
					source: "General",
					target: &result.OtherActions,
				},
			}

			for _, s := range searches {
				entries, err := searxngSearch(ctx, client, searchEndpoint, s.query, 3)
				if err != nil {
					continue
				}
				for _, e := range entries {
					*s.target = append(*s.target, govDBEntry{
						Source:  s.source,
						Title:   e.Title,
						URL:     e.URL,
						Snippet: e.Snippet,
					})
				}
			}

			return agent.ResultJSON(result), nil
		},
	)
}

func searxngSearch(ctx context.Context, client *http.Client, endpoint, query string, maxResults int) ([]searchResult, error) {
	u, err := url.Parse(endpoint + "/search")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("q", query)
	q.Set("format", "json")
	q.Set("categories", "general")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	var searxResp searxngResponse
	if err := json.Unmarshal(body, &searxResp); err != nil {
		return nil, err
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

	return results, nil
}
