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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestConnectorAccount_UpsertAndLoad(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderGitHub, key)
	require.NoError(t, err)

	account, err := insertConnectorAccount(ctx, client, scope, organizationID, connectorID, "acme", "Acme")
	require.NoError(t, err)
	require.NotEmpty(t, account.ID)

	loaded := &coredata.ConnectorAccount{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, scope, account.ID)
	}))
	assert.Equal(t, account.ID, loaded.ID)
	assert.Equal(t, connectorID, loaded.ConnectorID)
	assert.Equal(t, "acme", loaded.ExternalAccountID)
	assert.Equal(t, "Acme", loaded.Name)

	updated, err := insertConnectorAccount(ctx, client, scope, organizationID, connectorID, "acme", "Acme Inc")
	require.NoError(t, err)
	assert.Equal(t, account.ID, updated.ID)
	assert.Equal(t, "Acme Inc", updated.Name)

	other, err := insertConnectorAccount(ctx, client, scope, organizationID, connectorID, "other", "Other")
	require.NoError(t, err)
	assert.NotEqual(t, account.ID, other.ID)

	var accounts coredata.ConnectorAccounts

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.ConnectorAccountOrderField]{
			Field:     coredata.ConnectorAccountOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return accounts.LoadByConnectorID(ctx, conn, scope, connectorID, cursor)
	}))
	assert.Len(t, accounts, 2)
}

func TestConnectorAccount_LoadStandaloneByConnectorID(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderGitHub, key)
	require.NoError(t, err)

	standalone := &coredata.ConnectorAccount{}
	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return standalone.LoadStandaloneByConnectorID(ctx, conn, scope, connectorID)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	first, err := insertConnectorAccount(ctx, client, scope, organizationID, connectorID, "one", "One")
	require.NoError(t, err)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return standalone.LoadStandaloneByConnectorID(ctx, conn, scope, connectorID)
	}))
	assert.Equal(t, first.ID, standalone.ID)

	_, err = insertConnectorAccount(ctx, client, scope, organizationID, connectorID, "two", "Two")
	require.NoError(t, err)

	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return standalone.LoadStandaloneByConnectorID(ctx, conn, scope, connectorID)
	})
	require.ErrorIs(t, err, coredata.ErrMultipleConnectorAccounts)
}

func TestConnectorAccount_SyncStandaloneAccount(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	now := time.Now().UTC()
	cnnctr := &coredata.Connector{
		ID:             gid.New(scope.GetTenantID(), coredata.ConnectorEntityType),
		OrganizationID: organizationID,
		Provider:       coredata.ConnectorProviderGitHub,
		Protocol:       coredata.ConnectorProtocolOAuth2,
		Connection: &connector.OAuth2Connection{
			AccessToken: "test-token",
			TokenType:   "Bearer",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, cnnctr.SetSettings(coredata.GitHubConnectorSettings{Organization: "acme"}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return cnnctr.Insert(ctx, tx, scope, key)
	}))

	var synced *coredata.ConnectorAccount

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		synced, err = coredata.SyncStandaloneAccount(ctx, tx, scope, cnnctr, "acme", "acme")

		return err
	}))
	require.NotNil(t, synced)
	assert.Equal(t, "acme", synced.ExternalAccountID)

	require.NoError(t, cnnctr.SetSettings(coredata.GitHubConnectorSettings{Organization: "acme-inc"}))

	var updated *coredata.ConnectorAccount

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		updated, err = coredata.SyncStandaloneAccount(ctx, tx, scope, cnnctr, "acme-inc", "acme-inc")

		return err
	}))
	require.NotNil(t, updated)
	assert.Equal(t, synced.ID, updated.ID)
	assert.Equal(t, "acme-inc", updated.ExternalAccountID)

	var accounts coredata.ConnectorAccounts

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.ConnectorAccountOrderField]{
			Field:     coredata.ConnectorAccountOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return accounts.LoadByConnectorID(ctx, conn, scope, cnnctr.ID, cursor)
	}))
	assert.Len(t, accounts, 1)
}
