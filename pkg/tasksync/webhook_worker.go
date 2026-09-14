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

package tasksync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/tasksync/linear"
)

type webhookHandler struct {
	svc        *Service
	pg         *pg.Client
	logger     *log.Logger
	staleAfter time.Duration
}

func NewWebhookWorker(
	svc *Service,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.LinearWebhookEvent] {
	h := &webhookHandler{
		svc:        svc,
		pg:         svc.pg,
		logger:     logger,
		staleAfter: 5 * time.Minute,
	}

	return worker.New(
		"linear-webhook",
		h,
		logger,
		opts...,
	)
}

func (h *webhookHandler) Claim(ctx context.Context) (coredata.LinearWebhookEvent, error) {
	var item coredata.LinearWebhookEvent

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return item.ClaimNextForUpdateSkipLocked(ctx, tx, time.Now())
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.LinearWebhookEvent{}, worker.ErrNoTask
		}

		return coredata.LinearWebhookEvent{}, err
	}

	return item, nil
}

func (h *webhookHandler) Process(ctx context.Context, item coredata.LinearWebhookEvent) error {
	if err := h.handle(ctx, &item); err != nil {
		if failErr := h.fail(ctx, &item, err); failErr != nil {
			h.logger.ErrorCtx(ctx, "cannot fail Linear webhook event", log.Error(failErr))
		}

		return err
	}

	return nil
}

func (h *webhookHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleProcessingLinearWebhookEvents(ctx, conn, time.Now(), h.staleAfter)
		},
	)
}

func (h *webhookHandler) handle(ctx context.Context, item *coredata.LinearWebhookEvent) error {
	envelope, err := linear.ParseEnvelope(item.Envelope)
	if err != nil {
		return err
	}

	if err := envelope.Timestamp(time.Now()); err != nil {
		h.logger.WarnCtx(
			ctx,
			"skipping Linear webhook",
			log.String("delivery_id", item.DeliveryID),
			log.Error(err),
		)

		return h.markProcessed(ctx, item)
	}

	if err := h.svc.ApplyInboundIssue(ctx, envelope); err != nil {
		return fmt.Errorf("cannot apply Linear webhook: %w", err)
	}

	return h.markProcessed(ctx, item)
}

func (h *webhookHandler) markProcessed(ctx context.Context, item *coredata.LinearWebhookEvent) error {
	now := time.Now()
	item.ProcessedAt = &now
	item.ProcessingStartedAt = nil
	item.LastError = nil
	item.UpdatedAt = now

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return item.UpdateProcessingState(ctx, tx)
		},
	)
}

func (h *webhookHandler) fail(ctx context.Context, item *coredata.LinearWebhookEvent, processErr error) error {
	now := time.Now()
	message := processErr.Error()
	item.LastError = &message
	item.ProcessingStartedAt = nil
	item.UpdatedAt = now

	if item.AttemptCount >= coredata.LinearWebhookEventDefaultMaxAttempts {
		item.DeadLetteredAt = &now
	} else {
		next := now.Add(time.Minute)
		item.NextAttemptAt = &next
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return item.UpdateProcessingState(ctx, tx)
		},
	)
}

func (s *Service) EnqueueWebhook(ctx context.Context, deliveryID string, envelope []byte) (bool, error) {
	event := coredata.NewLinearWebhookEvent(deliveryID, envelope)

	var inserted bool

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			inserted, err = event.Insert(ctx, conn)
			if err != nil {
				return fmt.Errorf("cannot insert Linear webhook event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	return inserted, nil
}
