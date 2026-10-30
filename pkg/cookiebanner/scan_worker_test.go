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

package cookiebanner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/cookiebanner/scanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/netcheck"
)

func newScanWorker(t *testing.T, client *pg.Client) *ScanWorker {
	t.Helper()

	return NewCookieBannerScanWorker(
		NewService(client, false, 0),
		client,
		"localhost:9222",
		"https://api.example.com/api/cookie-banner/v1",
		log.NewLogger(log.WithOutput(io.Discard)),
	)
}

func insertRunningScan(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
) *coredata.CookieBannerScan {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	scan := &coredata.CookieBannerScan{
		ID:             gid.New(fx.scope.GetTenantID(), coredata.CookieBannerScanEntityType),
		OrganizationID: fx.organizationID,
		CookieBannerID: fx.banner.ID,
		State:          coredata.CookieBannerScanStatePending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := scan.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		return scan.MarkRunning(ctx, tx, fx.scope, now)
	}))

	return scan
}

func reloadScan(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	scanID gid.GID,
) coredata.CookieBannerScan {
	t.Helper()

	var scan coredata.CookieBannerScan

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return scan.LoadByID(ctx, conn, fx.scope, scanID)
	}))

	return scan
}

func TestCookieBannerScanHandler_Claim_StopsAtRunningLimit(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	insertRunningScan(t, ctx, client, fx)

	_, err := newScanWorker(t, client).handler.Claim(ctx)
	assert.ErrorIs(t, err, worker.ErrNoTask)
}

func TestCookieBannerScanHandler_Claim_TakesTheLeastRecentlyRunScan(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	pausedBanner := seedWorkerFixture(t, ctx, client)
	pendingBanner := seedWorkerFixture(t, ctx, client)

	// Year 2000 sorts ahead of any scan another test creates with the
	// current time, so these claims take our rows and not theirs. The
	// paused scan was requested first but ran after the pending one was
	// requested, so it waits its turn.
	requested := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	waiting := requested.Add(time.Hour)
	ran := requested.Add(2 * time.Hour)

	paused := &coredata.CookieBannerScan{
		ID:             gid.New(pausedBanner.scope.GetTenantID(), coredata.CookieBannerScanEntityType),
		OrganizationID: pausedBanner.organizationID,
		CookieBannerID: pausedBanner.banner.ID,
		State:          coredata.CookieBannerScanStatePending,
		CreatedAt:      requested,
		UpdatedAt:      requested,
	}
	pending := &coredata.CookieBannerScan{
		ID:             gid.New(pendingBanner.scope.GetTenantID(), coredata.CookieBannerScanEntityType),
		OrganizationID: pendingBanner.organizationID,
		CookieBannerID: pendingBanner.banner.ID,
		State:          coredata.CookieBannerScanStatePending,
		CreatedAt:      waiting,
		UpdatedAt:      waiting,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := paused.Insert(ctx, tx, pausedBanner.scope); err != nil {
			return err
		}

		if err := paused.MarkRunning(ctx, tx, pausedBanner.scope, requested); err != nil {
			return err
		}

		checkpoint := json.RawMessage(`{"visited_templates":["/"],"queue":[],"pages_scanned":1,"elapsed_ms":1}`)
		paused.Checkpoint = &checkpoint

		if err := paused.MarkPaused(ctx, tx, pausedBanner.scope, ran); err != nil {
			return err
		}

		return pending.Insert(ctx, tx, pendingBanner.scope)
	}))

	worker := newScanWorker(t, client)
	// Other tests may already hold a running scan. Raising the cap keeps
	// these claims on the ordering rule instead of the cluster limit.
	worker.handler.maxRunning = 1000

	first, err := worker.handler.Claim(ctx)
	require.NoError(t, err)
	assert.Equal(t, pending.ID, first.ID)
	assert.Equal(t, coredata.CookieBannerScanStateRunning, first.State)

	second, err := worker.handler.Claim(ctx)
	require.NoError(t, err)
	assert.Equal(t, paused.ID, second.ID)
}

