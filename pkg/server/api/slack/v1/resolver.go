// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package slack_v1

import (
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
	"go.probo.inc/probo/pkg/slack"
)

func NewMux(
	logger *log.Logger,
	slackSvc *slack.Service,
	inbox *slackchannel.InteractiveCommandInbox,
	slackbot *slackchannel.Service,
	installations *slackchannel.InstallationService,
) *chi.Mux {
	r := chi.NewMux()

	r.Group(
		func(r chi.Router) {
			r.Use(
				newSignatureMiddleware(
					logger,
					interactiveSigningSecrets(slackSvc, installations)...,
				),
			)
			r.Post("/interactive", SlackHandler(inbox, slackbot, logger))
		},
	)

	if slackbot != nil {
		r.Group(
			func(r chi.Router) {
				r.Use(
					newSignatureMiddleware(
						logger,
						slackbotSigningSecret(installations),
					),
				)
				r.Post("/commands", SlackCommandHandler(slackbot))
				r.Post("/events", SlackEventHandler(slackbot, logger))
			},
		)
	}

	return r
}

func interactiveSigningSecrets(
	slackSvc *slack.Service,
	installations *slackchannel.InstallationService,
) []string {
	secrets := make([]string, 0, 2)
	if installations != nil {
		secrets = append(secrets, installations.SigningSecret())
	}

	if slackSvc != nil {
		secrets = append(secrets, slackSvc.GetSlackSigningSecret())
	}

	return secrets
}

func slackbotSigningSecret(installations *slackchannel.InstallationService) string {
	if installations == nil {
		return ""
	}

	return installations.SigningSecret()
}
