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

package accessreview

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
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/timespan"
)

func TestInventorySyncFailureBackoff(t *testing.T) {
	t.Parallel()

	assert.Equal(t, time.Minute, inventorySyncFailureBackoff(0))
	assert.Equal(t, time.Minute, inventorySyncFailureBackoff(1))
	assert.Equal(t, 2*time.Minute, inventorySyncFailureBackoff(2))
	assert.Equal(t, 4*time.Minute, inventorySyncFailureBackoff(3))
	assert.Equal(t, inventorySyncFailureMaxBackoff, inventorySyncFailureBackoff(100))
}

func TestApplyInventorySyncJitter(t *testing.T) {
	t.Parallel()

	t.Run("zero base unchanged", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, time.Duration(0), applyInventorySyncJitter(0, func(int64) int64 { return 0 }))
	})

	t.Run("spreads within plus or minus twenty percent", func(t *testing.T) {
		t.Parallel()

		base := 100 * time.Minute

		low := applyInventorySyncJitter(base, func(n int64) int64 {
			assert.Equal(t, int64(base/inventorySyncJitterFraction)*2+1, n)
			return 0
		})
		high := applyInventorySyncJitter(base, func(n int64) int64 {
			return n - 1
		})

		assert.Equal(t, 80*time.Minute, low)
		assert.Equal(t, 120*time.Minute, high)
	})
}

func TestSyncNow_EnqueuesQueuedRun(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\nalice@example.com,Alice,Admin\n")
	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))

	run, err := svc.SyncNow(ctx, fx.scope, fx.sourceID)
	require.NoError(t, err)
	require.NotNil(t, run)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, run.Status)
	assert.Equal(t, fx.sourceID, run.AccessReviewSourceID)

	again, err := svc.SyncNow(ctx, fx.scope, fx.sourceID)
	require.NoError(t, err)
	assert.Equal(t, run.ID, again.ID, "second SyncNow must reuse the in-flight run")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return run.Delete(ctx, tx, fx.scope)
	}))
}

func TestInventorySyncScheduler_EnqueuesDueSource(t *testing.T) {
	// Not parallel: shares the cross-tenant due-source claim queue.
	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\nbob@example.com,Bob,Member\n")
	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
	require.NoError(t, holdDueInventorySyncSchedulesExcept(t, ctx, client, fx.sourceID))

	past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		source := &coredata.AccessReviewSource{}
		if err := source.LoadByID(ctx, tx, fx.scope, fx.sourceID); err != nil {
			return err
		}

		source.NextSyncAt = &past
		source.UpdatedAt = time.Now().UTC()

		return source.UpdateInventorySyncSchedule(ctx, tx, fx.scope)
	}))

	h := &inventorySyncSchedulerHandler{svc: svc, pg: client, logger: logger}
	claimed, err := h.Claim(ctx)
	require.NoError(t, err)
	assert.Equal(t, fx.sourceID, claimed.ID)

	inflight := &coredata.AccessSourceSyncRun{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return inflight.LoadInFlightBySourceID(ctx, conn, fx.scope, fx.sourceID)
	}))
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, inflight.Status)

	reloaded := &coredata.AccessReviewSource{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, fx.sourceID)
	}))
	require.NotNil(t, reloaded.NextSyncAt)
	assert.True(t, reloaded.NextSyncAt.After(time.Now().UTC()), "cadence must advance past now")

	// Still in flight + next_sync_at in the future: this source must not be due.
	due := &coredata.AccessReviewSource{}

	err = client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return due.LoadNextDueForInventorySyncSkipLocked(ctx, tx, time.Now().UTC())
	})
	if err == nil {
		assert.NotEqual(t, fx.sourceID, due.ID, "our source must not be reclaimable while in flight")
	} else {
		require.ErrorIs(t, err, coredata.ErrNoAccessReviewSourceInventorySyncDue)
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return inflight.Delete(ctx, tx, fx.scope)
	}))
}

func TestLoadNextDue_SkipsPastCadenceWhileQueued(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\n")
	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))

	past := time.Now().UTC().Add(-time.Hour)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		source := &coredata.AccessReviewSource{}
		if err := source.LoadByID(ctx, tx, fx.scope, fx.sourceID); err != nil {
			return err
		}

		source.NextSyncAt = &past
		source.UpdatedAt = time.Now().UTC()

		return source.UpdateInventorySyncSchedule(ctx, tx, fx.scope)
	}))

	_ = insertQueuedInventorySyncRun(t, ctx, client, fx)

	due := &coredata.AccessReviewSource{}

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return due.LoadNextDueForInventorySyncSkipLocked(ctx, tx, time.Now().UTC())
	})
	if err == nil {
		assert.NotEqual(t, fx.sourceID, due.ID, "QUEUED run must block due claim even when cadence is past")
	} else {
		require.ErrorIs(t, err, coredata.ErrNoAccessReviewSourceInventorySyncDue)
	}

	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
}

