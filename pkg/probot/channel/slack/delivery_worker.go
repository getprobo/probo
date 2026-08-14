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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	operationDeliveryFunc func(context.Context, *coredata.SlackDeliveryOperation) error

	operationClientResolver func(
		context.Context,
		*coredata.SlackDeliveryOperation,
	) (*Client, error)

	deliveryOperationHandler struct {
		pg         *pg.Client
		logger     *log.Logger
		staleAfter time.Duration
		retryBase  time.Duration
		retryMax   time.Duration
		now        func() time.Time
		deliver    operationDeliveryFunc
		client     operationClientResolver
	}
)

const (
	defaultDeliveryOperationMaxConcurrency = 5
	defaultDeliveryOperationRetryBase      = time.Minute
	defaultDeliveryOperationRetryMax       = time.Hour
)

var (
	_ worker.Handler[coredata.SlackDeliveryOperation] = (*deliveryOperationHandler)(nil)
	_ worker.StaleRecoverer                           = (*deliveryOperationHandler)(nil)
)

func NewDeliveryWorker(
	pgClient *pg.Client,
	installations *InstallationService,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.SlackDeliveryOperation] {
	workerOpts := append(
		[]worker.Option{worker.WithMaxConcurrency(defaultDeliveryOperationMaxConcurrency)},
		opts...,
	)
	h := &deliveryOperationHandler{
		pg:         pgClient,
		logger:     logger,
		staleAfter: 5 * time.Minute,
		retryBase:  defaultDeliveryOperationRetryBase,
		retryMax:   defaultDeliveryOperationRetryMax,
		now:        time.Now,
		client: func(
			ctx context.Context,
			operation *coredata.SlackDeliveryOperation,
		) (*Client, error) {
			if installations == nil {
				return nil, &permanentDeliveryError{
					err: fmt.Errorf("slack installation service is unavailable"),
				}
			}

			client, installation, err := installations.ClientByOrganizationID(
				ctx,
				coredata.NewScopeFromObjectID(operation.OrganizationID),
				operation.OrganizationID,
			)
			if err != nil {
				return nil, fmt.Errorf("cannot resolve current Slack installation: %w", err)
			}

			if installation.OrganizationID != operation.OrganizationID {
				return nil, &permanentDeliveryError{
					err: fmt.Errorf("slack installation organization mismatch"),
				}
			}

			return client, nil
		},
	}

	return worker.New("slack-delivery-operation-worker", h, logger, workerOpts...)
}

func (h *deliveryOperationHandler) Claim(
	ctx context.Context,
) (coredata.SlackDeliveryOperation, error) {
	var operation coredata.SlackDeliveryOperation

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return operation.ClaimNextForUpdateSkipLocked(ctx, tx, h.now())
		},
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return coredata.SlackDeliveryOperation{}, worker.ErrNoTask
	}

	if err != nil {
		return coredata.SlackDeliveryOperation{}, fmt.Errorf(
			"cannot claim Slack delivery operation: %w",
			err,
		)
	}

	return operation, nil
}

