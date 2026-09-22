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

package linear_v1

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

type webhookEnqueuer interface {
	EnqueueWebhook(ctx context.Context, deliveryID string, envelope []byte) (bool, error)
}

const (
	maxLinearWebhookBytes = 1 << 20
)

func NewMux(
	logger *log.Logger,
	svc webhookEnqueuer,
	webhookSecret string,
) *chi.Mux {
	r := chi.NewMux()
	r.Post("/webhooks", WebhookHandler(logger, svc, webhookSecret))

	return r
}

func WebhookHandler(
	logger *log.Logger,
	svc webhookEnqueuer,
	webhookSecret string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if webhookSecret == "" {
			httpserver.RenderError(w, http.StatusServiceUnavailable, errors.New("linear webhooks are not configured"))
			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxLinearWebhookBytes))
		if err != nil {
			status := http.StatusBadRequest
			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
				status = http.StatusRequestEntityTooLarge
			}

			httpserver.RenderError(w, status, err)

			return
		}

		signature := r.Header.Get("Linear-Signature")
		if !linear.VerifySignature(webhookSecret, signature, body) {
			httpserver.RenderError(w, http.StatusUnauthorized, errors.New("invalid Linear signature"))
			return
		}

		if _, err := linear.ParseEnvelope(body); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid json"))
			return
		}

		deliveryID := r.Header.Get("Linear-Delivery")
		if deliveryID == "" {
			httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing Linear delivery id"))
			return
		}

		if _, err := svc.EnqueueWebhook(ctx, deliveryID, body); err != nil {
			logger.ErrorCtx(ctx, "cannot persist Linear webhook", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
