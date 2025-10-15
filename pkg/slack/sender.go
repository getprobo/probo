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
	"net/http"
	"time"

	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
)

type (
	Sender struct {
		pg            *pg.Client
		logger        *log.Logger
		encryptionKey cipher.EncryptionKey
		interval      time.Duration
		httpClient    *http.Client
	}

	Config struct {
		Interval time.Duration
	}
)

func NewSender(pg *pg.Client, logger *log.Logger, encryptionKey cipher.EncryptionKey, cfg Config) *Sender {
	httpClientOpts := []httpclient.Option{
		httpclient.WithLogger(logger),
	}

	httpClient := httpclient.DefaultPooledClient(httpClientOpts...)

	return &Sender{
		pg:            pg,
		logger:        logger,
		encryptionKey: encryptionKey,
		interval:      cfg.Interval,
		httpClient:    httpClient,
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

				if sendErr := s.sendMessage(ctx, tx, message); sendErr != nil {
					errorMsg := sendErr.Error()
					message.Error = &errorMsg
					message.UpdatedAt = time.Now()

					if err := message.Update(ctx, tx); err != nil {
						return fmt.Errorf("cannot update slack message with error: %w", err)
					}

					s.logger.ErrorCtx(ctx, "error sending slack message", log.Error(sendErr), log.String("message_id", message.ID.String()))
					return nil
				}

				now := time.Now()
				message.SentAt = &now
				message.UpdatedAt = now

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

func (s *Sender) sendMessage(ctx context.Context, tx pg.Conn, message *coredata.SlackMessage) error {
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

	for _, c := range connectors {
		slackConn, ok := c.Connection.(*connector.SlackConnection)
		if !ok {
			return fmt.Errorf("slack connector must have SlackConnection type")
		}

		if slackConn.Settings.WebhookURL == "" {
			return fmt.Errorf("slack connector %s has no webhook URL", c.ID)
		}

		client := NewClient(slackConn.Settings.WebhookURL, s.httpClient)
		if err := client.PostMessage(ctx, message.Body); err != nil {
			return fmt.Errorf("failed to post message to Slack: %w", err)
		}
	}

	return nil
}
