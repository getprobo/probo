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

package connectorclient_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connectorclient"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
)

const hubAccountID = "111111111111"

type recordingSession struct {
	accountID string
}

func (s recordingSession) Cloud() string { return cloud.AWS }

func (s recordingSession) AccountID() string { return s.accountID }

func TestOpen_MemberAccountPassesExternalID(t *testing.T) {
	t.Parallel()

	const memberID = "222222222222"

	var seen string

	pgClient, seeded := seedAccount(
		t,
		coredata.ConnectorProviderAWS,
		coredata.ConnectorProtocolWorkloadIdentity,
		&connectorprotocol.WorkloadIdentityConnection{},
		memberID,
	)
	opener := connectorclient.New(
		pgClient,
		seeded.key,
		nil,
		workloadRegistry(t, &seen),
		connectorprotocol.NewConnectorRegistry(),
	)

	handle := openAccount(t, pgClient, opener, seeded.scope, seeded.account.ID)

	assert.Equal(t, memberID, seen)
	assert.Equal(t, coredata.ConnectorProviderAWS, handle.Provider)
	require.NotNil(t, handle.Session)
	assert.Nil(t, handle.HTTPClient)
	assert.Equal(t, memberID, handle.Session.AccountID())
}

func TestOpen_InitialAccountReachesHub(t *testing.T) {
	t.Parallel()

	var seen string

	pgClient, seeded := seedAccount(
		t,
		coredata.ConnectorProviderAWS,
		coredata.ConnectorProtocolWorkloadIdentity,
		&connectorprotocol.WorkloadIdentityConnection{},
		hubAccountID,
	)
	opener := connectorclient.New(
		pgClient,
		seeded.key,
		nil,
		workloadRegistry(t, &seen),
		connectorprotocol.NewConnectorRegistry(),
	)

	handle := openAccount(t, pgClient, opener, seeded.scope, seeded.account.ID)

	assert.Empty(t, seen)
	require.NotNil(t, handle.Session)
	assert.Empty(t, handle.Session.AccountID())
	assert.Nil(t, handle.HTTPClient)
}

func TestOpen_PersistsRotatedOAuthToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"refreshed","token_type":"Bearer","expires_in":3600}`))
	}))
	t.Cleanup(server.Close)

	pgClient, seeded := seedAccount(
		t,
		coredata.ConnectorProviderLinear,
		coredata.ConnectorProtocolOAuth2,
		&connectorprotocol.OAuth2Connection{
			AccessToken:  "stale",
			RefreshToken: "refresh-token",
			ExpiresAt:    time.Now().Add(-time.Hour),
			TokenType:    "Bearer",
		},
		"linear-workspace",
	)
	opener := oauthOpener(t, pgClient, seeded.key, server.URL)

	handle := openAccount(t, pgClient, opener, seeded.scope, seeded.account.ID)

	require.NotNil(t, handle.HTTPClient)
	assert.Nil(t, handle.Session)

	reloaded := reloadConnector(t, pgClient, seeded.scope, seeded.connector.ID, seeded.key)
	conn, ok := reloaded.Connection.(*connectorprotocol.OAuth2Connection)
	require.True(t, ok)
	assert.Equal(t, "refreshed", conn.AccessToken)
}

func TestOpen_PersistsFullCredentialAfterRollback(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"stale","refresh_token":"rotated-refresh","token_type":"DPoP","expires_in":3600}`))
	}))
	t.Cleanup(server.Close)

	pgClient, seeded := seedAccount(
		t,
		coredata.ConnectorProviderLinear,
		coredata.ConnectorProtocolOAuth2,
		&connectorprotocol.OAuth2Connection{
			AccessToken:  "stale",
			RefreshToken: "refresh-token",
			ExpiresAt:    time.Now().Add(-time.Hour),
			TokenType:    "Bearer",
		},
		"linear-workspace",
	)
	opener := oauthOpener(t, pgClient, seeded.key, server.URL)

	err := pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		_, openErr := opener.Open(ctx, tx, seeded.scope, seeded.account.ID)
		if openErr != nil {
			return openErr
		}

		return errors.New("rollback")
	})
	require.EqualError(t, err, "rollback")

	reloaded := reloadConnector(t, pgClient, seeded.scope, seeded.connector.ID, seeded.key)
	conn, ok := reloaded.Connection.(*connectorprotocol.OAuth2Connection)
	require.True(t, ok)
	assert.Equal(t, "stale", conn.AccessToken)
	assert.Equal(t, "rotated-refresh", conn.RefreshToken)
	assert.Equal(t, "DPoP", conn.TokenType)
	assert.True(t, conn.ExpiresAt.After(time.Now()))
}

func TestOpen_LeavesUnchangedOAuthToken(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called.Store(true)
	}))
	t.Cleanup(server.Close)

	pgClient, seeded := seedAccount(
		t,
		coredata.ConnectorProviderLinear,
		coredata.ConnectorProtocolOAuth2,
		&connectorprotocol.OAuth2Connection{
			AccessToken:  "kept",
			RefreshToken: "refresh-token",
			ExpiresAt:    time.Now().Add(time.Hour),
			TokenType:    "Bearer",
		},
		"linear-workspace",
	)
	stored := append([]byte(nil), seeded.connector.EncryptedConnection...)
	opener := oauthOpener(t, pgClient, seeded.key, server.URL)

	handle := openAccount(t, pgClient, opener, seeded.scope, seeded.account.ID)

	require.NotNil(t, handle.HTTPClient)
	assert.Nil(t, handle.Session)
	assert.False(t, called.Load())

	reloaded := &coredata.Connector{}

	require.NoError(t, pgClient.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadMetadataByID(ctx, conn, seeded.scope, seeded.connector.ID)
	}))
	assert.True(t, bytes.Equal(stored, reloaded.EncryptedConnection))
}

