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

package probod

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/thirdparty"
)

// buildTrackerAgentsConfig wires the tracker agents that share one LLM
// client and model: the tracker-mapping agent (catalog identification),
// the common-pattern enrichment agent (description research), and the
// third-party disambiguation agent that the tracker-mapping worker uses
// to promote patterns to org ThirdParties. All are opt-in: deployments
// that do not set `llm.tracker-mapping.provider` get zero configs (nil
// LLM client) so the workers run without agent fallback.
//
// The agents are sourced from the same `tracker-mapping` config slot
// because they share the LLM client, model, and lifecycle. The
// disambiguation agent has no Firecrawl/DB tools, so its config
// surface is narrower and it lives in the cross-domain pkg/thirdparty
// package.
func (impl *Implm) buildTrackerAgentsConfig(
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (cookiebanner.TrackerAgentsConfig, thirdparty.DisambiguationConfig, error) {
	if impl.cfg.Agents.TrackerMapping.Provider == "" {
		return cookiebanner.TrackerAgentsConfig{}, thirdparty.DisambiguationConfig{}, nil
	}

	agentCfg, llmClient, err := impl.resolveAgentClient(
		"tracker-mapping",
		impl.cfg.Agents.TrackerMapping,
		l,
		tp,
		r,
	)
	if err != nil {
		return cookiebanner.TrackerAgentsConfig{}, thirdparty.DisambiguationConfig{}, fmt.Errorf("cannot resolve tracker mapping agent client: %w", err)
	}

	mappingWorkerCfg := impl.cfg.TrackerMappingWorker
	enrichmentWorkerCfg := impl.cfg.CommonPatternEnrichmentWorker

	// The mapping and enrichment agents share one config slot but run
	// from separate workers with separate max-turns. AgentTimeout here
	// carries the mapping worker's value (also reused by the
	// disambiguation agent); the enrichment worker overrides it on its
	// own copy at registration.
	trackerAgentsCfg := cookiebanner.TrackerAgentsConfig{
		LLMClient:          llmClient,
		Model:              agentCfg.ModelName,
		FirecrawlAPIKey:    impl.cfg.Agents.Tools.FirecrawlAPIKey,
		MaxTokens:          agentCfg.MaxTokens,
		Temperature:        agentCfg.Temperature,
		AgentTimeout:       time.Duration(mappingWorkerCfg.AgentTimeout) * time.Second,
		MappingMaxTurns:    mappingWorkerCfg.AgentMaxTurns,
		EnrichmentMaxTurns: enrichmentWorkerCfg.AgentMaxTurns,
	}

	// The disambiguation agent emits a single id plus a short rationale,
	// so it keeps its own smaller token budget (left unset here) rather
	// than inheriting the mapping agent's. It shares the mapping worker's
	// timeout.
	disambiguationCfg := thirdparty.DisambiguationConfig{
		LLMClient:   llmClient,
		Model:       agentCfg.ModelName,
		Temperature: agentCfg.Temperature,
		Timeout:     time.Duration(mappingWorkerCfg.AgentTimeout) * time.Second,
	}

	return trackerAgentsCfg, disambiguationCfg, nil
}
