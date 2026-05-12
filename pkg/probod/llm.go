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
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/llm"
	llmanthropic "go.probo.inc/probo/pkg/llm/anthropic"
	llmopenai "go.probo.inc/probo/pkg/llm/openai"
)

// resolveAgentClient resolves the agent's effective config from defaults and
// builds an LLM client for it. The name parameter is used in the logger and
// in error messages.
func (impl *Implm) resolveAgentClient(
	name string,
	agent LLMAgentConfig,
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (LLMAgentConfig, *llm.Client, error) {
	resolved := impl.cfg.Agents.ResolveAgent(agent)
	providerCfg, ok := impl.cfg.Agents.Providers[resolved.Provider]
	if !ok {
		return LLMAgentConfig{}, nil, fmt.Errorf("unknown LLM provider %q for %s agent", resolved.Provider, name)
	}
	client, err := buildLLMClient(providerCfg, l.Named("llm."+name), tp, r)
	if err != nil {
		return LLMAgentConfig{}, nil, fmt.Errorf("cannot create %s LLM client: %w", name, err)
	}
	return resolved, client, nil
}

func buildLLMClient(cfg LLMProviderConfig, l *log.Logger, tp trace.TracerProvider, r prometheus.Registerer) (*llm.Client, error) {
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
