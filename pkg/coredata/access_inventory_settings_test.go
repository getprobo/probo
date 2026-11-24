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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/timespan"
)

func TestAccessInventorySettingsUpsertAndLoad(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessInventoryFixture(t, ctx, client)
	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC().Truncate(time.Millisecond)

	settingsID := gid.New(tenantID, coredata.AccessInventorySettingsEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		settings := &coredata.AccessInventorySettings{
			ID:                    settingsID,
			OrganizationID:        fx.organizationID,
			HideAfterMissingSyncs: coredata.DefaultHideAfterMissingSyncs,
			CreatedAt:             now,
			UpdatedAt:             now,
		}

		inserted, err := settings.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}

		assert.True(t, inserted)
		assert.Equal(t, settingsID, settings.ID)

		return nil
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded := &coredata.AccessInventorySettings{}
		if err := loaded.LoadByOrganizationID(ctx, conn, fx.scope, fx.organizationID); err != nil {
			return err
		}

		assert.Equal(t, settingsID, loaded.ID)
		assert.Equal(t, coredata.DefaultHideAfterMissingSyncs, loaded.HideAfterMissingSyncs)
		assert.Nil(t, loaded.ChangeEventRetention)
		assert.Nil(t, loaded.SnapshotRetention)

		return nil
	}))

	updatedAt := now.Add(time.Minute)
	retention := timespan.FromDuration(90 * 24 * time.Hour)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		settings := &coredata.AccessInventorySettings{
			ID:                    gid.New(tenantID, coredata.AccessInventorySettingsEntityType),
			OrganizationID:        fx.organizationID,
			HideAfterMissingSyncs: 5,
			ChangeEventRetention:  &retention,
			SnapshotRetention:     &retention,
			CreatedAt:             updatedAt,
			UpdatedAt:             updatedAt,
		}

		inserted, err := settings.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}

		assert.False(t, inserted)
		assert.Equal(t, settingsID, settings.ID)
		assert.Equal(t, 5, settings.HideAfterMissingSyncs)

		return nil
	}))
}

func TestAccessInventoryProviderSyncConfigUpsertAndResolve(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessInventoryFixture(t, ctx, client)
	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC().Truncate(time.Millisecond)
	providerKey := coredata.AccessInventoryProviderKeyFromConnector(coredata.ConnectorProviderGitHub)
	override := timespan.FromDuration(15 * time.Minute)
	configID := gid.New(tenantID, coredata.AccessInventoryProviderSyncConfigEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		config := &coredata.AccessInventoryProviderSyncConfig{
			ID:             configID,
			OrganizationID: fx.organizationID,
			ProviderKey:    providerKey,
			SyncCadence:    override,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		inserted, err := config.Upsert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}

		assert.True(t, inserted)

		return nil
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded := &coredata.AccessInventoryProviderSyncConfig{}
		if err := loaded.LoadByOrganizationIDAndProviderKey(
			ctx,
			conn,
			fx.scope,
			fx.organizationID,
			providerKey,
		); err != nil {
			return err
		}

		assert.Equal(t, configID, loaded.ID)
		assert.Equal(t, providerKey, loaded.ProviderKey)

		duration, ok := loaded.SyncCadence.ClockDuration()
		require.True(t, ok)
		assert.Equal(t, 15*time.Minute, duration)

		resolved := coredata.ResolveSyncCadence(providerKey, &loaded.SyncCadence)
		resolvedDuration, ok := resolved.ClockDuration()
		require.True(t, ok)
		assert.Equal(t, 15*time.Minute, resolvedDuration)

		var configs coredata.AccessInventoryProviderSyncConfigs
		cursor := page.NewCursor(
			10,
			nil,
			page.Head,
			page.OrderBy[coredata.AccessInventoryProviderSyncConfigOrderField]{
				Field:     coredata.AccessInventoryProviderSyncConfigOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)
		if err := configs.LoadByOrganizationID(ctx, conn, fx.scope, fx.organizationID, cursor); err != nil {
			return err
		}

		require.Len(t, configs, 1)
		assert.Equal(t, configID, configs[0].ID)

		return nil
	}))
}
