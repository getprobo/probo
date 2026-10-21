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
	"go.probo.inc/probo/pkg/page"
)

func TestReportDetectedTrackers_CountsSharedSampleWindow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	sample := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
	assert.Nil(t, sample.FrozenAt)

	hits := loadSampleHits(t, ctx, client, fx, sample.ID, "_ga")
	assert.Equal(t, 1, hits.ChromeHits)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyFirefox,
		Cookies: []DetectedCookie{{
			Name:   "_ga",
			Source: coredata.CookieSourceScript,
		}},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	sample = loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).FirefoxPageLoads)
	assert.Equal(t, 1, loadSampleHits(t, ctx, client, fx, sample.ID, "_ga").ChromeHits)
	assert.Equal(t, 0, loadSampleHits(t, ctx, client, fx, sample.ID, "_ga").FirefoxHits)

	insertDraftVersion(t, ctx, client, fx)

	_, err := svc.PublishCookieBannerVersion(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilySafari,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	sample = loadLatestSample(t, ctx, client, fx)
	require.NotNil(t, sample.FrozenAt)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).FirefoxPageLoads)
	assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).SafariPageLoads)
	assert.Equal(t, 1, loadSampleHits(t, ctx, client, fx, sample.ID, "_ga").ChromeHits)
	assert.Equal(t, 0, loadSampleHits(t, ctx, client, fx, sample.ID, "_ga").SafariHits)
}

func TestReportDetectedTrackers_CountsPatternOncePerReport(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	hostOnly := true
	domain := "example.com"
	notHostOnly := false

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{
			{
				Name:      "_ga",
				Source:    coredata.CookieSourceScript,
				HostOnly:  &hostOnly,
				SampleHit: true,
			},
			{
				Name:         "_ga",
				Source:       coredata.CookieSourceScript,
				CookieDomain: &domain,
				HostOnly:     &notHostOnly,
				SampleHit:    true,
			},
		},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	sample := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, 1, loadSampleHits(t, ctx, client, fx, sample.ID, "_ga").ChromeHits)
}

func TestPublishCookieBannerVersion_FreezesEmptySample(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	insertDraftVersion(t, ctx, client, fx)

	_, err := svc.PublishCookieBannerVersion(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	sample := loadLatestSample(t, ctx, client, fx)
	require.NotNil(t, sample.FrozenAt)
	assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	sample = loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)

	var pattern coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return pattern.LoadByBannerIDTypeAndPattern(ctx, conn, fx.scope, fx.banner.ID, coredata.TrackerTypeCookie, "_ga", nil)
	}))

	hits := loadSampleHits(t, ctx, client, fx, sample.ID, "_ga")
	assert.Equal(t, 0, hits.ChromeHits)
}

func TestReportDetectedTrackers_HitWithoutPageLoadStillCounts(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		Family: coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	sample := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
	assert.Equal(t, 1, loadSampleHits(t, ctx, client, fx, sample.ID, "_ga").ChromeHits)
}

func TestAddSampleHits_SumsOntoTarget(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	for _, name := range []string{"_ga", "_gid"} {
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView: true,
			Family:   coredata.BrowserFamilyChrome,
			Cookies: []DetectedCookie{{
				Name:      name,
				Source:    coredata.CookieSourceScript,
				SampleHit: true,
			}},
		}))
	}

	requireFlushSampleCounts(t, ctx, svc)

	ga := loadPattern(t, ctx, client, fx, "_ga")
	gidPattern := loadPattern(t, ctx, client, fx, "_gid")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return coredata.AddSampleHits(ctx, tx, fx.scope, ga.ID, gidPattern.ID)
	}))

	sample := loadLatestSample(t, ctx, client, fx)
	hits := loadSampleHits(t, ctx, client, fx, sample.ID, "_gid")
	assert.Equal(t, 2, hits.ChromeHits)
}

