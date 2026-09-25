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

package iam

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
)

func TestReactivateSCIMBridge(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	svc := OrganizationService{Service: &Service{pg: client}}

	states := []coredata.SCIMBridgeState{
		coredata.SCIMBridgeStateDisabled,
		coredata.SCIMBridgeStateFailed,
		coredata.SCIMBridgeStateSyncing,
		coredata.SCIMBridgeStatePending,
		coredata.SCIMBridgeStateActive,
	}

	for _, state := range states {
		t.Run(string(state), func(t *testing.T) {
			t.Parallel()

			scope, organizationID := seedResetBridgeOrg(t, ctx, client)
			bridge := insertResetBridge(t, ctx, client, scope, organizationID, state)

			reactivated, err := svc.ReactivateSCIMBridge(ctx, bridge.ID)
			require.NoError(t, err)

			assert.Equal(t, coredata.SCIMBridgeStateActive, reactivated.State)
			assert.Equal(t, 0, reactivated.ConsecutiveFailures)
			assert.Nil(t, reactivated.SyncError)
			require.NotNil(t, reactivated.NextSyncAt)
			assert.WithinDuration(t, time.Now(), *reactivated.NextSyncAt, 5*time.Second)
			assert.Equal(t, 9, reactivated.TotalFailureCount)
			assert.Equal(t, []string{"bot@example.com"}, reactivated.ExcludedUserNames)

			loaded := &coredata.SCIMBridge{}

			require.NoError(t, client.WithConn(
				ctx,
				func(ctx context.Context, conn pg.Querier) error {
					return loaded.LoadByID(ctx, conn, scope, bridge.ID)
				},
			))

			assert.Equal(t, coredata.SCIMBridgeStateActive, loaded.State)
			assert.Equal(t, 0, loaded.ConsecutiveFailures)
			assert.Nil(t, loaded.SyncError)
			require.NotNil(t, loaded.NextSyncAt)
			assert.WithinDuration(t, time.Now(), *loaded.NextSyncAt, 5*time.Second)
		})
	}

	t.Run("missing bridge", func(t *testing.T) {
		t.Parallel()

		scope, _ := seedResetBridgeOrg(t, ctx, client)
		missingID := gid.New(scope.GetTenantID(), coredata.SCIMBridgeEntityType)

		_, err := svc.ReactivateSCIMBridge(ctx, missingID)

		var notFound *ErrSCIMBridgeNotFound
		require.ErrorAs(t, err, &notFound)
	})
}

func seedResetBridgeOrg(t *testing.T, ctx context.Context, client *pg.Client) (coredata.Scoper, gid.GID) {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			org := &coredata.Organization{
				ID:        organizationID,
				TenantID:  tenantID,
				Name:      "SCIM bridge reset",
				CreatedAt: now,
				UpdatedAt: now,
			}

			return org.Insert(ctx, tx)
		},
	))

	return scope, organizationID
}

func insertResetBridge(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	state coredata.SCIMBridgeState,
) *coredata.SCIMBridge {
	t.Helper()

	now := time.Now()
	syncErr := "directory unavailable"

	var key cipher.EncryptionKey

	connectorID := gid.New(scope.GetTenantID(), coredata.ConnectorEntityType)
	bridge := &coredata.SCIMBridge{
		ID:                  gid.New(scope.GetTenantID(), coredata.SCIMBridgeEntityType),
		OrganizationID:      organizationID,
		ConnectorID:         &connectorID,
		Type:                coredata.SCIMBridgeTypeGoogleWorkspace,
		State:               state,
		ExcludedUserNames:   []string{"bot@example.com"},
		SyncError:           &syncErr,
		ConsecutiveFailures: 4,
		TotalFailureCount:   9,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if state != coredata.SCIMBridgeStateDisabled {
		future := now.Add(time.Hour)
		bridge.NextSyncAt = &future
	}

	require.NoError(t, client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{
				ID:             connectorID,
				OrganizationID: organizationID,
				Provider:       coredata.ConnectorProviderGoogleWorkspace,
				Protocol:       coredata.ConnectorProtocolOAuth2,
				Connection: &connector.OAuth2Connection{
					AccessToken: "test-token",
					TokenType:   "Bearer",
				},
				CreatedAt: now,
				UpdatedAt: now,
			}

			if err := cnnctr.Insert(ctx, tx, scope, key); err != nil {
				return err
			}

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

			bridge.ScimConfigurationID = config.ID

			return bridge.Insert(ctx, tx, scope)
		},
	))

	return bridge
}
