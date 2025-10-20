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

	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
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

						if updateErr := message.Update(ctx, tx); updateErr != nil {
							s.logger.ErrorCtx(ctx, "cannot update slack message after panic", log.Error(updateErr))
						}

						s.logger.ErrorCtx(ctx, "panic while sending slack message", log.String("error", panicErr), log.String("message_id", message.ID.String()))
						err = fmt.Errorf("panic recovered: %v", r)
					}
				}()

				err = message.LoadNextUnsentForUpdate(ctx, tx)
				if err != nil {
					return err
				}

				channelID, messageTS, sendErr := s.sendMessage(ctx, tx, message)
				message.ChannelID = channelID
				message.MessageTS = messageTS

				now := time.Now()
				message.UpdatedAt = now

				if sendErr != nil {
					errorMsg := sendErr.Error()
					message.Error = &errorMsg
					message.UpdatedAt = time.Now()

					if err := message.Update(ctx, tx); err != nil {
						return fmt.Errorf("cannot update slack message with error: %w", err)
					}

					s.logger.ErrorCtx(ctx, "error sending slack message", log.Error(sendErr), log.String("message_id", message.ID.String()))
					return nil
				}

				message.SentAt = &now

				if err := message.Update(ctx, tx); err != nil {
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
			s.logger.ErrorCtx(ctx, "failed to join Slack channel", log.Error(err))
		}
	}

	slackResp, err := client.CreateMessage(ctx, slackConn.AccessToken, slackConn.Settings.ChannelID, message.Body)
	if err != nil {
		s.logger.ErrorCtx(ctx, "failed to post message to Slack", log.Error(err))
		return nil, nil, fmt.Errorf("failed to post message to Slack: %w", err)
	}

	return &slackResp.Channel, &slackResp.TS, nil
}

func (s *Sender) batchUpdateMessages(ctx context.Context) error {
	for {
		err := s.pg.WithTx(
			ctx,
			func(tx pg.Conn) (err error) {
				update := &coredata.SlackMessageUpdate{}

				defer func() {
					if r := recover(); r != nil {
						panicErr := fmt.Sprintf("panic recovered: %v", r)
						update.Error = &panicErr
						update.UpdatedAt = time.Now()

						if updateErr := update.Update(ctx, tx); updateErr != nil {
							s.logger.ErrorCtx(ctx, "cannot update slack message update after panic", log.Error(updateErr))
						}

						s.logger.ErrorCtx(ctx, "panic while updating slack message", log.String("error", panicErr), log.String("update_id", update.ID.String()))
						err = fmt.Errorf("panic recovered: %v", r)
					}
				}()

				err = update.LoadNextUnsentForUpdate(ctx, tx)
				if err != nil {
					return err
				}

				message := &coredata.SlackMessage{}
				if err := message.LoadById(ctx, tx, coredata.NewScope(update.SlackMessageID.TenantID()), update.SlackMessageID); err != nil {
					return fmt.Errorf("cannot load slack message: %w", err)
				}

				updateErr := s.updateMessage(ctx, tx, message, update)

				now := time.Now()
				update.UpdatedAt = now

				if updateErr != nil {
					errorMsg := updateErr.Error()
					update.Error = &errorMsg
					update.UpdatedAt = time.Now()

					if err := update.Update(ctx, tx); err != nil {
						return fmt.Errorf("cannot update slack message update with error: %w", err)
					}

					s.logger.ErrorCtx(ctx, "error updating slack message", log.Error(updateErr), log.String("update_id", update.ID.String()))
					return nil
				}

				update.SentAt = &now

				if err := update.Update(ctx, tx); err != nil {
					return fmt.Errorf("cannot update slack message update: %w", err)
				}

				return nil
			},
		)

		if errors.Is(err, coredata.ErrNoUnsentSlackMessageUpdate{}) {
			return nil
		}

		if err != nil {
			return err
		}
	}
}

func (s *Sender) updateMessage(ctx context.Context, tx pg.Conn, message *coredata.SlackMessage, update *coredata.SlackMessageUpdate) error {
	if message.ChannelID == nil || message.MessageTS == nil {
		return fmt.Errorf("slack message has no channel ID or message TS")
	}

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

	if err := client.UpdateMessage(ctx, slackConn.AccessToken, *message.ChannelID, *message.MessageTS, update.Body); err != nil {
		s.logger.ErrorCtx(ctx, "failed to update message on Slack", log.Error(err))
		return fmt.Errorf("failed to update message on Slack: %w", err)
	}

	return nil
}
