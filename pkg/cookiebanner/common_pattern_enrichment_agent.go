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
	_ "embed"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/search"
	"go.probo.inc/probo/pkg/coredata"
)

//go:embed prompts/tracker_enrichment.txt.tmpl
var trackerEnrichmentPrompt string

// CommonPatternEnrichmentResult is the structured output the
// common-pattern enrichment agent returns.
type CommonPatternEnrichmentResult struct {
	Description string `json:"description" jsonschema:"A concise, factual, compliance-grade description of what this tracker stores or does and its purpose. One or two sentences. Name the operating company when known. Empty when the purpose cannot be substantiated from evidence."`
}

func buildCommonPatternEnrichmentAgent(
	cfg TrackerAgentsConfig,
	pgClient *pg.Client,
	logger *log.Logger,
) *agent.Agent {
	tools := []agent.Tool{
		searchThirdPartiesTool(pgClient),
	}

	if cfg.FirecrawlAPIKey != "" {
		tools = append(tools, search.FirecrawlSearchTool(cfg.FirecrawlAPIKey))
	}

	outputType, err := agent.NewOutputType[CommonPatternEnrichmentResult]("tracker_enrichment")
	if err != nil {
		panic(fmt.Sprintf("cookiebanner: cannot build tracker enrichment output type: %s", err))
	}

	maxTurns := cfg.EnrichmentMaxTurns
	if maxTurns < 1 {
		maxTurns = defaultEnrichmentMaxTurns
	}

	opts := []agent.Option{
		agent.WithInstructions(trackerEnrichmentPrompt),
		agent.WithModel(cfg.Model),
		agent.WithTools(tools...),
		agent.WithOutputType(outputType),
		agent.WithMaxTurns(maxTurns),
		agent.WithMaxTokens(resolveAgentMaxTokens(cfg.MaxTokens)),
		agent.WithLogger(logger),
	}

	if cfg.Temperature != nil {
		opts = append(opts, agent.WithTemperature(*cfg.Temperature))
	}

	return agent.New("common-pattern-enrichment", cfg.LLMClient, opts...)
}

// buildCommonPatternIdentificationPrompt builds the mapping-agent input
// for a global catalog pattern. Catalog rows carry no observed domains,
// so the prompt omits the <observed_domains> signal and relies on the
// pattern name, type, and naming conventions. It lets the enrichment
// worker reuse the mapping agent to attribute a vendor before describing.
func buildCommonPatternIdentificationPrompt(cp coredata.CommonTrackerPattern) string {
	return buildTrackerIdentificationPrompt(cp.Pattern, cp.TrackerType, cp.MatchType, cp.MaxAgeSeconds)
}

func buildEnrichmentPrompt(cp coredata.CommonTrackerPattern, thirdPartyName string) string {
	maxAge := "session"
	if cp.MaxAgeSeconds != nil {
		maxAge = fmt.Sprintf("%d seconds", *cp.MaxAgeSeconds)
	}

	prompt := fmt.Sprintf(
		"Describe the following tracker:\n\n"+
			"<pattern> %s </pattern>\n"+
			"<type> %s </type>\n"+
			"<match_type> %s </match_type>\n"+
			"<max_age> %s </max_age>\n",
		cp.Pattern,
		cp.TrackerType,
		cp.MatchType,
		maxAge,
	)

	if name := strings.TrimSpace(thirdPartyName); name != "" {
		prompt += fmt.Sprintf("<third_party> %s </third_party>\n", name)
	}

	return prompt
}