func TestOpen_NamesProviderWithoutASession(t *testing.T) {
	t.Parallel()

	registries := provider.NewRegistry()
	require.NoError(t, registries.Register(&provider.Registration{
		Provider:    coredata.ConnectorProviderAWS,
		DisplayName: "Amazon Web Services",
	}))

	pgClient, seeded := seedAccount(
		t,
		coredata.ConnectorProviderAWS,
		coredata.ConnectorProtocolWorkloadIdentity,
		&connectorprotocol.WorkloadIdentityConnection{},
		hubAccountID,
	)
	opener := connectorclient.New(
		pgClient,
		seeded.key,
		nil,
		registries,
		connectorprotocol.NewConnectorRegistry(),
	)

	var handle *connectorclient.Handle

	err := pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		var openErr error

		handle, openErr = opener.Open(ctx, tx, seeded.scope, seeded.account.ID)

		return openErr
	})

	require.Error(t, err)
	assert.ErrorContains(t, err, string(coredata.ConnectorProviderAWS))
	assert.Nil(t, handle)
}

func workloadRegistry(t *testing.T, seen *string) *provider.Registry {
	t.Helper()

	registries := provider.NewRegistry()
	err := registries.Register(&provider.Registration{
		Provider:    coredata.ConnectorProviderAWS,
		DisplayName: "Amazon Web Services",
		InitialAccountFunc: func(*coredata.Connector) (string, string, error) {
			return hubAccountID, hubAccountID, nil
		},
		WorkloadIdentity: &provider.WorkloadIdentityConfig{
			NewSession: func(
				_ context.Context,
				_ *identityfederation.Issuer,
				_ *coredata.Connector,
				accountID string,
			) (cloud.Session, error) {
				*seen = accountID

				return recordingSession{accountID: accountID}, nil
			},
			NewDriver: func(
				context.Context,
				cloud.Session,
				*coredata.Connector,
				*log.Logger,
			) (drivers.Driver, error) {
				return nil, nil
			},
		},
	})
	require.NoError(t, err)

	return registries
}

func oauthOpener(
	t *testing.T,
	pgClient *pg.Client,
	key cipher.EncryptionKey,
	tokenURL string,
) *connectorclient.Client {
	t.Helper()

	connectors := connectorprotocol.NewConnectorRegistry()
	require.NoError(t, connectors.Register(
		string(coredata.ConnectorProviderLinear),
		&connectorprotocol.OAuth2Connector{
			ClientID:     "client",
			ClientSecret: "secret",
			TokenURL:     tokenURL,
			AuthURL:      "https://example.com/auth",
			RedirectURI:  "https://example.com/callback",
		},
	))

	opener := connectorclient.New(
		pgClient,
		key,
		nil,
		provider.NewRegistry(),
		connectors,
	)
	connectorclient.SetHTTPOptions(opener, httpclient.WithSSRFAllowLoopback())

	return opener
}

type seededAccount struct {
	scope     coredata.Scoper
	connector *coredata.Connector
	account   *coredata.ConnectorAccount
	key       cipher.EncryptionKey
}

func seedAccount(
	t *testing.T,
	connectorProvider coredata.ConnectorProvider,
	protocol coredata.ConnectorProtocol,
	connection connectorprotocol.Connection,
	externalID string,
) (*pg.Client, seededAccount) {
	t.Helper()

	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	var key cipher.EncryptionKey

	cnnctr := &coredata.Connector{
		ID:             gid.New(tenantID, coredata.ConnectorEntityType),
		OrganizationID: organizationID,
		Provider:       connectorProvider,
		Protocol:       protocol,
		Name:           "Test",
		Connection:     connection,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	account := &coredata.ConnectorAccount{
		ID:                gid.New(tenantID, coredata.ConnectorAccountEntityType),
		OrganizationID:    organizationID,
		ConnectorID:       cnnctr.ID,
		ExternalAccountID: externalID,
		Name:              externalID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(t, pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Connector client test",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		if err := cnnctr.Insert(ctx, tx, scope, key); err != nil {
			return err
		}

		_, err := account.Upsert(ctx, tx, scope)

		return err
	}))

	return pgClient, seededAccount{
		scope:     scope,
		connector: cnnctr,
		account:   account,
		key:       key,
	}
}

func openAccount(
	t *testing.T,
	pgClient *pg.Client,
	opener *connectorclient.Client,
	scope coredata.Scoper,
	accountID gid.GID,
) *connectorclient.Handle {
	t.Helper()

	var handle *connectorclient.Handle

	require.NoError(t, pgClient.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		var err error

		handle, err = opener.Open(ctx, tx, scope, accountID)

		return err
	}))

	return handle
}

func reloadConnector(
	t *testing.T,
	pgClient *pg.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
	key cipher.EncryptionKey,
) *coredata.Connector {
	t.Helper()

	loaded := &coredata.Connector{}

	require.NoError(t, pgClient.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, scope, connectorID, key)
	}))

	return loaded
}
