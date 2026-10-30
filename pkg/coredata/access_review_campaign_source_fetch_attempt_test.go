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
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func insertFetchAttempt(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx accessEntryFixture,
	status coredata.AccessReviewCampaignSourceFetchStatus,
	updatedAt time.Time,
) *coredata.AccessReviewCampaignSourceFetchAttempt {
	t.Helper()

	attempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
		ID:                           gid.New(fx.scope.GetTenantID(), coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
		OrganizationID:               fx.organizationID,
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Status:                       status,
		CreatedAt:                    updatedAt,
		UpdatedAt:                    updatedAt,
	}

	err := client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return attempt.Insert(ctx, tx, fx.scope)
		},
	)
	require.NoError(t, err)

	return attempt
}

func loadFetchAttempts(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx accessEntryFixture,
) coredata.AccessReviewCampaignSourceFetchAttempts {
	t.Helper()

	var attempts coredata.AccessReviewCampaignSourceFetchAttempts

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.AccessReviewCampaignSourceFetchAttemptOrderField]{
			Field:     coredata.AccessReviewCampaignSourceFetchAttemptOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	err := client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return attempts.LoadByCampaignSourceID(ctx, conn, fx.scope, fx.campaignSourceID, cursor)
		},
	)
	require.NoError(t, err)

	return attempts
}

func TestAccessReviewCampaignSourceFetchAttempts_FailQueuedByCampaignID(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessReviewEntryFixture(t, ctx, client)
	other := seedAccessReviewEntryFixture(t, ctx, client)
	now := time.Now().UTC()

	insertFetchAttempt(t, ctx, client, fx, coredata.AccessReviewCampaignSourceFetchStatusFetching, now.Add(-time.Second))
	insertFetchAttempt(t, ctx, client, fx, coredata.AccessReviewCampaignSourceFetchStatusQueued, now)
	insertFetchAttempt(t, ctx, client, other, coredata.AccessReviewCampaignSourceFetchStatusQueued, now)

	err := client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var attempts coredata.AccessReviewCampaignSourceFetchAttempts

			return attempts.FailQueuedByCampaignID(ctx, tx, fx.scope, fx.campaignID, "cancelled", now)
		},
	)
	require.NoError(t, err)

	attempts := loadFetchAttempts(t, ctx, client, fx)
	require.Len(t, attempts, 2)
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusFetching, attempts[0].Status)
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusFailed, attempts[1].Status)
	require.NotNil(t, attempts[1].Error)
	assert.Equal(t, "cancelled", *attempts[1].Error)
	assert.NotNil(t, attempts[1].CompletedAt)

	otherAttempts := loadFetchAttempts(t, ctx, client, other)
	require.Len(t, otherAttempts, 1)
	assert.Equal(t, coredata.AccessReviewCampaignSourceFetchStatusQueued, otherAttempts[0].Status)
}

func campaignIDs(campaigns coredata.AccessReviewCampaigns) []gid.GID {
	ids := make([]gid.GID, 0, len(campaigns))
	for _, campaign := range campaigns {
		ids = append(ids, campaign.ID)
	}

	return ids
}

func TestAccessReviewCampaigns_LoadWithStaleFetchAttemptsForUpdateSkipLocked(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessReviewEntryFixture(t, ctx, client)
	// The old date keeps this sweep away from other tests' rows.
	staleAt := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	staleThreshold := staleAt.Add(time.Hour)

	var (
		locked    = make(chan error, 1)
		releaseCh = make(chan struct{})
		release   = sync.OnceFunc(func() { close(releaseCh) })
		done      = make(chan error, 1)
		ownLocked coredata.AccessReviewCampaigns
	)

	// Holds the campaign lock the way DeleteCampaign does before its cascade.
	go func() {
		done <- client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				campaign := &coredata.AccessReviewCampaign{ID: fx.campaignID}
				if err := campaign.LockForUpdate(ctx, tx, fx.scope); err != nil {
					locked <- err

					return err
				}

				locked <- nil

				<-releaseCh

				return ownLocked.LoadWithStaleFetchAttemptsForUpdateSkipLocked(ctx, tx, staleThreshold, 100)
			},
		)
	}()

	require.NoError(t, <-locked)
	t.Cleanup(release)

	// Inserted under the lock so no concurrent sweep can take it first.
	insertFetchAttempt(t, ctx, client, fx, coredata.AccessReviewCampaignSourceFetchStatusFetching, staleAt)

	sweepCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var swept coredata.AccessReviewCampaigns

	err := client.WithTx(
		sweepCtx,
		func(ctx context.Context, tx pg.Tx) error {
			return swept.LoadWithStaleFetchAttemptsForUpdateSkipLocked(ctx, tx, staleThreshold, 100)
		},
	)

	release()
	require.NoError(t, <-done)
	require.NoError(t, err, "the sweep must skip a locked campaign instead of waiting on it")

	assert.NotContains(t, campaignIDs(swept), fx.campaignID, "a locked campaign must be left for the next sweep")
	assert.Contains(t, campaignIDs(ownLocked), fx.campaignID, "the lock holder still sees its stale campaign")
}

func TestAccessReviewCampaigns_LoadWithStaleFetchAttemptsForUpdateSkipLocked_Limit(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	// Older than every other stale test and never committed, so no sweep but this one sees these rows.
	staleAt := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	fixtures := []accessEntryFixture{
		seedAccessReviewEntryFixture(t, ctx, client),
		seedAccessReviewEntryFixture(t, ctx, client),
	}
	errRollback := errors.New("rollback")

	var swept coredata.AccessReviewCampaigns

	err := client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, fx := range fixtures {
				attempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
					ID:                           gid.New(fx.scope.GetTenantID(), coredata.AccessReviewCampaignSourceFetchAttemptEntityType),
					OrganizationID:               fx.organizationID,
					AccessReviewCampaignSourceID: fx.campaignSourceID,
					Status:                       coredata.AccessReviewCampaignSourceFetchStatusFetching,
					CreatedAt:                    staleAt,
					UpdatedAt:                    staleAt,
				}
				if err := attempt.Insert(ctx, tx, fx.scope); err != nil {
					return err
				}
			}

			if err := swept.LoadWithStaleFetchAttemptsForUpdateSkipLocked(ctx, tx, staleAt.Add(time.Hour), 1); err != nil {
				return err
			}

			return errRollback
		},
	)
	require.ErrorIs(t, err, errRollback)

	assert.Len(t, swept, 1, "the sweep must lock at most limit campaigns")
}