func TestPublishCookieBannerVersion_FlushesPendingCounts(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}))

	insertDraftVersion(t, ctx, client, fx)

	_, err := svc.PublishCookieBannerVersion(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	sample := loadLatestSample(t, ctx, client, fx)
	require.NotNil(t, sample.FrozenAt)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
	assert.Equal(t, 1, loadSampleHits(t, ctx, client, fx, sample.ID, "_ga").ChromeHits)
}

func TestReportDetectedTrackers_SkipsNoopRepeatUpsert(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	req := ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, req))
	requireFlushSampleCounts(t, ctx, svc)

	patternBefore := loadPattern(t, ctx, client, fx, "_ga")
	detectionsBefore := loadDetectedTrackersByPattern(t, ctx, client, fx, "_ga")
	require.Len(t, detectionsBefore, 1)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, req))
	requireFlushSampleCounts(t, ctx, svc)

	patternAfter := loadPattern(t, ctx, client, fx, "_ga")
	detectionsAfter := loadDetectedTrackersByPattern(t, ctx, client, fx, "_ga")
	require.Len(t, detectionsAfter, 1)

	assert.Equal(t, patternBefore.UpdatedAt, patternAfter.UpdatedAt)
	assert.Equal(t, detectionsBefore[0].UpdatedAt, detectionsAfter[0].UpdatedAt)
	assert.Equal(t, detectionsBefore[0].ID, detectionsAfter[0].ID)

	sample := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, 2, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
	assert.Equal(t, 2, loadSampleHits(t, ctx, client, fx, sample.ID, "_ga").ChromeHits)
}

func TestRunSampleCountFlusher_FlushesOnShutdown(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	flusherCtx, cancel := context.WithCancel(ctx)
	done := make(chan error, 1)

	go func() {
		done <- svc.RunSampleCountFlusher(flusherCtx, log.NewLogger())
	}()

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
	}))

	cancel()
	require.NoError(t, <-done)

	sample := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
}

func TestStartCookieBannerSample_OpensAfterPublish(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	first := loadLatestSample(t, ctx, client, fx)

	insertDraftVersion(t, ctx, client, fx)
	_, err := svc.PublishCookieBannerVersion(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	started, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, started.ID)
	assert.Nil(t, started.FrozenAt)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilySafari,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	second := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, started.ID, second.ID)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, second.ID).SafariPageLoads)
	assert.Equal(t, 0, loadsOf(t, ctx, client, fx, second.ID).ChromePageLoads)
	assert.Equal(t, 1, loadSampleHits(t, ctx, client, fx, second.ID, "_ga").SafariHits)
	assert.Equal(t, 1, loadSampleHits(t, ctx, client, fx, first.ID, "_ga").ChromeHits)
	assert.Equal(t, 0, loadSampleHits(t, ctx, client, fx, first.ID, "_ga").SafariHits)
}

func TestStartCookieBannerSample_FreezesPreviousWindow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	first := loadLatestSample(t, ctx, client, fx)

	started, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyFirefox,
		Cookies: []DetectedCookie{{
			Name:      "_ga",
			Source:    coredata.CookieSourceScript,
			SampleHit: true,
		}},
	}))
	requireFlushSampleCounts(t, ctx, svc)

	reloadedFirst := loadSample(t, ctx, client, fx, first.ID)
	require.NotNil(t, reloadedFirst.FrozenAt)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, reloadedFirst.ID).ChromePageLoads)
	assert.Equal(t, 0, loadsOf(t, ctx, client, fx, reloadedFirst.ID).FirefoxPageLoads)

	second := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, started.ID, second.ID)
	assert.Nil(t, second.FrozenAt)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, second.ID).FirefoxPageLoads)
	assert.Equal(t, 0, loadsOf(t, ctx, client, fx, second.ID).ChromePageLoads)
}

func TestStopCookieBannerSample_FreezesOpenWindow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
	}))
	requireFlushSampleCounts(t, ctx, svc)

	stopped, err := svc.StopCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)
	require.NotNil(t, stopped.FrozenAt)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilySafari,
	}))
	requireFlushSampleCounts(t, ctx, svc)

	sample := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, stopped.ID, sample.ID)
	assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
	assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).SafariPageLoads)
}

