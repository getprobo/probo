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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/cookiebanner/scanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestRequestScan_RejectsASecondActiveScan(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	scan, err := svc.RequestScan(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)
	assert.Equal(t, coredata.CookieBannerScanStatePending, scan.State)

	_, err = svc.RequestScan(ctx, fx.scope, fx.banner.ID)
	assert.ErrorIs(t, err, ErrScanAlreadyInProgress)

	_, err = svc.RequestScan(ctx, fx.scope, gid.New(fx.scope.GetTenantID(), coredata.CookieBannerEntityType))
	assert.ErrorIs(t, err, ErrBannerNotFound)

	got, err := svc.GetScan(ctx, fx.scope, scan.ID)
	require.NoError(t, err)
	assert.Equal(t, scan.ID, got.ID)

	_, err = svc.GetScan(ctx, fx.scope, gid.New(fx.scope.GetTenantID(), coredata.CookieBannerScanEntityType))
	assert.ErrorIs(t, err, ErrScanNotFound)

	scans, err := svc.ListScans(
		ctx,
		fx.scope,
		fx.banner.ID,
		page.NewCursor(
			10,
			nil,
			page.Head,
			page.OrderBy[coredata.CookieBannerScanOrderField]{
				Field:     coredata.CookieBannerScanOrderFieldCreatedAt,
				Direction: page.OrderDirectionDesc,
			},
		),
	)
	require.NoError(t, err)
	require.Len(t, scans, 1)
	assert.Equal(t, scan.ID, scans[0].ID)

	count, err := svc.CountScans(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestIngestScanPage_StampsScannerRowsWithoutSamples(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	pattern := newExactPattern(fx, "_ga", fx.normalCategoryID, coredata.CookieSourceScript, nil)

	now := time.Now().UTC().Truncate(time.Microsecond)
	sdk := &coredata.DetectedTracker{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   fx.banner.ID,
		TrackerPatternID: &pattern.ID,
		TrackerType:      coredata.TrackerTypeCookie,
		Identifier:       "_ga",
		Source:           new(coredata.CookieSourceScript),
		ObservedBy:       coredata.DetectedTrackerObserverSDK,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	existing := &coredata.TrackerResource{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.TrackerResourceEntityType),
		OrganizationID:   fx.organizationID,
		CookieBannerID:   fx.banner.ID,
		CookieCategoryID: fx.uncategorisedID,
		ResourceType:     coredata.TrackerResourceTypeScript,
		Origin:           "https://www.googletagmanager.com",
		Path:             "/gtm.js",
		DisplayName:      "www.googletagmanager.com/gtm.js",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	scan := insertRunningScan(t, ctx, client, fx)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := pattern.Insert(ctx, tx, fx.scope); err != nil {
			return err
		}

		if _, err := sdk.InsertIfNotExists(ctx, tx, fx.scope); err != nil {
			return err
		}

		return existing.Insert(ctx, tx, fx.scope)
	}))

	banner := fx.banner
	banner.Capabilities.ResourceReporting = false

	var inserted int

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		inserted, err = svc.ingestScanPage(
			ctx,
			tx,
			&banner,
			scan.ID,
			scanner.Page{
				URL: "https://worker-test.example.com/",
				Cookies: []scanner.Cookie{
					{
						Name:        "_ga",
						Source:      coredata.CookieSourceHTTP,
						HTTPOnly:    new(true),
						Secure:      new(true),
						SameSite:    new(coredata.CookieSameSiteLax),
						Partitioned: new(true),
						SetByURL:    new("https://worker-test.example.com/login"),
					},
				},
				Resources: []scanner.Resource{
					{
						Type:   coredata.TrackerResourceTypeScript,
						Origin: "https://www.googletagmanager.com",
						Path:   "/gtm.js",
					},
					{
						Type:   coredata.TrackerResourceTypeScript,
						Origin: "https://cdn.example.com",
						Path:   "/pixel.js",
					},
				},
			},
		)

		return err
	}))
	assert.Equal(t, 1, inserted, "only the new resource is an insert; the cookie pattern and the known resource already exist")

	var reloaded coredata.CookieBanner

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, fx.banner.ID)
	}))
	assert.NotNil(t, reloaded.PatternAnalysisRequestedAt)

	var trackers coredata.DetectedTrackers

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		cursor := page.NewCursor(
			10,
			nil,
			page.Head,
			page.OrderBy[coredata.DetectedTrackerOrderField]{
				Field:     coredata.DetectedTrackerOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
		)

		return trackers.LoadByTrackerPatternID(ctx, conn, fx.scope, pattern.ID, cursor)
	}))
	require.Len(t, trackers, 2)

	var sdkRow, scannerRow *coredata.DetectedTracker

	for _, tracker := range trackers {
		switch tracker.ObservedBy {
		case coredata.DetectedTrackerObserverSDK:
			sdkRow = tracker
		case coredata.DetectedTrackerObserverScanner:
			scannerRow = tracker
		}
	}

	require.NotNil(t, sdkRow)
	assert.Nil(t, sdkRow.CookieBannerScanID)
	require.NotNil(t, scannerRow)
	require.NotNil(t, scannerRow.CookieBannerScanID)
	assert.Equal(t, scan.ID, *scannerRow.CookieBannerScanID)
	require.NotNil(t, scannerRow.HTTPOnly)
	assert.True(t, *scannerRow.HTTPOnly)
	require.NotNil(t, scannerRow.Secure)
	assert.True(t, *scannerRow.Secure)
	require.NotNil(t, scannerRow.SameSite)
	assert.Equal(t, coredata.CookieSameSiteLax, *scannerRow.SameSite)
	require.NotNil(t, scannerRow.Partitioned)
	assert.True(t, *scannerRow.Partitioned)
	require.NotNil(t, scannerRow.SetByURL)
	assert.Equal(t, "https://worker-test.example.com/login", *scannerRow.SetByURL)

	var stamped coredata.TrackerResource

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return stamped.LoadByID(ctx, conn, fx.scope, existing.ID)
	}))
	require.NotNil(t, stamped.CookieBannerScanID)
	assert.Equal(t, scan.ID, *stamped.CookieBannerScanID)

	var created coredata.TrackerResource

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return created.LoadByBannerTypeOriginPath(
			ctx,
			conn,
			fx.scope,
			fx.banner.ID,
			coredata.TrackerResourceTypeScript,
			"https://cdn.example.com",
			"/pixel.js",
		)
	}))
	require.NotNil(t, created.CookieBannerScanID)
	assert.Equal(t, scan.ID, *created.CookieBannerScanID)

	assert.Equal(t, 0, syncMapLen(&svc.pageLoadCounts))
	assert.Equal(t, 0, syncMapLen(&svc.patternHitCounts))
}

func syncMapLen(m *sync.Map) int {
	count := 0

	m.Range(func(_, _ any) bool {
		count++

		return true
	})

	return count
}
