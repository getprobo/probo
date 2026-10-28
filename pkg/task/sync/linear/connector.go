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

package linear

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
	linear "go.probo.inc/probo/pkg/task/sync/linear/client"
)

type LinearTeam struct {
	ID   string
	Name string
	Key  string
}

func (s *Service) linearClientForConnector(
	ctx context.Context,
	scope coredata.Scoper,
	dbConnector *coredata.Connector,
) (*linear.Client, *coredata.Connector, error) {
	oauthConn, ok := dbConnector.Connection.(*connectorprotocol.OAuth2Connection)
	if !ok {
		return nil, nil, fmt.Errorf("cannot use Linear connector: unsupported credential")
	}

	if missing := missingTaskSyncScopes(oauthConn.Scopes()); len(missing) > 0 {
		return nil, nil, NewLinearReconnectRequiredError(missing)
	}

	if err := s.connectorRegistry.ConfigureConnection(string(dbConnector.Provider), oauthConn); err != nil {
		return nil, nil, fmt.Errorf("cannot configure Linear connector: %w", err)
	}

	tokenBefore := oauthConn.AccessToken

	httpClient, err := s.oauthHTTP(ctx, oauthConn, dbConnector.Provider)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create Linear HTTP client: %w", err)
	}

	if oauthConn.AccessToken != tokenBefore {
		dbConnector.UpdatedAt = time.Now()

		err := s.pg.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return dbConnector.UpdateConnection(ctx, tx, scope, s.encryptionKey)
			},
		)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot persist refreshed Linear token: %w", err)
		}
	}

	api := linear.NewClient(httpClient, s.linearAPIBaseURL)
	api.SetConnectionGeneration(linearAssetGeneration(linearConnectionGenerationKey(oauthConn)))

	return api, dbConnector, nil
}

// linearConnectionGenerationKey identifies the OAuth grant that uploaded
// cached Linear assets. Refresh rotates the access token and keeps this key.
func linearConnectionGenerationKey(conn *connectorprotocol.OAuth2Connection) string {
	if conn == nil {
		return ""
	}

	if token := strings.TrimSpace(conn.RefreshToken); token != "" {
		return token
	}

	if conn.TokenURL != "" || conn.ClientID != "" {
		return conn.TokenURL + "\x00" + conn.ClientID
	}

	return ""
}
