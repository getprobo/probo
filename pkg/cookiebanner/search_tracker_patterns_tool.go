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
