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
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/slackbot"
)

func (impl *Implm) buildSlackbotBindingService(
	pgClient *pg.Client,
	baseURL *baseurl.BaseURL,
) *slackbot.BindingService {
	return slackbot.NewBindingService(
		pgClient,
		impl.cfg.Auth.Cookie.Secret,
		baseURL,
	)
}

func (impl *Implm) buildSlackbotHandler(
	ctx context.Context,
	pgClient *pg.Client,
	bindings *slackbot.BindingService,
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (*slackbot.Handler, error) {
	if !impl.cfg.Slackbot.Enabled {
		return nil, nil
	}

	signingSecret := impl.cfg.GetSlackbotSigningSecret()
	if signingSecret == "" {
		return nil, fmt.Errorf("probod.slackbot.signing-secret is required when slackbot is enabled")
	}
	if impl.cfg.Slackbot.BotToken == "" {
		return nil, fmt.Errorf("probod.slackbot.bot-token is required when slackbot is enabled")
	}

	agentCfg, llmClient, err := impl.resolveAgentClient("slackbot", impl.cfg.Agents.Slackbot, l, tp, r)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve slackbot agent client: %w", err)
	}

	slackClient := slackbot.NewClient(impl.cfg.Slackbot.BotToken, l.Named("slackbot.client"))
	rootAgent := slackbot.NewAgent(
		llmClient,
		l.Named("slackbot.agent"),
		agent.WithModel(agentCfg.ModelName),
		agent.WithTools(slackbot.Tools(slackClient)...),
	)

	return slackbot.NewHandler(
		signingSecret,
		rootAgent,
		slackClient,
		bindings,
		pgClient,
		ctx,
		l.Named("slackbot"),
	), nil
}
