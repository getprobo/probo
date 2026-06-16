// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	// config does not supply a value. The budget must exceed the number
	// of tool rounds each prompt authorizes (the mapping prompt allows
	// two DB searches plus up to three web searches, the enrichment
	// prompt one DB search plus up to three web searches) and still
	// leave a turn for the forced structured-output synthesis turn;
	// otherwise the loop trips MaxTurnsExceededError before emitting
	// JSON. Ten matches agent.DefaultMaxTurns and gives ample headroom.
	defaultMappingMaxTurns    = 10
	defaultEnrichmentMaxTurns = 10

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

	// trustedAttributionConfidence is the bar a catalog row must meet for
	// its third party to be adopted deterministically by another pattern
	// (the existing-link and matchByPattern paths). Only curated/seed rows
	// and operator links (confidence 1.0) clear it; agent (0.8) and
	// domain/sibling (0.7) attributions do not, so a single low-confidence
	// guess never becomes an authoritative precedent that auto-propagates
	// across organizations. Such rows are reused as hints only: the
	// pattern falls through to the evidence-guarded agent, which can
	// corroborate the guess (promoting the row to this tier) or override
	// it.
	trustedAttributionConfidence float32 = 0.9
)

//go:embed prompts/tracker_identification.txt.tmpl
var trackerIdentificationPrompt string

// Tracker-mapping evidence kinds. The agent must report which concrete
// evidence backs a vendor attribution; an attribution without one of the
// substantive kinds (i.e. "none" or empty) is discarded so the agent
// never attributes a vendor from general knowledge or vague similarity.
const (
	evidenceSourceDatabaseMatch    = "database_match"
	evidenceSourceNamingConvention = "naming_convention"
	evidenceSourceWebSearch        = "web_search"
	evidenceSourceBrowserPage      = "browser_page"
	evidenceSourceNone             = "none"
)

// TrackerMappingAgentResult is the structured output the tracker-mapping
// agent returns.
type TrackerMappingAgentResult struct {
	ThirdPartyName       string                      `json:"third_party_name" jsonschema:"Name of the company or service that sets this tracker (e.g. 'Google Analytics', 'Meta Pixel'). Empty string if truly unknown."`
	Category             coredata.ThirdPartyCategory `json:"category" jsonschema:"Third party category"`
	ThirdPartyConfidence float64                     `json:"third_party_confidence" jsonschema:"Confidence (0.0 to 1.0) in which company or service set this tracker, independent of whether the artifact is a classic web tracker. Set below 0.5 if unsure who set it."`
	EvidenceSource       string                      `json:"evidence_source" jsonschema:"The concrete evidence that backs the attribution: 'database_match' (exact pattern in the database), 'naming_convention' (the tracker's meaningful prefix or an embedded vendor name), 'web_search' (a web result naming the setter), 'browser_page' (a page you opened that names the setter), or 'none' when there is no concrete evidence. Must be 'none' whenever third_party_name is empty."`
	IsFirstParty         bool                        `json:"is_first_party" jsonschema:"True when the artifact has no third party at all: it is the scanned site's own tracker, a generic library or log key (e.g. 'loglevel'), a browser-extension key that embeds the scanned site's origin, or otherwise not attributable to any external vendor. Leave false when a vendor is or might be responsible."`
}

// evidenceSupportsAttribution reports whether the agent supplied a
// concrete evidence kind for a vendor attribution. An empty value or
// "none" (or any unrecognized value) does not support an attribution.
func evidenceSupportsAttribution(evidenceSource string) bool {
	switch evidenceSource {
	case
		evidenceSourceDatabaseMatch,
		evidenceSourceNamingConvention,
		evidenceSourceWebSearch,
		evidenceSourceBrowserPage:
		return true
	}

	return false
}

// buildTrackerMappingAgent builds the tracker-mapping agent. extraTools
// carries the browser read-only toolset when a headless Chrome endpoint
// is configured; it is empty otherwise, in which case the agent relies on
// the DB search tools and web search alone. The browser lets it open
// cookie-database and cookie-policy pages to read the true setter.
func buildTrackerMappingAgent(
	cfg TrackerMappingAgentConfig,
	pgClient *pg.Client,
	logger *log.Logger,
	extraTools []agent.Tool,
) *agent.Agent {
	tools := []agent.Tool{
		searchTrackerPatternsTool(pgClient),
		searchThirdPartiesTool(pgClient),
	}

	tools = append(tools, extraTools...)

	if cfg.FirecrawlAPIKey != "" {
		tools = append(tools, search.FirecrawlSearchTool(cfg.FirecrawlAPIKey))
	}

	outputType, err := agent.NewOutputType[TrackerMappingAgentResult]("tracker_identification")
	if err != nil {
		panic(fmt.Sprintf("cookiebanner: cannot build tracker identification output type: %s", err))
	}

	maxTurns := cfg.MaxTurns
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

func buildAgentPrompt(tp coredata.TrackerPattern, domains []string, siteDomain string) string {
	prompt := buildTrackerIdentificationPrompt(tp.Pattern, tp.TrackerType, tp.MatchType, tp.MaxAgeSeconds)

	if siteDomain != "" {
		prompt += fmt.Sprintf("<scanned_site> %s </scanned_site>\n", siteDomain)
	}

	if len(domains) > 0 {
		prompt += fmt.Sprintf("<observed_domains> %s </observed_domains>\n", strings.Join(domains, ", "))
	}

	return prompt
}