func TestSyncInventorySource_RecordsLastSuccessfulSyncAt(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvData := "email,full_name,role\ncarol@example.com,Carol,Admin\n"
	fx := seedInventorySyncFixture(t, ctx, client, csvData)

	run := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, run)
	require.NoError(t, svc.SyncInventorySource(ctx, fx.scope, run))

	source := &coredata.AccessReviewSource{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return source.LoadByID(ctx, conn, fx.scope, fx.sourceID)
	}))
	require.NotNil(t, source.LastSuccessfulSyncAt)
	assert.Equal(t, 0, source.SyncConsecutiveFailures)
	require.NotNil(t, source.NextSyncAt)
	assert.True(t, source.NextSyncAt.After(*source.LastSuccessfulSyncAt))
}

func TestInventorySyncHandler_FailureSchedulesBackoff(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedInventorySyncFixture(t, ctx, client, "full_name\nAlice\n")
	run := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, run)

	before := time.Now().UTC()
	h := &inventorySyncHandler{svc: svc, pg: client, logger: logger}
	require.NoError(t, h.handle(ctx, run))

	loaded := loadSyncRun(t, ctx, client, fx, run.ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusFailed, loaded.Status)

	source := &coredata.AccessReviewSource{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return source.LoadByID(ctx, conn, fx.scope, fx.sourceID)
	}))
	assert.Equal(t, 1, source.SyncConsecutiveFailures)
	require.NotNil(t, source.NextSyncAt)
	assert.True(t, source.NextSyncAt.After(before))
	assert.Nil(t, source.LastSuccessfulSyncAt)
}

func TestEnsureSources_DefersSyncUntilSyncNow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        orgID,
			TenantID:  tenantID,
			Name:      "Sync On Connect",
			CreatedAt: now,
			UpdatedAt: now,
		}

		return org.Insert(ctx, tx)
	}))

	csv := "email,full_name,role\n"
	results, err := svc.EnsureSources(
		ctx,
		scope,
		orgID,
		[]CreateAccessReviewSourceRequest{
			{
				OrganizationID: orgID,
				Name:           "CSV",
				CsvData:        &csv,
			},
		},
	)
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.True(t, results[0].Created)
	require.NotNil(t, results[0].Source.NextSyncAt)

	csvCadence, ok := coredata.DefaultSyncCadence(coredata.AccessInventoryProviderKeyCSV).ClockDuration()
	require.True(t, ok)
	assert.WithinDuration(
		t,
		now.Add(csvCadence),
		*results[0].Source.NextSyncAt,
		2*time.Second,
		"connect must defer next_sync_at by the provider cadence",
	)

	inflight := &coredata.AccessSourceSyncRun{}
	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return inflight.LoadInFlightBySourceID(ctx, conn, scope, results[0].Source.ID)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	run, err := svc.SyncNow(ctx, scope, results[0].Source.ID)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, run.Status)

	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, results[0].Source.ID))
}

func TestNeedsInitialInventorySync(t *testing.T) {
	t.Parallel()

	assert.True(t, NeedsInitialInventorySync(EnsuredAccessReviewSource{
		Source:  &coredata.AccessReviewSource{},
		Created: true,
	}))

	assert.True(t, NeedsInitialInventorySync(EnsuredAccessReviewSource{
		Source:  &coredata.AccessReviewSource{},
		Created: false,
	}), "recreate must retry SyncNow until first success")

	syncedAt := time.Now().UTC()
	assert.False(t, NeedsInitialInventorySync(EnsuredAccessReviewSource{
		Source:  &coredata.AccessReviewSource{LastSuccessfulSyncAt: &syncedAt},
		Created: false,
	}))
}

func TestResolveSourceSyncCadence_UsesOrgOverride(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\n")
	override := timespan.FromDuration(15 * time.Minute)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		config := &coredata.AccessInventoryProviderSyncConfig{
			ID:             gid.New(fx.scope.GetTenantID(), coredata.AccessInventoryProviderSyncConfigEntityType),
			OrganizationID: fx.organizationID,
			ProviderKey:    coredata.AccessInventoryProviderKeyCSV,
			SyncCadence:    override,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}
		_, err := config.Upsert(ctx, tx, fx.scope)

		return err
	}))

	source := &coredata.AccessReviewSource{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return source.LoadByID(ctx, conn, fx.scope, fx.sourceID)
	}))

	var cadence time.Duration

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var err error

		cadence, err = svc.resolveSourceSyncCadence(ctx, conn, fx.scope, source)

		return err
	}))
	assert.Equal(t, 15*time.Minute, cadence)
}

