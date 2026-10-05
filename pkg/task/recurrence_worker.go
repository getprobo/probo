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

package task

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type recurrenceHandler struct {
	pg     *pg.Client
	logger *log.Logger
}

var _ worker.Handler[coredata.Task] = (*recurrenceHandler)(nil)

func NewRecurrenceWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.Task] {
	return worker.New(
		"task-recurrence",
		&recurrenceHandler{
			pg:     pgClient,
			logger: logger,
		},
		logger,
		opts...,
	)
}

func (h *recurrenceHandler) Claim(ctx context.Context) (coredata.Task, error) {
	for {
		next, stopped, err := h.claimNext(ctx)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return coredata.Task{}, worker.ErrNoTask
			}

			return coredata.Task{}, err
		}

		if !stopped {
			return next, nil
		}
	}
}

func (h *recurrenceHandler) claimNext(ctx context.Context) (coredata.Task, bool, error) {
	var (
		next     *coredata.Task
		sourceID gid.GID
		stopped  bool
		stopErr  error
	)

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()

			var source coredata.Task
			if err := source.LoadNextDueRecurringForUpdateSkipLocked(ctx, tx, now); err != nil {
				return err
			}

			cloned, err := repeatLockedRecurringTask(ctx, tx, &source, now)
			if err != nil {
				// This deadline can never move forward, and it stays first
				// in the queue. Drop the interval so the next row can run.
				if !errors.Is(err, errRecurrenceCannotAdvance) {
					return err
				}

				previous := source

				source.Recurrence = nil

				source.UpdatedAt = now

				scope := coredata.NewScopeFromObjectID(source.ID)
				if err := source.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot stop recurring task %q: %w", source.ID, err)
				}

				if err := emitTaskUpdated(ctx, tx, scope, &previous, &source); err != nil {
					return fmt.Errorf("cannot emit stopped recurring task updated webhook: %w", err)
				}

				sourceID = source.ID

				stopErr = err

				stopped = true

				return nil
			}

			sourceID = source.ID

			next = cloned

			return nil
		},
	)
	if err != nil {
		return coredata.Task{}, false, err
	}

	if stopped {
		h.logger.ErrorCtx(
			ctx,
			"stopped recurring task that cannot advance",
			log.String("task_id", sourceID.String()),
			log.Error(stopErr),
		)

		return coredata.Task{}, true, nil
	}

	h.logger.InfoCtx(
		ctx,
		"created next recurring task",
		log.String("task_id", sourceID.String()),
		log.String("next_task_id", next.ID.String()),
	)

	return *next, false, nil
}

// The clone is committed in Claim so the row lock covers the insert.
func (h *recurrenceHandler) Process(context.Context, coredata.Task) error {
	return nil
}
