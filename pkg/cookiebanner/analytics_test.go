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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

var analyticsTestDay = time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)

func TestReportDetectedTrackers_CountsIntoTodayBucket(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.Add(15 * time.Hour)
	svc := newServiceAt(client, &now)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:   "_ga",
			Source: coredata.CookieSourceScript,
			Hit:    true,
		}},
	}))
	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyFirefox,
		Cookies: []DetectedCookie{{
			Name:   "_ga",
			Source: coredata.CookieSourceScript,
		}},
	}))
	requireFlushAnalytics(t, ctx, svc)

	assert.Equal(
		t,
		[]BrowserFamilyCount{
			{Family: coredata.BrowserFamilyChrome, Count: 1},
			{Family: coredata.BrowserFamilyFirefox, Count: 1},
		},
		pageLoads(t, ctx, svc, fx, 1),
	)
	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 1}},
		patternHits(t, ctx, client, svc, fx, "_ga", 1).Total,
	)
}

func TestReportDetectedTrackers_CountsPatternOncePerReport(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.Add(15 * time.Hour)
	svc := newServiceAt(client, &now)

	hostOnly := true
	domain := "example.com"
	notHostOnly := false

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{
			{
				Name:     "_ga",
				Source:   coredata.CookieSourceScript,
				HostOnly: &hostOnly,
				Hit:      true,
			},
			{
				Name:         "_ga",
				Source:       coredata.CookieSourceScript,
				CookieDomain: &domain,
				HostOnly:     &notHostOnly,
				Hit:          true,
			},
		},
	}))
	requireFlushAnalytics(t, ctx, svc)

	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 1}},
		patternHits(t, ctx, client, svc, fx, "_ga", DefaultAnalyticsDays).Total,
	)
}

func TestReportDetectedTrackers_HitWithoutPageLoadStillCounts(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.Add(15 * time.Hour)
	svc := newServiceAt(client, &now)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		Family: coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:   "_ga",
			Source: coredata.CookieSourceScript,
			Hit:    true,
		}},
	}))
	requireFlushAnalytics(t, ctx, svc)

	assert.Empty(t, pageLoads(t, ctx, svc, fx, DefaultAnalyticsDays))
	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 1}},
		patternHits(t, ctx, client, svc, fx, "_ga", DefaultAnalyticsDays).Total,
	)
}

func TestReportDetectedTrackers_SkipsNoopRepeatUpsert(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.Add(15 * time.Hour)
	svc := newServiceAt(client, &now)

	req := ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:   "_ga",
			Source: coredata.CookieSourceScript,
			Hit:    true,
		}},
	}

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, req))
	requireFlushAnalytics(t, ctx, svc)

	patternBefore := loadPattern(t, ctx, client, fx, "_ga")
	detectionsBefore := loadDetectedTrackersByPattern(t, ctx, client, fx, "_ga")
	require.Len(t, detectionsBefore, 1)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, req))
	requireFlushAnalytics(t, ctx, svc)

	patternAfter := loadPattern(t, ctx, client, fx, "_ga")
	detectionsAfter := loadDetectedTrackersByPattern(t, ctx, client, fx, "_ga")
	require.Len(t, detectionsAfter, 1)

	assert.Equal(t, patternBefore.UpdatedAt, patternAfter.UpdatedAt)
	assert.Equal(t, detectionsBefore[0].UpdatedAt, detectionsAfter[0].UpdatedAt)
	assert.Equal(t, detectionsBefore[0].ID, detectionsAfter[0].ID)

	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 2}},
		pageLoads(t, ctx, svc, fx, DefaultAnalyticsDays),
	)
	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 2}},
		patternHits(t, ctx, client, svc, fx, "_ga", DefaultAnalyticsDays).Total,
	)
}

