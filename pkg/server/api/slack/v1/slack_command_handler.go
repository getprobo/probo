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
	"context"
	"io"
	"mime"
	"net/http"
	"net/url"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

type slackSlashCommander interface {
	HandleSlashCommand(
		ctx context.Context,
		cmd slackchannel.SlashCommand,
	) slackchannel.SlashCommandResponse
}

func SlackCommandHandler(
	commander slackSlashCommander,
	signingSecrets []string,
	logger *log.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpserver.RenderJSON(
				w,
				http.StatusOK,
				ephemeralCommandError("cannot read request body"),
			)
			return
		}

		timestamp := r.Header.Get("X-Slack-Request-Timestamp")
		signature := r.Header.Get("X-Slack-Signature")
		if timestamp == "" || signature == "" {
			httpserver.RenderJSON(
				w,
				http.StatusBadRequest,
				ephemeralCommandError("missing Slack signature headers"),
			)
			return
		}

		if err := slackchannel.VerifyAnySignature(
			timestamp,
			signature,
			body,
			signingSecrets...,
		); err != nil {
			logger.ErrorCtx(ctx, "invalid Slack slash command signature", log.Error(err))
			httpserver.RenderJSON(
				w,
				http.StatusUnauthorized,
				ephemeralCommandError("invalid Slack signature"),
			)
			return
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/x-www-form-urlencoded" {
			httpserver.RenderJSON(
				w,
				http.StatusOK,
				ephemeralCommandError("unsupported content type"),
			)
			return
		}

		form, err := url.ParseQuery(string(body))
		if err != nil {
			httpserver.RenderJSON(
				w,
				http.StatusOK,
				ephemeralCommandError("cannot parse form"),
			)
			return
		}

		if commander == nil {
			httpserver.RenderJSON(
				w,
				http.StatusOK,
				ephemeralCommandError("Probot is not available in this workspace."),
			)
			return
		}

		httpserver.RenderJSON(
			w,
			http.StatusOK,
			commander.HandleSlashCommand(ctx, slackchannel.ParseSlashCommand(form)),
		)
	}
}

func ephemeralCommandError(text string) slackchannel.SlashCommandResponse {
	return slackchannel.SlashCommandResponse{
		ResponseType: slackchannel.SlashResponseTypeEphemeral,
		Text:         text,
	}
}
