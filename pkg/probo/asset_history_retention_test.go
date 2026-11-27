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

package probo

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestAssetHistoryRetentionHandler_DeletesNamedSnapshotsAndChangeEvents(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	now := time.Now().UTC().Round(time.Microsecond)
	org := insertHistoryOrganization(t, client, now)
	scope := coredata.NewScope(org.tenantID, gid.Nil)
	assetID := insertHistoryAsset(t, client, scope, org, now)

	name := "named snapshot"
	expiredSnapshot := insertHistorySnapshot(
		t,
		client,
		scope,
		org,
		assetID,
		now.Add(-DefaultAssetHistoryRetention).Add(-time.Second),
		&name,
	)
	boundarySnapshot := insertHistorySnapshot(
		t,
		client,
		scope,
		org,
		assetID,
		now.Add(-DefaultAssetHistoryRetention),
		nil,
	)
	recentSnapshot := insertHistorySnapshot(
		t,
		client,
		scope,
		org,
		assetID,
		now.Add(-time.Hour),
		nil,
	)

	insertHistoryChangeEvent(
		t,
		client,
		scope,
		org,
		assetID,
		now.Add(-DefaultAssetHistoryRetention).Add(-time.Second),
	)
	boundaryEvent := insertHistoryChangeEvent(
		t,
		client,
		scope,
		org,
		assetID,
		now.Add(-DefaultAssetHistoryRetention),
	)
	recentEvent := insertHistoryChangeEvent(
		t,
		client,
		scope,
		org,
		assetID,
		now.Add(-time.Hour),
	)

	h := &assetHistoryRetentionHandler{
		pg:        client,
		logger:    log.NewLogger(log.WithOutput(io.Discard)),
		retention: DefaultAssetHistoryRetention,
		batchSize: 1,
		now:       func() time.Time { return now },
	}

	require.NoError(t, h.Run(t.Context()))

	assert.Equal(
		t,
		[]gid.GID{boundarySnapshot, recentSnapshot},
		loadHistorySnapshotIDs(t, client, org.organizationID),
	)
	assert.Equal(t, 0, countHistoryCopies(t, client, expiredSnapshot))
	assert.Equal(t, 1, countHistoryCopies(t, client, recentSnapshot))
	assert.Equal(
		t,
		[]gid.GID{boundaryEvent, recentEvent},
		loadHistoryChangeEventIDs(t, client, org.organizationID),
	)
}

func TestAssetHistoryRetentionHandler_StopsWhenCanceled(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	// Ahead of the real clock. A parallel sweep deletes rows older than
	// 90 days from time.Now(), so this row stays out of its range.
	now := time.Now().UTC().Add(2 * DefaultAssetHistoryRetention).Round(time.Microsecond)
	org := insertHistoryOrganization(t, client, now)
	scope := coredata.NewScope(org.tenantID, gid.Nil)
	assetID := insertHistoryAsset(t, client, scope, org, now)
	name := "kept while canceled"
	kept := insertHistorySnapshot(
		t,
		client,
		scope,
		org,
		assetID,
		now.Add(-DefaultAssetHistoryRetention).Add(-time.Second),
		&name,
	)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	h := &assetHistoryRetentionHandler{
		pg:        client,
		logger:    log.NewLogger(log.WithOutput(io.Discard)),
		retention: DefaultAssetHistoryRetention,
		batchSize: 1,
		now:       func() time.Time { return now },
	}

	err := h.Run(ctx)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, []gid.GID{kept}, loadHistorySnapshotIDs(t, client, org.organizationID))
}

type historyOrganization struct {
	tenantID       gid.TenantID
	organizationID gid.GID
}

