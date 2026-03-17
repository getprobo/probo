// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	llmopenai "go.probo.inc/probo/pkg/llm/openai"
)

func buildLLMClient(cfg LLMConfig, l *log.Logger, tp trace.TracerProvider, r prometheus.Registerer) (*llm.Client, error) {
	provider := cfg.Provider
	if provider == "" {
		provider = "openai"
	}

	httpClient := httpclient.DefaultPooledClient(
		httpclient.WithLogger(l),
		httpclient.WithTracerProvider(tp),
		httpclient.WithRegisterer(r),
	)

	switch provider {
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
		return nil, fmt.Errorf("anthropic provider not yet wired; add import and construct here")
	case "bedrock":
		return nil, fmt.Errorf("bedrock provider not yet wired; requires aws.Config")
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %q", provider)
	}
}
