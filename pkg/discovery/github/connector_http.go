// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package github

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

// connectorHTTPClient loads a connector and returns an authenticated HTTP client.
func connectorHTTPClient(
	ctx context.Context,
	pgClient *pg.Client,
	scope coredata.Scoper,
	encryptionKey cipher.EncryptionKey,
	connectorRegistry *connector.ConnectorRegistry,
	providerRegistry *provider.Registry,
	connectorID gid.GID,
) (*http.Client, *coredata.Connector, error) {
	var dbConnector coredata.Connector

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := dbConnector.LoadByID(ctx, conn, scope, connectorID, encryptionKey); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	if dbConnector.Provider != coredata.ConnectorProviderGitHub {
		return nil, nil, fmt.Errorf("connector provider is %s, expected GITHUB", dbConnector.Provider)
	}

	var tokenBefore string

	oauth2Conn, isOAuth2 := dbConnector.Connection.(*connector.OAuth2Connection)
	if isOAuth2 {
		tokenBefore = oauth2Conn.AccessToken
	}

	var httpClient *http.Client

	if isOAuth2 && connectorRegistry != nil {
		refreshCfg := connectorRegistry.GetOAuth2RefreshConfig(string(dbConnector.Provider))
		if refreshCfg != nil {
			httpClient, err = oauth2Conn.RefreshableClient(ctx, *refreshCfg)
			if err != nil {
				return nil, nil, fmt.Errorf("cannot create refreshable HTTP client: %w", err)
			}
		}
	}

	if httpClient == nil {
		if err := providerRegistry.ApplyManagedAPIKey(&dbConnector); err != nil {
			return nil, nil, fmt.Errorf("cannot apply managed API key: %w", err)
		}

		httpClient, err = dbConnector.Connection.Client(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot create HTTP client: %w", err)
		}
	}

	if isOAuth2 && oauth2Conn.AccessToken != tokenBefore {
		dbConnector.UpdatedAt = time.Now()

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return dbConnector.Update(ctx, tx, scope, encryptionKey)
			},
		); err != nil {
			return nil, nil, fmt.Errorf("cannot persist refreshed connector token: %w", err)
		}
	}

	return httpClient, &dbConnector, nil
}
