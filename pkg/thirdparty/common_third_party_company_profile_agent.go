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

package thirdparty

import (
	_ "embed"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/search"
	"go.probo.inc/probo/pkg/coredata"
)

//go:embed prompts/common_third_party_company_profile.txt.tmpl
var companyProfilePrompt string

// CompanyProfileResult is the structured output of the company-profile
// agent (Agent A): the vendor's identity facts. website_url is the key
// signal the compliance-docs agent and the logo step depend on, so it is
// resolved here first.
type CompanyProfileResult struct {
	LegalName          EnrichedField `json:"legal_name" jsonschema:"The vendor's full legal company name including the entity suffix (e.g. 'Acme Technologies, Inc.')."`
	HeadquarterAddress EnrichedField `json:"headquarter_address" jsonschema:"The vendor's headquarters postal address (street, city, region, country)."`
	WebsiteURL         EnrichedField `json:"website_url" jsonschema:"The vendor's canonical primary marketing website URL (https scheme, no tracking query parameters, no trailing path)."`
}

// buildCompanyProfileAgent builds Agent A. extraTools carries the
// browser read-only toolset when a headless Chrome endpoint is
// configured; it is empty otherwise, in which case the agent relies on
// web_search alone. The browser lets it read footer/imprint/about/legal
// pages where the legal name and headquarters address live.
func buildCompanyProfileAgent(
	cfg EnrichmentConfig,
	logger *log.Logger,
	extraTools []agent.Tool,
) *agent.Agent {
	tools := append([]agent.Tool{}, extraTools...)

	if cfg.FirecrawlAPIKey != "" {
		tools = append(tools, search.FirecrawlSearchTool(cfg.FirecrawlAPIKey))
	}

	outputType, err := agent.NewOutputType[CompanyProfileResult]("common_third_party_company_profile")
	if err != nil {
		panic(fmt.Sprintf("thirdparty: cannot build company profile output type: %s", err))
	}

	opts := []agent.Option{
		agent.WithInstructions(companyProfilePrompt),
		agent.WithModel(cfg.Model),
		agent.WithOutputType(outputType),
		agent.WithMaxTurns(resolveEnrichmentMaxTurns(cfg.MaxTurns)),
		agent.WithMaxTokens(resolveEnrichmentMaxTokens(cfg.MaxTokens)),
		agent.WithLogger(logger),
	}

	if len(tools) > 0 {
		opts = append(opts, agent.WithTools(tools...))
	}

	if cfg.Temperature != nil {
		opts = append(opts, agent.WithTemperature(*cfg.Temperature))
	}

	return agent.New("common-third-party-company-profile", cfg.LLMClient, opts...)
}

// buildCompanyProfilePrompt renders the per-row input for Agent A. Any
// values already on the row are passed as hints so the agent confirms or
// corrects them rather than starting cold.
func buildCompanyProfilePrompt(party coredata.CommonThirdParty) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Research this company and return its profile.\n\n")
	fmt.Fprintf(&b, "<name> %s </name>\n", party.Name)

	if party.WebsiteURL != nil && strings.TrimSpace(*party.WebsiteURL) != "" {
		fmt.Fprintf(&b, "<known_website> %s </known_website>\n", *party.WebsiteURL)
	}

	if party.LegalName != nil && strings.TrimSpace(*party.LegalName) != "" {
		fmt.Fprintf(&b, "<known_legal_name> %s </known_legal_name>\n", *party.LegalName)
	}

	return b.String()
}