func insertHistoryOrganization(
	t *testing.T,
	client *pg.Client,
	now time.Time,
) historyOrganization {
	t.Helper()

	tenantID := gid.NewTenantID()
	org := historyOrganization{
		tenantID:       tenantID,
		organizationID: gid.New(tenantID, coredata.OrganizationEntityType),
	}

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at, asset_inventory_enabled)
			 VALUES ($1, $2, $3, $4, $5, false)`,
			org.organizationID,
			org.tenantID,
			"asset-history-"+org.organizationID.String(),
			now,
			now,
		)

		return err
	}))

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, org.organizationID)

			return err
		})
	})

	return org
}

func insertHistoryAsset(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	org historyOrganization,
	now time.Time,
) gid.GID {
	t.Helper()

	asset := &coredata.Asset{
		ID:             gid.New(org.tenantID, coredata.AssetEntityType),
		Name:           "retained asset",
		Amount:         1,
		OrganizationID: org.organizationID,
		Type:           coredata.AssetTypeVirtual,
		Status:         coredata.AssetStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		return asset.Insert(ctx, tx, scope)
	}))

	return asset.ID
}

func insertHistorySnapshot(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	org historyOrganization,
	assetID gid.GID,
	capturedAt time.Time,
	name *string,
) gid.GID {
	t.Helper()

	snapshot := &coredata.AssetSnapshot{
		ID:             gid.New(org.tenantID, coredata.AssetSnapshotEntityType),
		OrganizationID: org.organizationID,
		CapturedAt:     capturedAt,
		Name:           name,
	}

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		if err := snapshot.Insert(ctx, tx, scope); err != nil {
			return err
		}

		_, err := tx.Exec(
			ctx,
			`INSERT INTO asset_snapshot_assets (
				id, tenant_id, snapshot_id, organization_id, asset_id, name, amount,
				asset_type, status, asset_created_at, asset_updated_at,
				third_party_ids, business_function_ids, tag_ids
			) VALUES (
				$1, $2, $3, $4, $5, 'retained asset', 1,
				'VIRTUAL', 'ACTIVE', $6, $6,
				'{}', '{}', '{}'
			)`,
			gid.New(org.tenantID, coredata.AssetSnapshotAssetEntityType),
			org.tenantID,
			snapshot.ID,
			org.organizationID,
			assetID,
			capturedAt,
		)

		return err
	}))

	return snapshot.ID
}

func insertHistoryChangeEvent(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	org historyOrganization,
	assetID gid.GID,
	observedAt time.Time,
) gid.GID {
	t.Helper()

	event := &coredata.AssetChangeEvent{
		ID:             gid.New(org.tenantID, coredata.AssetChangeEventEntityType),
		OrganizationID: org.organizationID,
		AssetID:        assetID,
		ObservedAt:     observedAt,
		ChangeType:     coredata.AssetChangeTypeCreated,
		After:          []byte(`{"name":"retained asset"}`),
	}

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		return coredata.AssetChangeEvents{event}.Insert(ctx, tx, scope)
	}))

	return event.ID
}

func loadHistorySnapshotIDs(
	t *testing.T,
	client *pg.Client,
	organizationID gid.GID,
) []gid.GID {
	t.Helper()

	return loadHistoryIDs(
		t,
		client,
		`SELECT id FROM asset_snapshots WHERE organization_id = $1 ORDER BY captured_at, id`,
		organizationID,
	)
}

func loadHistoryChangeEventIDs(
	t *testing.T,
	client *pg.Client,
	organizationID gid.GID,
) []gid.GID {
	t.Helper()

	return loadHistoryIDs(
		t,
		client,
		`SELECT id FROM asset_change_events WHERE organization_id = $1 ORDER BY observed_at, id`,
		organizationID,
	)
}

func loadHistoryIDs(
	t *testing.T,
	client *pg.Client,
	query string,
	organizationID gid.GID,
) []gid.GID {
	t.Helper()

	var ids []gid.GID

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		rows, err := conn.Query(ctx, query, organizationID)
		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			var id gid.GID
			if err := rows.Scan(&id); err != nil {
				return err
			}

			ids = append(ids, id)
		}

		return rows.Err()
	}))

	return ids
}

func countHistoryCopies(t *testing.T, client *pg.Client, snapshotID gid.GID) int {
	t.Helper()

	var count int

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return conn.QueryRow(
			ctx,
			`SELECT COUNT(*) FROM asset_snapshot_assets WHERE snapshot_id = $1`,
			snapshotID,
		).Scan(&count)
	}))

	return count
}
