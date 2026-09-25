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
	"encoding/json"
	"errors"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

func SlackEventHandler(
	slackbot *slackchannel.Service,
	logger *log.Logger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var envelope slackchannel.Envelope
		if err := json.NewDecoder(r.Body).Decode(&envelope); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid json"))

			return
		}

		if envelope.Type == slackchannel.EnvelopeTypeURLVerification {
			if envelope.Challenge == "" {
				httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing challenge"))
				return
			}

			httpserver.RenderJSON(
				w,
				http.StatusOK,
				map[string]string{"challenge": envelope.Challenge},
			)

			return
		}

		if err := slackbot.EnqueueEvent(ctx, envelope); err != nil {
			if _, ok := errors.AsType[*slackchannel.InvalidEnvelopeError](err); ok {
				httpserver.RenderError(w, http.StatusBadRequest, err)
				return
			}

			logger.ErrorCtx(
				ctx,
				"cannot persist Slackbot event before acknowledgement",
				log.String("event_id", envelope.EventID),
				log.Error(err),
			)
			httpserver.RenderError(w, http.StatusInternalServerError, err)

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
