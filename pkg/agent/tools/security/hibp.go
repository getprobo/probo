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

package security

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

type hibpParams struct {
	Domain string `json:"domain" jsonschema:"description=The domain to check for known data breaches (e.g. example.com)"`
}

type breach struct {
	Name         string   `json:"Name"`
	BreachDate   string   `json:"BreachDate"`
	PwnCount     int      `json:"PwnCount"`
	DataClasses  []string `json:"DataClasses"`
	Description  string   `json:"Description"`
	IsVerified   bool     `json:"IsVerified"`
	IsSensitive  bool     `json:"IsSensitive"`
	IsRetired    bool     `json:"IsRetired"`
	IsSpamList   bool     `json:"IsSpamList"`
	IsMalware    bool     `json:"IsMalware"`
	IsSubscFree  bool     `json:"IsSubscriptionFree"`
	IsFabricated bool     `json:"IsFabricated"`
}

type hibpResult struct {
	Found       bool     `json:"found"`
	Count       int      `json:"count"`
	Breaches    []breach `json:"breaches,omitempty"`
	ErrorDetail string   `json:"error_detail,omitempty"`
}

func CheckBreachesTool() (agent.Tool, error) {
	return agent.FunctionTool[hibpParams](
		"check_breaches",
		"Check if a domain has been involved in known data breaches using the Have I Been Pwned API.",
		func(ctx context.Context, p hibpParams) (agent.ToolResult, error) {
			client := &http.Client{Timeout: 10 * time.Second}

			req, err := http.NewRequestWithContext(
				ctx,
				http.MethodGet,
				"https://haveibeenpwned.com/api/v3/breaches?domain="+url.QueryEscape(p.Domain),
				nil,
			)
			if err != nil {
				data, _ := json.Marshal(hibpResult{
					ErrorDetail: fmt.Sprintf("cannot create request: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			req.Header.Set("User-Agent", "Probo-Vendor-Assessment")

			resp, err := client.Do(req)
			if err != nil {
				data, _ := json.Marshal(hibpResult{
					ErrorDetail: fmt.Sprintf("cannot fetch breaches: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				data, _ := json.Marshal(hibpResult{
					ErrorDetail: fmt.Sprintf("cannot read response: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			if resp.StatusCode == http.StatusNotFound {
				data, _ := json.Marshal(hibpResult{Found: false, Count: 0})
				return agent.ToolResult{Content: string(data)}, nil
			}

			if resp.StatusCode != http.StatusOK {
				data, _ := json.Marshal(hibpResult{
					ErrorDetail: fmt.Sprintf("HIBP API returned status %d", resp.StatusCode),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			var breaches []breach
			if err := json.Unmarshal(body, &breaches); err != nil {
				data, _ := json.Marshal(hibpResult{
					ErrorDetail: fmt.Sprintf("cannot parse response: %s", err),
				})
				return agent.ToolResult{Content: string(data)}, nil
			}

			data, _ := json.Marshal(hibpResult{
				Found:    len(breaches) > 0,
				Count:    len(breaches),
				Breaches: breaches,
			})

			return agent.ToolResult{Content: string(data)}, nil
		},
	)
}
