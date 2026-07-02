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