func TestCookieBannerScanHandler_Process_CompletesWithDerivedCounts(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, ctx, client, fx)
	w := newScanWorker(t, client)

	w.handler.run = func(ctx context.Context, in scanner.Input) (*scanner.Result, error) {
		require.NoError(t, in.OnPage(
			ctx,
			scanner.Page{
				URL: fx.banner.Origin + "/",
				Cookies: []scanner.Cookie{
					{Name: "_ga", Source: coredata.CookieSourceHTTP},
				},
				Storage: []scanner.StorageItem{
					{Type: coredata.TrackerTypeLocalStorage, Key: "theme", ValueSize: 4},
				},
				Resources: []scanner.Resource{
					{
						Type:   coredata.TrackerResourceTypeScript,
						Origin: "https://cdn.example.com",
						Path:   "/a.js",
					},
				},
			},
			scanner.Checkpoint{PagesScanned: 1},
		))

		return &scanner.Result{
			PagesScanned: 3,
			Truncated:    true,
			Checkpoint:   scanner.Checkpoint{PagesScanned: 3},
		}, nil
	}

	require.NoError(t, w.handler.Process(ctx, *scan))

	got := reloadScan(t, ctx, client, fx, scan.ID)
	assert.Equal(t, coredata.CookieBannerScanStateCompleted, got.State)
	assert.True(t, got.Truncated)
	assert.Equal(t, 3, got.PagesScanned)
	assert.Equal(t, 1, got.CookiesFound)
	assert.Equal(t, 1, got.StorageFound)
	assert.Equal(t, 1, got.ResourcesFound)
	assert.NotNil(t, got.CompletedAt)
}

func TestCookieBannerScanHandler_Process_PausesOnShutdown(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, ctx, client, fx)
	w := newScanWorker(t, client)

	w.handler.run = func(ctx context.Context, _ scanner.Input) (*scanner.Result, error) {
		w.handler.signalShutdown()
		<-ctx.Done()

		return &scanner.Result{
			Paused:       true,
			PagesScanned: 2,
			Checkpoint: scanner.Checkpoint{
				PagesScanned: 2,
				Queue:        []scanner.QueueItem{{URL: "https://example.com/pricing", Depth: 1}},
			},
		}, nil
	}

	require.NoError(t, w.handler.Process(ctx, *scan))

	got := reloadScan(t, ctx, client, fx, scan.ID)
	assert.Equal(t, coredata.CookieBannerScanStatePaused, got.State)
	assert.Nil(t, got.FailureReason)
	assert.Nil(t, got.CompletedAt)
	require.NotNil(t, got.Checkpoint)
	assert.Contains(t, string(*got.Checkpoint), "https://example.com/pricing")
}

func TestCookieBannerScanHandler_Process_PausesWhenARunEnds(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, ctx, client, fx)
	w := newScanWorker(t, client)

	w.handler.run = func(context.Context, scanner.Input) (*scanner.Result, error) {
		return &scanner.Result{
			Paused:       true,
			PagesScanned: 10,
			Checkpoint: scanner.Checkpoint{
				PagesScanned: 10,
				Queue: []scanner.QueueItem{
					{URL: "https://example.com/stories/a", Depth: 1, EmbedHosts: []string{"www.youtube.com"}},
				},
				CoveredHosts: []string{"tally.so"},
			},
		}, nil
	}

	require.NoError(t, w.handler.Process(ctx, *scan))

	got := reloadScan(t, ctx, client, fx, scan.ID)
	assert.Equal(t, coredata.CookieBannerScanStatePaused, got.State)
	assert.Nil(t, got.FailureReason)
	assert.Nil(t, got.CompletedAt)
	assert.Equal(t, 10, got.PagesScanned)

	checkpoint, err := unmarshalCheckpoint(got.Checkpoint)
	require.NoError(t, err)
	require.NotNil(t, checkpoint)
	require.Len(t, checkpoint.Queue, 1)
	assert.Equal(t, []string{"www.youtube.com"}, checkpoint.Queue[0].EmbedHosts)
	assert.Equal(t, []string{"tally.so"}, checkpoint.CoveredHosts)
}

func TestCookieBannerScanHandler_Process_PausesWhenShutdownInterruptsAnError(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, ctx, client, fx)
	w := newScanWorker(t, client)

	w.handler.run = func(context.Context, scanner.Input) (*scanner.Result, error) {
		w.handler.signalShutdown()

		return nil, errors.New("browser closed")
	}

	require.NoError(t, w.handler.Process(ctx, *scan))

	got := reloadScan(t, ctx, client, fx, scan.ID)
	assert.Equal(t, coredata.CookieBannerScanStatePaused, got.State)
	assert.Nil(t, got.FailureReason)
}

