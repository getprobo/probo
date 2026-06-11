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

// buildTrackerAgents wires the three tracker agents from the probod
// config, each with its own LLM client and tuning: the tracker-mapping
// agent (catalog identification), the common-pattern enrichment agent
// (description research), and the third-party disambiguation agent. All
// are opt-in: when `llm.tracker-mapping.provider` is empty it returns
// zero configs (nil LLM clients) so callers run without agent fallback.
//
// The enrichment and disambiguation agents fall back to the
// tracker-mapping config when their own provider slot is empty, so a
// deployment that configures only `tracker-mapping` keeps wiring all
// three agents.
func (impl *Implm) buildTrackerAgents(
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (cookiebanner.TrackerMappingAgentConfig, cookiebanner.TrackerEnrichmentAgentConfig, thirdparty.DisambiguationAgentConfig, error) {
	if impl.cfg.Agents.TrackerMapping.Provider == "" {
		return cookiebanner.TrackerMappingAgentConfig{}, cookiebanner.TrackerEnrichmentAgentConfig{}, thirdparty.DisambiguationAgentConfig{}, nil
	}

	firecrawlAPIKey := impl.cfg.Agents.Tools.FirecrawlAPIKey

	mappingAgentCfg, mappingClient, err := impl.resolveAgentClient(
		"tracker-mapping",
		impl.cfg.Agents.TrackerMapping,
		l,
		tp,
		r,
	)
	if err != nil {
		return cookiebanner.TrackerMappingAgentConfig{}, cookiebanner.TrackerEnrichmentAgentConfig{}, thirdparty.DisambiguationAgentConfig{}, fmt.Errorf("cannot resolve tracker mapping agent client: %w", err)
	}

	mappingCfg := cookiebanner.TrackerMappingAgentConfig{
		LLMClient:       mappingClient,
		Model:           mappingAgentCfg.ModelName,
		FirecrawlAPIKey: firecrawlAPIKey,
		MaxTokens:       mappingAgentCfg.MaxTokens,
		Temperature:     mappingAgentCfg.Temperature,
		Timeout:         time.Duration(impl.cfg.TrackerMappingWorker.AgentTimeout) * time.Second,
		MaxTurns:        impl.cfg.TrackerMappingWorker.AgentMaxTurns,
	}

	enrichmentSlot := impl.cfg.Agents.TrackerEnrichment
	if enrichmentSlot.Provider == "" {
		enrichmentSlot = impl.cfg.Agents.TrackerMapping
	}

	enrichmentAgentCfg, enrichmentClient, err := impl.resolveAgentClient(
		"tracker-enrichment",
		enrichmentSlot,
		l,
		tp,
		r,
	)
	if err != nil {
		return cookiebanner.TrackerMappingAgentConfig{}, cookiebanner.TrackerEnrichmentAgentConfig{}, thirdparty.DisambiguationAgentConfig{}, fmt.Errorf("cannot resolve tracker enrichment agent client: %w", err)
	}

	enrichmentCfg := cookiebanner.TrackerEnrichmentAgentConfig{
		LLMClient:       enrichmentClient,
		Model:           enrichmentAgentCfg.ModelName,
		FirecrawlAPIKey: firecrawlAPIKey,
		MaxTokens:       enrichmentAgentCfg.MaxTokens,
		Temperature:     enrichmentAgentCfg.Temperature,
		Timeout:         time.Duration(impl.cfg.CommonPatternEnrichmentWorker.AgentTimeout) * time.Second,
		MaxTurns:        impl.cfg.CommonPatternEnrichmentWorker.AgentMaxTurns,
	}

	disambiguationSlot := impl.cfg.Agents.ThirdPartyDisambiguation
	if disambiguationSlot.Provider == "" {
		disambiguationSlot = impl.cfg.Agents.TrackerMapping
	}

	disambiguationAgentCfg, disambiguationClient, err := impl.resolveAgentClient(
		"third-party-disambiguation",
		disambiguationSlot,
		l,
		tp,
		r,
	)
	if err != nil {
		return cookiebanner.TrackerMappingAgentConfig{}, cookiebanner.TrackerEnrichmentAgentConfig{}, thirdparty.DisambiguationAgentConfig{}, fmt.Errorf("cannot resolve third party disambiguation agent client: %w", err)
	}

	disambiguationCfg := thirdparty.DisambiguationAgentConfig{
		LLMClient:   disambiguationClient,
		Model:       disambiguationAgentCfg.ModelName,
		MaxTokens:   disambiguationAgentCfg.MaxTokens,
		Temperature: disambiguationAgentCfg.Temperature,
		Timeout:     time.Duration(impl.cfg.TrackerMappingWorker.DisambiguationAgentTimeout) * time.Second,
	}

	return mappingCfg, enrichmentCfg, disambiguationCfg, nil
}
