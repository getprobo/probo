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
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/page"
)

func TestFetchSource_DualWritesInventory(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvData := "email,full_name,role\nalice@example.com,Alice,Admin\n"
	fx := seedFetchFixture(t, ctx, client, csvData)

	campaign, err := svc.GetCampaign(ctx, fx.scope, fx.campaignID)
	require.NoError(t, err)

	campaignSource := &coredata.AccessReviewCampaignSource{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return campaignSource.LoadByID(ctx, conn, fx.scope, fx.campaignSourceID)
	}))

	count, err := svc.FetchSource(ctx, fx.scope, campaign, campaignSource)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	var entries coredata.AccessReviewEntries
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var err error
		count, err = entries.CountByCampaignID(
			ctx,
			conn,
			fx.scope,
			fx.campaignID,
			&coredata.AccessReviewEntryFilter{},
		)

		return err
	}))
	assert.Equal(t, 1, count)

	inventoryFX := inventorySyncFixture{
		scope:          fx.scope,
		organizationID: fx.organizationID,
		sourceID:       fx.sourceID,
	}
	grants := loadInventoryGrants(t, ctx, client, inventoryFX)
	require.Len(t, grants, 1)
	assert.Equal(t, "alice@example.com", grants[0].PrincipalKey)
	assert.Equal(t, DefaultResourceKey, grants[0].ResourceKey)
	assert.Equal(t, "Admin", grants[0].EntitlementKey)
	assert.Equal(t, coredata.AccessInventoryGrantPresencePresent, grants[0].Presence)

	run, err := svc.GetLatestAccessSourceSyncRunForSourceID(ctx, fx.scope, fx.sourceID)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusSuccess, run.Status)
	assert.Equal(t, 1, run.GrantsUpsertedCount)

	events := loadInventoryChangeEvents(t, ctx, client, inventoryFX)
	require.Len(t, events, 1)
	assert.Equal(t, coredata.AccessInventoryChangeTypeAppeared, events[0].ChangeType)
	require.NotNil(t, events[0].SyncRunID)
	assert.Equal(t, run.ID, *events[0].SyncRunID)
}

func TestFetchSource_SkipsDualWriteWhenSyncInFlight(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvData := "email,full_name,role\nbob@example.com,Bob,Member\n"
	fx := seedFetchFixture(t, ctx, client, csvData)

	inventoryFX := inventorySyncFixture{
		scope:          fx.scope,
		organizationID: fx.organizationID,
		sourceID:       fx.sourceID,
	}
	queued := insertQueuedInventorySyncRun(t, ctx, client, inventoryFX)

	campaign, err := svc.GetCampaign(ctx, fx.scope, fx.campaignID)
	require.NoError(t, err)

	campaignSource := &coredata.AccessReviewCampaignSource{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return campaignSource.LoadByID(ctx, conn, fx.scope, fx.campaignSourceID)
	}))

	count, err := svc.FetchSource(ctx, fx.scope, campaign, campaignSource)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	assert.Empty(t, loadInventoryGrants(t, ctx, client, inventoryFX))

	run, err := svc.GetLatestAccessSourceSyncRunForSourceID(ctx, fx.scope, fx.sourceID)
	require.NoError(t, err)
	assert.Equal(t, queued.ID, run.ID)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, run.Status)
}

func TestStartCampaign_EnqueuesInventorySync(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvData := "email,full_name,role\ncarol@example.com,Carol,Viewer\n"
	fx := seedFetchFixture(t, ctx, client, csvData)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		campaign := &coredata.AccessReviewCampaign{}
		if err := campaign.LoadByID(ctx, tx, fx.scope, fx.campaignID); err != nil {
			return err
		}

		campaign.Status = coredata.AccessReviewCampaignStatusDraft
		campaign.StartedAt = nil
		campaign.UpdatedAt = campaign.CreatedAt

		return campaign.Update(ctx, tx, fx.scope)
	}))

	started, err := svc.StartCampaign(ctx, fx.scope, fx.campaignID)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessReviewCampaignStatusInProgress, started.Status)

	run, err := svc.GetLatestAccessSourceSyncRunForSourceID(ctx, fx.scope, fx.sourceID)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessSourceSyncRunStatusQueued, run.Status)
	assert.Equal(t, fx.sourceID, run.AccessReviewSourceID)

	var attempts coredata.AccessReviewCampaignSourceFetchAttempts
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			10,
			nil,
			page.Head,
			page.OrderBy[coredata.AccessReviewCampaignSourceFetchAttemptOrderField]{
				Field:     coredata.AccessReviewCampaignSourceFetchAttemptOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		return attempts.LoadByCampaignSourceID(ctx, conn, fx.scope, fx.campaignSourceID, cursor)
	}))
	require.Len(t, attempts, 1)
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusQueued, attempts[0].Status)
}

