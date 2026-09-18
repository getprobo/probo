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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

// newConnectorAccount builds an account row on connectorID. A nil
// externalAccountID is the "the credential is the account" shape.
func newConnectorAccount(
	scope coredata.Scoper,
	organizationID gid.GID,
	connectorID gid.GID,
	externalAccountID *string,
	name string,
) *coredata.ConnectorAccount {
	now := time.Now().UTC()

	return &coredata.ConnectorAccount{
		ID:                gid.New(scope.GetTenantID(), coredata.ConnectorAccountEntityType),
		OrganizationID:    organizationID,
		ConnectorID:       connectorID,
		ExternalAccountID: externalAccountID,
		Name:              name,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// upsertConnectorAccount routes to the arbiter that matches the row shape:
// ON CONFLICT never matches a NULL, so the implicit row needs the partial
// index instead of the composite one.
func upsertConnectorAccount(
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	account *coredata.ConnectorAccount,
) error {
	return client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if account.ExternalAccountID == nil {
			return account.UpsertImplicit(ctx, tx, scope)
		}

		return account.Upsert(ctx, tx, scope)
	})
}

func countConnectorAccounts(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	connectorID gid.GID,
) int {
	t.Helper()

	var count int

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		accounts := &coredata.ConnectorAccounts{}

		var err error

		count, err = accounts.CountByConnectorID(ctx, conn, scope, connectorID)

		return err
	}))

	return count
}

// TestConnectorAccount_Upsert covers the two arbiters, because they are the
// only part of this entity the end-to-end suite cannot see: everything else
// about account lifecycle is exercised through the API in
// e2e/console/connector_account_test.go.
func TestConnectorAccount_Upsert(t *testing.T) {
	t.Parallel()

	t.Run("an account is recorded once per vendor identifier", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		scope, organizationID := seedConnectorOrg(t, ctx, client)

		var key cipher.EncryptionKey

		connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderAWS, key)
		require.NoError(t, err)

		externalID := "123456789012"

		first := newConnectorAccount(scope, organizationID, connectorID, &externalID, "Old name")
		require.NoError(t, upsertConnectorAccount(ctx, client, scope, first))

		second := newConnectorAccount(scope, organizationID, connectorID, &externalID, "New name")
		require.NoError(t, upsertConnectorAccount(ctx, client, scope, second))

		assert.Equal(t, 1, countConnectorAccounts(t, ctx, client, scope, connectorID))
		// The row keeps its id, so a source already pointing at it still does.
		assert.Equal(t, first.ID, second.ID)
		assert.Equal(t, "New name", second.Name)
	})

	t.Run("a credential the vendor never names keeps one implicit row", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		scope, organizationID := seedConnectorOrg(t, ctx, client)

		var key cipher.EncryptionKey

		connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderSlack, key)
		require.NoError(t, err)

		// ON CONFLICT never matches a NULL, so this arbitrates on the partial
		// index; without it a second call would insert beside the first.
		for _, name := range []string{"Slack", "Slack renamed"} {
			account := newConnectorAccount(scope, organizationID, connectorID, nil, name)
			require.NoError(t, upsertConnectorAccount(ctx, client, scope, account))
		}

		assert.Equal(t, 1, countConnectorAccounts(t, ctx, client, scope, connectorID))
	})
}
