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

package cmdutil

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace/noop"
	"go.probo.inc/probo/pkg/llm"
	llmanthropic "go.probo.inc/probo/pkg/llm/anthropic"
	llmopenai "go.probo.inc/probo/pkg/llm/openai"
	"go.probo.inc/probo/pkg/probodconfig"
)

// resolveAgentClient resolves an agent's effective config from defaults
// and builds an LLM client for it. The name is used in the logger name
// and error messages. proboctl runs synchronously with no tracing or
// metrics, so it duplicates probod's wiring with a no-op tracer and a
// throwaway registry rather than sharing a package.
func resolveAgentClient(
	agents probodconfig.AgentsConfig,
	name string,
	agent probodconfig.LLMAgentConfig,
	l *log.Logger,
) (probodconfig.LLMAgentConfig, *llm.Client, error) {
	resolved := agents.ResolveAgent(agent)

	providerCfg, ok := agents.Providers[resolved.Provider]
	if !ok {
		return probodconfig.LLMAgentConfig{}, nil, fmt.Errorf("unknown LLM provider %q for %s agent", resolved.Provider, name)
	}

	client, err := buildLLMClient(providerCfg, l.Named("llm."+name))
	if err != nil {
		return probodconfig.LLMAgentConfig{}, nil, fmt.Errorf("cannot create %s LLM client: %w", name, err)
	}

	return resolved, client, nil
}

// buildLLMClient creates an LLM client for the given provider config.
func buildLLMClient(cfg probodconfig.LLMProviderConfig, l *log.Logger) (*llm.Client, error) {
	providerType := cfg.Type
	if providerType == "" {
		providerType = "openai"
	}

	httpClient := httpclient.DefaultPooledClient(
		httpclient.WithLogger(l),
		httpclient.WithTracerProvider(noop.NewTracerProvider()),
		httpclient.WithRegisterer(prometheus.NewRegistry()),
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
		), nil
	case "bedrock":
		return nil, fmt.Errorf("bedrock provider not yet wired; requires aws.Config")
	default:
		return nil, fmt.Errorf("unsupported LLM provider type: %q", providerType)
	}
}
