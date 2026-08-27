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

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func insertSource(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	connectorID *gid.GID,
) (*coredata.AccessReviewSource, bool) {
	t.Helper()

	now := time.Now().UTC()
	source := &coredata.AccessReviewSource{
		ID:             gid.New(scope.GetTenantID(), coredata.AccessReviewSourceEntityType),
		OrganizationID: organizationID,
		ConnectorID:    connectorID,
		Name:           "owner test source",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	var inserted bool

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		inserted, err = source.Insert(ctx, tx, scope)

		return err
	}))

	return source, inserted
}

// TestAccessReviewSourceInsert_IdempotentPerConnector pins the
// index-arbitrated idempotency CreateSource relies on: the second
// insert against the same connector is skipped, while CSV sources
// (nil connector) always insert.
func TestAccessReviewSourceInsert_IdempotentPerConnector(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderGitHub, key)
	require.NoError(t, err)

	first, inserted := insertSource(t, ctx, client, scope, organizationID, &connectorID)
	require.True(t, inserted)

	_, inserted = insertSource(t, ctx, client, scope, organizationID, &connectorID)
	require.False(t, inserted, "second insert for the same connector must be skipped")

	loaded := &coredata.AccessReviewSource{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByConnectorID(ctx, conn, scope, connectorID)
	}))
	require.Equal(t, first.ID, loaded.ID)

	_, inserted = insertSource(t, ctx, client, scope, organizationID, nil)
	require.True(t, inserted, "CSV sources never conflict")

	_, inserted = insertSource(t, ctx, client, scope, organizationID, nil)
	require.True(t, inserted, "CSV sources never conflict")
}

// TestAccessReviewSourceDeleteReturningConnectorID pins the connector
// handoff read: the delete returns the connector referenced at delete
// time and ErrResourceNotFound for a missing source.
func TestAccessReviewSourceDeleteReturningConnectorID(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderGitHub, key)
	require.NoError(t, err)

	source, inserted := insertSource(t, ctx, client, scope, organizationID, &connectorID)
	require.True(t, inserted)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		returned, err := source.DeleteReturningConnectorID(ctx, tx, scope)
		if err != nil {
			return err
		}

		require.NotNil(t, returned)
		require.Equal(t, connectorID, *returned)

		return nil
	}))

	err = client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := source.DeleteReturningConnectorID(ctx, tx, scope)

		return err
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

// TestSCIMBridgeConnectorUnique pins the schema half of single
// ownership on the bridge side: two bridges cannot reference one
// connector.
func TestSCIMBridgeConnectorUnique(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderGoogleWorkspace, key)
	require.NoError(t, err)

	now := time.Now().UTC()

	// A configuration is unique per organization, so the second bridge
	// needs its own organization; the connector unique index must refuse
	// it all the same.
	insertBridge := func(scope coredata.Scoper, organizationID gid.GID) error {
		return client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			config := &coredata.SCIMConfiguration{
				ID:             gid.New(scope.GetTenantID(), coredata.SCIMConfigurationEntityType),
				OrganizationID: organizationID,
				HashedToken:    []byte{0x01},
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := config.Insert(ctx, tx, scope); err != nil {
				return err
			}

			bridge := &coredata.SCIMBridge{
				ID:                  gid.New(scope.GetTenantID(), coredata.SCIMBridgeEntityType),
				OrganizationID:      organizationID,
				ScimConfigurationID: config.ID,
				ConnectorID:         &connectorID,
				Type:                coredata.SCIMBridgeTypeGoogleWorkspace,
				State:               coredata.SCIMBridgeStateActive,
				ExcludedUserNames:   []string{},
				CreatedAt:           now,
				UpdatedAt:           now,
			}

			return bridge.Insert(ctx, tx, scope)
		})
	}

	require.NoError(t, insertBridge(scope, organizationID))

	otherScope, otherOrganizationID := seedConnectorOrg(t, ctx, client)

	err = insertBridge(otherScope, otherOrganizationID)
	require.Error(t, err)

	var pgErr *pgconn.PgError

	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23505", pgErr.Code)
	require.Equal(t, "idx_iam_scim_bridges_connector_id", pgErr.ConstraintName)
}

// TestConnectorDeleteRefusedUnderBridge pins the FK flip: deleting a
// connector out from under a live bridge is refused as in-use instead
// of silently unbinding the bridge (the old ON DELETE SET NULL).
func TestConnectorDeleteRefusedUnderBridge(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := seedConnectorOrg(t, ctx, client)

	var key cipher.EncryptionKey

	connectorID, err := insertConnector(ctx, client, scope, organizationID, coredata.ConnectorProviderGoogleWorkspace, key)
	require.NoError(t, err)

	now := time.Now().UTC()

	err = client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		config := &coredata.SCIMConfiguration{
			ID:             gid.New(scope.GetTenantID(), coredata.SCIMConfigurationEntityType),
			OrganizationID: organizationID,
			HashedToken:    []byte{0x01},
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := config.Insert(ctx, tx, scope); err != nil {
			return err
		}

		bridge := &coredata.SCIMBridge{
			ID:                  gid.New(scope.GetTenantID(), coredata.SCIMBridgeEntityType),
			OrganizationID:      organizationID,
			ScimConfigurationID: config.ID,
			ConnectorID:         &connectorID,
			Type:                coredata.SCIMBridgeTypeGoogleWorkspace,
			State:               coredata.SCIMBridgeStateActive,
			ExcludedUserNames:   []string{},
			CreatedAt:           now,
			UpdatedAt:           now,
		}

		return bridge.Insert(ctx, tx, scope)
	})
	require.NoError(t, err)

	err = client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		connector := &coredata.Connector{ID: connectorID}

		return connector.Delete(ctx, tx, scope)
	})
	require.ErrorIs(t, err, coredata.ErrResourceInUse)
}
