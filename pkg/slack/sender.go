// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
)

type (
	Sender struct {
		pg            *pg.Client
		logger        *log.Logger
		encryptionKey cipher.EncryptionKey
		interval      time.Duration
	}

	Config struct {
		Interval time.Duration
	}
)

func NewSender(pg *pg.Client, logger *log.Logger, encryptionKey cipher.EncryptionKey, cfg Config) *Sender {
	return &Sender{
		pg:            pg,
		logger:        logger,
		encryptionKey: encryptionKey,
		interval:      cfg.Interval,
	}
}

func (s *Sender) Run(ctx context.Context) error {
LOOP:
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(s.interval):
		ctx := context.Background()
		if err := s.batchSendMessages(ctx); err != nil {
			s.logger.ErrorCtx(ctx, "cannot send slack message", log.Error(err))
		}

		if err := s.batchUpdateMessages(ctx); err != nil {
			s.logger.ErrorCtx(ctx, "cannot update slack message", log.Error(err))
		}

		goto LOOP
	}
}

func (s *Sender) batchSendMessages(ctx context.Context) error {
	for {
		err := s.pg.WithTx(
			ctx,
			func(tx pg.Conn) (err error) {
				message := &coredata.SlackMessage{}

				defer func() {
					if r := recover(); r != nil {
						panicErr := fmt.Sprintf("panic recovered: %v", r)
						message.Error = &panicErr
						message.UpdatedAt = time.Now()

						scope := coredata.NewScope(message.ID.TenantID())
						if updateErr := message.Update(ctx, tx, scope); updateErr != nil {
							s.logger.ErrorCtx(ctx, "cannot update slack message after panic", log.Error(updateErr))
						}

						s.logger.ErrorCtx(ctx, "panic while sending slack message", log.String("error", panicErr), log.String("message_id", message.ID.String()))
						err = fmt.Errorf("panic recovered: %v", r)
					}
				}()

				err = message.LoadNextInitalUnsentForUpdate(ctx, tx)
				if err != nil {
					return err
				}

				scope := coredata.NewScope(message.ID.TenantID())
				channelID, messageTS, sendErr := s.sendMessage(ctx, tx, message)

				now := time.Now()
				message.UpdatedAt = now

				if channelID != nil && messageTS != nil {
					message.ChannelID = channelID
					message.MessageTS = messageTS
					message.UpdatedAt = now

					if err := message.UpdateChannelAndTSByInitialMessageID(ctx, tx, scope, message.ID, *channelID, *messageTS, now); err != nil {
						return fmt.Errorf("cannot update all messages with initial message id: %w", err)
					}
				}

				if sendErr != nil {
					errorMsg := sendErr.Error()
					message.Error = &errorMsg

					if err := message.Update(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot update slack message with error: %w", err)
					}

					s.logger.ErrorCtx(ctx, "error sending slack message", log.Error(sendErr), log.String("message_id", message.ID.String()))
					return nil
				}

				message.SentAt = &now

				if err := message.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update slack message: %w", err)
				}

				return nil
			},
		)

		if errors.Is(err, coredata.ErrNoUnsentSlackMessage{}) {
			return nil
		}

		if err != nil {
			return err
		}
	}
}

func (s *Sender) sendMessage(ctx context.Context, tx pg.Conn, message *coredata.SlackMessage) (*string, *string, error) {
	tenantID := message.ID.TenantID()
	scope := coredata.NewScope(tenantID)

	var connectors coredata.Connectors
	if err := connectors.LoadAllByOrganizationIDProtocolAndProvider(
		ctx,
		tx,
		scope,
		message.OrganizationID,
		coredata.ConnectorProtocolOAuth2,
		coredata.ConnectorProviderSlack,
		s.encryptionKey,
	); err != nil {
		return nil, nil, fmt.Errorf("cannot load slack connectors: %w", err)
	}

	if len(connectors) == 0 {
		return nil, nil, fmt.Errorf("no slack connectors configured for organization")
	}

	c := connectors[0]
	if c.Connection == nil {
		return nil, nil, fmt.Errorf("slack connector has nil connection")
	}

	slackConn, ok := c.Connection.(*connector.SlackConnection)
	if !ok {
		return nil, nil, fmt.Errorf("slack connector must have SlackConnection type, got %T", c.Connection)
	}

	if slackConn.Settings.ChannelID == "" {
		return nil, nil, fmt.Errorf("slack connector %s has no channel ID", c.ID)
	}

	if slackConn.AccessToken == "" {
		return nil, nil, fmt.Errorf("slack connector %s has no access token", c.ID)
	}

	client := NewClient(s.logger)

	if message.Type == coredata.SlackMessageTypeWelcome {
		if err := client.JoinChannel(ctx, slackConn.AccessToken, slackConn.Settings.ChannelID); err != nil {
			s.logger.ErrorCtx(ctx, "cannot join Slack channel", log.Error(err))
		}
	}

	slackResp, err := client.CreateMessage(ctx, slackConn.AccessToken, slackConn.Settings.ChannelID, message.Body)
	if err != nil {
		s.logger.ErrorCtx(ctx, "cannot post message to Slack", log.Error(err))
		return nil, nil, fmt.Errorf("cannot post message to Slack: %w", err)
	}

	return &slackResp.Channel, &slackResp.TS, nil
}

