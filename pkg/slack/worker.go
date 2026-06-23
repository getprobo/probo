// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package slack

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

type (
	sendingHandler struct {
		pg            *pg.Client
		logger        *log.Logger
		encryptionKey cipher.EncryptionKey
		staleAfter    time.Duration
	}

	SendingWorkerOption func(*sendingHandler)
)

var (
	_ worker.Handler[coredata.SlackMessage] = (*sendingHandler)(nil)
	_ worker.StaleRecoverer                 = (*sendingHandler)(nil)
)

func WithSendingWorkerStaleAfter(d time.Duration) SendingWorkerOption {
	return func(h *sendingHandler) { h.staleAfter = d }
}

func NewSendingWorker(
	pgClient *pg.Client,
	logger *log.Logger,
	encryptionKey cipher.EncryptionKey,
	handlerOpts []SendingWorkerOption,
	workerOpts ...worker.Option,
) *worker.Worker[coredata.SlackMessage] {
	h := &sendingHandler{
		pg:            pgClient,
		logger:        logger,
		encryptionKey: encryptionKey,
		staleAfter:    5 * time.Minute,
	}

	for _, opt := range handlerOpts {
		opt(h)
	}

	return worker.New(
		"slack-sending-worker",
		h,
		logger,
		workerOpts...,
	)
}

func (h *sendingHandler) Claim(ctx context.Context) (coredata.SlackMessage, error) {
	var message coredata.SlackMessage

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := message.LoadNextClaimableForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			message.ProcessingStartedAt = &now
			message.UpdatedAt = now

			scope := coredata.NewScope(message.ID.TenantID())
			if err := message.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot mark slack message as processing: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrNoUnsentSlackMessage{}) {
			return coredata.SlackMessage{}, worker.ErrNoTask
		}

		return coredata.SlackMessage{}, err
	}

	return message, nil
}

func (h *sendingHandler) Process(ctx context.Context, message coredata.SlackMessage) error {
	isInitial := message.ID == message.InitialSlackMessageID

	var (
		channelID *string
		messageTS *string
		sendErr   error
	)

	if isInitial {
		channelID, messageTS, sendErr = h.sendMessage(ctx, &message)
	} else {
		sendErr = h.updateMessage(ctx, &message)
	}

	scope := coredata.NewScope(message.ID.TenantID())

	if commitErr := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			now := time.Now()
			message.UpdatedAt = now
			message.ProcessingStartedAt = nil

			if sendErr != nil {
				errorMsg := sendErr.Error()
				message.Error = &errorMsg

				if err := message.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update slack message with error: %w", err)
				}

				return nil
			}

			if isInitial && channelID != nil && messageTS != nil {
				message.ChannelID = channelID
				message.MessageTS = messageTS

				if err := message.UpdateChannelAndTSByInitialMessageID(ctx, tx, scope, message.ID, *channelID, *messageTS, now); err != nil {
					return fmt.Errorf("cannot update all messages with initial message id: %w", err)
				}
			}

			message.SentAt = &now

			if err := message.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update slack message: %w", err)
			}

			return nil
		},
	); commitErr != nil {
		return commitErr
	}

	if sendErr != nil {
		h.logger.ErrorCtx(ctx, "error processing slack message", log.Error(sendErr), log.String("message_id", message.ID.String()))
		return sendErr
	}

	return nil
}

func (h *sendingHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleProcessingSlackMessages(ctx, conn, h.staleAfter)
		},
	)
}

func (h *sendingHandler) loadSlackConnection(ctx context.Context, message *coredata.SlackMessage) (*connector.SlackConnection, error) {
	scope := coredata.NewScope(message.ID.TenantID())

	var c coredata.Connector

	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return c.LoadOneByOrganizationIDAndProvider(
				ctx,
				conn,
				scope,
				h.encryptionKey,
				message.OrganizationID,
				coredata.ConnectorProviderSlack,
			)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, fmt.Errorf("no connector configured for organization")
		}

		return nil, err
	}

	if c.Connection == nil {
		return nil, fmt.Errorf("connector has nil connection")
	}

	slackConn, ok := c.Connection.(*connector.SlackConnection)
	if !ok {
		return nil, fmt.Errorf("unexpected connection type %T", c.Connection)
	}

	if slackConn.AccessToken == "" {
		return nil, fmt.Errorf("connector %s has no access token", c.ID)
	}

	return slackConn, nil
}

func (h *sendingHandler) sendMessage(ctx context.Context, message *coredata.SlackMessage) (*string, *string, error) {
	slackConn, err := h.loadSlackConnection(ctx, message)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot send slack message: %w", err)
	}

	if slackConn.Settings.ChannelID == "" {
		return nil, nil, fmt.Errorf("cannot send slack message: connector has no channel ID")
	}

	client := NewClient(h.logger)

	if message.Type == coredata.SlackMessageTypeWelcome {
		if err := client.JoinChannel(ctx, slackConn.AccessToken, slackConn.Settings.ChannelID); err != nil {
			h.logger.ErrorCtx(ctx, "cannot join Slack channel", log.Error(err))
		}
	}

	slackResp, err := client.CreateMessage(ctx, slackConn.AccessToken, slackConn.Settings.ChannelID, message.Body)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot post message to Slack", log.Error(err))
		return nil, nil, fmt.Errorf("cannot post message to Slack: %w", err)
	}

	return &slackResp.Channel, &slackResp.TS, nil
}

func (h *sendingHandler) updateMessage(ctx context.Context, message *coredata.SlackMessage) error {
	if message.ChannelID == nil || message.MessageTS == nil {
		return fmt.Errorf("cannot update slack message: missing channel ID or message TS")
	}

	slackConn, err := h.loadSlackConnection(ctx, message)
	if err != nil {
		return fmt.Errorf("cannot update slack message: %w", err)
	}

	client := NewClient(h.logger)

	if err := client.UpdateMessage(ctx, slackConn.AccessToken, *message.ChannelID, *message.MessageTS, message.Body); err != nil {
		h.logger.ErrorCtx(ctx, "cannot update message on Slack", log.Error(err))
		return fmt.Errorf("cannot update message on Slack: %w", err)
	}

	return nil
}
