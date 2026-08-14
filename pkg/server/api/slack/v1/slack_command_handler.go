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

package slack_v1

import (
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"go.gearno.de/kit/httpserver"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

func SlackCommandHandler(slackbot *slackchannel.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpserver.RenderJSON(
				w,
				http.StatusOK,
				slackchannel.SlashCommandResponse{
					ResponseType: slackchannel.SlashResponseTypeEphemeral,
					Text:         "cannot read request body",
				},
			)

			return
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/x-www-form-urlencoded" {
			httpserver.RenderJSON(
				w,
				http.StatusOK,
				slackchannel.SlashCommandResponse{
					ResponseType: slackchannel.SlashResponseTypeEphemeral,
					Text:         "unsupported content type",
				},
			)

			return
		}

		form, err := url.ParseQuery(string(body))
		if err != nil {
			httpserver.RenderJSON(
				w,
				http.StatusOK,
				slackchannel.SlashCommandResponse{
					ResponseType: slackchannel.SlashResponseTypeEphemeral,
					Text:         "cannot parse form",
				},
			)

			return
		}

		command := slackchannel.SlashCommand{
			Command:     strings.TrimSpace(form.Get("command")),
			Text:        strings.TrimSpace(form.Get("text")),
			UserID:      strings.TrimSpace(form.Get("user_id")),
			UserName:    strings.TrimSpace(form.Get("user_name")),
			TeamID:      strings.TrimSpace(form.Get("team_id")),
			TeamDomain:  strings.TrimSpace(form.Get("team_domain")),
			ResponseURL: strings.TrimSpace(form.Get("response_url")),
		}

		httpserver.RenderJSON(
			w,
			http.StatusOK,
			slackbot.HandleSlashCommand(ctx, command),
		)
	}
}
