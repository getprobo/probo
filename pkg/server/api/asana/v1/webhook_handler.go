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

package asana_v1

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/asana"
)

type webhookEnqueuer interface {
	StoreAsanaHookSecret(ctx context.Context, connectorID gid.GID, nonce string, secret string) error
	AsanaHookSecret(ctx context.Context, connectorID gid.GID, nonce string) (string, error)
	EnqueueAsanaWebhook(ctx context.Context, connectorID gid.GID, body []byte) (bool, error)
}

const maxAsanaWebhookBytes = 1 << 20

func NewMux(logger *log.Logger, svc webhookEnqueuer) *chi.Mux {
	r := chi.NewMux()
	r.Post("/webhooks/{connectorID}/{nonce}", WebhookHandler(logger, svc))

	return r
}

func WebhookHandler(logger *log.Logger, svc webhookEnqueuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		connectorID, err := gid.ParseGID(chi.URLParam(r, "connectorID"))
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid connector id"))

			return
		}

		nonce := chi.URLParam(r, "nonce")
		if nonce == "" {
			httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing webhook nonce"))

			return
		}

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxAsanaWebhookBytes))
		if err != nil {
			status := http.StatusBadRequest
			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
				status = http.StatusRequestEntityTooLarge
			}

			httpserver.RenderError(w, status, err)

			return
		}

		if secret := r.Header.Get(asana.HeaderHookSecret); secret != "" {
			if err := svc.StoreAsanaHookSecret(ctx, connectorID, nonce, secret); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					httpserver.RenderError(w, http.StatusUnauthorized, errors.New("invalid Asana signature"))

					return
				}

				logger.ErrorCtx(ctx, "cannot store Asana webhook secret", log.Error(err))
				httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

				return
			}

			w.Header().Set(asana.HeaderHookSecret, secret)
			w.WriteHeader(http.StatusOK)

			return
		}

		secret, err := svc.AsanaHookSecret(ctx, connectorID, nonce)
		if err != nil {
			httpserver.RenderError(w, http.StatusUnauthorized, errors.New("invalid Asana signature"))

			return
		}

		if !asana.VerifySignature(secret, r.Header.Get(asana.HeaderHookSignature), body) {
			httpserver.RenderError(w, http.StatusUnauthorized, errors.New("invalid Asana signature"))

			return
		}

		if _, err := svc.EnqueueAsanaWebhook(ctx, connectorID, body); err != nil {
			logger.ErrorCtx(ctx, "cannot persist Asana webhook", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
