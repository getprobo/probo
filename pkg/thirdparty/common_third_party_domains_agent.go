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

package thirdparty

import (
	_ "embed"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/search"
)

//go:embed prompts/common_third_party_domains.txt.tmpl
var domainsPrompt string

// DomainsResult is the structured output of the domain-discovery agent
// (Agent C): the registrable domains the vendor itself owns and operates,
// including the marketing site, app, public API hosts, and CDN/asset
// hosts. The worker reduces these to eTLD+1, applies an ownership gate,
// and writes the survivors to common_third_party_domains so the
// tracker-mapping domain step can attribute trackers to this vendor.
type (
	DomainsResult struct {
		Domains []DomainCandidate `json:"domains" jsonschema:"The domains the vendor owns and operates. Empty when none can be confirmed."`
	}

	DomainCandidate struct {
		Domain     string  `json:"domain" jsonschema:"A domain the vendor owns and operates (a registrable domain such as 'intercomcdn.com' or a full host such as 'api.vendor.com'). Never a third-party or shared-infrastructure host the vendor does not own."`
		Confidence float64 `json:"confidence" jsonschema:"Confidence from 0.0 to 1.0 that the vendor owns this domain. Use 0 when ownership is not confirmed."`
		SourceURL  string  `json:"source_url" jsonschema:"The URL where the vendor's ownership of this domain was observed, or an empty string."`
	}
)

// buildCommonThirdPartyDomainsAgent builds Agent C. extraTools carries
// the browser read-only toolset when a headless Chrome endpoint is
// configured; it is empty otherwise, in which case the agent relies on
// web_search alone.
func buildCommonThirdPartyDomainsAgent(
	cfg EnrichmentConfig,
	logger *log.Logger,
	extraTools []agent.Tool,
) *agent.Agent {
	tools := append([]agent.Tool{}, extraTools...)

	if cfg.FirecrawlAPIKey != "" {
		tools = append(tools, search.FirecrawlSearchTool(cfg.FirecrawlAPIKey))
	}

	outputType, err := agent.NewOutputType[DomainsResult]("common_third_party_domains")
	if err != nil {
		panic(fmt.Sprintf("thirdparty: cannot build domains output type: %s", err))
	}

	opts := []agent.Option{
		agent.WithInstructions(domainsPrompt),
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

	return agent.New("common-third-party-domains", cfg.LLMClient, opts...)
}

// buildCommonThirdPartyDomainsPrompt renders the per-row input for Agent
// C, seeding it with the vendor name and the website resolved by Agent A
// so it can anchor ownership to the vendor's own domain.
func buildCommonThirdPartyDomainsPrompt(name, websiteURL string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Find the domains owned and operated by this vendor.\n\n")
	fmt.Fprintf(&b, "<name> %s </name>\n", name)

	if w := strings.TrimSpace(websiteURL); w != "" {
		fmt.Fprintf(&b, "<website> %s </website>\n", w)
	}

	return b.String()
}
