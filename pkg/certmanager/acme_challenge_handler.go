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

package certmanager

import (
	"context"
	"net/http"
	"strings"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

type ACMEChallengeHandler struct {
	pg     *pg.Client
	logger *log.Logger
}

func NewACMEChallengeHandler(
	pg *pg.Client,
	logger *log.Logger,
) *ACMEChallengeHandler {
	return &ACMEChallengeHandler{
		pg:     pg,
		logger: logger.Named("certmanager.acme-challenge-handler"),
	}
}

func (h *ACMEChallengeHandler) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			next.ServeHTTP(w, r)
			return
		}

		token := strings.TrimPrefix(r.URL.Path, "/.well-known/acme-challenge/")
		if token == "" {
			http.NotFound(w, r)
			return
		}

		keyAuth, err := h.getKeyAuthForToken(r.Context(), token)
		if err != nil {
			h.logger.WarnCtx(
				r.Context(),
				"cannot get key auth for token",
				log.String("token", token),
				log.Error(err),
			)

			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(keyAuth))
	})
}

func (h *ACMEChallengeHandler) getKeyAuthForToken(ctx context.Context, token string) (string, error) {
	var keyAuth string

	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			domain := &coredata.CustomDomain{}
			if err := domain.LoadByHTTPChallengeToken(ctx, conn, coredata.NewNoScope(), token); err != nil {
				return err
			}

			if domain.HTTPChallengeKeyAuth == nil {
				return http.ErrNotSupported
			}

			keyAuth = *domain.HTTPChallengeKeyAuth

			return nil
		},
	)

	return keyAuth, err
}
