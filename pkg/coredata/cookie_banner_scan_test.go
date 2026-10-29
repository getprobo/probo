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
	"fmt"
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
		return first.MarkRunning(ctx, tx, fx.scope, time.Now())
	}))

	whileRunning := newPendingCookieBannerScan(fx)
	err = insertCookieBannerScan(t, ctx, client, fx, whileRunning)
	assert.ErrorIs(t, err, coredata.ErrResourceAlreadyExists, "a second scan must be rejected while one is running")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
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

func TestCookieBannerScan_Pause_BlocksAnotherScanAndKeepsStartedAt(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	startedAt := time.Now().Add(-time.Minute).UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return scan.MarkRunning(ctx, tx, fx.scope, startedAt)
	}))

	checkpoint := json.RawMessage(`{"visited_templates":["/"],"queue":[],"pages_scanned":1,"elapsed_ms":1000,"truncated":false}`)
	scan.Checkpoint = &checkpoint
	scan.PagesScanned = 1

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return scan.MarkPaused(ctx, tx, fx.scope, time.Now())
	}))

	second := newPendingCookieBannerScan(fx)
	err := insertCookieBannerScan(t, ctx, client, fx, second)
	assert.ErrorIs(t, err, coredata.ErrResourceAlreadyExists, "a second scan must be rejected while one is paused")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return scan.MarkRunning(ctx, tx, fx.scope, time.Now())
	}))

	var resumed coredata.CookieBannerScan

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return resumed.LoadByID(ctx, conn, fx.scope, scan.ID)
	}))
	assert.Equal(t, coredata.CookieBannerScanStateRunning, resumed.State)
	require.NotNil(t, resumed.StartedAt)
	assert.True(t, resumed.StartedAt.Equal(startedAt), "resuming must keep the original start")
	assert.Equal(t, 1, resumed.PagesScanned)
	require.NotNil(t, resumed.Checkpoint)
	assert.JSONEq(t, string(checkpoint), string(*resumed.Checkpoint))
}

func TestFailStaleCookieBannerScans_PausesCheckpointedRun(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := scan.MarkRunning(ctx, tx, fx.scope, time.Now()); err != nil {
			return err
		}

		checkpoint := json.RawMessage(`{"visited_templates":["/pricing"],"queue":[{"url":"https://example.com/about","depth":1}],"pages_scanned":1,"elapsed_ms":5000,"truncated":false}`)
		scan.Checkpoint = &checkpoint

		return scan.SaveCheckpoint(ctx, tx, fx.scope, time.Now().Add(-time.Hour))
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return coredata.FailStaleCookieBannerScans(ctx, conn, 20*time.Minute)
	}))

	var reloaded coredata.CookieBannerScan

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, scan.ID)
	}))
	assert.Equal(t, coredata.CookieBannerScanStatePaused, reloaded.State)
	assert.Nil(t, reloaded.FailureReason)
	assert.Nil(t, reloaded.CompletedAt)
	require.NotNil(t, reloaded.Checkpoint)
	assert.JSONEq(t, string(*scan.Checkpoint), string(*reloaded.Checkpoint))
}

func TestFailStaleCookieBannerScans_KeepsPausedScan(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := scan.MarkRunning(ctx, tx, fx.scope, time.Now()); err != nil {
			return err
		}

		checkpoint := json.RawMessage(`{"visited_templates":["/"],"queue":[],"pages_scanned":1,"elapsed_ms":1000,"truncated":false}`)
		scan.Checkpoint = &checkpoint

		return scan.MarkPaused(ctx, tx, fx.scope, time.Now().Add(-time.Hour))
	}))

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return coredata.FailStaleCookieBannerScans(ctx, conn, 20*time.Minute)
	}))

	var reloaded coredata.CookieBannerScan

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, scan.ID)
	}))
	assert.Equal(t, coredata.CookieBannerScanStatePaused, reloaded.State)
	assert.Nil(t, reloaded.FailureReason)
	assert.Nil(t, reloaded.CompletedAt)
}

