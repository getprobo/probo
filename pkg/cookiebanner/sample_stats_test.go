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
	assert.Equal(t, 1, sample.ChromePageLoads)
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
	assert.Equal(t, 1, sample.ChromePageLoads)
	assert.Equal(t, 1, sample.FirefoxPageLoads)
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
	assert.Equal(t, 1, sample.ChromePageLoads)
	assert.Equal(t, 1, sample.FirefoxPageLoads)
	assert.Equal(t, 0, sample.SafariPageLoads)
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
	assert.Equal(t, 0, sample.ChromePageLoads)

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
	assert.Equal(t, 0, sample.ChromePageLoads)

	var pattern coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return pattern.LoadByBannerIDTypeAndPattern(ctx, conn, fx.scope, fx.banner.ID, coredata.TrackerTypeCookie, "_ga", nil)
	}))

	var hits coredata.TrackerPatternSampleHits

	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return hits.LoadBySampleAndPattern(ctx, conn, fx.scope, sample.ID, pattern.ID)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
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
	assert.Equal(t, 0, sample.ChromePageLoads)
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
	assert.Equal(t, 1, sample.ChromePageLoads)
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
	assert.Equal(t, 2, sample.ChromePageLoads)
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
	assert.Equal(t, 1, sample.ChromePageLoads)
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
	assert.Equal(t, 1, second.SafariPageLoads)
	assert.Equal(t, 0, second.ChromePageLoads)
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
	assert.Equal(t, 1, reloadedFirst.ChromePageLoads)
	assert.Equal(t, 0, reloadedFirst.FirefoxPageLoads)

	second := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, started.ID, second.ID)
	assert.Nil(t, second.FrozenAt)
	assert.Equal(t, 1, second.FirefoxPageLoads)
	assert.Equal(t, 0, second.ChromePageLoads)
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
	assert.Equal(t, 1, sample.ChromePageLoads)
	assert.Equal(t, 0, sample.SafariPageLoads)
}

func TestReportDetectedTrackers_ConsentingPageLoads(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	acceptAll := coredata.CookieConsentActionAcceptAll
	rejectAll := coredata.CookieConsentActionRejectAll
	acknowledge := coredata.CookieConsentActionAcknowledge

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView:      true,
		ConsentAction: &acceptAll,
		Family:        coredata.BrowserFamilyChrome,
	}))
	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView:      true,
		ConsentAction: &acknowledge,
		Family:        coredata.BrowserFamilyFirefox,
	}))
	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView:      true,
		ConsentAction: &rejectAll,
		Family:        coredata.BrowserFamilySafari,
	}))
	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.BrowserFamilyEdge,
	}))
	requireFlushSampleCounts(t, ctx, svc)

	sample := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, 1, sample.ChromePageLoads)
	assert.Equal(t, 1, sample.FirefoxPageLoads)
	assert.Equal(t, 1, sample.SafariPageLoads)
	assert.Equal(t, 1, sample.EdgePageLoads)
	assert.Equal(t, 1, sample.ChromeConsentingPageLoads)
	assert.Equal(t, 1, sample.FirefoxConsentingPageLoads)
	assert.Equal(t, 0, sample.SafariConsentingPageLoads)
	assert.Equal(t, 0, sample.EdgeConsentingPageLoads)
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

func loadSampleHits(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	sampleID gid.GID,
	pattern string,
) coredata.TrackerPatternSampleHits {
	t.Helper()

	tp := loadPattern(t, ctx, client, fx, pattern)

	var hits coredata.TrackerPatternSampleHits

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return hits.LoadBySampleAndPattern(ctx, conn, fx.scope, sampleID, tp.ID)
	}))

	return hits
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
