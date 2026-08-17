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
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.opentelemetry.io/otel/trace"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/probot"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

func (impl *Implm) buildSlackbot(
	pgClient *pg.Client,
	bindings identitybinding.Gate,
	installations *slackchannel.InstallationService,
	bindPrompts *slackchannel.BindPromptService,
	l *log.Logger,
	tp trace.TracerProvider,
	r prometheus.Registerer,
) (*slackchannel.Service, *agent.Agent, error) {
	if !impl.cfg.Slackbot.Enabled {
		return nil, nil, nil
	}

	signingSecret := impl.cfg.GetSlackbotSigningSecret()
	if signingSecret == "" {
		return nil, nil, fmt.Errorf("probod.slackbot.signing-secret is required when slackbot is enabled")
	}

	if impl.cfg.Slackbot.ClientID == "" ||
		impl.cfg.Slackbot.ClientSecret == "" ||
		impl.cfg.Slackbot.RedirectURI == "" {
		return nil, nil, fmt.Errorf("probod.slackbot OAuth client configuration is required when slackbot is enabled")
	}

	agentCfg, llmClient, err := impl.resolveAgentClient("slackbot", impl.cfg.Agents.Slackbot, l, tp, r)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot resolve slackbot agent client: %w", err)
	}

	opts := []agent.Option{agent.WithModel(agentCfg.ModelName)}
	if agentCfg.Temperature != nil {
		opts = append(opts, agent.WithTemperature(*agentCfg.Temperature))
	}
	if agentCfg.MaxTokens != nil {
		opts = append(opts, agent.WithMaxTokens(*agentCfg.MaxTokens))
	}

	rootAgent := probot.NewAgent(
		llmClient,
		l.Named("slackbot.agent"),
		opts...,
	)

	slackbot := slackchannel.NewService(
		bindings,
		installations,
		pgClient,
		l.Named("slackbot"),
	)
	slackbot.SetBindPrompts(bindPrompts)

	return slackbot, rootAgent, nil
}

func (impl *Implm) buildSlackbotInstallationService(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	endpoints provider.Endpoints,
	l *log.Logger,
) *slackchannel.InstallationService {
	if !impl.cfg.Slackbot.Enabled {
		return nil
	}

	return slackchannel.NewInstallationService(
		pgClient,
		encryptionKey,
		slackchannel.InstallationConfig{
			ClientID:      impl.cfg.Slackbot.ClientID,
			ClientSecret:  impl.cfg.Slackbot.ClientSecret,
			RedirectURI:   impl.cfg.Slackbot.RedirectURI,
			AuthURL:       endpoints.Auth,
			TokenURL:      endpoints.Token,
			APIBaseURL:    endpoints.APIBase,
			StateSecret:   impl.cfg.Auth.Cookie.Secret,
			SigningSecret: impl.cfg.GetSlackbotSigningSecret(),
		},
		l.Named("slackbot.installations"),
	)
}
