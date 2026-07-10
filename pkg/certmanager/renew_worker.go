// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

type renewHandler struct {
	pg            *pg.Client
	encryptionKey cipher.EncryptionKey
	logger        *log.Logger
}

var (
	_ worker.Handler[coredata.Certificate] = (*renewHandler)(nil)
	_ worker.StaleRecoverer                = (*renewHandler)(nil)
)

func NewRenewWorker(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.Certificate] {
	h := &renewHandler{
		pg:            pgClient,
		encryptionKey: encryptionKey,
		logger:        logger,
	}

	return worker.New(
		"certificate-renew-worker",
		h,
		logger,
		opts...,
	)
}

func (h *renewHandler) Claim(ctx context.Context) (coredata.Certificate, error) {
	var certificate coredata.Certificate

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := certificate.LoadNextForRenewalForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			certificate.Status = coredata.CertificateStatusRenewing
			if err := certificate.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot update certificate status: %w", err)
			}

			h.logger.InfoCtx(
				ctx,
				"queued certificate for renewal",
				log.String("hostname", certificate.Hostname),
			)

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.Certificate{}, worker.ErrNoTask
		}

		return coredata.Certificate{}, err
	}

	return certificate, nil
}

func (h *renewHandler) Process(_ context.Context, _ coredata.Certificate) error {
	return nil
}

func (h *renewHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var caches coredata.CachedCertificates

			cacheCount, err := caches.CountAll(ctx, conn)
			if err != nil {
				return fmt.Errorf("cannot count certificate cache: %w", err)
			}

			if cacheCount == 0 {
				h.logger.InfoCtx(ctx, "certificate cache is empty, rebuilding from certificates")

				warmer := NewCacheStore(h.pg, h.encryptionKey, h.logger)
				if err := warmer.WarmCache(ctx); err != nil {
					return fmt.Errorf("cannot rebuild certificate cache: %w", err)
				}

				h.logger.InfoCtx(ctx, "certificate cache rebuilt successfully")
			}

			if err := caches.CleanExpired(ctx, conn); err != nil {
				return fmt.Errorf("cannot clean certificate cache: %w", err)
			}

			return nil
		},
	)
}
