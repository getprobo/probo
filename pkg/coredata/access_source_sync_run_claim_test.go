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
	// Not parallel: shares the cross-tenant QUEUED claim queue with the
	// SkipLocked sibling below.
	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	createdAt := time.Now().UTC()
	runID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusQueued, createdAt)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		claimed := &coredata.AccessSourceSyncRun{}
		if err := claimed.LoadNextQueuedForUpdateSkipLocked(ctx, tx); err != nil {
			return err
		}

		assert.Equal(t, runID, claimed.ID)
		assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, claimed.Status)

		startedAt := createdAt
		claimed.Status = coredata.AccessSourceSyncRunStatusRunning
		claimed.StartedAt = &startedAt
		claimed.UpdatedAt = createdAt

		scope := coredata.NewScope(claimed.ID.TenantID(), gid.Nil)

		return claimed.Update(ctx, tx, scope)
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		ours := &coredata.AccessSourceSyncRun{}
		if err := ours.LoadByID(ctx, conn, fx.scope, runID); err != nil {
			return err
		}

		assert.Equal(t, coredata.AccessSourceSyncRunStatusRunning, ours.Status)

		return nil
	}))
}

func TestAccessSourceSyncRun_LoadNextQueuedForUpdateSkipLocked_SkipsSourceWithRunning(t *testing.T) {
	// Not parallel: shares the cross-tenant QUEUED claim queue.
	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	createdAt := time.Now().UTC()
	_ = insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusRunning, createdAt)
	_ = insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusQueued, createdAt.Add(time.Second))

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		claimed := &coredata.AccessSourceSyncRun{}

		return claimed.LoadNextQueuedForUpdateSkipLocked(ctx, tx)
	})
	assert.ErrorIs(t, err, coredata.ErrNoAccessSourceSyncRunAvailable)
}

func TestAccessSourceSyncRun_LoadNextQueuedForUpdateSkipLocked_OneQueuedPerSource(t *testing.T) {
	// Not parallel: shares the cross-tenant QUEUED claim queue.
	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	olderAt := time.Date(2, time.January, 1, 0, 0, 0, 0, time.UTC)
	newerAt := olderAt.Add(time.Second)
	olderID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusQueued, olderAt)
	_ = insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusQueued, newerAt)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		claimed := &coredata.AccessSourceSyncRun{}
		if err := claimed.LoadNextQueuedForUpdateSkipLocked(ctx, tx); err != nil {
			return err
		}

		assert.Equal(t, olderID, claimed.ID)

		return nil
	}))
}

func TestAccessSourceSyncRun_LoadNextQueuedForUpdateSkipLocked_SkipLocked(t *testing.T) {
	// Not parallel: pins exact claim IDs against the shared QUEUED queue.
	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	// Far-past stamps so these two rows head the global claim queue.
	unlockedAt := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	lockedAt := unlockedAt.Add(time.Second)
	unlockedID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusQueued, unlockedAt)
	lockedID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusQueued, lockedAt)

	var (
		locked    = make(chan error, 1)
		releaseCh = make(chan struct{})
		release   = sync.OnceFunc(func() { close(releaseCh) })
		done      = make(chan error, 1)
	)

	go func() {
		done <- client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var held gid.GID
			if err := tx.QueryRow(
				ctx,
				`SELECT id FROM access_source_sync_runs WHERE id = @id FOR UPDATE`,
				pgx.StrictNamedArgs{"id": lockedID},
			).Scan(&held); err != nil {
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

		assert.Equal(t, unlockedID, claimed.ID, "SKIP LOCKED must claim the unlocked candidate")
		assert.NotEqual(t, lockedID, claimed.ID)

		return nil
	})

	release()
	require.NoError(t, <-done)
	require.NoError(t, err, "claim must skip the locked row instead of waiting on it")
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

		scope := coredata.NewScope(stale[0].ID.TenantID(), gid.Nil)

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

