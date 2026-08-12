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
	searchThirdPartiesParams struct {
		Query string `json:"query" jsonschema:"Search fragment to match against known third party names (e.g. 'Google', 'Meta', 'Hotjar')"`
	}

	searchThirdPartiesResult struct {
		Name string `json:"name"`
		// Slug is the catalog's deduplication key, derived from the name.
		// Exposing it lets the agent see that two spellings of one company
		// would become two separate catalog entries.
		Slug       string `json:"slug"`
		Category   string `json:"category"`
		WebsiteURL string `json:"website_url,omitempty"`
	}
)

// searchThirdPartiesSearchLimit caps how many catalog matches the agent
// sees. It is deliberately above the default LoadAll cap: results are
// ordered by name, so a narrow cap on a broad fragment can hide the
// specific entry the agent needs behind alphabetically earlier ones, and a
// missed match means a duplicate catalog row gets created.
const searchThirdPartiesSearchLimit = 50

func searchThirdPartiesTool(pgClient *pg.Client) agent.Tool {
	return agent.FunctionTool(
		"search_third_parties",
		"Search the internal catalog of known third parties (companies/services) by name fragment. Returns each match's exact name, its slug (the catalog's deduplication key), category, and website URL. Use this before naming a vendor: when a match is the same company, return its exact name verbatim so the tracker links to the existing catalog entry instead of creating a duplicate.",
		func(ctx context.Context, p searchThirdPartiesParams) (agent.ToolResult, error) {
			if p.Query == "" {
				return agent.ResultError("query is required"), nil
			}

			var out []searchThirdPartiesResult

			if err := pgClient.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					var parties coredata.CommonThirdParties
					if err := parties.LoadAllWithLimit(
						ctx,
						conn,
						coredata.NewCommonThirdPartyFilter(&p.Query),
						searchThirdPartiesSearchLimit,
					); err != nil {
						return err
					}

					out = make([]searchThirdPartiesResult, len(parties))
					for i, tp := range parties {
						out[i] = searchThirdPartiesResult{
							Name:     tp.Name,
							Slug:     tp.Slug,
							Category: string(tp.Category),
						}
						if tp.WebsiteURL != nil {
							out[i].WebsiteURL = *tp.WebsiteURL
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
