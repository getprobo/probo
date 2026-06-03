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

// Package agentsbuild builds LLM clients and tracker-agent configuration
// from the shared probodconfig types. It is the single wiring used by
// both probod (background workers) and proboctl (synchronous operator
// commands) so the two executables build agents identically.
package agentsbuild

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/llm"
	llmanthropic "go.probo.inc/probo/pkg/llm/anthropic"
	llmopenai "go.probo.inc/probo/pkg/llm/openai"
	"go.probo.inc/probo/pkg/probodconfig"
	"go.probo.inc/probo/pkg/thirdparty"
)

// BuildLLMClient creates an LLM client for the given provider config.
func BuildLLMClient(
	cfg probodconfig.LLMProviderConfig,
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (*llm.Client, error) {
	providerType := cfg.Type
	if providerType == "" {
		providerType = "openai"
	}

	httpClient := httpclient.DefaultPooledClient(
		httpclient.WithLogger(l),
		httpclient.WithTracerProvider(tp),
		httpclient.WithRegisterer(r),
	)

	switch providerType {
	case "openai":
		p := llmopenai.NewProvider(
			cfg.APIKey,
			llmopenai.WithHTTPClient(httpClient),
		)

		return llm.NewClient(
			p,
			"openai",
			llm.WithLogger(l),
			llm.WithTracerProvider(tp),
		), nil
	case "anthropic":
		p := llmanthropic.NewProvider(
			cfg.APIKey,
			llmanthropic.WithHTTPClient(httpClient),
		)

		return llm.NewClient(
			p,
			"anthropic",
			llm.WithLogger(l),
			llm.WithTracerProvider(tp),
		), nil
	case "bedrock":
		return nil, fmt.Errorf("bedrock provider not yet wired; requires aws.Config")
	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %q", providerType)
	}
}

// ResolveAgentClient resolves an agent's effective config from defaults
// and builds an LLM client for it. The name is used in the logger name
// and error messages.
func ResolveAgentClient(
	agents probodconfig.AgentsConfig,
	name string,
	agent probodconfig.LLMAgentConfig,
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (probodconfig.LLMAgentConfig, *llm.Client, error) {
	resolved := agents.ResolveAgent(agent)

	providerCfg, ok := agents.Providers[resolved.Provider]
	if !ok {
		return probodconfig.LLMAgentConfig{}, nil, fmt.Errorf("unknown LLM provider %q for %s agent", resolved.Provider, name)
	}

	client, err := BuildLLMClient(providerCfg, l.Named("llm."+name), tp, r)
	if err != nil {
		return probodconfig.LLMAgentConfig{}, nil, fmt.Errorf("cannot create %s LLM client: %w", name, err)
	}

	return resolved, client, nil
}

// BuildTrackerAgentsConfig wires the tracker agents that share one LLM
// client and model: the tracker-mapping agent (catalog identification),
// the common-pattern enrichment agent (description research), and the
// third-party disambiguation agent. All are opt-in: when
// `llm.tracker-mapping.provider` is empty it returns zero configs (nil
// LLM client) so callers run without agent fallback.
func BuildTrackerAgentsConfig(
	cfg probodconfig.Config,
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (cookiebanner.TrackerAgentsConfig, thirdparty.DisambiguationConfig, error) {
	if cfg.Agents.TrackerMapping.Provider == "" {
		return cookiebanner.TrackerAgentsConfig{}, thirdparty.DisambiguationConfig{}, nil
	}

	agentCfg, llmClient, err := ResolveAgentClient(
		cfg.Agents,
		"tracker-mapping",
		cfg.Agents.TrackerMapping,
		l,
		tp,
		r,
	)
	if err != nil {
		return cookiebanner.TrackerAgentsConfig{}, thirdparty.DisambiguationConfig{}, fmt.Errorf("cannot resolve tracker mapping agent client: %w", err)
	}

	mappingWorkerCfg := cfg.TrackerMappingWorker
	enrichmentWorkerCfg := cfg.CommonPatternEnrichmentWorker

	// The mapping and enrichment agents share one config slot but run
	// from separate workers with separate max-turns. AgentTimeout here
	// carries the mapping worker's value (also reused by the
	// disambiguation agent); the enrichment worker overrides it on its
	// own copy at registration.
	trackerAgentsCfg := cookiebanner.TrackerAgentsConfig{
		LLMClient:          llmClient,
		Model:              agentCfg.ModelName,
		FirecrawlAPIKey:    cfg.Agents.Tools.FirecrawlAPIKey,
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