func TestCookieBannerScanHandler_Process_MapsFailureReasons(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		err    error
		reason coredata.CookieBannerScanFailureReason
	}{
		{
			name:   "no pages",
			err:    scanner.ErrNoPageLoaded,
			reason: coredata.CookieBannerScanFailureReasonNoPagesLoaded,
		},
		{
			name:   "unreachable",
			err:    scanner.ErrUnreachable,
			reason: coredata.CookieBannerScanFailureReasonUnreachable,
		},
		{
			name:   "blocked",
			err:    scanner.ErrBlocked,
			reason: coredata.CookieBannerScanFailureReasonBlocked,
		},
		{
			name:   "private host",
			err:    netcheck.ErrBlockedAddress,
			reason: coredata.CookieBannerScanFailureReasonBlocked,
		},
		{
			name:   "internal",
			err:    errors.New("boom"),
			reason: coredata.CookieBannerScanFailureReasonInternal,
		},
		{
			name:   "banner not installed",
			err:    scanner.ErrBannerNotInstalled,
			reason: coredata.CookieBannerScanFailureReasonBannerNotInstalled,
		},
		{
			name:   "different banner",
			err:    scanner.ErrDifferentBanner,
			reason: coredata.CookieBannerScanFailureReasonDifferentBanner,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := test.PGClient(t)
			ctx := context.Background()
			fx := seedWorkerFixture(t, ctx, client)
			scan := insertRunningScan(t, ctx, client, fx)
			w := newScanWorker(t, client)

			w.handler.run = func(context.Context, scanner.Input) (*scanner.Result, error) {
				if errors.Is(tc.err, scanner.ErrNoPageLoaded) {
					return &scanner.Result{}, fmt.Errorf("scan: %w", tc.err)
				}

				return nil, fmt.Errorf("scan: %w", tc.err)
			}

			require.NoError(t, w.handler.Process(ctx, *scan))

			got := reloadScan(t, ctx, client, fx, scan.ID)
			assert.Equal(t, coredata.CookieBannerScanStateFailed, got.State)
			require.NotNil(t, got.FailureReason)
			assert.Equal(t, tc.reason, *got.FailureReason)
		})
	}
}

func TestCookieBannerScanHandler_Process_KeepsPagesWhenALaterPageFails(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, ctx, client, fx)
	w := newScanWorker(t, client)

	w.handler.run = func(ctx context.Context, in scanner.Input) (*scanner.Result, error) {
		require.NoError(t, in.OnPage(
			ctx,
			scanner.Page{
				Cookies: []scanner.Cookie{{Name: "_ga", Source: coredata.CookieSourceHTTP}},
			},
			scanner.Checkpoint{PagesScanned: 2},
		))

		return nil, errors.New("cannot accept scanned page")
	}

	require.NoError(t, w.handler.Process(ctx, *scan))

	got := reloadScan(t, ctx, client, fx, scan.ID)
	assert.Equal(t, coredata.CookieBannerScanStateFailed, got.State)
	assert.Equal(t, 2, got.PagesScanned)
}

func TestCookieBannerScanHandler_Process_DropsWorkWhenOwnershipIsLost(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, ctx, client, fx)
	w := newScanWorker(t, client)

	w.handler.run = func(ctx context.Context, in scanner.Input) (*scanner.Result, error) {
		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return scan.MarkFailed(ctx, tx, fx.scope, coredata.CookieBannerScanFailureReasonInterrupted, time.Now())
		}))

		require.NoError(t, in.OnPage(
			ctx,
			scanner.Page{
				Cookies: []scanner.Cookie{{Name: "lost", Source: coredata.CookieSourceHTTP}},
			},
			scanner.Checkpoint{PagesScanned: 1},
		))

		return &scanner.Result{PagesScanned: 1, Checkpoint: scanner.Checkpoint{PagesScanned: 9}}, nil
	}

	require.NoError(t, w.handler.Process(ctx, *scan))

	got := reloadScan(t, ctx, client, fx, scan.ID)
	assert.Equal(t, coredata.CookieBannerScanStateFailed, got.State)
	require.NotNil(t, got.FailureReason)
	assert.Equal(t, coredata.CookieBannerScanFailureReasonInterrupted, *got.FailureReason)
	assert.Nil(t, got.Checkpoint)

	var pattern coredata.TrackerPattern

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return pattern.FindMatchingPattern(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.TrackerTypeCookie,
			"lost",
		)
	})
	assert.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func TestCookieBannerScanHandler_Process_DiscardsOutcomeWhenHeartbeatLosesTheScan(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, context.Background(), client, fx)
	w := newScanWorker(t, client)
	w.handler.heartbeatInterval = 10 * time.Millisecond

	w.handler.run = func(ctx context.Context, _ scanner.Input) (*scanner.Result, error) {
		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return scan.MarkFailed(ctx, tx, fx.scope, coredata.CookieBannerScanFailureReasonInterrupted, time.Now())
		}))

		<-ctx.Done()

		return &scanner.Result{
			Paused:     true,
			Checkpoint: scanner.Checkpoint{PagesScanned: 9},
		}, nil
	}

	require.NoError(t, w.handler.Process(ctx, *scan))

	got := reloadScan(t, context.Background(), client, fx, scan.ID)
	assert.Equal(t, coredata.CookieBannerScanStateFailed, got.State)
	assert.Nil(t, got.Checkpoint, "a lost scan must not be paused over the top of the failure")
}

