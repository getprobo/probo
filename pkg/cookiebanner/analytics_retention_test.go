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
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestAnalyticsRetentionHandler_DeletesDaysOutsideMaxWindow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	now := analyticsTestDay.Add(12 * time.Hour)
	svc := newServiceAt(client, &now)

	reportOn := func(day time.Time, family coredata.BrowserFamily) {
		now = day.Add(12 * time.Hour)

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

	firstKeptDay := analyticsTestDay.AddDate(0, 0, 1-MaxAnalyticsDays)

	reportOn(firstKeptDay.AddDate(0, 0, -2), coredata.BrowserFamilyEdge)
	reportOn(firstKeptDay.AddDate(0, 0, -1), coredata.BrowserFamilyFirefox)
	reportOn(firstKeptDay, coredata.BrowserFamilySafari)
	reportOn(analyticsTestDay, coredata.BrowserFamilyChrome)
	requireFlushAnalytics(t, ctx, svc)

	require.Len(t, storedPageLoads(t, ctx, client, fx), 4)
	require.Len(t, storedPatternHits(t, ctx, client, fx, "_ga"), 4)

	h := &analyticsRetentionHandler{
		pg:        client,
		logger:    log.NewLogger(log.WithOutput(io.Discard)),
		batchSize: 1,
		now:       func() time.Time { return analyticsTestDay.Add(12 * time.Hour) },
	}
	require.NoError(t, h.Run(ctx))

	kept := []BrowserFamilyCount{
		{Family: coredata.BrowserFamilyChrome, Count: 1},
		{Family: coredata.BrowserFamilySafari, Count: 1},
	}

	assert.Equal(t, kept, storedPageLoads(t, ctx, client, fx))
	assert.Equal(t, kept, storedPatternHits(t, ctx, client, fx, "_ga"))
}

// storedPageLoads reads every stored day, ignoring the API window that
// retention is aligned with.
func storedPageLoads(
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
			analyticsTestDay.AddDate(-1, 0, 0),
		)
		if err != nil {
			return err
		}

		rows = loaded

		return nil
	}))

	return PageLoadsByFamily(rows)
}

func storedPatternHits(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	pattern string,
) []BrowserFamilyCount {
	t.Helper()

	tp := loadPattern(t, ctx, client, fx, pattern)

	var rows []coredata.DailyPatternHit

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded, err := coredata.LoadDailyPatternHitsByTrackerPatternIDs(
			ctx,
			conn,
			fx.scope,
			[]gid.GID{tp.ID},
			analyticsTestDay.AddDate(-1, 0, 0),
		)
		if err != nil {
			return err
		}

		rows = loaded

		return nil
	}))

	total, _, _ := SplitPatternHits(rows)

	return total
}
