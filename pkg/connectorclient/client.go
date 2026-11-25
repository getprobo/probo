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

// Package connectorclient opens one connector account as a cloud session or
// an HTTP client.
package connectorclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

type (
	// Client opens a connector account. It holds the credentials and
	// registries Open needs, and nothing about a particular account.
	Client struct {
		pg            *pg.Client
		encryptionKey cipher.EncryptionKey
		federation    *identityfederation.Issuer
		providers     *provider.Registry
		connectors    *connectorprotocol.Registry
		httpOptions   []httpclient.Option
	}

	// Handle is one opened connector account. Session and HTTPClient are
	// mutually exclusive: a workload-identity account has a session, and an
	// HTTP credential has a client.
	Handle struct {
		Provider   coredata.ConnectorProvider
		Connector  *coredata.Connector
		Account    *coredata.ConnectorAccount
		Session    cloud.Session
		HTTPClient *http.Client
	}
)

func New(
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	federation *identityfederation.Issuer,
	providers *provider.Registry,
	connectors *connectorprotocol.Registry,
) *Client {
	return &Client{
		pg:            pgClient,
		encryptionKey: encryptionKey,
		federation:    federation,
		providers:     providers,
		connectors:    connectors,
	}
}

// Open loads the connector account and its connector, then returns a session
// or an HTTP client. A refreshed OAuth credential is committed in its own
// transaction before Open returns.
func (c *Client) Open(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	accountID gid.GID,
) (*Handle, error) {
	account := &coredata.ConnectorAccount{}
	if err := account.LoadByID(ctx, tx, scope, accountID); err != nil {
		return nil, fmt.Errorf("cannot load connector account: %w", err)
	}

	dbConnector := &coredata.Connector{}
	if err := dbConnector.LoadByID(ctx, tx, scope, account.ConnectorID, c.encryptionKey); err != nil {
		return nil, fmt.Errorf("cannot load connector %s: %w", account.ConnectorID, err)
	}

	switch dbConnector.Connection.(type) {
	case *connectorprotocol.WorkloadIdentityConnection:
		return c.openSession(ctx, dbConnector, account)

	case connectorprotocol.HTTPConnection:
		return c.openHTTP(ctx, scope, dbConnector, account)

	default:
		return nil, fmt.Errorf(
			"cannot open %s connector account: unsupported credential",
			dbConnector.Provider,
		)
	}
}

func (c *Client) openSession(
	ctx context.Context,
	dbConnector *coredata.Connector,
	account *coredata.ConnectorAccount,
) (*Handle, error) {
	session, err := provider.OpenSession(
		ctx,
		c.federation,
		c.providers,
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

func (c *Client) openHTTP(
	ctx context.Context,
	scope coredata.Scoper,
	dbConnector *coredata.Connector,
	account *coredata.ConnectorAccount,
) (*Handle, error) {
	oauth2Conn := connectorprotocol.OAuth2ConnectionOf(dbConnector.Connection)
	before := oauthCredentialFrom(oauth2Conn)

	httpClient, err := provider.HTTPClient(
		ctx,
		c.connectors,
		c.providers,
		dbConnector,
		c.httpOptions...,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open HTTP client for %s connector: %w", dbConnector.Provider, err)
	}

	if httpClient == nil {
		return nil, fmt.Errorf("cannot open %s connector account: empty HTTP client", dbConnector.Provider)
	}

	// A rotating provider has already consumed the old refresh token, so
	// this write commits on its own and survives the caller's rollback.
	if before.changed(oauth2Conn) {
		dbConnector.UpdatedAt = time.Now()

		if err := c.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return dbConnector.UpdateConnection(ctx, tx, scope, c.encryptionKey)
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
