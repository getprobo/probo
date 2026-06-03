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
	searchThirdPartiesParams struct {
		Query string `json:"query" jsonschema:"Search fragment to match against known third party names (e.g. 'Google', 'Meta', 'Hotjar')"`
	}

	searchThirdPartiesResult struct {
		Name       string `json:"name"`
		Category   string `json:"category"`
		WebsiteURL string `json:"website_url,omitempty"`
	}
)

func searchThirdPartiesTool(pgClient *pg.Client) agent.Tool {
	return agent.FunctionTool(
		"search_third_parties",
		"Search the internal database of known third parties (companies/services) by name fragment. Returns matching third party names, categories, and website URLs. Use this to find the exact name of a known third party to link the tracker to.",
		func(ctx context.Context, p searchThirdPartiesParams) (agent.ToolResult, error) {
			if p.Query == "" {
				return agent.ResultError("query is required"), nil
			}

			var out []searchThirdPartiesResult

			if err := pgClient.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					var parties coredata.CommonThirdParties
					if err := parties.LoadAll(
						ctx,
						conn,
						coredata.NewCommonThirdPartyFilter(&p.Query),
					); err != nil {
						return err
					}

					out = make([]searchThirdPartiesResult, len(parties))
					for i, tp := range parties {
						out[i] = searchThirdPartiesResult{
							Name:     tp.Name,
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
