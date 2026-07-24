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

package itam

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
		logger: logger.Named("itam.garbage_collector"),
	}

	return worker.New(
		"itam.garbage_collector",
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
			var token coredata.DeviceEnrollmentToken

			tokensDeleted, err := token.DeleteExpired(ctx, tx, now)
			if err != nil {
				return fmt.Errorf("cannot delete expired device enrollment tokens: %w", err)
			}

			h.logger.InfoCtx(
				ctx,
				"itam garbage collector cleaned up",
				log.Int64("device_enrollment_tokens_deleted", tokensDeleted),
			)

			return nil
		},
	)
}
