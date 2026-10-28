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
	"sync"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

type asanaWebhookHandler struct {
	svc               *Service
	pg                *pg.Client
	logger            *log.Logger
	staleAfter        time.Duration
	heartbeatInterval time.Duration
}

func NewAsanaWebhookWorker(
	svc *Service,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.AsanaWebhookEvent] {
	staleAfter := 5 * time.Minute
	h := &asanaWebhookHandler{
		svc:               svc,
		pg:                svc.pg,
		logger:            logger,
		staleAfter:        staleAfter,
		heartbeatInterval: outboundLeaseHeartbeatInterval(staleAfter, webhookHeartbeatInterval),
	}

	return worker.New(
		"asana-webhook",
		h,
		logger,
		opts...,
	)
}

func (h *asanaWebhookHandler) Claim(ctx context.Context) (coredata.AsanaWebhookEvent, error) {
	var item coredata.AsanaWebhookEvent

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return item.ClaimNextForUpdateSkipLocked(ctx, tx, time.Now())
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.AsanaWebhookEvent{}, worker.ErrNoTask
		}

		return coredata.AsanaWebhookEvent{}, err
	}

	return item, nil
}

func (h *asanaWebhookHandler) Process(ctx context.Context, item coredata.AsanaWebhookEvent) error {
	handler := *h
	handler.logger = h.logger.With(log.String("delivery_id", item.DeliveryID))

	runCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	stopHeartbeat := handler.startHeartbeat(runCtx, cancel, item)
	defer stopHeartbeat()

	if err := handler.handle(runCtx, &item); err != nil {
		if processingLeaseLost(runCtx, err) {
			handler.logger.InfoCtx(ctx, "lost Asana webhook processing lease")

			return nil
		}

		if failErr := handler.fail(ctx, &item, err); failErr != nil {
			if errors.Is(failErr, coredata.ErrProcessingLeaseLost) {
				handler.logger.InfoCtx(ctx, "lost Asana webhook processing lease")

				return nil
			}

			handler.logger.ErrorCtx(ctx, "cannot fail Asana webhook event", log.Error(failErr))
		}

		return err
	}

	return nil
}

func (h *asanaWebhookHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleProcessingAsanaWebhookEvents(ctx, conn, time.Now(), h.staleAfter)
		},
	)
}

func (h *asanaWebhookHandler) startHeartbeat(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	item coredata.AsanaWebhookEvent,
) func() {
	if item.ProcessingOwnerToken == nil || *item.ProcessingOwnerToken == "" {
		panic("Asana webhook event heartbeat requires an owner token")
	}

	stop := make(chan struct{})
	exited := make(chan struct{})

	var once sync.Once

	lease := coredata.AsanaWebhookEvent{
		DeliveryID:           item.DeliveryID,
		ProcessingOwnerToken: item.ProcessingOwnerToken,
	}

	go func() {
		defer close(exited)

		ticker := time.NewTicker(h.heartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := h.pg.WithConn(
					context.WithoutCancel(ctx),
					func(ctx context.Context, conn pg.Querier) error {
						return lease.TouchLease(ctx, conn, time.Now())
					},
				)
				if err == nil {
					continue
				}

				select {
				case <-stop:
					return
				default:
				}

				if errors.Is(err, coredata.ErrProcessingLeaseLost) {
					cancel(coredata.ErrProcessingLeaseLost)
				} else {
					cancel(fmt.Errorf("cannot heartbeat Asana webhook event: %w", err))
				}

				return
			}
		}
	}()

	return func() {
		once.Do(
			func() {
				close(stop)
				<-exited
			},
		)
	}
}

func (h *asanaWebhookHandler) handle(ctx context.Context, item *coredata.AsanaWebhookEvent) error {
	if err := h.svc.ApplyAsanaWebhook(ctx, item.Envelope); err != nil {
		return fmt.Errorf("cannot apply Asana webhook: %w", err)
	}

	return h.markProcessed(ctx, item)
}

func (h *asanaWebhookHandler) markProcessed(ctx context.Context, item *coredata.AsanaWebhookEvent) error {
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

func (h *asanaWebhookHandler) fail(ctx context.Context, item *coredata.AsanaWebhookEvent, processErr error) error {
	now := time.Now()
	message := processErr.Error()
	item.LastError = &message
	item.ProcessingStartedAt = nil
	item.UpdatedAt = now

	if item.AttemptCount >= coredata.AsanaWebhookEventDefaultMaxAttempts {
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