func TestStopCookieBannerSample_RequiresOpenSample(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	_, err := svc.StopCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.ErrorIs(t, err, ErrSampleNotFound)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyChrome,
	}))
	requireFlushSampleCounts(t, ctx, svc)

	stopped, err := svc.StopCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)
	require.NotNil(t, stopped.FrozenAt)

	_, err = svc.StopCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.ErrorIs(t, err, ErrSampleNotFound)
}

func TestFreezeCookieBannerSample_LeavesNewerOpenSample(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	first, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	second, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, second.ID)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return coredata.FreezeCookieBannerSample(ctx, tx, fx.scope, first.ID)
	}))

	reloadedFirst := loadSample(t, ctx, client, fx, first.ID)
	require.NotNil(t, reloadedFirst.FrozenAt)

	open := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, second.ID, open.ID)
	assert.Nil(t, open.FrozenAt)
}

func TestReportDetectedTrackers_ExpectedPageLoads(t *testing.T) {
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
		svc := NewService(client, false, 0)

		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &acceptAll,
			Family:        coredata.BrowserFamilyChrome,
			Presentation:  PresentationOptIn,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:     true,
			Family:       coredata.BrowserFamilyFirefox,
			Presentation: PresentationOptIn,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &rejectAll,
			Family:        coredata.BrowserFamilySafari,
			Presentation:  PresentationOptIn,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &acknowledge,
			Family:        coredata.BrowserFamilyEdge,
			Presentation:  PresentationOptIn,
		}))
		requireFlushSampleCounts(t, ctx, svc)

		sample := loadLatestSample(t, ctx, client, fx)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).FirefoxPageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).SafariPageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).EdgePageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromeConsentingPageLoads)
		assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).FirefoxConsentingPageLoads)
		assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).SafariConsentingPageLoads)
		assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).EdgeConsentingPageLoads)
	})

	t.Run("opt-out counts every view except reject and gpc", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		fx := seedWorkerFixture(t, ctx, client)
		svc := NewService(client, false, 0)

		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:     true,
			Family:       coredata.BrowserFamilyChrome,
			Presentation: PresentationOptOut,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &acceptAll,
			Family:        coredata.BrowserFamilyFirefox,
			Presentation:  PresentationOptOut,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &rejectAll,
			Family:        coredata.BrowserFamilySafari,
			Presentation:  PresentationOptOut,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &gpc,
			Family:        coredata.BrowserFamilyEdge,
			Presentation:  PresentationOptOut,
		}))
		requireFlushSampleCounts(t, ctx, svc)

		sample := loadLatestSample(t, ctx, client, fx)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).FirefoxPageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).SafariPageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).EdgePageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromeConsentingPageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).FirefoxConsentingPageLoads)
		assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).SafariConsentingPageLoads)
		assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).EdgeConsentingPageLoads)
	})

	t.Run("notice counts acknowledge and a missing action", func(t *testing.T) {
		t.Parallel()

		client := test.PGClient(t)
		ctx := context.Background()
		fx := seedWorkerFixture(t, ctx, client)
		svc := NewService(client, false, 0)

		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:     true,
			Family:       coredata.BrowserFamilyChrome,
			Presentation: PresentationNotice,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &acknowledge,
			Family:        coredata.BrowserFamilyFirefox,
			Presentation:  PresentationNotice,
		}))
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView:      true,
			ConsentAction: &rejectAll,
			Family:        coredata.BrowserFamilySafari,
			Presentation:  PresentationNotice,
		}))
		requireFlushSampleCounts(t, ctx, svc)

		sample := loadLatestSample(t, ctx, client, fx)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromePageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).FirefoxPageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).SafariPageLoads)
		assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).EdgePageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).ChromeConsentingPageLoads)
		assert.Equal(t, 1, loadsOf(t, ctx, client, fx, sample.ID).FirefoxConsentingPageLoads)
		assert.Equal(t, 0, loadsOf(t, ctx, client, fx, sample.ID).SafariConsentingPageLoads)
	})
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

