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

type (
	hibpParams struct {
		Domain string `json:"domain" jsonschema:"The domain to check for known data breaches (e.g. example.com)"`
	}

	breach struct {
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

	hibpResult struct {
		Found       bool     `json:"found"`
		Count       int      `json:"count"`
		Breaches    []breach `json:"breaches,omitempty"`
		ErrorDetail string   `json:"error_detail,omitempty"`
	}
)

func CheckBreachesTool() agent.Tool {
	return agent.FunctionTool(
		"check_breaches",
		"Check if a domain has been involved in known data breaches using the Have I Been Pwned API.",
		func(ctx context.Context, p hibpParams) (agent.ToolResult, error) {
			client := &http.Client{Timeout: 10 * time.Second}

			hibpURL, err := url.Parse("https://haveibeenpwned.com/api/v3/breaches")
			if err != nil {
				return agent.ResultJSON(
					hibpResult{
						ErrorDetail: fmt.Sprintf("cannot parse HIBP URL: %s", err),
					},
				), nil
			}

			q := hibpURL.Query()
			q.Set("domain", p.Domain)
			hibpURL.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, hibpURL.String(), nil)
			if err != nil {
				return agent.ResultJSON(
					hibpResult{
						ErrorDetail: fmt.Sprintf("cannot create request: %s", err),
					},
				), nil
			}

			req.Header.Set("User-Agent", "Probo-Vendor-Assessment")

			resp, err := client.Do(req)
			if err != nil {
				return agent.ResultJSON(
					hibpResult{
						ErrorDetail: fmt.Sprintf("cannot fetch breaches: %s", err),
					},
				), nil
			}

			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return agent.ResultJSON(
					hibpResult{
						ErrorDetail: fmt.Sprintf("cannot read response: %s", err),
					},
				), nil
			}

			if resp.StatusCode == http.StatusNotFound {
				return agent.ResultJSON(hibpResult{Found: false, Count: 0}), nil
			}

			if resp.StatusCode != http.StatusOK {
				return agent.ResultJSON(
					hibpResult{
						ErrorDetail: fmt.Sprintf("HIBP API returned status %d", resp.StatusCode),
					},
				), nil
			}

			var breaches []breach
			if err := json.Unmarshal(body, &breaches); err != nil {
				return agent.ResultJSON(
					hibpResult{
						ErrorDetail: fmt.Sprintf("cannot parse response: %s", err),
					},
				), nil
			}

			return agent.ResultJSON(
				hibpResult{
					Found:    len(breaches) > 0,
					Count:    len(breaches),
					Breaches: breaches,
				},
			), nil
		},
	)
}
