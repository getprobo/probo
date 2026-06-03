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
	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/thirdparty"
	"go.probo.inc/probo/pkg/vetting"
)

// buildThirdPartyVetter wires the third-party vetting agent. Unset
// third-party-vetter fields inherit from the default agent config
// (AGENT_DEFAULT_*), same as evidence-describer and probo.
func (impl *Implm) buildThirdPartyVetter(
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (thirdparty.Vetter, error) {
	agentCfg, llmClient, err := impl.resolveAgentClient("third-party-vetter", impl.cfg.Agents.ThirdPartyVetter, l, tp, r)
	if err != nil {
		return nil, err
	}

	maxTokens := vetting.DefaultMaxTokens
	if agentCfg.MaxTokens != nil {
		maxTokens = *agentCfg.MaxTokens
	}

	return vetting.NewAssessor(vetting.Config{
		Client:          llmClient,
		Model:           agentCfg.ModelName,
		MaxTokens:       maxTokens,
		ChromeAddr:      impl.cfg.ChromeDPAddr,
		FirecrawlAPIKey: impl.cfg.Agents.Tools.FirecrawlAPIKey,
		Logger:          l.Named("third-party-vetter"),
	}), nil
}
