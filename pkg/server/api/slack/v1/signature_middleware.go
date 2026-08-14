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
	"bytes"
	"errors"
	"io"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

const maxSlackRequestBytes = 1 << 20

func newSignatureMiddleware(
	logger *log.Logger,
	signingSecrets ...string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSlackRequestBytes))
				if err != nil {
					status := http.StatusBadRequest
					if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
						status = http.StatusRequestEntityTooLarge
					}

					httpserver.RenderError(w, status, err)

					return
				}

				timestamp := r.Header.Get("X-Slack-Request-Timestamp")
				signature := r.Header.Get("X-Slack-Signature")

				if timestamp == "" || signature == "" {
					httpserver.RenderError(
						w,
						http.StatusBadRequest,
						errors.New("missing Slack signature headers"),
					)

					return
				}

				if err := slackchannel.VerifyAnySignature(
					timestamp,
					signature,
					body,
					signingSecrets...,
				); err != nil {
					logger.ErrorCtx(r.Context(), "invalid Slack signature", log.Error(err))
					httpserver.RenderError(w, http.StatusUnauthorized, err)

					return
				}

				r.Body = io.NopCloser(bytes.NewReader(body))
				next.ServeHTTP(w, r)
			},
		)
	}
}
