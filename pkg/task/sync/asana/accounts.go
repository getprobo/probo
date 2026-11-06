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

package asana

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	client "go.probo.inc/probo/pkg/task/sync/asana/client"
)

type asanaAccount struct {
	client    *client.Client
	connector *coredata.Connector
	settings  coredata.AsanaConnectorSettings
}

func (s *Service) loadAsanaConnectors(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]*coredata.Connector, error) {
	var connectors []*coredata.Connector

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			provider := coredata.ConnectorProviderAsana
			filter := coredata.NewConnectorProviderFilter(&provider)

			loaded, err := page.LoadAll(
				ctx,
				page.OrderBy[coredata.ConnectorOrderField]{
					Field:     coredata.ConnectorOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
				func(ctx context.Context, cursor *page.Cursor[coredata.ConnectorOrderField]) ([]*coredata.Connector, error) {
					var batch coredata.Connectors
					if err := batch.LoadByOrganizationIDWithoutDecryptedConnection(
						ctx,
						conn,
						scope,
						organizationID,
						cursor,
						filter,
					); err != nil {
						return nil, err
					}

					return batch, nil
				},
			)
			if err != nil {
				return err
			}

			connectors = loaded

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load Asana connectors: %w", err)
	}

	return connectors, nil
}

func (s *Service) asanaAccounts(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]asanaAccount, error) {
	connectors, err := s.loadAsanaConnectors(ctx, scope, organizationID)
	if err != nil {
		return nil, err
	}

	if len(connectors) == 0 {
		return nil, ErrAsanaNotConnected
	}

	accounts := make([]asanaAccount, 0, len(connectors))

	var lastErr error

	for _, connector := range connectors {
		if err := connector.DecryptConnection(s.encryptionKey); err != nil {
			s.skipAsanaConnector(ctx, connector.ID, fmt.Errorf("cannot decrypt Asana connector: %w", err), &lastErr)
			continue
		}

		account, err := s.asanaAccountForConnector(ctx, scope, connector)
		if err != nil {
			s.skipAsanaConnector(ctx, connector.ID, err, &lastErr)
			continue
		}

		accounts = append(accounts, account)
	}

	if len(accounts) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("cannot use Asana connectors: %w", lastErr)
		}

		return nil, ErrAsanaNotConnected
	}

	return accounts, nil
}

func (s *Service) skipAsanaConnector(
	ctx context.Context,
	connectorID gid.GID,
	cause error,
	lastErr *error,
) {
	*lastErr = cause

	s.warnSkippedAsanaConnector(ctx, connectorID, cause)
}

func (s *Service) warnSkippedAsanaConnector(ctx context.Context, connectorID gid.GID, cause error) {
	if s.logger == nil {
		return
	}

	s.logger.WarnCtx(
		ctx,
		"skipping Asana connector",
		log.String("connector_id", connectorID.String()),
		log.Error(cause),
	)
}

func (s *Service) asanaAccountForConnector(
	ctx context.Context,
	scope coredata.Scoper,
	connector *coredata.Connector,
) (asanaAccount, error) {
	settings, err := coredata.ConnectorSettings[coredata.AsanaConnectorSettings](connector)
	if err != nil {
		return asanaAccount{}, fmt.Errorf("cannot read Asana settings: %w", err)
	}

	oauthConn, ok := connector.Connection.(*connectorprotocol.OAuth2Connection)
	if !ok {
		return asanaAccount{}, fmt.Errorf("cannot use Asana connector: unsupported credential")
	}

	if err := s.connectorRegistry.ConfigureConnection(string(connector.Provider), oauthConn); err != nil {
		return asanaAccount{}, fmt.Errorf("cannot configure Asana connector: %w", err)
	}

	tokenBefore := oauthConn.AccessToken

	httpClient, err := s.oauthHTTP(ctx, oauthConn, connector.Provider)
	if err != nil {
		return asanaAccount{}, fmt.Errorf("cannot create Asana HTTP client: %w", err)
	}

	if oauthConn.AccessToken != tokenBefore {
		connector.UpdatedAt = time.Now()

		err := s.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := connector.UpdateConnection(ctx, tx, scope, s.encryptionKey); err != nil {
					return fmt.Errorf("cannot persist refreshed Asana token: %w", err)
				}

				return nil
			},
		)
		if err != nil {
			return asanaAccount{}, fmt.Errorf("cannot persist refreshed Asana token: %w", err)
		}
	}

	if s.asanaAPIBaseURL == "" {
		return asanaAccount{}, fmt.Errorf("cannot create Asana client: api base url is empty")
	}

	return asanaAccount{
		client:    client.NewClient(httpClient, s.asanaAPIBaseURL),
		connector: connector,
		settings:  settings,
	}, nil
}

func (s *Service) asanaAccountForProject(
	ctx context.Context,
	accounts []asanaAccount,
	projectGID string,
) (asanaAccount, client.ProjectSetup, error) {
	var lastErr error

	for _, account := range accounts {
		setup, err := account.client.ProjectSetup(ctx, projectGID)
		if err != nil {
			if client.IsNotFound(err) {
				continue
			}

			lastErr = err

			continue
		}

		if setup.WorkspaceGID != "" && setup.WorkspaceGID != account.settings.WorkspaceGID {
			continue
		}

		return account, setup, nil
	}

	if lastErr != nil {
		return asanaAccount{}, client.ProjectSetup{}, fmt.Errorf("cannot load Asana project: %w", lastErr)
	}

	return asanaAccount{}, client.ProjectSetup{}, ErrAsanaProjectNotFound
}

func (s *Service) openAsanaSecret(sealed string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(sealed)
	if err != nil {
		return "", fmt.Errorf("cannot decode Asana webhook secret: %w", err)
	}

	plain, err := cipher.Decrypt(raw, s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("cannot decrypt Asana webhook secret: %w", err)
	}

	return string(plain), nil
}

func (s *Service) asanaClientByConnectorID(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (asanaAccount, error) {
	connector := &coredata.Connector{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := connector.LoadByID(ctx, conn, scope, connectorID, s.encryptionKey); err != nil {
				return err
			}

			return nil
		},
	)
	if err != nil {
		return asanaAccount{}, fmt.Errorf("cannot load Asana connector: %w", err)
	}

	if connector.Provider != coredata.ConnectorProviderAsana {
		return asanaAccount{}, fmt.Errorf("cannot use connector %q for Asana sync", connector.Provider)
	}

	account, err := s.asanaAccountForConnector(ctx, scope, connector)
	if err != nil {
		return asanaAccount{}, fmt.Errorf("cannot open Asana connector: %w", err)
	}

	return account, nil
}