func (s *Sender) batchUpdateMessages(ctx context.Context) error {
	for {
		err := s.pg.WithTx(
			ctx,
			func(tx pg.Conn) (err error) {
				updateMessage := &coredata.SlackMessage{}

				defer func() {
					if r := recover(); r != nil {
						panicErr := fmt.Sprintf("panic recovered: %v", r)
						updateMessage.Error = &panicErr
						updateMessage.UpdatedAt = time.Now()

						scope := coredata.NewScope(updateMessage.ID.TenantID())
						if updateErr := updateMessage.Update(ctx, tx, scope); updateErr != nil {
							s.logger.ErrorCtx(ctx, "cannot update slack message after panic", log.Error(updateErr))
						}

						s.logger.ErrorCtx(ctx, "panic while updating slack message", log.String("error", panicErr), log.String("message_id", updateMessage.ID.String()))
						err = fmt.Errorf("panic recovered: %v", r)
					}
				}()

				err = updateMessage.LoadNextUpdateUnsentForUpdate(ctx, tx)
				if err != nil {
					return err
				}

				scope := coredata.NewScope(updateMessage.ID.TenantID())
				updateErr := s.updateMessage(ctx, tx, updateMessage)

				now := time.Now()
				updateMessage.UpdatedAt = now

				if updateErr != nil {
					errorMsg := updateErr.Error()
					updateMessage.Error = &errorMsg
					updateMessage.UpdatedAt = time.Now()

					if err := updateMessage.Update(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot update slack message with error: %w", err)
					}

					s.logger.ErrorCtx(ctx, "error updating slack message", log.Error(updateErr), log.String("message_id", updateMessage.ID.String()))
					return nil
				}

				updateMessage.SentAt = &now

				if err := updateMessage.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update slack message: %w", err)
				}

				return nil
			},
		)

		if errors.Is(err, coredata.ErrNoUnsentSlackMessage{}) {
			return nil
		}

		if err != nil {
			return err
		}
	}
}

func (s *Sender) updateMessage(ctx context.Context, tx pg.Conn, updateMessage *coredata.SlackMessage) error {
	if updateMessage.ChannelID == nil || updateMessage.MessageTS == nil {
		return fmt.Errorf("slack message has no channel ID or message TS")
	}

	tenantID := updateMessage.ID.TenantID()
	scope := coredata.NewScope(tenantID)

	var connectors coredata.Connectors
	if err := connectors.LoadAllByOrganizationIDProtocolAndProvider(
		ctx,
		tx,
		scope,
		updateMessage.OrganizationID,
		coredata.ConnectorProtocolOAuth2,
		coredata.ConnectorProviderSlack,
		s.encryptionKey,
	); err != nil {
		return fmt.Errorf("cannot load slack connectors: %w", err)
	}

	if len(connectors) == 0 {
		return fmt.Errorf("no slack connectors configured for organization")
	}

	c := connectors[0]
	if c.Connection == nil {
		return fmt.Errorf("slack connector has nil connection")
	}

	slackConn, ok := c.Connection.(*connector.SlackConnection)
	if !ok {
		return fmt.Errorf("slack connector must have SlackConnection type, got %T", c.Connection)
	}

	if slackConn.AccessToken == "" {
		return fmt.Errorf("slack connector %s has no access token", c.ID)
	}

	client := NewClient(s.logger)

	if err := client.UpdateMessage(ctx, slackConn.AccessToken, *updateMessage.ChannelID, *updateMessage.MessageTS, updateMessage.Body); err != nil {
		s.logger.ErrorCtx(ctx, "cannot update message on Slack", log.Error(err))
		return fmt.Errorf("cannot update message on Slack: %w", err)
	}

	return nil
}
