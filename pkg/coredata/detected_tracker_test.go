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
	"go.probo.inc/probo/pkg/page"
)

func newDetectedTracker(
	fx trackerPatternFixture,
	patternID gid.GID,
	identifier string,
	observedBy coredata.DetectedTrackerObserver,
) *coredata.DetectedTracker {
	now := time.Now().UTC().Truncate(time.Microsecond)

	return &coredata.DetectedTracker{
		ID:               gid.New(fx.scope.GetTenantID(), coredata.DetectedTrackerEntityType),
		CookieBannerID:   fx.cookieBannerID,
		TrackerPatternID: &patternID,
		TrackerType:      coredata.TrackerTypeCookie,
		Identifier:       identifier,
		Source:           new(coredata.CookieSourceHTTP),
		ObservedBy:       observedBy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func TestDetectedTracker_InsertIfNotExists_SeparatesObservers(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)
	tp := seedTrackerPattern(t, ctx, client, fx, "_ga", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceScript)

	insert := func(dt *coredata.DetectedTracker) bool {
		var inserted bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			inserted, err = dt.InsertIfNotExists(ctx, tx, fx.scope)

			return err
		}))

		return inserted
	}

	assert.True(t, insert(newDetectedTracker(fx, tp.ID, "_ga", coredata.DetectedTrackerObserverSDK)))
	assert.True(t, insert(newDetectedTracker(fx, tp.ID, "_ga", coredata.DetectedTrackerObserverScanner)), "the scanner row must not collide with the SDK row")
	assert.False(t, insert(newDetectedTracker(fx, tp.ID, "_ga", coredata.DetectedTrackerObserverSDK)), "a repeated SDK detection must not insert a second row")
}

func TestDetectedTracker_Upsert_RefreshesScannerAttributes(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)
	tp := seedTrackerPattern(t, ctx, client, fx, "_session", coredata.TrackerPatternMatchTypeExact, coredata.CookieSourceHTTP)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	first := newDetectedTracker(fx, tp.ID, "_session", coredata.DetectedTrackerObserverScanner)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := first.Upsert(ctx, tx, fx.scope)
		return err
	}))

	second := newDetectedTracker(fx, tp.ID, "_session", coredata.DetectedTrackerObserverScanner)
	second.HTTPOnly = new(true)
	second.SameSite = new(coredata.CookieSameSiteLax)
	second.SetByURL = new("https://tracker-pattern-test.example.com/login")
	second.CookieBannerScanID = &scan.ID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := second.Upsert(ctx, tx, fx.scope)
		return err
	}))

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

		return trackers.LoadByTrackerPatternID(ctx, conn, fx.scope, tp.ID, cursor)
	}))

	require.Len(t, trackers, 1)
	got := trackers[0]
	assert.Equal(t, first.ID, got.ID)
	assert.Equal(t, coredata.DetectedTrackerObserverScanner, got.ObservedBy)
	require.NotNil(t, got.HTTPOnly)
	assert.True(t, *got.HTTPOnly)
	require.NotNil(t, got.SameSite)
	assert.Equal(t, coredata.CookieSameSiteLax, *got.SameSite)
	require.NotNil(t, got.CookieBannerScanID)
	assert.Equal(t, scan.ID, *got.CookieBannerScanID)
}

func TestTrackerResource_Upsert_StampsScanOnlyWhenSet(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedTrackerPatternFixture(t, ctx, client)

	scan := newPendingCookieBannerScan(fx)
	require.NoError(t, insertCookieBannerScan(t, ctx, client, fx, scan))

	newResource := func(scanID *gid.GID) *coredata.TrackerResource {
		now := time.Now().UTC().Truncate(time.Microsecond)

		return &coredata.TrackerResource{
			ID:                 gid.New(fx.scope.GetTenantID(), coredata.TrackerResourceEntityType),
			OrganizationID:     fx.organizationID,
			CookieBannerID:     fx.cookieBannerID,
			CookieCategoryID:   fx.cookieCategoryID,
			ResourceType:       coredata.TrackerResourceTypeScript,
			Origin:             "https://www.googletagmanager.com",
			Path:               "/gtm.js",
			DisplayName:        "www.googletagmanager.com/gtm.js",
			CookieBannerScanID: scanID,
			CreatedAt:          now,
			UpdatedAt:          now,
		}
	}

	upsert := func(tr *coredata.TrackerResource) bool {
		var inserted bool

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var err error

			inserted, err = tr.Upsert(ctx, tx, fx.scope)

			return err
		}))

		return inserted
	}

	sdkReport := newResource(nil)
	require.True(t, upsert(sdkReport))
	sdkUpdatedAt := sdkReport.UpdatedAt

	assert.False(t, upsert(newResource(nil)), "a repeated SDK report must stay a no-op")

	scannerReport := newResource(&scan.ID)
	scannerReport.UpdatedAt = sdkUpdatedAt.Add(time.Second)
	assert.False(t, upsert(scannerReport), "a scanner report of an existing resource is not an insert")

	var reloaded coredata.TrackerResource

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return reloaded.LoadByID(ctx, conn, fx.scope, sdkReport.ID)
	}))
	require.NotNil(t, reloaded.CookieBannerScanID)
	assert.Equal(t, scan.ID, *reloaded.CookieBannerScanID)
	assert.True(t, reloaded.UpdatedAt.After(sdkUpdatedAt), "stamping the scan must refresh updated_at")
}
