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
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

func SlackHandler(
	inbox *slackchannel.InteractiveCommandInbox,
	logger *log.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpserver.RenderJSON(
				w,
				http.StatusBadRequest,
				slackchannel.InteractiveResponse{
					Success: false,
					Message: "cannot read request body",
				},
			)

			return
		}

		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/x-www-form-urlencoded" {
			httpserver.RenderJSON(
				w,
				http.StatusBadRequest,
				slackchannel.InteractiveResponse{
					Success: false,
					Message: "unsupported content type",
				},
			)

			return
		}

		form, err := url.ParseQuery(string(body))
		if err != nil {
			httpserver.RenderJSON(
				w,
				http.StatusBadRequest,
				slackchannel.InteractiveResponse{
					Success: false,
					Message: "cannot parse form",
				},
			)

			return
		}

		rawPayload := []byte(form.Get("payload"))
		var payload slackchannel.InteractivePayload
		if err := json.Unmarshal(rawPayload, &payload); err != nil {
			httpserver.RenderJSON(
				w,
				http.StatusBadRequest,
				slackchannel.InteractiveResponse{
					Success: false,
					Message: "cannot parse Slack payload",
				},
			)

			return
		}

		if payload.Team.ID == "" || payload.User.ID == "" {
			httpserver.RenderJSON(
				w,
				http.StatusBadRequest,
				slackchannel.InteractiveResponse{
					Success: false,
					Message: "cannot parse Slack payload",
				},
			)

			return
		}

		if _, err := inbox.Enqueue(ctx, rawPayload); err != nil {
			logger.ErrorCtx(ctx, "cannot persist Slack interactive command", log.Error(err))

			httpserver.RenderJSON(
				w,
				http.StatusServiceUnavailable,
				slackchannel.InteractiveResponse{
					Success: false,
					Message: "cannot accept interactive command",
				},
			)

			return
		}

		httpserver.RenderJSON(
			w,
			http.StatusOK,
			slackchannel.InteractiveResponse{Success: true},
		)
	}
}
