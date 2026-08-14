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
	"net"
	"net/http"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

type (
	deliveryFunc func(context.Context, *coredata.SlackbotMessage) error

	DeliverySuccessHook interface {
		OnSlackbotDeliverySuccess(
			ctx context.Context,
			tx pg.Tx,
			scope coredata.Scoper,
			message *coredata.SlackbotMessage,
		) error
	}

	notificationHandler struct {
		pg            *pg.Client
		installations *InstallationService
		successHook   DeliverySuccessHook
		logger        *log.Logger
		staleAfter    time.Duration
		retryBase     time.Duration
		retryMax      time.Duration
		now           func() time.Time
		deliver       deliveryFunc
	}
)

const (
	defaultNotificationRetryBase = time.Minute
	defaultNotificationRetryMax  = time.Hour
)

var (
	_ worker.Handler[coredata.SlackbotMessage] = (*notificationHandler)(nil)
	_ worker.StaleRecoverer                    = (*notificationHandler)(nil)
)

func NewNotificationWorker(
	pgClient *pg.Client,
	installations *InstallationService,
	successHook DeliverySuccessHook,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.SlackbotMessage] {
	h := &notificationHandler{
		pg:            pgClient,
		installations: installations,
		successHook:   successHook,
		logger:        logger,
		staleAfter:    5 * time.Minute,
		retryBase:     defaultNotificationRetryBase,
		retryMax:      defaultNotificationRetryMax,
		now:           time.Now,
	}

	return worker.New(
		"slackbot-notification-worker",
		h,
		logger,
		opts...,
	)
}

func (h *notificationHandler) Claim(ctx context.Context) (coredata.SlackbotMessage, error) {
	var message coredata.SlackbotMessage

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return message.ClaimNextForUpdateSkipLocked(ctx, tx, h.now())
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.SlackbotMessage{}, worker.ErrNoTask
		}

		return coredata.SlackbotMessage{}, fmt.Errorf("cannot claim Slackbot message: %w", err)
	}

	return message, nil
}

func (h *notificationHandler) Process(ctx context.Context, message coredata.SlackbotMessage) error {
	logger := h.logger.With(
		log.String("message_id", message.ID.String()),
		log.String("organization_id", message.OrganizationID.String()),
	)
	isInitial := message.ID == message.InitialSlackbotMessageID

	var deliveryErr error
	if h.deliver != nil {
		deliveryErr = h.deliver(ctx, &message)
	} else if isInitial {
		deliveryErr = h.deliverInitial(ctx, &message)
	} else {
		deliveryErr = h.deliverRevision(ctx, &message)
	}

	scope := coredata.NewScopeFromObjectID(message.ID)

	err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := h.now()
			message.UpdatedAt = now
			message.ProcessingStartedAt = nil

			if deliveryErr != nil {
				message.LastError = new(deliveryErr.Error())
				if isTransientDeliveryError(deliveryErr) &&
					message.AttemptCount < message.MaxAttempts {
					nextAttemptAt := now.Add(h.retryDelay(message.AttemptCount, deliveryErr))
					message.NextAttemptAt = &nextAttemptAt
					message.Error = nil
				} else {
					message.NextAttemptAt = nil
					message.Error = new(deliveryErr.Error())
				}

				if err := message.UpdateDeliveryState(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot persist Slackbot delivery error: %w", err)
				}

				return nil
			}

			message.SentAt = &now
			message.NextAttemptAt = nil
			message.LastError = nil

			message.Error = nil
			if isInitial {
				if err := message.PropagateDeliveryReferenceToRevisions(
					ctx,
					tx,
					scope,
				); err != nil {
					return fmt.Errorf("cannot propagate Slackbot delivery reference: %w", err)
				}
			}

			if err := message.UpdateDeliveryState(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot mark Slackbot message as sent: %w", err)
			}

			if h.successHook != nil {
				if err := h.successHook.OnSlackbotDeliverySuccess(
					ctx,
					tx,
					scope,
					&message,
				); err != nil {
					return fmt.Errorf("cannot run Slackbot delivery success hook: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		if deliveryErr == nil {
			if retryErr := h.persistPostDeliveryFailure(
				ctx,
				&message,
				scope,
				err,
			); retryErr != nil {
				return fmt.Errorf(
					"cannot persist Slackbot message result after transaction error %v: %w",
					err,
					retryErr,
				)
			}
		}

		return fmt.Errorf("cannot persist Slackbot message result: %w", err)
	}

	if deliveryErr != nil {
		logger.ErrorCtx(
			ctx,
			"cannot deliver Slackbot notification",
			log.Int("attempt_count", message.AttemptCount),
			log.Bool("dead_lettered", message.Error != nil),
			log.Error(deliveryErr),
		)

		return deliveryErr
	}

	logger.InfoCtx(ctx, "delivered Slackbot notification", log.Int("attempt_count", message.AttemptCount))

	return nil
}

func (h *notificationHandler) persistPostDeliveryFailure(
	ctx context.Context,
	message *coredata.SlackbotMessage,
	scope coredata.Scoper,
	persistErr error,
) error {
	return h.pg.WithTx(
		context.WithoutCancel(ctx),
		func(ctx context.Context, tx pg.Tx) error {
			now := h.now()
			message.SentAt = nil
			message.ProcessingStartedAt = nil

			message.LastError = new(persistErr.Error())
			if message.AttemptCount < message.MaxAttempts {
				nextAttemptAt := now.Add(h.retryDelay(message.AttemptCount, persistErr))
				message.NextAttemptAt = &nextAttemptAt
				message.Error = nil
			} else {
				message.NextAttemptAt = nil
				message.Error = new(persistErr.Error())
			}

			message.UpdatedAt = now

			return message.UpdateDeliveryState(ctx, tx, scope)
		},
	)
}

func (h *notificationHandler) RecoverStale(ctx context.Context) error {
	err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleProcessingSlackbotMessages(
				ctx,
				conn,
				h.staleAfter,
			)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot recover stale Slackbot notifications: %w", err)
	}

	return nil
}

