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

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

func drainInFlightInventorySyncRunsForSource(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	sourceID gid.GID,
) error {
	t.Helper()

	return client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`
DELETE FROM access_source_sync_runs
WHERE access_review_source_id = @access_review_source_id
  AND status IN ('QUEUED', 'RUNNING')
`,
			pgx.NamedArgs{"access_review_source_id": sourceID},
		)

		return err
	})
}

// holdDueInventorySyncSchedulesExcept parks every other due source so the
// unscoped scheduler claim can see only exceptSourceID. Shared CI DBs may
// still carry NULL/past next_sync_at from parallel suites.
func holdDueInventorySyncSchedulesExcept(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	exceptSourceID gid.GID,
) error {
	t.Helper()

	return client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`
UPDATE access_review_sources
SET next_sync_at = NOW() + INTERVAL '30 days'
WHERE id <> @except_source_id
  AND (next_sync_at IS NULL OR next_sync_at <= NOW())
`,
			pgx.NamedArgs{"except_source_id": exceptSourceID},
		)

		return err
	})
}

func TestInventorySyncHandler_ClaimAndProcess(t *testing.T) {
	// Not parallel: shares the cross-tenant QUEUED claim queue.
	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvData := "email,full_name,role\ndave@example.com,Dave,Viewer\n"
	fx := seedInventorySyncFixture(t, ctx, client, csvData)
	require.NoError(t, drainInFlightInventorySyncRunsForSource(t, ctx, client, fx.sourceID))
	// Far-past created_at wins ORDER BY against leftover QUEUED rows.
	queued := insertQueuedInventorySyncRunAt(t, ctx, client, fx, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))

	h := &inventorySyncHandler{
		svc:    svc,
		pg:     client,
		logger: logger,
	}

	claimed, err := h.Claim(ctx)
	require.NoError(t, err)
	assert.Equal(t, queued.ID, claimed.ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusRunning, claimed.Status)
	require.NotNil(t, claimed.StartedAt)

	require.NoError(t, h.Process(ctx, claimed))

	loaded := loadSyncRun(t, ctx, client, fx, queued.ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusSuccess, loaded.Status)
	assert.Equal(t, 1, loaded.GrantsUpsertedCount)

	grants := loadInventoryGrants(t, ctx, client, fx)
	require.Len(t, grants, 1)
	assert.Equal(t, "dave@example.com", grants[0].PrincipalKey)
	assert.Equal(t, "Viewer", grants[0].EntitlementKey)

	events := loadInventoryChangeEvents(t, ctx, client, fx)
	require.Len(t, events, 1)
	assert.Equal(t, coredata.AccessInventoryChangeTypeAppeared, events[0].ChangeType)
}

func TestInventorySyncHandler_ProcessFailedDriver(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	// Missing email column makes CSV ListAccounts fail.
	fx := seedInventorySyncFixture(t, ctx, client, "full_name\nAlice\n")
	run := insertQueuedInventorySyncRun(t, ctx, client, fx)
	markSyncRunRunning(t, ctx, client, fx, run)

	h := &inventorySyncHandler{
		svc:    svc,
		pg:     client,
		logger: logger,
	}

	require.NoError(t, h.handle(ctx, run))

	loaded := loadSyncRun(t, ctx, client, fx, run.ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusFailed, loaded.Status)
	require.NotNil(t, loaded.Error)
	assert.Equal(t, sourceFetchFailureMessage, *loaded.Error)
}

func TestInventorySyncHandler_RecoverStale(t *testing.T) {
	// Not parallel: shares the cross-tenant stale RUNNING sweep.
	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	staleAt := time.Date(2011, time.January, 1, 0, 0, 0, 0, time.UTC)
	fx := seedInventorySyncFixture(t, ctx, client, "email,full_name\neve@example.com,Eve\n")
	run := insertQueuedInventorySyncRun(t, ctx, client, fx)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		run.Status = coredata.AccessSourceSyncRunStatusRunning
		run.StartedAt = &staleAt
		run.UpdatedAt = staleAt

		return run.Update(ctx, tx, fx.scope)
	}))

	h := &inventorySyncHandler{
		svc:        svc,
		pg:         client,
		logger:     logger,
		staleAfter: time.Since(staleAt.Add(time.Hour)),
	}

	before := time.Now().UTC()
	require.NoError(t, h.RecoverStale(ctx))

	loaded := loadSyncRun(t, ctx, client, fx, run.ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusFailed, loaded.Status)
	require.NotNil(t, loaded.Error)
	assert.Equal(t, staleInventorySyncMessage, *loaded.Error)

	source := &coredata.AccessReviewSource{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return source.LoadByID(ctx, conn, fx.scope, fx.sourceID)
	}))
	assert.Equal(t, 1, source.SyncConsecutiveFailures)
	require.NotNil(t, source.NextSyncAt)
	assert.True(t, source.NextSyncAt.After(before))
}