func loadLatestSample(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
) coredata.CookieBannerSample {
	t.Helper()

	var sample coredata.CookieBannerSample

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return sample.LoadLatestByCookieBannerID(ctx, conn, fx.scope, fx.banner.ID)
	}))

	return sample
}

func loadSample(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	sampleID gid.GID,
) coredata.CookieBannerSample {
	t.Helper()

	var sample coredata.CookieBannerSample

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return sample.LoadByID(ctx, conn, fx.scope, sampleID)
	}))

	return sample
}

type loadCounts struct {
	ChromePageLoads            int
	EdgePageLoads              int
	FirefoxPageLoads           int
	SafariPageLoads            int
	OtherPageLoads             int
	ChromeConsentingPageLoads  int
	EdgeConsentingPageLoads    int
	FirefoxConsentingPageLoads int
	SafariConsentingPageLoads  int
	OtherConsentingPageLoads   int
}

func loadsOf(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	sampleID gid.GID,
) loadCounts {
	t.Helper()

	var rows []coredata.SamplePageLoad

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded, err := coredata.LoadSamplePageLoads(ctx, conn, fx.scope, sampleID)
		if err != nil {
			return err
		}

		rows = loaded

		return nil
	}))

	var counts loadCounts

	for _, row := range rows {
		allowed := pageLoadGranted(Presentation(row.Presentation), consentActionPtr(row.ConsentAction))
		switch row.BrowserFamily {
		case coredata.BrowserFamilyChrome:
			counts.ChromePageLoads += row.Count
			if allowed {
				counts.ChromeConsentingPageLoads += row.Count
			}
		case coredata.BrowserFamilyEdge:
			counts.EdgePageLoads += row.Count
			if allowed {
				counts.EdgeConsentingPageLoads += row.Count
			}
		case coredata.BrowserFamilyFirefox:
			counts.FirefoxPageLoads += row.Count
			if allowed {
				counts.FirefoxConsentingPageLoads += row.Count
			}
		case coredata.BrowserFamilySafari:
			counts.SafariPageLoads += row.Count
			if allowed {
				counts.SafariConsentingPageLoads += row.Count
			}
		default:
			counts.OtherPageLoads += row.Count
			if allowed {
				counts.OtherConsentingPageLoads += row.Count
			}
		}
	}

	return counts
}

type hitCounts struct {
	ChromeHits  int
	EdgeHits    int
	FirefoxHits int
	SafariHits  int
	OtherHits   int
}

func loadSampleHits(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	sampleID gid.GID,
	pattern string,
) hitCounts {
	t.Helper()

	tp := loadPattern(t, ctx, client, fx, pattern)

	var rows []coredata.SamplePatternHit

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded, err := coredata.LoadSamplePatternHits(ctx, conn, fx.scope, sampleID, tp.ID)
		if err != nil {
			return err
		}

		rows = loaded

		return nil
	}))

	var counts hitCounts

	for _, row := range rows {
		switch row.BrowserFamily {
		case coredata.BrowserFamilyChrome:
			counts.ChromeHits += row.Count
		case coredata.BrowserFamilyEdge:
			counts.EdgeHits += row.Count
		case coredata.BrowserFamilyFirefox:
			counts.FirefoxHits += row.Count
		case coredata.BrowserFamilySafari:
			counts.SafariHits += row.Count
		default:
			counts.OtherHits += row.Count
		}
	}

	return counts
}

func requireFlushSampleCounts(t *testing.T, ctx context.Context, svc *Service) {
	t.Helper()

	require.NoError(t, svc.FlushSampleCounts(ctx))
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

func TestListCookieBannerSamples_OrdersNewestFirst(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	first, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)
	second, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.CookieBannerSampleOrderField]{
			Field:     coredata.CookieBannerSampleOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		},
	)

	samples, err := svc.ListCookieBannerSamples(ctx, fx.scope, fx.banner.ID, cursor)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(samples), 2)
	assert.Equal(t, second.ID, samples[0].ID)
	assert.Equal(t, first.ID, samples[1].ID)
}
