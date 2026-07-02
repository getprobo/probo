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

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/llm"
	llmanthropic "go.probo.inc/probo/pkg/llm/anthropic"
	llmopenai "go.probo.inc/probo/pkg/llm/openai"
)

// buildLLMClient creates an LLM client for the given provider config.
func (impl *Implm) buildLLMClient(
	cfg LLMProviderConfig,
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

	client, err := impl.buildLLMClient(providerCfg, l.Named("llm."+name), tp, r)
	if err != nil {
		return LLMAgentConfig{}, nil, fmt.Errorf("cannot create %s LLM client: %w", name, err)
	}

	return resolved, client, nil
}
