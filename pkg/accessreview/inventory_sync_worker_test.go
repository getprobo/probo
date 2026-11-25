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
)

func TestInventorySyncHandler_ClaimAndProcess(t *testing.T) {
	// Not parallel: shares the cross-tenant QUEUED claim queue.
	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvData := "email,full_name,role\ndave@example.com,Dave,Viewer\n"
	fx := seedInventorySyncFixture(t, ctx, client, csvData)
	queued := insertQueuedInventorySyncRun(t, ctx, client, fx)

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
		pg:         client,
		logger:     logger,
		staleAfter: time.Since(staleAt.Add(time.Hour)),
	}

	require.NoError(t, h.RecoverStale(ctx))

	loaded := loadSyncRun(t, ctx, client, fx, run.ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusFailed, loaded.Status)
	require.NotNil(t, loaded.Error)
	assert.Equal(t, staleInventorySyncMessage, *loaded.Error)
}
