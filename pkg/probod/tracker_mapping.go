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
	"go.probo.inc/probo/pkg/thirdparty"
)

// buildTrackerMappingConfig wires the tracker-mapping agent (catalog
// identification) and the third-party disambiguation agent that the
// tracker-mapping worker uses to promote patterns to org ThirdParties.
// Both are opt-in: deployments that do not set
// `llm.tracker-mapping.provider` get zero configs (nil LLM client) so
// the worker runs without agent fallback.
//
// Both agents are sourced from the same `tracker-mapping` config slot
// because they share the LLM client, model, and lifecycle. The
// disambiguation agent has no Firecrawl/DB tools, so its config
// surface is narrower and it lives in the cross-domain pkg/thirdparty
// package.
func (impl *Implm) buildTrackerMappingConfig(
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (cookiebanner.TrackerMappingConfig, thirdparty.DisambiguationConfig, error) {
	if impl.cfg.Agents.TrackerMapping.Provider == "" {
		return cookiebanner.TrackerMappingConfig{}, thirdparty.DisambiguationConfig{}, nil
	}

	agentCfg, llmClient, err := impl.resolveAgentClient(
		"tracker-mapping",
		impl.cfg.Agents.TrackerMapping,
		l,
		tp,
		r,
	)
	if err != nil {
		return cookiebanner.TrackerMappingConfig{}, thirdparty.DisambiguationConfig{}, fmt.Errorf("cannot resolve tracker mapping agent client: %w", err)
	}

	mappingCfg := cookiebanner.TrackerMappingConfig{
		LLMClient:       llmClient,
		Model:           agentCfg.ModelName,
		FirecrawlAPIKey: impl.cfg.Agents.Tools.FirecrawlAPIKey,
	}

	disambiguationCfg := thirdparty.DisambiguationConfig{
		LLMClient: llmClient,
		Model:     agentCfg.ModelName,
	}

	return mappingCfg, disambiguationCfg, nil
}
