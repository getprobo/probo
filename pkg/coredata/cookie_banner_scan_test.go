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
)

func newPendingCookieBannerScan(fx trackerPatternFixture) *coredata.CookieBannerScan {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return &coredata.CookieBannerScan{
		ID:             gid.New(fx.scope.GetTenantID(), coredata.CookieBannerScanEntityType),
		OrganizationID: fx.organizationID,
		CookieBannerID: fx.cookieBannerID,
		State:          coredata.CookieBannerScanStatePending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func insertCookieBannerScan(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx trackerPatternFixture,
	scan *coredata.CookieBannerScan,
) error {
	t.Helper()

	return client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return scan.Insert(ctx, tx, fx.scope)
	})
}

func TestCookieBannerScan_Insert_OneActivePerBanner(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	first := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, first))

	second := newPendingCookieBannerScan(fx)
	err := insertCookieBannerScan(t, ctx, client, fx, second)
	assert.ErrorIs(t, err, coredata.ErrResourceAlreadyExists, "a second scan must be rejected while one is pending")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := first.MarkRunning(ctx, tx, fx.scope, time.Now()); err != nil {
			return err
		}

		first.PagesScanned = 3
		first.CookiesFound = 7

		return first.MarkCompleted(ctx, tx, fx.scope, time.Now())
	}))

	third := newPendingCookieBannerScan(fx)
	assert.NoError(t, insertCookieBannerScan(t, ctx, client, fx, third), "a new scan must be accepted once the previous one completed")

	var latest coredata.CookieBannerScan
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return latest.LoadLatestByCookieBannerID(ctx, conn, fx.scope, fx.cookieBannerID)
	}))
	assert.Equal(t, third.ID, latest.ID)

	var completed coredata.CookieBannerScan
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return completed.LoadByID(ctx, conn, fx.scope, first.ID)
	}))
	assert.Equal(t, coredata.CookieBannerScanStateCompleted, completed.State)
	assert.Equal(t, 3, completed.PagesScanned)
	assert.Equal(t, 7, completed.CookiesFound)
	assert.NotNil(t, completed.StartedAt)
	assert.NotNil(t, completed.CompletedAt)
}

func TestCookieBannerScan_MarkCompleted_RejectsStaleFailedScan(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	startedAt := time.Now().Add(-time.Hour)
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return scan.MarkRunning(ctx, tx, fx.scope, startedAt)
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return coredata.FailStaleCookieBannerScans(ctx, conn, 20*time.Minute)
	}))

	var reloaded coredata.CookieBannerScan
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, scan.ID)
	}))
	assert.Equal(t, coredata.CookieBannerScanStateFailed, reloaded.State)
	require.NotNil(t, reloaded.FailureReason)
	assert.Equal(t, coredata.CookieBannerScanFailureReasonInterrupted, *reloaded.FailureReason)

	err := client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return scan.MarkCompleted(ctx, tx, fx.scope, time.Now())
	})
	assert.ErrorIs(t, err, coredata.ErrResourceNotFound, "a late worker must not complete a scan stale recovery already failed")
}

func TestFailStaleCookieBannerScans_KeepsRecentRunningScan(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return scan.MarkRunning(ctx, tx, fx.scope, time.Now())
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return coredata.FailStaleCookieBannerScans(ctx, conn, 20*time.Minute)
	}))

	var reloaded coredata.CookieBannerScan
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, scan.ID)
	}))
	assert.Equal(t, coredata.CookieBannerScanStateRunning, reloaded.State)
	assert.Nil(t, reloaded.FailureReason)
}