func TestStartCampaign_SkipsInventoryEnqueueWhenSourceDeleted(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvData := "email,full_name,role\neve@example.com,Eve,Viewer\n"
	fx := seedFetchFixture(t, ctx, client, csvData)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		campaign := &coredata.AccessReviewCampaign{}
		if err := campaign.LoadByID(ctx, tx, fx.scope, fx.campaignID); err != nil {
			return err
		}

		campaign.Status = coredata.AccessReviewCampaignStatusDraft
		campaign.StartedAt = nil
		campaign.UpdatedAt = campaign.CreatedAt

		if err := campaign.Update(ctx, tx, fx.scope); err != nil {
			return err
		}

		source := &coredata.AccessReviewSource{ID: fx.sourceID}

		return source.Delete(ctx, tx, fx.scope)
	}))

	started, err := svc.StartCampaign(ctx, fx.scope, fx.campaignID)
	require.NoError(t, err)
	assert.Equal(t, coredata.AccessReviewCampaignStatusInProgress, started.Status)

	_, err = svc.GetLatestAccessSourceSyncRunForSourceID(ctx, fx.scope, fx.sourceID)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	var attempts coredata.AccessReviewCampaignSourceFetchAttempts
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			10,
			nil,
			page.Head,
			page.OrderBy[coredata.AccessReviewCampaignSourceFetchAttemptOrderField]{
				Field:     coredata.AccessReviewCampaignSourceFetchAttemptOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		return attempts.LoadByCampaignSourceID(ctx, conn, fx.scope, fx.campaignSourceID, cursor)
	}))
	require.Len(t, attempts, 1)
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusQueued, attempts[0].Status)
}

func TestDualWriteInventory_SkipsWhenSourceDeleted(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	fx := seedFetchFixture(t, ctx, client, fetchFixtureCSV)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		source := &coredata.AccessReviewSource{ID: fx.sourceID}

		return source.Delete(ctx, tx, fx.scope)
	}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return svc.dualWriteInventoryFromCampaignFetch(
			ctx,
			tx,
			fx.scope,
			fx.sourceID,
			[]drivers.AccountRecord{{Email: "alice@example.com", Roles: []string{"Admin"}}},
			time.Now().UTC(),
		)
	}))

	inventoryFX := inventorySyncFixture{
		scope:          fx.scope,
		organizationID: fx.organizationID,
		sourceID:       fx.sourceID,
	}
	assert.Empty(t, loadInventoryGrants(t, ctx, client, inventoryFX))
}

func TestDualWriteInventory_SkipsStaleRoster(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := log.NewLogger(log.WithOutput(io.Discard))
	svc := NewService(client, cipher.EncryptionKey{}, nil, provider.NewBuiltinRegistry(), logger)

	csvData := "email,full_name,role\ndave@example.com,Dave,Admin\n"
	fx := seedFetchFixture(t, ctx, client, csvData)

	observedAt := time.Now().UTC().Add(-time.Minute)
	newerSyncAt := observedAt.Add(30 * time.Second)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		source := &coredata.AccessReviewSource{}
		if err := source.LoadByID(ctx, tx, fx.scope, fx.sourceID); err != nil {
			return err
		}

		source.LastSuccessfulSyncAt = &newerSyncAt
		source.UpdatedAt = newerSyncAt

		return source.UpdateInventorySyncSchedule(ctx, tx, fx.scope)
	}))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return svc.dualWriteInventoryFromCampaignFetch(
			ctx,
			tx,
			fx.scope,
			fx.sourceID,
			[]drivers.AccountRecord{{Email: "dave@example.com", Roles: []string{"Admin"}}},
			observedAt,
		)
	}))

	inventoryFX := inventorySyncFixture{
		scope:          fx.scope,
		organizationID: fx.organizationID,
		sourceID:       fx.sourceID,
	}
	assert.Empty(t, loadInventoryGrants(t, ctx, client, inventoryFX))

	_, err := svc.GetLatestAccessSourceSyncRunForSourceID(ctx, fx.scope, fx.sourceID)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}
