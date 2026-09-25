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
	"go.probo.inc/probo/pkg/task/sync/linear"
)

const webhookHeartbeatInterval = 30 * time.Second

type webhookHandler struct {
	svc               *Service
	pg                *pg.Client
	logger            *log.Logger
	staleAfter        time.Duration
	heartbeatInterval time.Duration
}

func NewWebhookWorker(
	svc *Service,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.LinearWebhookEvent] {
	staleAfter := 5 * time.Minute
	h := &webhookHandler{
		svc:               svc,
		pg:                svc.pg,
		logger:            logger,
		staleAfter:        staleAfter,
		heartbeatInterval: outboundLeaseHeartbeatInterval(staleAfter, webhookHeartbeatInterval),
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
	handler := *h
	handler.logger = h.logger.With(log.String("delivery_id", item.DeliveryID))

	runCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	stopHeartbeat := handler.startHeartbeat(runCtx, cancel, item)
	defer stopHeartbeat()

	if err := handler.handle(runCtx, &item); err != nil {
		if processingLeaseLost(runCtx, err) {
			handler.logger.InfoCtx(ctx, "lost Linear webhook processing lease")

			return nil
		}

		if failErr := handler.fail(ctx, &item, err); failErr != nil {
			if errors.Is(failErr, coredata.ErrProcessingLeaseLost) {
				handler.logger.InfoCtx(ctx, "lost Linear webhook processing lease")

				return nil
			}

			handler.logger.ErrorCtx(ctx, "cannot fail Linear webhook event", log.Error(failErr))
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

func (h *webhookHandler) startHeartbeat(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	item coredata.LinearWebhookEvent,
) func() {
	if item.ProcessingOwnerToken == nil || *item.ProcessingOwnerToken == "" {
		panic("Linear webhook event heartbeat requires an owner token")
	}

	stop := make(chan struct{})
	exited := make(chan struct{})

	var once sync.Once

	lease := coredata.LinearWebhookEvent{
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
					cancel(fmt.Errorf("cannot heartbeat Linear webhook event: %w", err))
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
	issueID, organizationID, keep, err := webhookIssueIdentity(envelope)
	if err != nil {
		return false, err
	}

	if !keep {
		return false, nil
	}

	event := coredata.NewLinearWebhookEvent(deliveryID, envelope)

	var inserted bool

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			links := coredata.TaskExternalLinks{}
			if err := links.LoadByExternalID(
				ctx,
				conn,
				coredata.NewNoScope(),
				coredata.ConnectorProviderLinear,
				issueID,
			); err != nil {
				return fmt.Errorf("cannot load task external links: %w", err)
			}

			if _, ok := PickLinkForLinearOrganization(links, organizationID); !ok {
				return nil
			}

			var insertErr error

			inserted, insertErr = event.Insert(ctx, conn)
			if insertErr != nil {
				return fmt.Errorf("cannot insert Linear webhook event: %w", insertErr)
			}

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	return inserted, nil
}

func webhookIssueIdentity(body []byte) (issueID, organizationID string, keep bool, err error) {
	envelope, err := linear.ParseEnvelope(body)
	if err != nil {
		return "", "", false, fmt.Errorf("cannot parse Linear webhook: %w", err)
	}

	if envelope.Type != "Issue" || envelope.OrganizationID == "" {
		return "", "", false, nil
	}

	data, err := envelope.IssueData()
	if err != nil || data.ID == "" {
		return "", "", false, nil
	}

	return data.ID, envelope.OrganizationID, true, nil
}
