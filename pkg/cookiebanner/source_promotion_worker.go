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

package cookiebanner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	bannerSourcePromotionMaxAttempts = 3
	bannerSourcePromotionStaleAfter  = 10 * time.Minute
)

type bannerSourcePromotionTask struct {
	banner *coredata.CookieBanner
}

type bannerSourcePromotionHandler struct {
	svc    *Service
	pg     *pg.Client
	logger *log.Logger
}

func NewBannerSourcePromotionWorker(
	svc *Service,
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[bannerSourcePromotionTask] {
	h := &bannerSourcePromotionHandler{
		svc:    svc,
		pg:     pgClient,
		logger: logger,
	}

	opts = append([]worker.Option{worker.WithMaxConcurrency(1)}, opts...)

	return worker.New(
		"banner-source-promotion-worker",
		h,
		logger,
		opts...,
	)
}

func (h *bannerSourcePromotionHandler) Claim(ctx context.Context) (bannerSourcePromotionTask, error) {
	var task bannerSourcePromotionTask

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var banner coredata.CookieBanner
			if err := banner.LoadNextForSourcePromotionForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			if err := banner.ClearSourcePromotionRequested(ctx, tx); err != nil {
				return err
			}

			task.banner = &banner

			return nil
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return bannerSourcePromotionTask{}, worker.ErrNoTask
		}

		return bannerSourcePromotionTask{}, err
	}

	return task, nil
}

func (h *bannerSourcePromotionHandler) Process(ctx context.Context, task bannerSourcePromotionTask) error {
	if task.banner == nil {
		return nil
	}

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := h.svc.promoteSiblingCookieSources(ctx, tx, task.banner); err != nil {
				return err
			}

			return task.banner.ResetSourcePromotionAttempts(ctx, tx)
		},
	)
	if err != nil {
		if task.banner.SourcePromotionAttempts < bannerSourcePromotionMaxAttempts {
			if retryErr := h.pg.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					return task.banner.RequeueSourcePromotionRequested(ctx, tx)
				},
			); retryErr != nil {
				h.logger.ErrorCtx(ctx, "cannot requeue banner source promotion", log.Error(retryErr))
			}
		} else {
			h.logger.ErrorCtx(
				ctx,
				"banner source promotion exhausted its attempts",
				log.Error(err),
				log.Int("attempts", task.banner.SourcePromotionAttempts),
			)
		}

		return err
	}

	return nil
}

func (h *bannerSourcePromotionHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleSourcePromotions(
				ctx,
				conn,
				bannerSourcePromotionStaleAfter,
				bannerSourcePromotionMaxAttempts,
			); err != nil {
				return fmt.Errorf("cannot reset stale banner source promotions: %w", err)
			}

			return nil
		},
	)
}
