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

package accessreview

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

type inventorySyncSchedulerHandler struct {
	svc    *Service
	pg     *pg.Client
	logger *log.Logger
}

func NewInventorySyncSchedulerWorker(
	svc *Service,
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AccessReviewSource] {
	h := &inventorySyncSchedulerHandler{
		svc:    svc,
		pg:     pgClient,
		logger: logger,
	}

	return worker.New(
		"inventory-sync-scheduler",
		h,
		logger,
		opts...,
	)
}

func (h *inventorySyncSchedulerHandler) Claim(
	ctx context.Context,
) (coredata.AccessReviewSource, error) {
	var source coredata.AccessReviewSource

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()
			if err := source.LoadNextDueForInventorySyncSkipLocked(ctx, tx, now); err != nil {
				return err
			}

			scope := coredata.NewScopeFromObjectID(source.ID)

			if _, err := h.svc.enqueueInventorySyncRun(ctx, tx, scope, &source); err != nil {
				return fmt.Errorf("cannot enqueue inventory sync run: %w", err)
			}

			if err := h.svc.advanceInventorySyncCadence(ctx, tx, scope, &source, now); err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrNoAccessReviewSourceInventorySyncDue) {
			return coredata.AccessReviewSource{}, worker.ErrNoTask
		}

		return coredata.AccessReviewSource{}, fmt.Errorf("cannot claim inventory sync schedule: %w", err)
	}

	return source, nil
}

func (h *inventorySyncSchedulerHandler) Process(
	ctx context.Context,
	source coredata.AccessReviewSource,
) error {
	h.logger.InfoCtx(
		ctx,
		"enqueued inventory sync from cadence",
		log.String("access_review_source_id", source.ID.String()),
	)

	return nil
}