func (h *notificationHandler) deliverInitial(
	ctx context.Context,
	message *coredata.SlackbotMessage,
) error {
	if message.ChannelID == nil || *message.ChannelID == "" {
		return permanent(fmt.Errorf("slackbot message has no channel ID"))
	}

	client, err := h.clientForMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("cannot load Slackbot client for initial message: %w", err)
	}

	text, blocks := slackMessageBody(message.Body)

	ref, err := client.CreateMessageWithBlocks(
		ctx,
		*message.ChannelID,
		text,
		blocks,
		message.ClientMsgID,
	)
	if err != nil {
		return fmt.Errorf("cannot create Slackbot message: %w", err)
	}

	message.ChannelID = &ref.Channel
	message.MessageTS = &ref.TS

	return nil
}

func (h *notificationHandler) deliverRevision(
	ctx context.Context,
	message *coredata.SlackbotMessage,
) error {
	if message.ChannelID == nil || *message.ChannelID == "" ||
		message.MessageTS == nil || *message.MessageTS == "" {
		return permanent(fmt.Errorf("slackbot message revision has no delivery reference"))
	}

	client, err := h.clientForMessage(ctx, message)
	if err != nil {
		return fmt.Errorf("cannot load Slackbot client for revision: %w", err)
	}

	text, blocks := slackMessageBody(message.Body)
	if err := client.UpdateMessage(
		ctx,
		*message.ChannelID,
		*message.MessageTS,
		text,
		blocks,
	); err != nil {
		return fmt.Errorf("cannot update Slackbot message: %w", err)
	}

	return nil
}

func (h *notificationHandler) clientForMessage(
	ctx context.Context,
	message *coredata.SlackbotMessage,
) (*Client, error) {
	if h.installations == nil {
		return nil, permanent(fmt.Errorf("slackbot installation service is unavailable"))
	}

	client, installation, err := h.installations.ClientByOrganizationID(
		ctx,
		coredata.NewScopeFromObjectID(message.OrganizationID),
		message.OrganizationID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load Slack client: %w", err)
	}

	if installation.OrganizationID != message.OrganizationID {
		return nil, permanent(fmt.Errorf("slackbot installation organization mismatch"))
	}

	return client, nil
}

func slackMessageBody(body map[string]any) (string, []any) {
	text, _ := body["text"].(string)
	blocks, _ := body["blocks"].([]any)

	return text, blocks
}

func (h *notificationHandler) retryDelay(attempt int, err error) time.Duration {
	return slackAPIRetryDelay(attempt, h.retryBase, h.retryMax, err)
}

func isTransientDeliveryError(err error) bool {
	if isPermanent(err) {
		return false
	}

	if apiErr, ok := errors.AsType[*APIError](err); ok {
		switch apiErr.Code {
		case "fatal_error", "internal_error", "ratelimited",
			"request_timeout", "service_unavailable":
			return true
		}

		return apiErr.StatusCode == http.StatusRequestTimeout ||
			apiErr.StatusCode == http.StatusTooEarly ||
			apiErr.StatusCode == http.StatusTooManyRequests ||
			apiErr.StatusCode >= http.StatusInternalServerError
	}

	if _, ok := errors.AsType[net.Error](err); ok {
		return true
	}

	if errors.Is(err, ErrSlackbotNotInstalled) {
		return false
	}

	return true
}
