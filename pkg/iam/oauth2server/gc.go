// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package oauth2server

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	DefaultGCInterval = 5 * time.Minute
)

type GarbageCollector = worker.Worker[struct{}]

type gcHandler struct {
	pg        *pg.Client
	logger    *log.Logger
	lastRunAt atomic.Int64
}

func NewGarbageCollector(
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *GarbageCollector {
	h := &gcHandler{
		pg:     pgClient,
		logger: logger.Named("oauth2server.garbage_collector"),
	}

	return worker.New(
		"oauth2server.garbage_collector",
		h,
		logger,
		append(
			[]worker.Option{
				worker.WithInterval(DefaultGCInterval),
				worker.WithMaxConcurrency(1),
			},
			opts...,
		)...,
	)
}

func (h *gcHandler) Claim(_ context.Context) (struct{}, error) {
	now := time.Now().UnixNano()
	last := h.lastRunAt.Load()

	if last > 0 && now-last < int64(DefaultGCInterval) {
		return struct{}{}, worker.ErrNoTask
	}

	if !h.lastRunAt.CompareAndSwap(last, now) {
		return struct{}{}, worker.ErrNoTask
	}

	return struct{}{}, nil
}

func (h *gcHandler) Process(ctx context.Context, _ struct{}) error {
	return h.cleanup(ctx)
}

func (h *gcHandler) cleanup(ctx context.Context) error {
	now := time.Now()

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var authCode coredata.OAuth2AuthorizationCode
			authCodesDeleted, err := authCode.DeleteExpired(ctx, tx, now)
			if err != nil {
				return fmt.Errorf("cannot delete expired authorization codes: %w", err)
			}

			var accessToken coredata.OAuth2AccessToken
			accessTokensDeleted, err := accessToken.DeleteExpired(ctx, tx, now)
			if err != nil {
				return fmt.Errorf("cannot delete expired access tokens: %w", err)
			}

			var refreshToken coredata.OAuth2RefreshToken
			refreshTokensDeleted, err := refreshToken.DeleteExpired(ctx, tx, now)
			if err != nil {
				return fmt.Errorf("cannot delete expired refresh tokens: %w", err)
			}

			var deviceCode coredata.OAuth2DeviceCode
			deviceCodesDeleted, err := deviceCode.DeleteExpired(ctx, tx, now)
			if err != nil {
				return fmt.Errorf("cannot delete expired device codes: %w", err)
			}

			h.logger.InfoCtx(
				ctx,
				"oauth2 server garbage collector cleaned up",
				log.Int64("authorization_codes_deleted", authCodesDeleted),
				log.Int64("access_tokens_deleted", accessTokensDeleted),
				log.Int64("refresh_tokens_deleted", refreshTokensDeleted),
				log.Int64("device_codes_deleted", deviceCodesDeleted),
			)

			return nil
		},
	)
}
