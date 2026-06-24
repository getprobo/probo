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

package cookiebanner

import (
	"context"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	searchPatternsParams struct {
		Query string `json:"query" jsonschema:"Search fragment to match against known cookie/tracker pattern names and descriptions (e.g. '_ga', 'matomo', 'facebook')"`
	}

	searchPatternsResult struct {
		Pattern        string  `json:"pattern"`
		Description    string  `json:"description"`
		TrackerType    string  `json:"tracker_type"`
		ThirdPartyName string  `json:"third_party_name,omitempty"`
		Confidence     float32 `json:"confidence"`
	}
)

func searchTrackerPatternsTool(pgClient *pg.Client) agent.Tool {
	return agent.FunctionTool(
		"search_tracker_patterns",
		"Search the internal database of known cookie and tracker patterns by name fragment or description keyword. Returns matching patterns with their linked third party name and confidence score. Use this first to find similar known patterns before resorting to web search.",
		func(ctx context.Context, p searchPatternsParams) (agent.ToolResult, error) {
			if p.Query == "" {
				return agent.ResultError("query is required"), nil
			}

			var out []searchPatternsResult

			if err := pgClient.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					var patterns coredata.CommonTrackerPatterns

					results, err := patterns.FindByKeyword(ctx, conn, p.Query, 10)
					if err != nil {
						return err
					}

					out = make([]searchPatternsResult, len(results))
					for i, r := range results {
						out[i] = searchPatternsResult{
							Pattern:     r.Pattern,
							Description: r.Description,
							TrackerType: string(r.TrackerType),
							Confidence:  r.Confidence,
						}
						if r.ThirdPartyName != nil {
							out[i].ThirdPartyName = *r.ThirdPartyName
						}
					}

					return nil
				},
			); err != nil {
				return agent.ResultErrorf("search failed: %s", err), nil
			}

			return agent.ResultJSON(out), nil
		},
	)
}
