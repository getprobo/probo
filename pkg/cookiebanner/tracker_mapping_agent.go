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
	_ "embed"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/agent/tools/search"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// defaultAgentTimeout caps a single mapping or enrichment agent run
	// when the worker config does not supply one. It guards against a
	// hung LLM provider or a slow web search.
	defaultAgentTimeout = 45 * time.Second

	// defaultMappingMaxTurns and defaultEnrichmentMaxTurns bound the
	// agent reasoning loop (LLM call + tool round-trips) when the worker
	// config does not supply a value.
	defaultMappingMaxTurns    = 4
	defaultEnrichmentMaxTurns = 3

	// defaultAgentMaxTokens caps the output of the mapping and
	// enrichment agents when the agent config carries no max-tokens
	// budget. Both final outputs are tiny structured JSON, but the
	// budget must leave ample headroom for reasoning models (e.g. the
	// GPT-5 family): their reasoning tokens count against max_tokens,
	// so too small a budget gets consumed by reasoning and truncates
	// the JSON, surfacing as "unexpected end of JSON input".
	defaultAgentMaxTokens = 4096

	agentThirdPartyConfidenceThreshold = 0.6
	// agentSourceConfidence is the fixed confidence stored on catalog
	// rows the agent attributes to a third party. The agent's own
	// confidence now gauges the attribution (see ThirdPartyConfidence)
	// rather than the pattern, so the stored row confidence is a
	// constant like the other heuristic signals (domain, sibling).
	agentSourceConfidence = 0.8
)

//go:embed prompts/tracker_identification.txt.tmpl
var trackerIdentificationPrompt string

// TrackerMappingAgentResult is the structured output the tracker-mapping
// agent returns.
type TrackerMappingAgentResult struct {
	ThirdPartyName       string                      `json:"third_party_name" jsonschema:"Name of the company or service that sets this tracker (e.g. 'Google Analytics', 'Meta Pixel'). Empty string if truly unknown."`
	Category             coredata.ThirdPartyCategory `json:"category" jsonschema:"Third party category"`
	ThirdPartyConfidence float64                     `json:"third_party_confidence" jsonschema:"Confidence (0.0 to 1.0) in which company or service set this tracker, independent of whether the artifact is a classic web tracker. Set below 0.5 if unsure who set it."`
}

func buildTrackerMappingAgent(
	cfg TrackerAgentsConfig,
	pgClient *pg.Client,
	logger *log.Logger,
) *agent.Agent {
	tools := []agent.Tool{
		searchTrackerPatternsTool(pgClient),
		searchThirdPartiesTool(pgClient),
	}

	if cfg.FirecrawlAPIKey != "" {
		tools = append(tools, search.FirecrawlSearchTool(cfg.FirecrawlAPIKey))
	}

	outputType, err := agent.NewOutputType[TrackerMappingAgentResult]("tracker_identification")
	if err != nil {
		panic(fmt.Sprintf("cookiebanner: cannot build tracker identification output type: %s", err))
	}

	maxTurns := cfg.MappingMaxTurns
	if maxTurns < 1 {
		maxTurns = defaultMappingMaxTurns
	}

	opts := []agent.Option{
		agent.WithInstructionsFunc(trackerMappingInstructions),
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

	return agent.New("tracker-mapping", cfg.LLMClient, opts...)
}

// resolveAgentMaxTokens returns the configured max-tokens budget for the
// mapping and enrichment agents, falling back to defaultAgentMaxTokens
// when none is set.
func resolveAgentMaxTokens(configured *int) int {
	if configured != nil && *configured > 0 {
		return *configured
	}

	return defaultAgentMaxTokens
}

func trackerMappingInstructions(_ context.Context, _ *agent.Agent) string {
	categories := coredata.ThirdPartyCategories()

	parts := make([]string, len(categories))
	for i, c := range categories {
		parts[i] = string(c)
	}

	return strings.Replace(
		trackerIdentificationPrompt,
		"{{.Categories}}",
		strings.Join(parts, ", "),
		1,
	)
}

// buildTrackerIdentificationPrompt renders the base mapping-agent input
// shared by live tracker patterns and global catalog patterns: the four
// XML signal tags plus the max-age preamble. Callers append any extra
// signals (e.g. observed domains) to the returned prompt.
func buildTrackerIdentificationPrompt(
	pattern string,
	trackerType coredata.TrackerType,
	matchType coredata.TrackerPatternMatchType,
	maxAgeSeconds *int,
) string {
	maxAge := "session"
	if maxAgeSeconds != nil {
		maxAge = fmt.Sprintf("%d seconds", *maxAgeSeconds)
	}

	return fmt.Sprintf(
		"Identify the following tracker:\n\n"+
			"<pattern> %s </pattern>\n"+
			"<type> %s </type>\n"+
			"<match_type> %s </match_type>\n"+
			"<max_age> %s </max_age>\n",
		pattern,
		trackerType,
		matchType,
		maxAge,
	)
}

func buildAgentPrompt(tp coredata.TrackerPattern, domains []string) string {
	prompt := buildTrackerIdentificationPrompt(tp.Pattern, tp.TrackerType, tp.MatchType, tp.MaxAgeSeconds)

	if len(domains) > 0 {
		prompt += fmt.Sprintf("<observed_domains> %s </observed_domains>\n", strings.Join(domains, ", "))
	}

	return prompt
}