func TestAccessInventoryProviderSyncConfig_RejectsMonthCadence(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\n")
	now := time.Now().UTC()

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		config := &coredata.AccessInventoryProviderSyncConfig{
			ID:             gid.New(fx.scope.GetTenantID(), coredata.AccessInventoryProviderSyncConfigEntityType),
			OrganizationID: fx.organizationID,
			ProviderKey:    coredata.AccessInventoryProviderKeyCSV,
			SyncCadence:    timespan.TimeSpan{Months: 1},
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		_, err := config.Upsert(ctx, tx, fx.scope)

		return err
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "representable as a fixed clock interval")
}

func TestApplyInventorySyncJitter_CapsFailureBackoff(t *testing.T) {
	t.Parallel()

	base := inventorySyncFailureMaxBackoff
	// Force +20% jitter so uncapped delay would exceed the max.
	jittered := applyInventorySyncJitter(base, func(n int64) int64 {
		return n - 1
	})
	assert.Greater(t, jittered, inventorySyncFailureMaxBackoff)

	delay := min(jittered, inventorySyncFailureMaxBackoff)

	assert.Equal(t, inventorySyncFailureMaxBackoff, delay)
}

func TestUpdateSource_EnqueuesSyncOnCSVUpload(t *testing.T) {
	// Not parallel: leaves a cross-tenant QUEUED run until cleanup.
	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\nalice@example.com,Alice,Admin\n")
	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
	csvV2 := "email,full_name,role\nbob@example.com,Bob,Member\n"

	updated, err := svc.UpdateSource(
		ctx,
		fx.scope,
		UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: fx.sourceID,
			CsvData:              new(&csvV2),
		},
	)
	require.NoError(t, err)
	require.NotNil(t, updated.CsvData)
	assert.Equal(t, csvV2, *updated.CsvData)

	first := &coredata.AccessSourceSyncRun{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return first.LoadInFlightBySourceID(ctx, conn, fx.scope, fx.sourceID)
	}))
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, first.Status)

	csvV3 := "email,full_name,role\ncarol@example.com,Carol,Admin\n"
	_, err = svc.UpdateSource(
		ctx,
		fx.scope,
		UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: fx.sourceID,
			CsvData:              new(&csvV3),
		},
	)
	require.NoError(t, err)

	runs := loadInventorySyncRunsBySource(t, ctx, client, fx)
	require.Len(t, runs, 2)
	assert.Equal(t, first.ID, runs[0].ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, runs[0].Status)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, runs[1].Status)
	assert.NotEqual(t, runs[0].ID, runs[1].ID, "CSV re-upload must enqueue a fresh QUEUED run")

	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
}

func TestUpdateSource_EnqueuesQueuedWhileRunning(t *testing.T) {
	// Not parallel: leaves a cross-tenant QUEUED run until cleanup.
	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\nalice@example.com,Alice,Admin\n")
	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))

	running := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, running)

	csvV2 := "email,full_name,role\nbob@example.com,Bob,Member\n"
	_, err := svc.UpdateSource(
		ctx,
		fx.scope,
		UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: fx.sourceID,
			CsvData:              new(&csvV2),
		},
	)
	require.NoError(t, err)

	runs := loadInventorySyncRunsBySource(t, ctx, client, fx)
	require.Len(t, runs, 2)
	assert.Equal(t, running.ID, runs[0].ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusRunning, runs[0].Status)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, runs[1].Status)

	// Claim must wait on RUNNING: the follow-up QUEUED stays unclaimed.
	next := &coredata.AccessSourceSyncRun{}
	err = client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return next.LoadNextQueuedForUpdateSkipLocked(ctx, tx)
	})
	if err == nil {
		assert.NotEqual(t, runs[1].ID, next.ID, "follow-up QUEUED must wait while RUNNING")
		assert.NotEqual(t, fx.sourceID, next.AccessReviewSourceID)
	} else {
		require.ErrorIs(t, err, coredata.ErrNoAccessSourceSyncRunAvailable)
	}

	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
}

