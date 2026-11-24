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
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func insertAccessSourceSyncRun(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx accessInventoryFixture,
	status coredata.AccessSourceSyncRunStatus,
	updatedAt time.Time,
) gid.GID {
	t.Helper()

	tenantID := fx.scope.GetTenantID()
	runID := gid.New(tenantID, coredata.AccessSourceSyncRunEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		run := &coredata.AccessSourceSyncRun{
			ID:                   runID,
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			Status:               status,
			GrantsUpsertedCount:  0,
			CreatedAt:            updatedAt,
			UpdatedAt:            updatedAt,
		}

		return run.Insert(ctx, tx, fx.scope)
	}))

	return runID
}

func TestAccessSourceSyncRun_LoadNextQueuedForUpdateSkipLocked(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	now := time.Now().UTC()
	runID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusQueued, now)
	_ = insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusRunning, now)

	var claimedID gid.GID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		claimed := &coredata.AccessSourceSyncRun{}
		if err := claimed.LoadNextQueuedForUpdateSkipLocked(ctx, tx); err != nil {
			return err
		}

		assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, claimed.Status)
		assert.NotEqual(t, gid.TenantID{}, claimed.TenantID)

		startedAt := now
		claimed.Status = coredata.AccessSourceSyncRunStatusRunning
		claimed.StartedAt = &startedAt
		claimed.UpdatedAt = now
		claimedID = claimed.ID

		scope := coredata.NewScope(claimed.TenantID, gid.Nil)

		return claimed.Update(ctx, tx, scope)
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		scope := coredata.NewScope(claimedID.TenantID(), gid.Nil)
		loaded := &coredata.AccessSourceSyncRun{}
		if err := loaded.LoadByID(ctx, conn, scope, claimedID); err != nil {
			return err
		}

		assert.Equal(t, coredata.AccessSourceSyncRunStatusRunning, loaded.Status)
		require.NotNil(t, loaded.StartedAt)

		// Our fixture row remains claimable if a different queued row was taken.
		ours := &coredata.AccessSourceSyncRun{}
		if err := ours.LoadByID(ctx, conn, fx.scope, runID); err != nil {
			return err
		}

		assert.Contains(
			t,
			[]coredata.AccessSourceSyncRunStatus{
				coredata.AccessSourceSyncRunStatusQueued,
				coredata.AccessSourceSyncRunStatusRunning,
			},
			ours.Status,
		)

		return nil
	}))
}

func TestAccessSourceSyncRun_LoadNextQueuedForUpdateSkipLocked_SkipLocked(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	createdAt := time.Now().UTC()
	runID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusQueued, createdAt)

	var (
		locked    = make(chan error, 1)
		releaseCh = make(chan struct{})
		release   = sync.OnceFunc(func() { close(releaseCh) })
		done      = make(chan error, 1)
	)

	// Hold a row lock on this specific QUEUED run the way a worker claim would,
	// without racing other tests' older queued rows via LoadNextQueued.
	go func() {
		done <- client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var lockedID gid.GID
			if err := tx.QueryRow(
				ctx,
				`SELECT id FROM access_source_sync_runs WHERE id = @id FOR UPDATE`,
				pgx.StrictNamedArgs{"id": runID},
			).Scan(&lockedID); err != nil {
				locked <- err

				return err
			}

			locked <- nil

			<-releaseCh

			return nil
		})
	}()

	require.NoError(t, <-locked)
	t.Cleanup(release)

	sweepCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := client.WithTx(sweepCtx, func(ctx context.Context, tx pg.Tx) error {
		claimed := &coredata.AccessSourceSyncRun{}
		if err := claimed.LoadNextQueuedForUpdateSkipLocked(ctx, tx); err != nil {
			return err
		}

		assert.NotEqual(t, runID, claimed.ID)

		return nil
	})

	release()
	require.NoError(t, <-done)

	if err != nil {
		require.ErrorIs(t, err, coredata.ErrNoAccessSourceSyncRunAvailable)
	}
}