func TestReportDetectedTrackers_PageLoadGrant(t *testing.T) {
	t.Parallel()

	acceptAll := coredata.CookieConsentActionAcceptAll
	rejectAll := coredata.CookieConsentActionRejectAll
	acknowledge := coredata.CookieConsentActionAcknowledge
	gpc := coredata.CookieConsentActionGPC

	t.Run("opt-in counts only accept all", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		fx := seedWorkerFixture(t, ctx, client)
		now := analyticsTestDay.Add(15 * time.Hour)
		svc := newServiceAt(client, &now)

		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &acceptAll,
			Family:        coredata.BrowserFamilyChrome,
			Presentation:  coredata.CookieBannerPresentationOptIn,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:     true,
			Family:       coredata.BrowserFamilyFirefox,
			Presentation: coredata.CookieBannerPresentationOptIn,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &rejectAll,
			Family:        coredata.BrowserFamilySafari,
			Presentation:  coredata.CookieBannerPresentationOptIn,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &acknowledge,
			Family:        coredata.BrowserFamilyEdge,
			Presentation:  coredata.CookieBannerPresentationOptIn,
		}))
		requireFlushAnalytics(t, ctx, svc)

		assert.Equal(
			t,
			[]BrowserFamilyCount{
				{Family: coredata.BrowserFamilyChrome, Count: 1},
				{Family: coredata.BrowserFamilyEdge, Count: 1},
				{Family: coredata.BrowserFamilyFirefox, Count: 1},
				{Family: coredata.BrowserFamilySafari, Count: 1},
			},
			pageLoads(t, ctx, svc, fx, DefaultAnalyticsDays),
		)
		assert.Equal(
			t,
			[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 1}},
			grantedPageLoads(t, ctx, client, fx),
		)
	})

	t.Run("opt-out counts every view except reject and gpc", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		fx := seedWorkerFixture(t, ctx, client)
		now := analyticsTestDay.Add(15 * time.Hour)
		svc := newServiceAt(client, &now)

		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:     true,
			Family:       coredata.BrowserFamilyChrome,
			Presentation: coredata.CookieBannerPresentationOptOut,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &acceptAll,
			Family:        coredata.BrowserFamilyFirefox,
			Presentation:  coredata.CookieBannerPresentationOptOut,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &rejectAll,
			Family:        coredata.BrowserFamilySafari,
			Presentation:  coredata.CookieBannerPresentationOptOut,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &gpc,
			Family:        coredata.BrowserFamilyEdge,
			Presentation:  coredata.CookieBannerPresentationOptOut,
		}))
		requireFlushAnalytics(t, ctx, svc)

		assert.Equal(
			t,
			[]BrowserFamilyCount{
				{Family: coredata.BrowserFamilyChrome, Count: 1},
				{Family: coredata.BrowserFamilyEdge, Count: 1},
				{Family: coredata.BrowserFamilyFirefox, Count: 1},
				{Family: coredata.BrowserFamilySafari, Count: 1},
			},
			pageLoads(t, ctx, svc, fx, DefaultAnalyticsDays),
		)
		assert.Equal(
			t,
			[]BrowserFamilyCount{
				{Family: coredata.BrowserFamilyChrome, Count: 1},
				{Family: coredata.BrowserFamilyFirefox, Count: 1},
			},
			grantedPageLoads(t, ctx, client, fx),
		)
	})

	t.Run("notice counts acknowledge and a missing action", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		fx := seedWorkerFixture(t, ctx, client)
		now := analyticsTestDay.Add(15 * time.Hour)
		svc := newServiceAt(client, &now)

		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:     true,
			Family:       coredata.BrowserFamilyChrome,
			Presentation: coredata.CookieBannerPresentationNotice,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &acknowledge,
			Family:        coredata.BrowserFamilyFirefox,
			Presentation:  coredata.CookieBannerPresentationNotice,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &rejectAll,
			Family:        coredata.BrowserFamilySafari,
			Presentation:  coredata.CookieBannerPresentationNotice,
		}))
		requireFlushAnalytics(t, ctx, svc)

		assert.Equal(
			t,
			[]BrowserFamilyCount{
				{Family: coredata.BrowserFamilyChrome, Count: 1},
				{Family: coredata.BrowserFamilyFirefox, Count: 1},
				{Family: coredata.BrowserFamilySafari, Count: 1},
			},
			pageLoads(t, ctx, svc, fx, DefaultAnalyticsDays),
		)
		assert.Equal(
			t,
			[]BrowserFamilyCount{
				{Family: coredata.BrowserFamilyChrome, Count: 1},
				{Family: coredata.BrowserFamilyFirefox, Count: 1},
			},
			grantedPageLoads(t, ctx, client, fx),
		)
	})

	t.Run("force opt-in denies an unconsented opt-out report", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		fx := seedWorkerFixture(t, ctx, client)
		fx.banner.ForceOptIn = true

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return fx.banner.Update(ctx, tx, fx.scope)
		}))

		now := analyticsTestDay.Add(15 * time.Hour)
		svc := newServiceAt(client, &now)

		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:     true,
			Family:       coredata.BrowserFamilyChrome,
			Presentation: coredata.CookieBannerPresentationOptOut,
		}))
		requireFlushAnalytics(t, ctx, svc)

		assert.Equal(
			t,
			[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 1}},
			pageLoads(t, ctx, svc, fx, DefaultAnalyticsDays),
		)
		assert.Empty(t, grantedPageLoads(t, ctx, client, fx))
	})
}

