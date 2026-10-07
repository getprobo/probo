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

func TestAccessReviewSourceInsert_AccountOrCSV(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderGitHub, key)
	require.NoError(t, err)

	account, err := insertConnectorAccount(ctx, client, scope, organizationID, connectorID, "acme", "Acme")
	require.NoError(t, err)

	now := time.Now().UTC()

	t.Run("csv both null inserts", func(t *testing.T) {
		t.Parallel()

		source := &coredata.AccessReviewSource{
			ID:             gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
			OrganizationID: organizationID,
			Name:           "csv",
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			inserted, err := source.Insert(ctx, tx, scope)
			if err != nil {
				return err
			}

			require.True(t, inserted)

			return nil
		}))
	})

	t.Run("account only inserts", func(t *testing.T) {
		t.Parallel()

		source := &coredata.AccessReviewSource{
			ID:                 gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
			OrganizationID:     organizationID,
			ConnectorAccountID: &account.ID,
			Name:               "account only",
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			inserted, err := source.Insert(ctx, tx, scope)
			if err != nil {
				return err
			}

			require.True(t, inserted)

			return nil
		}))

		loaded := &coredata.AccessReviewSource{}

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return loaded.LoadByID(ctx, conn, scope, source.ID)
		}))
		require.NotNil(t, loaded.ConnectorAccountID)
		assert.Equal(t, account.ID, *loaded.ConnectorAccountID)
	})
}

func TestAccessReviewSourceInsert_TwoAccountsOnOneConnector(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderGitHub, key)
	require.NoError(t, err)

	firstAccount, err := insertConnectorAccount(ctx, client, scope, organizationID, connectorID, "a1", "Account 1")
	require.NoError(t, err)

	secondAccount, err := insertConnectorAccount(ctx, client, scope, organizationID, connectorID, "a2", "Account 2")
	require.NoError(t, err)

	now := time.Now().UTC()
	insertForAccount := func(accountID gid.GID, name string) (*coredata.AccessReviewSource, bool) {
		t.Helper()

		source := &coredata.AccessReviewSource{
			ID:                 gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
			OrganizationID:     organizationID,
			ConnectorAccountID: &accountID,
			Name:               name,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		var inserted bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			inserted, err = source.Insert(ctx, tx, scope)

			return err
		}))

		return source, inserted
	}

	first, inserted := insertForAccount(firstAccount.ID, "source a1")
	require.True(t, inserted)

	second, inserted := insertForAccount(secondAccount.ID, "source a2")
	require.True(t, inserted, "a second account on the same connector must insert")
	assert.NotEqual(t, first.ID, second.ID)

	_, inserted = insertForAccount(firstAccount.ID, "source a1 again")
	require.False(t, inserted, "a duplicate account source must return created false")

	loaded := &coredata.AccessReviewSource{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByConnectorAccountID(ctx, conn, scope, firstAccount.ID)
	}))
	assert.Equal(t, first.ID, loaded.ID)
	require.NotNil(t, loaded.ConnectorAccountID)
	assert.Equal(t, firstAccount.ID, *loaded.ConnectorAccountID)

	byID := &coredata.AccessReviewSource{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return byID.LoadByID(ctx, conn, scope, first.ID)
	}))
	require.NotNil(t, byID.ConnectorAccountID)
	assert.Equal(t, firstAccount.ID, *byID.ConnectorAccountID)
}
