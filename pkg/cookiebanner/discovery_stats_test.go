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
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestReportDetectedTrackers_CountsSharedDiscoveryWindow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.DiscoveryBrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:         "_ga",
			Source:       coredata.CookieSourceScript,
			DiscoveryHit: true,
		}},
	}))

	stats := loadDiscoveryStats(t, ctx, client, fx)
	assert.Equal(t, 1, stats.ChromePageLoads)
	assert.Nil(t, stats.FrozenAt)

	hits := loadDiscoveryHits(t, ctx, client, fx, "_ga")
	assert.Equal(t, 1, hits.ChromeHits)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.DiscoveryBrowserFamilyFirefox,
		Cookies: []DetectedCookie{{
			Name:   "_ga",
			Source: coredata.CookieSourceScript,
		}},
	}))

	stats = loadDiscoveryStats(t, ctx, client, fx)
	assert.Equal(t, 1, stats.ChromePageLoads)
	assert.Equal(t, 1, stats.FirefoxPageLoads)
	assert.Equal(t, 1, loadDiscoveryHits(t, ctx, client, fx, "_ga").ChromeHits)
	assert.Equal(t, 0, loadDiscoveryHits(t, ctx, client, fx, "_ga").FirefoxHits)

	insertDraftVersion(t, ctx, client, fx)

	_, err := svc.PublishCookieBannerVersion(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.DiscoveryBrowserFamilySafari,
		Cookies: []DetectedCookie{{
			Name:         "_ga",
			Source:       coredata.CookieSourceScript,
			DiscoveryHit: true,
		}},
	}))

	stats = loadDiscoveryStats(t, ctx, client, fx)
	require.NotNil(t, stats.FrozenAt)
	assert.Equal(t, 1, stats.ChromePageLoads)
	assert.Equal(t, 1, stats.FirefoxPageLoads)
	assert.Equal(t, 0, stats.SafariPageLoads)
	assert.Equal(t, 1, loadDiscoveryHits(t, ctx, client, fx, "_ga").ChromeHits)
	assert.Equal(t, 0, loadDiscoveryHits(t, ctx, client, fx, "_ga").SafariHits)
}

func TestReportDetectedTrackers_PageViewStartsSharedWindow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.DiscoveryBrowserFamilyChrome,
	}))

	stats := loadDiscoveryStats(t, ctx, client, fx)
	assert.Equal(t, 1, stats.ChromePageLoads)
	assert.Nil(t, stats.FrozenAt)
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
		Family:   coredata.DiscoveryBrowserFamilyChrome,
		Cookies: []DetectedCookie{
			{
				Name:         "_ga",
				Source:       coredata.CookieSourceScript,
				HostOnly:     &hostOnly,
				DiscoveryHit: true,
			},
			{
				Name:         "_ga",
				Source:       coredata.CookieSourceScript,
				CookieDomain: &domain,
				HostOnly:     &notHostOnly,
				DiscoveryHit: true,
			},
		},
	}))

	assert.Equal(t, 1, loadDiscoveryHits(t, ctx, client, fx, "_ga").ChromeHits)
}

func TestPublishCookieBannerVersion_FreezesEmptyDiscoveryStats(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	insertDraftVersion(t, ctx, client, fx)

	_, err := svc.PublishCookieBannerVersion(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	stats := loadDiscoveryStats(t, ctx, client, fx)
	require.NotNil(t, stats.FrozenAt)
	assert.Equal(t, 0, stats.ChromePageLoads)
	assert.Equal(t, 0, stats.EdgePageLoads)
	assert.Equal(t, 0, stats.FirefoxPageLoads)
	assert.Equal(t, 0, stats.SafariPageLoads)
	assert.Equal(t, 0, stats.OtherPageLoads)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		PageView: true,
		Family:   coredata.DiscoveryBrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:         "_ga",
			Source:       coredata.CookieSourceScript,
			DiscoveryHit: true,
		}},
	}))

	stats = loadDiscoveryStats(t, ctx, client, fx)
	assert.Equal(t, 0, stats.ChromePageLoads)

	var pattern coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return pattern.LoadByBannerIDTypeAndPattern(ctx, conn, fx.scope, fx.banner.ID, coredata.TrackerTypeCookie, "_ga", nil)
	}))

	var hits coredata.TrackerPatternDiscoveryHits

	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return hits.LoadByTrackerPatternID(ctx, conn, fx.scope, pattern.ID)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func TestReportDetectedTrackers_HitWithoutPageLoadDoesNotCount(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
		Family: coredata.DiscoveryBrowserFamilyChrome,
		Cookies: []DetectedCookie{{
			Name:         "_ga",
			Source:       coredata.CookieSourceScript,
			DiscoveryHit: true,
		}},
	}))

	var stats coredata.CookieBannerDiscoveryStats

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return stats.LoadByCookieBannerID(ctx, conn, fx.scope, fx.banner.ID)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)

	var pattern coredata.TrackerPattern

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return pattern.LoadByBannerIDTypeAndPattern(ctx, conn, fx.scope, fx.banner.ID, coredata.TrackerTypeCookie, "_ga", nil)
	}))

	var hits coredata.TrackerPatternDiscoveryHits

	err = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return hits.LoadByTrackerPatternID(ctx, conn, fx.scope, pattern.ID)
	})
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}

func TestAddDiscoveryHits_SumsOntoTarget(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	for _, name := range []string{"_ga", "_gid"} {
		require.NoError(t, svc.ReportDetectedTrackers(ctx, fx.banner.ID, ReportDetectedTrackersRequest{
			PageView: true,
			Family:   coredata.DiscoveryBrowserFamilyChrome,
			Cookies: []DetectedCookie{{
				Name:         name,
				Source:       coredata.CookieSourceScript,
				DiscoveryHit: true,
			}},
		}))
	}

	ga := loadPattern(t, ctx, client, fx, "_ga")
	gidPattern := loadPattern(t, ctx, client, fx, "_gid")

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return coredata.AddDiscoveryHits(ctx, tx, fx.scope, ga.ID, gidPattern.ID)
	}))

	hits := loadDiscoveryHits(t, ctx, client, fx, "_gid")
	assert.Equal(t, 2, hits.ChromeHits)
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

func loadDiscoveryStats(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
) coredata.CookieBannerDiscoveryStats {
	t.Helper()

	var stats coredata.CookieBannerDiscoveryStats

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return stats.LoadByCookieBannerID(ctx, conn, fx.scope, fx.banner.ID)
	}))

	return stats
}

func loadDiscoveryHits(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	pattern string,
) coredata.TrackerPatternDiscoveryHits {
	t.Helper()

	tp := loadPattern(t, ctx, client, fx, pattern)

	var hits coredata.TrackerPatternDiscoveryHits

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return hits.LoadByTrackerPatternID(ctx, conn, fx.scope, tp.ID)
	}))

	return hits
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