func TestFlushAnalytics_KeepsDayAcrossMidnight(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.Add(-10 * time.Second)
	svc := newServiceAt(client, &now)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
	}))

	now = analyticsTestDay.Add(5 * time.Second)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
	}))
	requireFlushAnalytics(t, ctx, svc)

	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 1}},
		pageLoads(t, ctx, svc, fx, 1),
	)
	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 2}},
		pageLoads(t, ctx, svc, fx, 2),
	)

	requireFlushAnalytics(t, ctx, svc)

	assert.Equal(t, 1, syncMapLen(&svc.pageLoadCounts))
}

func TestRunAnalyticsFlusher_FlushesOnShutdown(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.Add(15 * time.Hour)
	svc := newServiceAt(client, &now)

	flusherCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)

	go func() {
		done <- svc.RunAnalyticsFlusher(flusherCtx, log.NewLogger())
	}()

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
	}))

	cancel()
	require.NoError(t, <-done)

	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 1}},
		pageLoads(t, ctx, svc, fx, DefaultAnalyticsDays),
	)
}

func TestPublishCookieBannerVersion_KeepsDailyCounts(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.Add(15 * time.Hour)
	svc := newServiceAt(client, &now)

	req := ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:   "_ga",
			Source: coredata.CookieSourceScript,
			Hit:    true,
		}},
	}

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, req))
	requireFlushAnalytics(t, ctx, svc)

	insertDraftVersion(t, ctx, client, fx)

	_, err := svc.PublishCookieBannerVersion(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, req))
	requireFlushAnalytics(t, ctx, svc)

	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 2}},
		pageLoads(t, ctx, svc, fx, DefaultAnalyticsDays),
	)
	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 2}},
		patternHits(t, ctx, client, svc, fx, "_ga", DefaultAnalyticsDays).Total,
	)
}

func TestAddDailyPatternHits_SumsEveryDay(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.AddDate(0, 0, -1).Add(15 * time.Hour)
	svc := newServiceAt(client, &now)

	reportCookie := func(name string) {
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView: true,
			Family:   coredata.BrowserFamilyChrome,
			Cookies: []DetectedCookie{{
				Name:   name,
				Source: coredata.CookieSourceScript,
				Hit:    true,
			}},
		}))
	}

	reportCookie("_ga")

	now = analyticsTestDay.Add(15 * time.Hour)

	reportCookie("_ga")
	reportCookie("_gid")
	requireFlushAnalytics(t, ctx, svc)

	ga := loadPattern(t, ctx, client, fx, "_ga")
	gidPattern := loadPattern(t, ctx, client, fx, "_gid")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return coredata.AddDailyPatternHits(ctx, tx, fx.scope, ga.ID, gidPattern.ID)
	}))

	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 2}},
		patternHits(t, ctx, client, svc, fx, "_gid", 1).Total,
	)
	assert.Equal(
		t,
		[]BrowserFamilyCount{{Family: coredata.BrowserFamilyChrome, Count: 3}},
		patternHits(t, ctx, client, svc, fx, "_gid", 2).Total,
	)
}

func TestGetPageLoads_WindowEndsToday(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.AddDate(0, 0, -7).Add(12 * time.Hour)
	svc := newServiceAt(client, &now)

	report := func(family coredata.BrowserFamily) {
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView: true,
			Family:   family,
			Cookies: []DetectedCookie{{
				Name:   "_ga",
				Source: coredata.CookieSourceScript,
				Hit:    true,
			}},
		}))
	}

	report(coredata.BrowserFamilyChrome)

	now = analyticsTestDay.AddDate(0, 0, -6).Add(12 * time.Hour)

	report(coredata.BrowserFamilyFirefox)

	now = analyticsTestDay.Add(12 * time.Hour)

	report(coredata.BrowserFamilySafari)
	requireFlushAnalytics(t, ctx, svc)

	assert.Equal(
		t,
		[]BrowserFamilyCount{
			{Family: coredata.BrowserFamilyFirefox, Count: 1},
			{Family: coredata.BrowserFamilySafari, Count: 1},
		},
		pageLoads(t, ctx, svc, fx, 7),
	)
	assert.Equal(
		t,
		[]BrowserFamilyCount{
			{Family: coredata.BrowserFamilyChrome, Count: 1},
			{Family: coredata.BrowserFamilyFirefox, Count: 1},
			{Family: coredata.BrowserFamilySafari, Count: 1},
		},
		pageLoads(t, ctx, svc, fx, 8),
	)
	assert.Equal(
		t,
		[]BrowserFamilyCount{
			{Family: coredata.BrowserFamilyFirefox, Count: 1},
			{Family: coredata.BrowserFamilySafari, Count: 1},
		},
		patternHits(t, ctx, client, svc, fx, "_ga", 7).Total,
	)
}

