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
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/asana"
)

func (s *Service) ensureAsanaWebhook(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
	projectGID string,
) {
	if err := s.registerAsanaWebhook(ctx, scope, connectorID, projectGID); err != nil && s.logger != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot register Asana webhook",
			log.String("project_gid", projectGID),
			log.Error(err),
		)
	}
}

func (s *Service) registerAsanaWebhook(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
	projectGID string,
) error {
	account, err := s.asanaClientByConnectorID(ctx, scope, connectorID)
	if err != nil {
		return err
	}

	if account.settings.Webhooks != nil {
		if existing, ok := account.settings.Webhooks[projectGID]; ok && existing.GID != "" {
			return nil
		}
	}

	nonce, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("cannot create Asana webhook nonce: %w", err)
	}

	pending := &coredata.AsanaWebhookHandshake{
		Nonce:            nonce.String(),
		ConnectorID:      connectorID,
		SecretCiphertext: []byte{},
		CreatedAt:        time.Now(),
	}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := pending.Insert(ctx, conn); err != nil {
				return fmt.Errorf("cannot reserve Asana webhook nonce: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	target, err := s.asanaWebhookTarget(connectorID, nonce.String())
	if err != nil {
		s.deleteAsanaHandshake(ctx, nonce.String())

		return err
	}

	webhookGID, err := account.client.CreateWebhook(ctx, projectGID, target)
	if err != nil {
		s.deleteAsanaHandshake(ctx, nonce.String())

		return fmt.Errorf("cannot create Asana webhook: %w", err)
	}

	handshake := &coredata.AsanaWebhookHandshake{}

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := handshake.Load(ctx, conn, nonce.String()); err != nil {
				return fmt.Errorf("cannot load Asana webhook handshake: %w", err)
			}

			return nil
		},
	)
	if err != nil || handshake.ConnectorID != connectorID || len(handshake.SecretCiphertext) == 0 {
		_ = account.client.DeleteWebhook(ctx, webhookGID)
		s.deleteAsanaHandshake(ctx, nonce.String())

		if err != nil {
			return err
		}

		return fmt.Errorf("cannot load Asana webhook handshake: secret was not received")
	}

	secret := base64.StdEncoding.EncodeToString(handshake.SecretCiphertext)

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			current, err := s.asanaClientByConnectorID(ctx, scope, connectorID)
			if err != nil {
				return err
			}

			if current.settings.Webhooks == nil {
				current.settings.Webhooks = map[string]coredata.AsanaWebhookRegistration{}
			}

			current.settings.Webhooks[projectGID] = coredata.AsanaWebhookRegistration{
				GID:    webhookGID,
				Nonce:  nonce.String(),
				Secret: secret,
			}
			if err := current.connector.SetSettings(&current.settings); err != nil {
				return err
			}

			current.connector.UpdatedAt = time.Now()
			if err := current.connector.UpdateSettings(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot store Asana webhook: %w", err)
			}

			handshake.Nonce = nonce.String()
			if err := handshake.Delete(ctx, tx); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		_ = account.client.DeleteWebhook(ctx, webhookGID)
		s.deleteAsanaHandshake(ctx, nonce.String())

		return err
	}

	return nil
}

func (s *Service) asanaWebhookTarget(connectorID gid.GID, nonce string) (string, error) {
	target, err := url.JoinPath(
		s.baseURL,
		"api",
		"asana",
		"v1",
		"webhooks",
		url.PathEscape(connectorID.String()),
		url.PathEscape(nonce),
	)
	if err != nil {
		return "", fmt.Errorf("cannot build Asana webhook url: %w", err)
	}

	return target, nil
}

func (s *Service) StoreAsanaHookSecret(
	ctx context.Context,
	connectorID gid.GID,
	nonce string,
	secret string,
) error {
	sealed, err := cipher.Encrypt([]byte(secret), s.encryptionKey)
	if err != nil {
		return fmt.Errorf("cannot encrypt Asana webhook secret: %w", err)
	}

	handshake := &coredata.AsanaWebhookHandshake{
		Nonce:            nonce,
		ConnectorID:      connectorID,
		SecretCiphertext: sealed,
		CreatedAt:        time.Now(),
	}

	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := handshake.StoreSecret(ctx, conn); err != nil {
				return fmt.Errorf("cannot store Asana webhook secret: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) deleteAsanaHandshake(ctx context.Context, nonce string) {
	handshake := &coredata.AsanaWebhookHandshake{Nonce: nonce}
	_ = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return handshake.Delete(ctx, conn)
		},
	)
}

func (s *Service) AsanaHookSecret(
	ctx context.Context,
	connectorID gid.GID,
	nonce string,
) (string, error) {
	var secret string

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			handshake := &coredata.AsanaWebhookHandshake{}

			err := handshake.Load(ctx, conn, nonce)
			if err == nil && handshake.ConnectorID == connectorID && len(handshake.SecretCiphertext) > 0 {
				plain, err := cipher.Decrypt(handshake.SecretCiphertext, s.encryptionKey)
				if err != nil {
					return fmt.Errorf("cannot decrypt Asana webhook secret: %w", err)
				}

				secret = string(plain)

				return nil
			}

			scope := coredata.NewScopeFromObjectID(connectorID)

			connector := &coredata.Connector{}
			if err := connector.LoadMetadataByID(ctx, conn, scope, connectorID); err != nil {
				return fmt.Errorf("cannot load Asana connector: %w", err)
			}

			settings, err := coredata.ConnectorSettings[coredata.AsanaConnectorSettings](connector)
			if err != nil {
				return err
			}

			for _, webhook := range settings.Webhooks {
				if webhook.Nonce != nonce || webhook.Secret == "" {
					continue
				}

				plain, err := s.openAsanaSecret(webhook.Secret)
				if err != nil {
					return err
				}

				secret = plain

				return nil
			}

			return fmt.Errorf("cannot find Asana webhook secret")
		},
	)
	if err != nil {
		return "", err
	}

	if secret == "" {
		return "", fmt.Errorf("cannot find Asana webhook secret")
	}

	return secret, nil
}

func (s *Service) EnqueueAsanaWebhook(
	ctx context.Context,
	connectorID gid.GID,
	body []byte,
) (bool, error) {
	batch, err := asana.ParseEvents(body)
	if err != nil {
		return false, err
	}

	if len(batch.Events) == 0 {
		return false, nil
	}

	envelope, err := json.Marshal(asanaInbox{
		ConnectorID: connectorID.String(),
		Body:        body,
	})
	if err != nil {
		return false, fmt.Errorf("cannot marshal Asana webhook envelope: %w", err)
	}

	sum := sha256.Sum256(envelope)
	event := coredata.NewAsanaWebhookEvent(hex.EncodeToString(sum[:]), envelope)

	var inserted bool

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var insertErr error

			inserted, insertErr = event.Insert(ctx, conn)
			if insertErr != nil {
				return fmt.Errorf("cannot insert Asana webhook event: %w", insertErr)
			}

			return nil
		},
	)
	if err != nil {
		return false, err
	}

	return inserted, nil
}
