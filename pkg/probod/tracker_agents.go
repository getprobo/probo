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

package probod

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/cookiebanner"
)

// buildTrackerAgents wires the tracker agents from the probod config,
// each with its own LLM client and tuning: the tracker-mapping agent
// (catalog identification) and the common-pattern enrichment agent
// (description research). Both are opt-in: when
// `llm.tracker-mapping.provider` is empty it returns zero configs
// (nil LLM clients) so callers run without agent fallback.
//
// The enrichment agent falls back to the tracker-mapping config when
// its own provider slot is empty, so a deployment that configures
// only `tracker-mapping` keeps wiring both agents.
func (impl *Implm) buildTrackerAgents(
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (cookiebanner.TrackerMappingAgentConfig, cookiebanner.TrackerEnrichmentAgentConfig, error) {
	if impl.cfg.Agents.TrackerMapping.Provider == "" {
		return cookiebanner.TrackerMappingAgentConfig{}, cookiebanner.TrackerEnrichmentAgentConfig{}, nil
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
		return cookiebanner.TrackerMappingAgentConfig{}, cookiebanner.TrackerEnrichmentAgentConfig{}, fmt.Errorf("cannot resolve tracker mapping agent client: %w", err)
	}

	mappingCfg := cookiebanner.TrackerMappingAgentConfig{
		LLMClient:       mappingClient,
		Model:           mappingAgentCfg.ModelName,
		FirecrawlAPIKey: firecrawlAPIKey,
		ChromeAddr:      impl.cfg.ChromeDPAddr,
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
		return cookiebanner.TrackerMappingAgentConfig{}, cookiebanner.TrackerEnrichmentAgentConfig{}, fmt.Errorf("cannot resolve tracker enrichment agent client: %w", err)
	}

	enrichmentCfg := cookiebanner.TrackerEnrichmentAgentConfig{
		LLMClient:       enrichmentClient,
		Model:           enrichmentAgentCfg.ModelName,
		FirecrawlAPIKey: firecrawlAPIKey,
		ChromeAddr:      impl.cfg.ChromeDPAddr,
		MaxTokens:       enrichmentAgentCfg.MaxTokens,
		Temperature:     enrichmentAgentCfg.Temperature,
		Timeout:         time.Duration(impl.cfg.CommonPatternEnrichmentWorker.AgentTimeout) * time.Second,
		MaxTurns:        impl.cfg.CommonPatternEnrichmentWorker.AgentMaxTurns,
	}

	return mappingCfg, enrichmentCfg, nil
}