func TestAccessSourceSyncRuns_LoadStaleRunningForUpdateSkipLocked_SkipLocked(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	staleAt := time.Date(2001, time.March, 1, 0, 0, 0, 0, time.UTC)
	staleThreshold := staleAt.Add(time.Hour)

	lockedID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusRunning, staleAt)
	unlockedID := insertAccessSourceSyncRun(
		t,
		ctx,
		client,
		fx,
		coredata.AccessSourceSyncRunStatusRunning,
		staleAt.Add(time.Second),
	)

	var (
		locked    = make(chan error, 1)
		releaseCh = make(chan struct{})
		release   = sync.OnceFunc(func() { close(releaseCh) })
		done      = make(chan error, 1)
		ownLocked coredata.AccessSourceSyncRuns
	)

	go func() {
		done <- client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var held gid.GID
			if err := tx.QueryRow(
				ctx,
				`SELECT id FROM access_source_sync_runs WHERE id = @id FOR UPDATE`,
				pgx.StrictNamedArgs{"id": lockedID},
			).Scan(&held); err != nil {
				locked <- err

				return err
			}

			locked <- nil

			<-releaseCh

			return ownLocked.LoadStaleRunningForUpdateSkipLocked(ctx, tx, staleThreshold, 100)
		})
	}()

	require.NoError(t, <-locked)
	t.Cleanup(release)

	sweepCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var swept coredata.AccessSourceSyncRuns

	err := client.WithTx(sweepCtx, func(ctx context.Context, tx pg.Tx) error {
		return swept.LoadStaleRunningForUpdateSkipLocked(ctx, tx, staleThreshold, 100)
	})

	release()
	require.NoError(t, <-done)
	require.NoError(t, err, "the sweep must skip a locked stale run instead of waiting on it")

	sweptIDs := make([]gid.GID, 0, len(swept))
	for _, run := range swept {
		sweptIDs = append(sweptIDs, run.ID)
	}

	assert.NotContains(t, sweptIDs, lockedID, "a locked stale run must be left for the next sweep")
	assert.Contains(t, sweptIDs, unlockedID, "the unlocked stale run must still be recovered")

	ownIDs := make([]gid.GID, 0, len(ownLocked))
	for _, run := range ownLocked {
		ownIDs = append(ownIDs, run.ID)
	}

	assert.Contains(t, ownIDs, lockedID, "the lock holder still sees its stale run")
}

func TestAccessSourceSyncRun_UpdateRejectsTerminalRows(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedAccessInventoryFixture(t, ctx, client)
	now := time.Now().UTC()
	runID := insertAccessSourceSyncRun(t, ctx, client, fx, coredata.AccessSourceSyncRunStatusSuccess, now)

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		run := &coredata.AccessSourceSyncRun{}
		if err := run.LoadByID(ctx, tx, fx.scope, runID); err != nil {
			return err
		}

		run.Status = coredata.AccessSourceSyncRunStatusFailed
		run.UpdatedAt = now.Add(time.Minute)

		return run.Update(ctx, tx, fx.scope)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded := &coredata.AccessSourceSyncRun{}
		if err := loaded.LoadByID(ctx, conn, fx.scope, runID); err != nil {
			return err
		}

		assert.Equal(t, coredata.AccessSourceSyncRunStatusSuccess, loaded.Status)

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
	eventID := gid.New(tenantID, coredata.AccessInventoryChangeEventEntityType)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		event := &coredata.AccessInventoryChangeEvent{
			ID:                   eventID,
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

	fxOther := seedAccessInventoryFixture(t, ctx, client)
	otherRunID := insertAccessSourceSyncRun(t, ctx, client, fxOther, coredata.AccessSourceSyncRunStatusSuccess, now)

	err = client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
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
			SyncRunID:            &otherRunID,
			After:                &afterPayload,
			IdempotencyKey:       "cross-source-sync-run|" + otherRunID.String(),
			CreatedAt:            now,
		}

		_, err := event.Insert(ctx, tx, fx.scope)

		return err
	})
	require.Error(t, err, "composite FK must reject a sync run from another source/tenant")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		run := &coredata.AccessSourceSyncRun{ID: runID}

		return run.Delete(ctx, tx, fx.scope)
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded := &coredata.AccessInventoryChangeEvent{}
		if err := loaded.LoadByID(ctx, conn, fx.scope, eventID); err != nil {
			return err
		}

		assert.Nil(t, loaded.SyncRunID, "ON DELETE SET NULL must clear only sync_run_id")
		assert.Equal(t, fx.organizationID, loaded.OrganizationID)
		assert.Equal(t, fx.sourceID, loaded.AccessReviewSourceID)

		return nil
	}))
}
