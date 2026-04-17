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

//go:build !e2e

package probod

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/agents/vetting"
	"go.probo.inc/probo/pkg/probo"
)

// buildVendorAssessor wires the real LLM/browser-driven vendor assessor.
// The e2e-tagged twin substitutes a deterministic stub.
func (impl *Implm) buildVendorAssessor(
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (probo.VendorAssessor, error) {
	agentCfg, llmClient, err := impl.resolveAgentClient("vendor-assessor", impl.cfg.Agents.VendorAssessor, l, tp, r)
	if err != nil {
		return nil, err
	}

	maxTokens := vetting.DefaultMaxTokens
	if agentCfg.MaxTokens != nil {
		maxTokens = *agentCfg.MaxTokens
	}

	return vetting.NewAssessor(vetting.Config{
		Client:         llmClient,
		Model:          agentCfg.ModelName,
		MaxTokens:      maxTokens,
		ChromeAddr:     impl.cfg.ChromeDPAddr,
		SearchEndpoint: impl.cfg.SearchEndpoint,
		Logger:         l.Named("vendor-assessor"),
	}), nil
}