func TestUpdateSource_NameOnlyDoesNotEnqueueSync(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\ncarol@example.com,Carol,Admin\n")
	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
	name := "Renamed CSV Roster"

	_, err := svc.UpdateSource(
		ctx,
		fx.scope,
		UpdateAccessReviewSourceRequest{
			AccessReviewSourceID: fx.sourceID,
			Name:                 new(&name),
		},
	)
	require.NoError(t, err)

	inflight := &coredata.AccessSourceSyncRun{}
	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return inflight.LoadInFlightBySourceID(ctx, conn, fx.scope, fx.sourceID)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func TestSyncNow_EnqueuesProboMembershipsSource(t *testing.T) {
	// Not parallel: leaves a cross-tenant QUEUED run until cleanup.
	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedMembershipsInventorySyncFixture(t, ctx, client)
	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))

	run, err := svc.SyncNow(ctx, fx.scope, fx.sourceID)
	require.NoError(t, err)
	require.NotNil(t, run)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, run.Status)

	var cadence time.Duration
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		source := &coredata.AccessReviewSource{}
		if err := source.LoadByID(ctx, conn, fx.scope, fx.sourceID); err != nil {
			return err
		}

		var err error
		cadence, err = svc.resolveSourceSyncCadence(ctx, conn, fx.scope, source)

		return err
	}))
	assert.Equal(t, time.Hour, cadence)

	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
}

func TestInventorySyncScheduler_EnqueuesDueMembershipsSource(t *testing.T) {
	// Not parallel: shares the cross-tenant due-source claim queue.
	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedMembershipsInventorySyncFixture(t, ctx, client)
	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
	require.NoError(t, holdDueInventorySyncSchedulesExcept(t, ctx, client, fx.sourceID))

	past := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		source := &coredata.AccessReviewSource{}
		if err := source.LoadByID(ctx, tx, fx.scope, fx.sourceID); err != nil {
			return err
		}

		source.NextSyncAt = &past
		source.UpdatedAt = time.Now().UTC()

		return source.UpdateInventorySyncSchedule(ctx, tx, fx.scope)
	}))

	h := &inventorySyncSchedulerHandler{svc: svc, pg: client, logger: logger}
	claimed, err := h.Claim(ctx)
	require.NoError(t, err)
	assert.Equal(t, fx.sourceID, claimed.ID)

	inflight := &coredata.AccessSourceSyncRun{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return inflight.LoadInFlightBySourceID(ctx, conn, fx.scope, fx.sourceID)
	}))
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, inflight.Status)

	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
}

func TestResolveInventoryProviderKey_EmptyCSVIsMemberships(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	empty := ""
	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name,role\n")
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		source := &coredata.AccessReviewSource{}
		if err := source.LoadByIDForUpdate(ctx, tx, fx.scope, fx.sourceID); err != nil {
			return err
		}

		source.CsvData = &empty
		source.UpdatedAt = time.Now().UTC()

		return source.Update(ctx, tx, fx.scope)
	}))

	source := &coredata.AccessReviewSource{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return source.LoadByID(ctx, conn, fx.scope, fx.sourceID)
	}))

	var key coredata.AccessInventoryProviderKey
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var err error
		key, err = svc.resolveInventoryProviderKey(ctx, conn, fx.scope, source)

		return err
	}))
	assert.Equal(t, coredata.AccessInventoryProviderKeyProboMemberships, key)
}

func seedMembershipsInventorySyncFixture(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
) inventorySyncFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	fx := inventorySyncFixture{
		scope:          coredata.NewScope(tenantID, gid.Nil),
		organizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		sourceID:       gid.New(tenantID, coredata.AccessReviewSourceEntityType),
	}
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        fx.organizationID,
			TenantID:  tenantID,
			Name:      "Memberships Sync",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		nextSyncAt := now.Add(24 * time.Hour)
		source := &coredata.AccessReviewSource{
			ID:             fx.sourceID,
			OrganizationID: fx.organizationID,
			Name:           "Probo Memberships",
			NextSyncAt:     &nextSyncAt,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		_, err := source.Insert(ctx, tx, fx.scope)

		return err
	}))

	return fx
}

func loadInventorySyncRunsBySource(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx inventorySyncFixture,
) []*coredata.AccessSourceSyncRun {
	t.Helper()

	var runs coredata.AccessSourceSyncRuns
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			100,
			nil,
			page.Head,
			page.OrderBy[coredata.AccessSourceSyncRunOrderField]{
				Field:     coredata.AccessSourceSyncRunOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		return runs.LoadBySourceID(ctx, conn, fx.scope, fx.sourceID, cursor)
	}))

	return runs
}