func TestAccessSourceSyncRuns_LoadStaleRunningForUpdateSkipLocked(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	staleAt := time.Date(2001, time.February, 1, 0, 0, 0, 0, time.UTC)
	staleThreshold := staleAt.Add(time.Hour)
	freshAt := time.Now().UTC()

	staleID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusRunning, staleAt)
	_ = insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusRunning, freshAt)
	_ = insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusQueued, staleAt)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var stale coredata.AccessSourceSyncRuns
		if err := stale.LoadStaleRunningForUpdateSkipLocked(ctx, tx, staleThreshold, 100); err != nil {
			return err
		}

		require.Len(t, stale, 1)
		assert.Equal(t, staleID, stale[0].ID)

		errMsg := "stale sync run recovered"
		completedAt := time.Now().UTC()
		stale[0].Status = coredata.AccessSourceSyncRunStatusFailed
		stale[0].Error = &errMsg
		stale[0].CompletedAt = &completedAt
		stale[0].UpdatedAt = completedAt

		scope := coredata.NewScope(stale[0].TenantID, gid.Nil)

		return stale[0].Update(ctx, tx, scope)
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded := &coredata.AccessSourceSyncRun{}
		if err := loaded.LoadByID(ctx, conn, fx.scope, staleID); err != nil {
			return err
		}

		assert.Equal(t, coredata.AccessSourceSyncRunStatusFailed, loaded.Status)
		require.NotNil(t, loaded.Error)
		assert.Equal(t, "stale sync run recovered", *loaded.Error)

		return nil
	}))
}

func TestAccessInventoryChangeEvent_SyncRunForeignKey(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	grantID := seedAccessInventoryGrant(t, ctx, client, fx)
	tenantID := fx.scope.GetTenantID()
	now := time.Now().UTC()
	runID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusSuccess, now)
	afterPayload := json.RawMessage(`{"presence":"PRESENT"}`)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		event := &coredata.AccessInventoryChangeEvent{
			ID:                   gid.New(tenantID, coredata.AccessInventoryChangeEventEntityType),
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			GrantID:              &grantID,
			PrincipalKey:         "change-event@example.com",
			ResourceKey:          "system",
			EntitlementKey:       "member",
			ChangeType:           coredata.AccessInventoryChangeTypeAppeared,
			ObservedAt:           now,
			SyncRunID:            &runID,
			After:                &afterPayload,
			IdempotencyKey:       "with-sync-run|" + runID.String(),
			CreatedAt:            now,
		}

		inserted, err := event.Insert(ctx, tx, fx.scope)
		if err != nil {
			return err
		}

		require.True(t, inserted)

		loaded := &coredata.AccessInventoryChangeEvent{}
		if err := loaded.LoadByID(ctx, tx, fx.scope, event.ID); err != nil {
			return err
		}

		require.NotNil(t, loaded.SyncRunID)
		assert.Equal(t, runID, *loaded.SyncRunID)

		return nil
	}))

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		missingRunID := gid.New(tenantID, coredata.AccessSourceSyncRunEntityType)
		event := &coredata.AccessInventoryChangeEvent{
			ID:                   gid.New(tenantID, coredata.AccessInventoryChangeEventEntityType),
			OrganizationID:       fx.organizationID,
			AccessReviewSourceID: fx.sourceID,
			GrantID:              &grantID,
			PrincipalKey:         "change-event@example.com",
			ResourceKey:          "system",
			EntitlementKey:       "member",
			ChangeType:           coredata.AccessInventoryChangeTypeAppeared,
			ObservedAt:           now,
			SyncRunID:            &missingRunID,
			After:                &afterPayload,
			IdempotencyKey:       "missing-sync-run|" + missingRunID.String(),
			CreatedAt:            now,
		}

		_, err := event.Insert(ctx, tx, fx.scope)

		return err
	})
	require.Error(t, err)
	assert.False(t, errors.Is(err, coredata.ErrNoAccessSourceSyncRunAvailable))
}