func TestGetPageLoads_IsolatesTenants(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	other := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.Add(15 * time.Hour)
	svc := newServiceAt(client, &now)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:   "_ga",
			Source: coredata.CookieSourceScript,
			Hit:    true,
		}},
	}))
	requireFlushAnalytics(t, ctx, svc)

	loads, err := svc.GetPageLoads(ctx, other.scope, []gid.GID{fx.banner.ID}, DefaultAnalyticsDays)
	require.NoError(t, err)
	assert.Empty(t, loads[fx.banner.ID])

	ga := loadPattern(t, ctx, client, fx, "_ga")

	splits, err := svc.GetHitSplits(ctx, other.scope, ga.ID, DefaultAnalyticsDays)
	require.NoError(t, err)
	assert.Empty(t, splits.Total)
}

func TestGetPageLoads_RejectsDaysOutOfRange(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)
	patternID := gid.New(fx.scope.GetTenantID(), coredata.TrackerPatternEntityType)

	for _, days := range []int{0, MaxAnalyticsDays + 1} {
		var validationErrs validator.ValidationErrors

		_, err := svc.GetPageLoads(ctx, fx.scope, []gid.GID{fx.banner.ID}, days)
		assert.ErrorAs(t, err, &validationErrs)

		_, err = svc.GetHitSplits(ctx, fx.scope, patternID, days)
		assert.ErrorAs(t, err, &validationErrs)
	}

	_, err := svc.GetPageLoads(ctx, fx.scope, []gid.GID{fx.banner.ID}, MaxAnalyticsDays)
	require.NoError(t, err)
}

func newServiceAt(client *pg.Client, now *time.Time) *Service {
	svc := NewService(client, false, 0)
	svc.now = func() time.Time { return *now }

	return svc
}

func pageLoads(
	t *testing.T,
	ctx context.Context,
	svc *Service,
	fx workerFixture,
	days int,
) []BrowserFamilyCount {
	t.Helper()

	loads, err := svc.GetPageLoads(ctx, fx.scope, []gid.GID{fx.banner.ID}, days)
	require.NoError(t, err)

	return loads[fx.banner.ID]
}

func patternHits(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	svc *Service,
	fx workerFixture,
	pattern string,
	days int,
) HitSplits {
	t.Helper()

	tp := loadPattern(t, ctx, client, fx, pattern)

	splits, err := svc.GetHitSplits(ctx, fx.scope, tp.ID, days)
	require.NoError(t, err)

	return splits
}

func grantedPageLoads(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
) []BrowserFamilyCount {
	t.Helper()

	var rows []coredata.DailyPageLoad

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded, err := coredata.LoadDailyPageLoadsByCookieBannerIDs(
			ctx,
			conn,
			fx.scope,
			[]gid.GID{fx.banner.ID},
			analyticsTestDay,
		)
		if err != nil {
			return err
		}

		rows = loaded

		return nil
	}))

	granted := make([]coredata.DailyPageLoad, 0, len(rows))

	for _, row := range rows {
		var action *coredata.CookieConsentAction
		if row.ConsentAction != "" {
			action = &row.ConsentAction
		}

		if pageLoadGranted(row.Presentation, action) {
			granted = append(granted, row)
		}
	}

	return PageLoadsByFamily(granted)
}

func insertDraftVersion(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
) {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		version := &coredata.CookieBannerVersion{
			ID:             gid.New(fx.scope.GetTenantID(), coredata.CookieBannerVersionEntityType),
			OrganizationID: fx.organizationID,
			CookieBannerID: fx.banner.ID,
			Version:        1,
			State:          coredata.CookieBannerVersionStateDraft,
			Snapshot:       json.RawMessage(`{}`),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		return version.Insert(ctx, tx, fx.scope)
	}))
}

func requireFlushAnalytics(t *testing.T, ctx context.Context, svc *Service) {
	t.Helper()

	require.NoError(t, svc.FlushAnalytics(ctx))
}

func loadPattern(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	pattern string,
) coredata.TrackerPattern {
	t.Helper()

	var tp coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return tp.LoadByBannerIDTypeAndPattern(ctx, conn, fx.scope, fx.banner.ID, coredata.TrackerTypeCookie, pattern, nil)
	}))

	return tp
}
