// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package certmanager

import (
	"context"
	"net/http"
	"strings"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

type ACMEChallengeHandler struct {
	pg            *pg.Client
	encryptionKey cipher.EncryptionKey
	logger        *log.Logger
}

func NewACMEChallengeHandler(
	pg *pg.Client,
	encryptionKey cipher.EncryptionKey,
	logger *log.Logger,
) *ACMEChallengeHandler {
	return &ACMEChallengeHandler{
		pg:            pg,
		encryptionKey: encryptionKey,
		logger:        logger.Named("certmanager.acme-challenge-handler"),
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
		func(conn pg.Conn) error {
			domain := &coredata.CustomDomain{}
			if err := domain.LoadByHTTPChallengeToken(ctx, conn, coredata.NewNoScope(), h.encryptionKey, token); err != nil {
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