func TestCookieBannerScanHandler_Process_ResumesFromCheckpoint(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, ctx, client, fx)

	raw := json.RawMessage(`{"visited_templates":["https://example.com/"],"queue":[{"url":"https://example.com/pricing","depth":1}],"pages_scanned":4,"elapsed_ms":20}`)
	scan.Checkpoint = &raw

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return scan.SaveCheckpoint(ctx, conn, fx.scope, time.Now())
	}))

	w := newScanWorker(t, client)
	w.handler.run = func(_ context.Context, in scanner.Input) (*scanner.Result, error) {
		require.NotNil(t, in.Checkpoint)
		assert.Equal(t, "https://example.com/pricing", in.Checkpoint.Queue[0].URL)
		assert.Equal(t, 4, in.Checkpoint.PagesScanned)

		return &scanner.Result{PagesScanned: 4, Checkpoint: *in.Checkpoint}, nil
	}

	require.NoError(t, w.handler.Process(ctx, *scan))

	got := reloadScan(t, ctx, client, fx, scan.ID)
	assert.Equal(t, coredata.CookieBannerScanStateCompleted, got.State)
	assert.Equal(t, 4, got.PagesScanned)
}

func TestCookieBannerScanHandler_Process_SeedsConsentFromThePublishedVersion(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, ctx, client, fx)

	snapshot, err := json.Marshal(coredata.CookieBannerVersionSnapshot{
		CookiePolicyURL:   "https://example.com/cookies",
		ConsentExpiryDays: 180,
		DefaultLanguage:   "en",
		Categories: []coredata.CookieBannerVersionSnapshotCategory{
			{Name: "Uncategorised", Slug: "uncategorised", Kind: coredata.CookieCategoryKindUncategorised},
			{Name: "Analytics", Slug: "analytics", Kind: coredata.CookieCategoryKindNormal},
			{Name: "Marketing", Slug: "marketing", Kind: coredata.CookieCategoryKindNormal},
		},
	})
	require.NoError(t, err)

	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		version := &coredata.CookieBannerVersion{
			ID:             gid.New(fx.scope.GetTenantID(), coredata.CookieBannerVersionEntityType),
			OrganizationID: fx.organizationID,
			CookieBannerID: fx.banner.ID,
			Version:        3,
			State:          coredata.CookieBannerVersionStatePublished,
			Snapshot:       snapshot,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		return version.Insert(ctx, tx, fx.scope)
	}))

	w := newScanWorker(t, client)
	w.handler.run = func(_ context.Context, in scanner.Input) (*scanner.Result, error) {
		require.NotNil(t, in.Consent)
		assert.Equal(t, 3, in.Consent.Version)
		assert.Equal(t, []string{"analytics", "marketing"}, in.Consent.CategorySlugs)
		assert.Equal(t, "ws://localhost:9222", in.BrowserURL)

		return &scanner.Result{PagesScanned: 1}, nil
	}

	require.NoError(t, w.handler.Process(ctx, *scan))
}

func TestCookieBannerScanHandler_Process_SkipsConsentWithoutAPublishedVersion(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	scan := insertRunningScan(t, ctx, client, fx)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(
			ctx,
			`UPDATE cookie_banners SET state = $1 WHERE id = $2`,
			coredata.CookieBannerStateInactive,
			fx.banner.ID,
		)

		return err
	}))

	w := newScanWorker(t, client)
	w.handler.run = func(_ context.Context, in scanner.Input) (*scanner.Result, error) {
		assert.Nil(t, in.Consent)

		return &scanner.Result{PagesScanned: 1}, nil
	}

	require.NoError(t, w.handler.Process(ctx, *scan))
}
