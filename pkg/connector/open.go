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

package connector

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// Handle is one opened connector account. Session and HTTPClient are
// mutually exclusive: a workload-identity account has a session, and an
// HTTP credential has a client.
type Handle struct {
	Provider   coredata.ConnectorProvider
	Connector  *coredata.Connector
	Account    *coredata.ConnectorAccount
	Session    cloud.Session
	HTTPClient *http.Client
}

// Open loads the connector account and its connector, then returns a session
// or an HTTP client. The load releases its connection before the provider
// call. A refreshed OAuth credential is committed on its own transaction
// after that, so the caller does not hold a connection while the commit
// waits for one, and a rollback of the caller's transaction leaves it.
func (s *Service) Open(
	ctx context.Context,
	scope coredata.Scoper,
	accountID gid.GID,
) (*Handle, error) {
	account := &coredata.ConnectorAccount{}
	dbConnector := &coredata.Connector{}

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := account.LoadByID(ctx, conn, scope, accountID); err != nil {
			return fmt.Errorf("cannot load connector account: %w", err)
		}

		if err := dbConnector.LoadByID(ctx, conn, scope, account.ConnectorID, s.encryptionKey); err != nil {
			return fmt.Errorf("cannot load connector %s: %w", account.ConnectorID, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	switch dbConnector.Connection.(type) {
	case *connectorprotocol.WorkloadIdentityConnection:
		return s.openSession(ctx, dbConnector, account)

	case connectorprotocol.HTTPConnection:
		return s.openHTTP(ctx, scope, dbConnector, account)

	default:
		return nil, fmt.Errorf(
			"cannot open %s connector account: unsupported credential",
			dbConnector.Provider,
		)
	}
}

func (s *Service) openSession(
	ctx context.Context,
	dbConnector *coredata.Connector,
	account *coredata.ConnectorAccount,
) (*Handle, error) {
	session, err := provider.OpenSession(
		ctx,
		s.federation,
		s.providerRegistry,
		dbConnector,
		account.ExternalAccountID,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open connector account: %w", err)
	}

	if session == nil {
		return nil, fmt.Errorf("cannot open %s connector account: empty session", dbConnector.Provider)
	}

	return &Handle{
		Provider:  dbConnector.Provider,
		Connector: dbConnector,
		Account:   account,
		Session:   session,
	}, nil
}

func (s *Service) openHTTP(
	ctx context.Context,
	scope coredata.Scoper,
	dbConnector *coredata.Connector,
	account *coredata.ConnectorAccount,
) (*Handle, error) {
	oauth2Conn := connectorprotocol.OAuth2ConnectionOf(dbConnector.Connection)
	before := oauthCredentialFrom(oauth2Conn)

	httpClient, err := provider.HTTPClient(
		ctx,
		s.connectorRegistry,
		s.providerRegistry,
		dbConnector,
		s.httpOptions...,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open HTTP client for %s connector: %w", dbConnector.Provider, err)
	}

	if httpClient == nil {
		return nil, fmt.Errorf("cannot open %s connector account: empty HTTP client", dbConnector.Provider)
	}

	// The provider has already consumed the old refresh token. The load
	// connection is already released, so this commit does not wait for a
	// pool connection while another transaction is held.
	if before.changed(oauth2Conn) {
		dbConnector.UpdatedAt = time.Now()

		if err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return dbConnector.UpdateConnection(ctx, tx, scope, s.encryptionKey)
		}); err != nil {
			return nil, fmt.Errorf("cannot persist refreshed token for connector %s: %w", dbConnector.ID, err)
		}
	}

	return &Handle{
		Provider:   dbConnector.Provider,
		Connector:  dbConnector,
		Account:    account,
		HTTPClient: httpClient,
	}, nil
}

type oauthCredential struct {
	accessToken  string
	refreshToken string
	expiresAt    time.Time
	tokenType    string
}

func oauthCredentialFrom(conn *connectorprotocol.OAuth2Connection) oauthCredential {
	if conn == nil {
		return oauthCredential{}
	}

	return oauthCredential{
		accessToken:  conn.AccessToken,
		refreshToken: conn.RefreshToken,
		expiresAt:    conn.ExpiresAt,
		tokenType:    conn.TokenType,
	}
}

func (c oauthCredential) changed(conn *connectorprotocol.OAuth2Connection) bool {
	if conn == nil {
		return false
	}

	return c.accessToken != conn.AccessToken ||
		c.refreshToken != conn.RefreshToken ||
		!c.expiresAt.Equal(conn.ExpiresAt) ||
		c.tokenType != conn.TokenType
}
