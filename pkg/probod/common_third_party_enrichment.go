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
