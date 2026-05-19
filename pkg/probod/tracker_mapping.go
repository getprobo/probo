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

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/cookiebanner"
)

// buildTrackerMappingConfig wires the tracker-mapping agent. It is opt-in:
// deployments that do not set `llm.tracker-mapping.provider` get a zero
// config (nil LLM client) so the worker runs without agent fallback.
func (impl *Implm) buildTrackerMappingConfig(
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (cookiebanner.TrackerMappingConfig, error) {
	if impl.cfg.Agents.TrackerMapping.Provider == "" {
		return cookiebanner.TrackerMappingConfig{}, nil
	}

	agentCfg, llmClient, err := impl.resolveAgentClient(
		"tracker-mapping",
		impl.cfg.Agents.TrackerMapping,
		l,
		tp,
		r,
	)
	if err != nil {
		return cookiebanner.TrackerMappingConfig{}, fmt.Errorf("cannot resolve tracker mapping agent client: %w", err)
	}

	return cookiebanner.TrackerMappingConfig{
		LLMClient:         llmClient,
		Model:             agentCfg.ModelName,
		SearchEndpoint:    impl.cfg.SearchEndpoint,
		FirecrawlEndpoint: impl.cfg.Firecrawl.Endpoint,
		FirecrawlAPIKey:   impl.cfg.Firecrawl.APIKey,
	}, nil
}