func (h *deliveryOperationHandler) Process(
	ctx context.Context,
	operation coredata.SlackDeliveryOperation,
) error {
	logger := h.logger.With(
		log.String("operation_id", operation.ID.String()),
		log.String("organization_id", operation.OrganizationID.String()),
		log.String("operation_kind", operation.OperationKind.String()),
	)

	var deliveryErr error
	if h.deliver != nil {
		deliveryErr = h.deliver(ctx, &operation)
	} else {
		deliveryErr = h.deliverOperation(ctx, &operation)
	}

	now := h.now()
	operation.ProcessingStartedAt = nil

	operation.UpdatedAt = now
	if deliveryErr == nil {
		operation.CompletedAt = &now
		operation.NextAttemptAt = nil
		operation.LastError = nil
	} else {
		operation.LastError = new(deliveryErr.Error())
		if isTransientDeliveryError(deliveryErr) &&
			operation.AttemptCount < operation.MaxAttempts {
			next := now.Add(h.retryDelay(operation.AttemptCount, deliveryErr))
			operation.NextAttemptAt = &next
		} else {
			operation.NextAttemptAt = nil
			operation.DeadLetteredAt = &now
		}
	}

	scope := coredata.NewScopeFromObjectID(operation.OrganizationID)

	err := h.pg.WithConn(
		context.WithoutCancel(ctx),
		func(ctx context.Context, conn pg.Querier) error {
			return operation.UpdateDeliveryState(ctx, conn, scope)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot persist Slack delivery operation result: %w", err)
	}

	if deliveryErr != nil {
		logger.ErrorCtx(
			ctx,
			"cannot deliver Slack operation",
			log.Int("attempt_count", operation.AttemptCount),
			log.Bool("dead_lettered", operation.DeadLetteredAt != nil),
			log.Error(deliveryErr),
		)

		return deliveryErr
	}

	logger.InfoCtx(ctx, "delivered Slack operation", log.Int("attempt_count", operation.AttemptCount))

	return nil
}

func (h *deliveryOperationHandler) RecoverStale(ctx context.Context) error {
	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleSlackDeliveryOperations(
				ctx,
				conn,
				h.now(),
				h.staleAfter,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot recover stale Slack delivery operations: %w", err)
	}

	return nil
}

func (h *deliveryOperationHandler) deliverOperation(
	ctx context.Context,
	operation *coredata.SlackDeliveryOperation,
) error {
	client, err := h.client(ctx, operation)
	if err != nil {
		return fmt.Errorf("cannot resolve Slack delivery client: %w", err)
	}

	switch operation.OperationKind {
	case coredata.SlackDeliveryOperationKindPostMessage:
		channel, text, threadTS, ok := postMessageOperationPayload(operation.Payload)
		if !ok || operation.ClientMsgID == nil {
			return &permanentDeliveryError{err: fmt.Errorf("invalid post message operation payload")}
		}

		_, err := client.CreateMessage(ctx, channel, text, threadTS, *operation.ClientMsgID)
		if err != nil {
			return fmt.Errorf("cannot deliver Slack post message operation: %w", err)
		}
	case coredata.SlackDeliveryOperationKindAddReaction:
		channel, reaction, timestamp, ok := reactionOperationPayload(operation.Payload)
		if !ok {
			return &permanentDeliveryError{err: fmt.Errorf("invalid reaction operation payload")}
		}

		err := client.AddReaction(ctx, channel, reaction, timestamp)
		if apiErr, ok := errors.AsType[*APIError](err); ok && apiErr.Code == "already_reacted" {
			return nil
		}

		if err != nil {
			return fmt.Errorf("cannot deliver Slack reaction operation: %w", err)
		}
	default:
		return &permanentDeliveryError{err: fmt.Errorf("unsupported Slack delivery operation kind")}
	}

	return nil
}

func (h *deliveryOperationHandler) retryDelay(attempt int, err error) time.Duration {
	return slackAPIRetryDelay(attempt, h.retryBase, h.retryMax, err)
}

func postMessageOperationPayload(payload map[string]any) (string, string, string, bool) {
	channel, channelOK := payload["channel"].(string)
	text, textOK := payload["text"].(string)
	threadTS, threadOK := payload["thread_ts"].(string)

	return channel, text, threadTS, channelOK && channel != "" && textOK && threadOK
}

func reactionOperationPayload(payload map[string]any) (string, string, string, bool) {
	channel, channelOK := payload["channel"].(string)
	reaction, reactionOK := payload["reaction"].(string)
	timestamp, timestampOK := payload["timestamp"].(string)

	return channel, reaction, timestamp,
		channelOK && channel != "" && reactionOK && reaction != "" &&
			timestampOK && timestamp != ""
}