func TestCookieBannerScan_Touch_RefreshesOnlyRunning(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return scan.Touch(ctx, conn, fx.scope, time.Now())
	})
	assert.ErrorIs(t, err, coredata.ErrResourceNotFound, "a pending scan has no heartbeat")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return scan.MarkRunning(ctx, tx, fx.scope, time.Now())
	}))

	touchedAt := time.Now().Add(time.Minute).UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return scan.Touch(ctx, conn, fx.scope, touchedAt)
	}))

	var reloaded coredata.CookieBannerScan

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, scan.ID)
	}))
	assert.True(t, reloaded.UpdatedAt.Equal(touchedAt))
	assert.Equal(t, coredata.CookieBannerScanStateRunning, reloaded.State)

	checkpoint := json.RawMessage(`{"visited_templates":["/"],"queue":[],"pages_scanned":1,"elapsed_ms":1000,"truncated":false}`)
	scan.Checkpoint = &checkpoint

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return scan.MarkPaused(ctx, tx, fx.scope, time.Now())
	}))

	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return scan.Touch(ctx, conn, fx.scope, time.Now())
	})
	assert.ErrorIs(t, err, coredata.ErrResourceNotFound, "a paused scan must not take a heartbeat")
}

func TestCountRunningCookieBannerScans_CountsOnlyRunningAcrossTenants(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	first := seedTrackerPatternFixture(t, ctx, client)
	second := seedTrackerPatternFixture(t, ctx, client)

	firstScan := newPendingCookieBannerScan(first)
	secondScan := newPendingCookieBannerScan(second)

	// Repeatable read freezes the baseline, so a scan another test
	// commits while this one runs cannot change the deltas.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if _, err := tx.Exec(ctx, `SET TRANSACTION ISOLATION LEVEL REPEATABLE READ`); err != nil {
			return err
		}

		before, err := coredata.CountRunningCookieBannerScans(ctx, tx)
		if err != nil {
			return err
		}

		if err := firstScan.Insert(ctx, tx, first.scope); err != nil {
			return err
		}

		if err := secondScan.Insert(ctx, tx, second.scope); err != nil {
			return err
		}

		pending, err := coredata.CountRunningCookieBannerScans(ctx, tx)
		if err != nil {
			return err
		}

		if pending != before {
			return fmt.Errorf("pending scans counted: before %d, after %d", before, pending)
		}

		if err := firstScan.MarkRunning(ctx, tx, first.scope, time.Now()); err != nil {
			return err
		}

		oneRunning, err := coredata.CountRunningCookieBannerScans(ctx, tx)
		if err != nil {
			return err
		}

		if oneRunning != before+1 {
			return fmt.Errorf("one running scan counted as %d, baseline %d", oneRunning, before)
		}

		if err := secondScan.MarkRunning(ctx, tx, second.scope, time.Now()); err != nil {
			return err
		}

		twoRunning, err := coredata.CountRunningCookieBannerScans(ctx, tx)
		if err != nil {
			return err
		}

		if twoRunning != before+2 {
			return fmt.Errorf("running scans on two tenants counted as %d, baseline %d", twoRunning, before)
		}

		secondScan.PagesScanned = 1

		if err := secondScan.MarkCompleted(ctx, tx, second.scope, time.Now()); err != nil {
			return err
		}

		checkpoint := json.RawMessage(`{"visited_templates":["/"],"queue":[],"pages_scanned":1,"elapsed_ms":1,"truncated":false}`)
		firstScan.Checkpoint = &checkpoint

		if err := firstScan.MarkPaused(ctx, tx, first.scope, time.Now()); err != nil {
			return err
		}

		settled, err := coredata.CountRunningCookieBannerScans(ctx, tx)
		if err != nil {
			return err
		}

		if settled != before {
			return fmt.Errorf("paused and completed scans still counted: before %d, after %d", before, settled)
		}

		return nil
	}))
}
