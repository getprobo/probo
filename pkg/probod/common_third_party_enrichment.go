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
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/thirdparty"
)

// buildCommonThirdPartyEnrichmentConfig wires the common-third-party
// enrichment worker config: the LLM client for its two agents plus the
// worker tuning, browser endpoint, and logo-storage dependencies. It is
// opt-in: a deployment that does not set
// `llm.common-third-party-enrichment.provider` gets a zero config (nil
// LLM client), so the worker runs as a no-op and the caller skips
// registration.
func (impl *Implm) buildCommonThirdPartyEnrichmentConfig(
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
	fileManager *filemanager.Service,
) (thirdparty.EnrichmentConfig, error) {
	if impl.cfg.Agents.CommonThirdPartyEnrichment.Provider == "" {
		return thirdparty.EnrichmentConfig{}, nil
	}

	agentCfg, llmClient, err := impl.resolveAgentClient(
		"common-third-party-enrichment",
		impl.cfg.Agents.CommonThirdPartyEnrichment,
		l,
		tp,
		r,
	)
	if err != nil {
		return thirdparty.EnrichmentConfig{}, fmt.Errorf("cannot resolve common third party enrichment agent client: %w", err)
	}

	workerCfg := impl.cfg.CommonThirdPartyEnrichmentWorker

	return thirdparty.EnrichmentConfig{
		LLMClient:           llmClient,
		Model:               agentCfg.ModelName,
		MaxTokens:           agentCfg.MaxTokens,
		Temperature:         agentCfg.Temperature,
		FirecrawlAPIKey:     impl.cfg.Agents.Tools.FirecrawlAPIKey,
		ChromeAddr:          impl.cfg.ChromeDPAddr,
		AgentTimeout:        time.Duration(workerCfg.AgentTimeout) * time.Second,
		MaxTurns:            workerCfg.AgentMaxTurns,
		ConfidenceThreshold: workerCfg.ConfidenceThreshold,
		StaleAfter:          time.Duration(workerCfg.StaleAfter) * time.Second,
		MaxAttempts:         workerCfg.MaxAttempts,
		FileManager:         fileManager,
		Bucket:              impl.cfg.AWS.Bucket,
	}, nil
}
