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

package probot

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
)

type (
	messageWorkerHandler struct {
		pg           *pg.Client
		capabilities *CapabilityRegistry
		deliveries   MessageDelivery
		logger       *log.Logger
		staleAfter   time.Duration
		retryBase    time.Duration
		retryMax     time.Duration
		now          func() time.Time
	}

	permanentMessageError struct {
		err error
	}
)

const (
	defaultMessageWorkerMaxConcurrency = 4
	defaultMessageRetryBase            = time.Second
	defaultMessageRetryMax             = 5 * time.Minute
	defaultMessageStaleAfter           = 5 * time.Minute
)

var (
	_ worker.Handler[coredata.BotMessage] = (*messageWorkerHandler)(nil)
	_ worker.StaleRecoverer               = (*messageWorkerHandler)(nil)
)

func (e *permanentMessageError) Error() string {
	return e.err.Error()
}

func (e *permanentMessageError) Unwrap() error {
	return e.err
}

func NewMessageWorker(
	pgClient *pg.Client,
	capabilities *CapabilityRegistry,
	deliveries MessageDelivery,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.BotMessage] {
	h := &messageWorkerHandler{
		pg:           pgClient,
		capabilities: capabilities,
		deliveries:   deliveries,
		logger:       logger,
		staleAfter:   defaultMessageStaleAfter,
		retryBase:    defaultMessageRetryBase,
		retryMax:     defaultMessageRetryMax,
		now:          time.Now,
	}

	workerOpts := append(
		[]worker.Option{worker.WithMaxConcurrency(defaultMessageWorkerMaxConcurrency)},
		opts...,
	)

	return worker.New(
		"probot-message-worker",
		h,
		logger,
		workerOpts...,
	)
}

func (h *messageWorkerHandler) Claim(ctx context.Context) (coredata.BotMessage, error) {
	var message coredata.BotMessage

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return message.ClaimNextForUpdateSkipLocked(ctx, tx, h.now())
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.BotMessage{}, worker.ErrNoTask
		}

		return coredata.BotMessage{}, fmt.Errorf("cannot claim bot message: %w", err)
	}

	return message, nil
}

func (h *messageWorkerHandler) Process(ctx context.Context, message coredata.BotMessage) error {
	logger := h.logger.With(log.String("message_id", message.ID.String()))

	err := h.deliver(ctx, message)
	processingErr := err

	persistErr := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			now := h.now()
			message.ProcessingStartedAt = nil
			message.UpdatedAt = now

			if processingErr == nil {
				message.ProcessedAt = &now
				message.NextAttemptAt = nil
				message.LastError = nil
				message.DeadLetteredAt = nil

				return message.UpdateProcessingState(ctx, conn)
			}

			message.LastError = new(processingErr.Error())
			if isPermanentMessageError(processingErr) {
				message.AttemptCount = message.MaxAttempts
				message.NextAttemptAt = nil
				message.DeadLetteredAt = &now
			} else if message.AttemptCount < message.MaxAttempts {
				nextAttemptAt := now.Add(h.retryDelay(message.AttemptCount))
				message.NextAttemptAt = &nextAttemptAt
			} else {
				message.NextAttemptAt = nil
				message.DeadLetteredAt = &now
			}

			return message.UpdateProcessingState(ctx, conn)
		},
	)
	if persistErr != nil {
		return fmt.Errorf("cannot persist bot message outcome: %w", persistErr)
	}

	if processingErr != nil {
		logger.ErrorCtx(
			ctx,
			"cannot process bot message",
			log.Int("attempt_count", message.AttemptCount),
			log.Bool("dead_lettered", message.DeadLetteredAt != nil),
			log.Error(processingErr),
		)

		return processingErr
	}

	logger.InfoCtx(ctx, "processed bot message", log.Int("attempt_count", message.AttemptCount))

	return nil
}

func (h *messageWorkerHandler) deliver(
	ctx context.Context,
	message coredata.BotMessage,
) error {
	if h.capabilities == nil {
		return &permanentMessageError{err: fmt.Errorf("bot message capabilities are unavailable")}
	}

	if h.deliveries == nil {
		return &permanentMessageError{err: fmt.Errorf("bot message delivery is unavailable")}
	}

	var attributes map[string]any
	if len(message.Attributes) > 0 {
		if err := json.Unmarshal(message.Attributes, &attributes); err != nil {
			return &permanentMessageError{
				err: fmt.Errorf("cannot decode bot message attributes: %w", err),
			}
		}
	}

	result, err := h.capabilities.BuildOutboundMessage(
		ctx,
		message.Capability,
		message.OrganizationID,
		message.MessageType,
		attributes,
	)
	if err != nil {
		if errors.Is(err, ErrCapabilityNotFound) || errors.Is(err, ErrCapabilityInvalidInput) {
			return &permanentMessageError{err: err}
		}

		return fmt.Errorf("cannot build outbound bot message: %w", err)
	}

	if err := h.deliveries.DeliverOutbound(
		ctx,
		OutboundDelivery{
			OrganizationID:   message.OrganizationID,
			Purpose:          message.Purpose,
			SourceEventID:    message.ID.String(),
			SubjectNamespace: message.SubjectNamespace,
			SubjectKey:       message.SubjectKey,
			Capability:       message.Capability,
			MessageType:      message.MessageType,
			Attributes:       attributes,
			Result:           result,
		},
	); err != nil {
		return fmt.Errorf("cannot deliver outbound bot message: %w", err)
	}

	return nil
}

func (h *messageWorkerHandler) RecoverStale(ctx context.Context) error {
	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleProcessingBotMessages(
				ctx,
				conn,
				h.now(),
				h.staleAfter,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot recover stale bot messages: %w", err)
	}

	return nil
}

func (h *messageWorkerHandler) retryDelay(attempt int) time.Duration {
	return ExponentialRetryDelay(attempt, h.retryBase, h.retryMax)
}

func isPermanentMessageError(err error) bool {
	_, ok := errors.AsType[*permanentMessageError](err)

	return ok
}
