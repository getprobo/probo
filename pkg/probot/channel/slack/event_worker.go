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

package slack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/probot"
)

type (
	EventProcessor interface {
		ProcessEvent(ctx context.Context, envelope Envelope) error
	}

	eventWorkerHandler struct {
		pg         *pg.Client
		processor  EventProcessor
		logger     *log.Logger
		staleAfter time.Duration
		retryBase  time.Duration
		retryMax   time.Duration
		now        func() time.Time
	}
)

const (
	defaultEventWorkerMaxConcurrency = 4
	defaultEventRetryBase            = time.Second
	defaultEventRetryMax             = 5 * time.Minute
)

var (
	_ worker.Handler[coredata.SlackbotEvent] = (*eventWorkerHandler)(nil)
	_ worker.StaleRecoverer                  = (*eventWorkerHandler)(nil)
)

func NewEventWorker(
	pgClient *pg.Client,
	processor EventProcessor,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.SlackbotEvent] {
	h := &eventWorkerHandler{
		pg:         pgClient,
		processor:  processor,
		logger:     logger,
		staleAfter: 5 * time.Minute,
		retryBase:  defaultEventRetryBase,
		retryMax:   defaultEventRetryMax,
		now:        time.Now,
	}

	workerOpts := append(
		[]worker.Option{worker.WithMaxConcurrency(defaultEventWorkerMaxConcurrency)},
		opts...,
	)

	return worker.New(
		"slackbot-event-worker",
		h,
		logger,
		workerOpts...,
	)
}

func (h *eventWorkerHandler) Claim(ctx context.Context) (coredata.SlackbotEvent, error) {
	var event coredata.SlackbotEvent

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return event.ClaimNextForUpdateSkipLocked(ctx, tx, h.now())
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.SlackbotEvent{}, worker.ErrNoTask
		}

		return coredata.SlackbotEvent{}, fmt.Errorf("cannot claim Slackbot event: %w", err)
	}

	return event, nil
}

func (h *eventWorkerHandler) Process(ctx context.Context, event coredata.SlackbotEvent) error {
	logger := h.logger.With(
		log.String("event_id", event.EventID),
		log.String("event_record_id", event.ID.String()),
	)

	var envelope Envelope

	err := json.Unmarshal(event.Envelope, &envelope)
	if err != nil {
		err = permanent(fmt.Errorf("cannot decode persisted Slackbot event: %w", err))
	} else if h.processor == nil {
		err = permanent(fmt.Errorf("slackbot event processor is unavailable"))
	} else {
		err = h.processor.ProcessEvent(ctx, envelope)
	}

	processingErr := err

	persistErr := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			now := h.now()
			event.ProcessingStartedAt = nil
			event.UpdatedAt = now

			if processingErr == nil {
				event.ProcessedAt = &now
				event.NextAttemptAt = nil
				event.LastError = nil
				event.DeadLetteredAt = nil

				return event.UpdateProcessingState(ctx, conn)
			}

			event.LastError = new(processingErr.Error())
			if isPermanent(processingErr) {
				event.AttemptCount = event.MaxAttempts
				event.NextAttemptAt = nil
				event.DeadLetteredAt = &now
			} else if event.AttemptCount < event.MaxAttempts {
				nextAttemptAt := now.Add(h.retryDelay(event.AttemptCount))
				event.NextAttemptAt = &nextAttemptAt
			} else {
				event.NextAttemptAt = nil
				event.DeadLetteredAt = &now
			}

			return event.UpdateProcessingState(ctx, conn)
		},
	)
	if persistErr != nil {
		return fmt.Errorf("cannot persist Slackbot event outcome: %w", persistErr)
	}

	if processingErr != nil {
		logger.ErrorCtx(
			ctx,
			"cannot process Slackbot event",
			log.Int("attempt_count", event.AttemptCount),
			log.Bool("dead_lettered", event.DeadLetteredAt != nil),
			log.Error(processingErr),
		)

		return processingErr
	}

	logger.InfoCtx(ctx, "processed Slackbot event", log.Int("attempt_count", event.AttemptCount))

	return nil
}

func (h *eventWorkerHandler) RecoverStale(ctx context.Context) error {
	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleProcessingSlackbotEvents(
				ctx,
				conn,
				h.now(),
				h.staleAfter,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot recover stale Slackbot events: %w", err)
	}

	return nil
}

func (h *eventWorkerHandler) retryDelay(attempt int) time.Duration {
	return probot.ExponentialRetryDelay(attempt, h.retryBase, h.retryMax)
}
